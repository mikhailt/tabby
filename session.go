package main

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type IgnoreMap map[string]*regexp.Regexp

var ignore IgnoreMap

func file_is_saved(file string) bool {
	return strings.Contains(file, string(os.PathSeparator))
}

func name_is_ignored(name string) bool {
	for _, re := range ignore {
		if re != nil && re.Match([]byte(name)) {
			return true
		}
	}
	return false
}

func get_stack_set() (map[string]int, []string, int) {
	m := make(map[string]int)
	list := make([]string, STACK_SIZE)
	list_size := 0
	get_stack_set_add_file(cur_file, m, list, &list_size)
	for {
		file := file_stack_pop()
		if file == "" {
			break
		}
		get_stack_set_add_file(file, m, list, &list_size)
	}
	return m, list, list_size
}

func get_file_info(file string) string {
	rec := file_map[file]
	return file + ":" + strconv.Itoa(rec.sel_be) + ":" + strconv.Itoa(rec.sel_en) + "\n"
}

func session_save() {
	file_save_current()
	file, err := os.OpenFile(os.Getenv("HOME")+"/.tabby", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		tabby_log("unable to save session")
		return
	}
	defer file.Close()

	file.Truncate(0)
	stack_set, list, list_size := get_stack_set()

	for k := range file_map {
		if _, found := stack_set[k]; !found && file_is_saved(k) {
			file.WriteString(get_file_info(k))
		}
	}

	for y := list_size - 1; y >= 0; y-- {
		file.WriteString(get_file_info(list[y]))
	}
}

func session_open_and_read_file(name string) bool {
	read_ok, buf := open_file_read_to_buf(name, false)
	if !read_ok {
		return false
	}
	if add_file_record(name, buf, true) {
		file_stack_push(name)
		return true
	}
	return false
}

func session_restore() {
	reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabby")
	if file == nil {
		return
	}
	defer file.Close()

	var str string
	for next_string_from_reader(reader, &str) {
		split_str := strings.SplitN(str, ":", 3)
		if session_open_and_read_file(split_str[0]) {
			be, _ := strconv.Atoi(split_str[1])
			en, _ := strconv.Atoi(split_str[2])
			file_map[split_str[0]].sel_be = be
			file_map[split_str[0]].sel_en = en
		}
	}

	ignore = make(IgnoreMap)
	reader2, file2 := take_reader_from_file(os.Getenv("HOME") + "/.tabbyignore")
	if file2 == nil {
		return
	}
	defer file2.Close()

	for next_string_from_reader(reader2, &str) {
		ignore[str], _ = regexp.Compile(str)
	}
}

func take_reader_from_file(name string) (*bufio.Reader, *os.File) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		tabby_log("unable to Open file for reading: " + name)
		return nil, nil
	}
	return bufio.NewReader(file), file
}

func next_string_from_reader(reader *bufio.Reader, s *string) bool {
	str, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	*s = str[:len(str)-1]
	return true
}

func get_stack_set_add_file(file string, m map[string]int, l []string, s *int) {
	if !file_is_saved(file) {
		return
	}
	_, found := m[file]
	if !found {
		m[file] = 1
		l[*s] = file
		*s++
	}
}