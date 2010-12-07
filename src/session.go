package main

import (
	"os"
	"bufio"
	"regexp"
)

type IgnoreMap map[string]*regexp.Regexp

var ignore IgnoreMap

func name_is_ignored(name string) bool {
	for _, re := range ignore {
	  if nil == re {
	    continue
	  }
		if re.Match([]byte(name)) {
			return true
		}
	}
	return false
}

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

func session_open_and_read_file(name string) {
	read_ok, buf := open_file_read_to_buf(name, false)
	if false == read_ok {
		return
	}
	if add_file_record(name, buf, true) {
		file_stack_push(name)
	}
}

func session_restore() {
	reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabby")
	defer file.Close()
	var str string
	for next_string_from_reader(reader, &str) {
		session_open_and_read_file(str)
	}
	ignore = make(IgnoreMap)
	reader, file = take_reader_from_file(os.Getenv("HOME") + "/.tabbyignore")
	for next_string_from_reader(reader, &str) {
		ignore[str], _ = regexp.Compile(str)
	}
}

func take_reader_from_file(name string) (*bufio.Reader, *os.File) {
	file, _ := os.Open(name, os.O_CREAT|os.O_RDONLY, 0644)
	if nil == file {
		println("tabby: unable to Open file for reading: ", name)
		return nil, nil
	}
	return bufio.NewReader(file), file
}

func next_string_from_reader(reader *bufio.Reader, s *string) bool {
	str, err := reader.ReadString('\n')
	if nil != err {
		return false
	}
	*s = str[:len(str)-1]
	return true
}
