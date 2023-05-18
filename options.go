// Package main is the main package of the program.
package main

import (
	"os"
	"strings"
	"strconv"
)

// Options is a struct that holds all the options for the program.
type Options struct {
	show_error, show_search, space_not_tab          bool
	ohp_position, ihp_position, vvp_position        int
	window_width, window_height, window_x, window_y int
	font                                            string
	tabsize                                         int
}

// new_options is a function that initializes the options struct with default values.
func new_options() (o Options) {
	o.show_search = true
	o.show_error = true
	o.ihp_position = 150
	o.ohp_position = 670
	o.vvp_position = 375
	o.window_width, o.window_height = 800, 510
	o.window_x, o.window_y = 0, 0
	o.font = "Monospace Regular 10"
	o.tabsize = 2
	return
}

// atoi is a function that converts a string to an integer.
func atoi(s string) (i int) {
	i, _ = strconv.Atoi(s)
	return
}

// load_options is a function that loads the options from a file.
func load_options() {
	// Get the reader and the file.
	reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabbyoptions")
	defer file.Close()
	var str string
	// Loop through the lines of the file.
	for next_string_from_reader(reader, &str) {
		args := strings.Split(compact_space(str), "\t")
		switch args[0] {
		case "space_not_tab":
			opt.space_not_tab, _ = strconv.ParseBool(args[1])
		case "show_search":
			opt.show_search, _ = strconv.ParseBool(args[1])
		case "show_error":
			opt.show_error, _ = strconv.ParseBool(args[1])
		case "ihp_position":
			opt.ihp_position = atoi(args[1])
		case "ohp_position":
			opt.ohp_position = atoi(args[1])
		case "vvp_position":
			opt.vvp_position = atoi(args[1])
		case "alloc_window":
			opt.window_width, opt.window_height, opt.window_x, opt.window_y = atoi(args[1]),
				atoi(args[2]), atoi(args[3]), atoi(args[4])
		case "font":
			opt.font = args[1]
		case "tabsize":
			opt.tabsize = atoi(args[1])
		}
	}
}

// save_options is a function that saves the options to a file.
func save_options() {
	file, _ := os.OpenFile(os.Getenv("HOME")+"/.tabbyoptions", os.O_CREATE|os.O_WRONLY, 0644)
	if nil == file {
		tabby_log("unable to save options")
		return
	}
	file.Truncate(0)
	file.WriteString("show_search\t" + strconv.FormatBool(opt.show_search) + "\n")
	file.WriteString("show_error\t" + strconv.FormatBool(opt.show_error) + "\n")
	file.WriteString("space_not_tab\t" + strconv.FormatBool(opt.space_not_tab) + "\n")
	file.WriteString("ihp_position\t" + strconv.Itoa(opt.ihp_position) + "\n")
	file.WriteString("ohp_position\t" + strconv.Itoa(opt.ohp_position) + "\n")
	file.WriteString("vvp_position\t" + strconv.Itoa(opt.vvp_position) + "\n")
	file.WriteString("alloc_window\t" + strconv.Itoa(opt.window_width) + "\t" +
		strconv.Itoa(opt.window_height) + "\t" + strconv.Itoa(opt.window_x) + "\t" +
		strconv.Itoa(opt.window_y) + "\n")
	file.WriteString("font\t" + opt.font + "\n")
	file.WriteString("tabsize\t" + strconv.Itoa(opt.tabsize) + "\n")
	file.Close()
}

// compact_space is a function that removes extra spaces and tabs from a string.
func compact_space(s string) string {
	s = strings.TrimSpace(s)
	n := replace_space(s)
	for n != s {
		s = n
		n = replace_space(s)
	}
	return s
}

// replace_space is a function that replaces extra spaces and tabs in a string.
func replace_space(s string) string {
	return strings.Replace(strings.Replace(strings.Replace(s, "  ", " ", -1),
		"\t ", "\t", -1),
		" \t", "\t", -1)
}

// opt is the options struct.
var opt Options = new_options()

// window_event_cb is a function that handles window events.
func window_event_cb() {
	main_window.GetSize(&opt.window_width, &opt.window_height)
	main_window.GetPosition(&opt.window_x, &opt.window_y)
	// TODO: Decide where to place these initialization.
	source_view.ModifyFontEasy(opt.font)
	options_set_tabsize(opt.tabsize)
}

// ohp_cb is a function that handles the position of the horizontal scroll bar.
func ohp_cb(pos int) {
	opt.ohp_position = pos
}

// ihp_cb is a function that handles the position of the horizontal scroll bar.
func ihp_cb(pos int) {
	opt.ihp_position = pos
}

// vvp_cb is a function that handles the position of the vertical scroll bar.
func vvp_cb(pos int) {
	opt.vvp_position = pos
}

// options_set_tabsize is a function that sets the tab size for the source view.
func options_set_tabsize(s int) {
	opt.tabsize = s
	source_view.SetIndentWidth(s)
	source_view.SetTabWidth(uint(s))
}