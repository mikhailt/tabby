package main

import (
	"os"
	"bufio"
)

func session_save() {
	file, _ := os.Open(os.Getenv("HOME")+"/.tabby", os.O_CREAT|os.O_WRONLY, 0644)
	if nil == file {
		println("tabby: unable to save session")
		return
	}
	file.Truncate(0)
	for k, _ := range file_map {
		file.WriteString(k + "\n")
	}
	file.Close()
}

func session_restore() {
	file, _ := os.Open(os.Getenv("HOME")+"/.tabby", os.O_RDONLY, 0644)
	if nil == file {
		println("tabby: unable to restore session")
		return
	}
	reader := bufio.NewReader(file)
	for {
		str, err := reader.ReadString('\n')
		if nil != err {
			break
		}
		str = str[:len(str)-1]
		read_ok, buf := open_file_read_to_buf(str)
		if false == read_ok {
			continue
		}
		if add_file_record(str, buf, true) {
			file_stack_push(str)
		}
	}
	file.Close()
}
