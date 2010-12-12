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

// Returns set of files contained in stack + cur_file. Deletes all the files 
// from stack as a side effect. Also returns reverse list of files from stack
// without duplications preceeded by cur_file.
func get_stack_set() (map[string]int, []string, int) {
	m := make(map[string]int)
	list := make([]string, STACK_SIZE)
	m[cur_file] = 1
	list[0] = cur_file
	list_size := 1
	for {
		file := file_stack_pop()
		if "" == file {
			break
		}
		_, found := m[file]
		if false == found {
			list[list_size] = file
			list_size++
		}
		m[file] = 1
	}
	return m, list, list_size
}

func session_save() {
	file, _ := os.Open(os.Getenv("HOME")+"/.tabby", os.O_CREAT|os.O_WRONLY, 0644)
	if nil == file {
		println("tabby: unable to save session")
		return
	}
	file.Truncate(0)
	stack_set, list, list_size := get_stack_set()
	// Dump all the files not contained in file_stack.
	for k, _ := range file_map {
		_, found := stack_set[k]
		if false == found {
			file.WriteString(k + "\n")
		}
	}
	// Dump files from stack in the right order. Last file should be last in the
	// list of files in .tabby file.
	for y := list_size - 1; y >= 0; y-- {
		file.WriteString(list[y] + "\n")
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
