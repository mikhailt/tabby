diff --git a/main.go b/main.go
index e69de29..a7b0e97 100644
--- a/main.go
+++ b/main.go
@@ -1 +1,88 @@
+// open_files_from_args opens files from command line arguments.
 func open_files_from_args() {
 	for _, s := range tabby_args {
 		open_file_from_args(prefixed_path(s), *pfocus_line)
 	}
 }
 
+// tabby_server starts a tabby server to listen for and process incoming requests.
 func tabby_server() {
 	var focus_line int
 	buf := make([]byte, 1024)
 
 	for {
 		c, _ := listener.Accept()
 		if nil != c {
 			nread, err := c.Read(buf)
 			if 0 >= nread {
 				tabby_log("server: read from unix socket: " + err.Error())
 				c.Close()
 				continue
 			}
 
+// provide_tabby_server checks if a new tabby instance is needed and returns true if so.
 func provide_tabby_server(cnt int) bool {
 	if cnt > 3 {
 		return true
 	}
 	if *pstandalone {
 		return true
 	}
 
+// init_args initializes command line arguments and returns true if a new tabby instance is needed.
 func init_args() bool {
 	pfocus_line = flag.Int("f", 1, "Focus line")
 	pstandalone = flag.Bool("s", false, "Forces to open new instance of tabby.")
 	flag.Parse()
 	tabby_args = flag.Args()
 
 	return provide_tabby_server(0)
 }
 
+// pack_tabby_args packs command line arguments into a single string.
 func pack_tabby_args() string {
 	res := strconv.Itoa(*pfocus_line) + "\n"
 	for _, s := range tabby_args {
 		res += prefixed_path(s) + "\n"
 	}
 	res += "\n"
 	return res
 }
 
+// simplified_path simplifies the file path by removing redundant characters.
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
 
+// prefixed_path adds the current working directory as a prefix to a relative file path.
 func prefixed_path(file string) string {
 	if '/' != file[0] {
 		// Relative file name.
 		wd, err := os.Getwd()
 		if "" == wd {
 			tabby_log(err.Error())
 		} else {
 			file = wd + "/" + file
 		}
 	}
 	return file
 }
 
+// open_file_from_args opens a file from command line arguments and sets the focus line.
 func open_file_from_args(file string, focus_line int) bool {
 	split_file := strings.SplitN(file, ":", 2)
 	if len(split_file) >= 2 {
 		focus_line, _ = strconv.Atoi(split_file[1])
 	}
 	file = simplified_path(split_file[0])
 	if false == session_open_and_read_file(file) {
 	  return false
 	}
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
 	} else {
 		return false
 	}
 	return true
 }