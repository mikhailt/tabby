package main

import (
	"strings"
	"os"
)

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
		for ; '/' != res[prev_slash]; prev_slash-- {}
		res = res[:prev_slash+1] + res[i+4:]
	}
	return res
}

func open_file_from_args(file string, focus_line int) {
	if '/' != file[0] {
		// Relative file name.
		wd, _ := os.Getwd()
		if "" == wd {
			tabby_log("Getwd failed")
			return
		}
		file = wd + "/" + file
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
}