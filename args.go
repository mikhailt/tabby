package main

import (
	"flag"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mattn/go-gtk/gdk"
)

var listener net.Listener
var tabby_args []string
var pfocus_line = flag.Int("f", 1, "Focus line")
var pstandalone = flag.Bool("s", false, "Forces to open new instance of tabby.")

func init() {
	flag.Parse()
	tabby_args = flag.Args()
}

func simplifiedPath(file string) string {
	res := file
	for {
		i := strings.Index(res, "/./")
		if i == -1 {
			break
		}
		res = res[:i+1] + res[i+3:]
	}
	for {
		i := strings.Index(res, "/../")
		if i == -1 {
			break
		}
		prevSlash := i - 1
		for res[prevSlash] != '/' {
			prevSlash--
		}
		res = res[:prevSlash+1] + res[i+4:]
	}
	return res
}

func prefixedPath(file string) string {
	if file[0] != '/' {
		wd, err := os.Getwd()
		if err == nil {
			file = wd + "/" + file
		}
	}
	return file
}

func provideTabbyServer(cnt int) bool {
	if cnt > 3 {
		return true
	}
	if *pstandalone {
		return true
	}

	if runtime.GOOS == "windows" {
		return true
	}
	user := os.Getenv("USER")
	socketName := "/tmp/tabby-" + user
	var err error
	listener, err = net.Listen("unix", socketName)
	if err != nil {
		conn, err := net.Dial("unix", socketName)
		if err == nil {
			defer conn.Close()
			if len(tabby_args) > 0 {
				_, _ = conn.Write([]byte(packTabbyArgs()))
			}
			return false
		}
		os.Remove(socketName)
		return provideTabbyServer(cnt + 1)
	}
	go tabbyServer()
	return true
}

func packTabbyArgs() string {
	res := strconv.Itoa(*pfocus_line) + "\n"
	for _, s := range tabby_args {
		res += prefixedPath(s) + "\n"
	}
	res += "\n"
	return res
}

func openFileFromArgs(file string, focusLine int) bool {
	split := strings.SplitN(file, ":", 2)
	if len(split) >= 2 {
		focusLine, _ = strconv.Atoi(split[1])
	}
	file = simplifiedPath(split[0])
	if !sessionOpenAndReadFile(file) {
		return false
	}
	rec, found := fileMap[file]
	if !found {
		return false
	}
	curLine := 1
	var y int
	for y = 0; y < len(rec.buf); y++ {
		if curLine == focusLine {
			break
		}
		if rec.buf[y] == '\n' {
			curLine++
		}
	}
	rec.selBe = y
	rec.selEn = y
	return true
}

func openFilesFromArgs() {
	for _, s := range tabby_args {
		openFileFromArgs(prefixedPath(s), *pfocus_line)
	}
}

func tabbyServer() {
	var focusLine int
	buf := make([]byte, 1024)

	for {
		c, err := listener.Accept()
		if err != nil {
			return
		}
		n, err := c.Read(buf)
		if n <= 0 {
			tabbyLog("server: read from unix socket: " + err.Error())
			c.Close()
			continue
		}

		gdk.ThreadsEnter()
		defer gdk.ThreadsLeave()

		openedCnt := 0
		s := buf[:]
		for cnt := 0; ; cnt++ {
			en := strings.Index(string(s), "\n")
			if en == 0 {
				break
			}
			if cnt == 0 {
				focusLine, _ = strconv.Atoi(string(s[:en]))
			} else {
				if openFileFromArgs(string(s[:en]), focusLine) {
					openedCnt++
				}
			}
			s = s[en+1:]
		}
		if openedCnt > 0 {
			mainWindow.Present()
			fileTreeStore()
			newFile := fileStackPop()
			fileSaveCurrent()
			fileSwitchTo(newFile)
		}

		c.Close()
	}
}

func initArgs() bool {
	return provideTabbyServer(0)
}