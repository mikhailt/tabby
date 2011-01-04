package main

import (
	"gdk"
	"strings"
	"os"
	"flag"
	"net"
	"strconv"
)

var listener net.Listener
var tabby_args []string
var pfocus_line *int
var pstandalone *bool

func open_files_from_args() {
	for _, s := range tabby_args {
		open_file_from_args(s, *pfocus_line)
	}
}

func tabby_server() {
	var focus_line int
	buf := make([]byte, 1024)
	
	for {
		c, e := listener.Accept()
		if nil != c {
			nread, err := c.Read(buf)
			if -1 == nread {
				tabby_log(err.String())
				c.Close()
				continue
			}
			
			// At this point buf contains '\n' separated file names preceeded by focus
			// line number. Double '\n' at the end of list.
			
			gdk.ThreadsEnter()
			
			s := buf[:]
			for cnt := 0; ; cnt++ {
				en := strings.Index(string(s), "\n")
				if 0 == en {
					break
				}
				if 0 == cnt {
					focus_line, _ = strconv.Atoi(string(s[:en]))
				} else {
					open_file_from_args(string(s[:en]), focus_line)
				}
				s = s[en+1:]
			}
			file_tree_store()
			new_file := file_stack_pop()
			file_save_current()
			file_switch_to(new_file)
			
			gdk.ThreadsLeave()
			
			c.Close()
		} else {
			tabby_log(e.String())
		}
	}
}

func provide_tabby_server(cnt int) bool {
	if cnt > 3 {
		return true
	}
	user := os.Getenv("USER")
	socket_name := "/tmp/tabby-" + user
	listener, _ = net.Listen("unix", socket_name)
	if nil == listener {
		// tabby server already exists, trying to connect to it. Or do nothing if
		// -s (for standalone) was specified or if args are empty.
		if (0 == len(tabby_args)) || (*pstandalone) {
			return true
		}
		conn, _ := net.Dial("unix", "", socket_name)
		if nil == conn {
			// Server exists but we cannot connect to it. Delete socket file then
			// and repeat the logic.
			os.Remove(socket_name)
			return provide_tabby_server(cnt + 1)
		}
		// Dial succeeded.
		conn.Write([]byte(pack_tabby_args()))
		conn.Close()
		return false
	}
	// Ok, this instance of tabby becomes a server.
	go tabby_server()
	return true
}

func init_args() bool {
	pfocus_line = flag.Int("f", 1, "Focus line")
	pstandalone = flag.Bool("s", false, "Forces to open new instance of tabby.")
	flag.Parse()
	tabby_args = flag.Args()

	return provide_tabby_server(0)
}

func pack_tabby_args() string {
	res := strconv.Itoa(*pfocus_line) + "\n"
	for _, s := range tabby_args {
		res += s + "\n"
	}
	res += "\n"
	return res
}

func simplified_path(file string) string {
	res := file
	for {
		i := strings.Index(res, "/./")
		if -1 == i {
			break
		}
		res = res[:i+1] + res[i+3:]
	}
	for {
		i := strings.Index(res, "/../")
		if -1 == i {
			break
		}
		prev_slash := i - 1
		for ; '/' != res[prev_slash]; prev_slash-- {
		}
		res = res[:prev_slash+1] + res[i+4:]
	}
	return res
}

func open_file_from_args(file string, focus_line int) {
	if '/' != file[0] {
		// Relative file name.
		wd, err := os.Getwd()
		if "" == wd {
			tabby_log(err.String())
			return
		}
		file = wd + "/" + file
	}
	file = simplified_path(file)
	session_open_and_read_file(file)
	rec, found := file_map[file]
	if found {
		cur_line := 1
		var y int
		for y = 0; y < len(rec.buf); y++ {
			if cur_line == focus_line {
				break
			}
			if rec.buf[y] == '\n' {
				cur_line++
			}
		}
		rec.sel_be = y
		rec.sel_en = y
	}
}
