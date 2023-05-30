package main

import (
	"os"
	"strings"
	"strconv"
)

type Options struct {
	show_error, show_search, space_not_tab bool
	ohp_position, ihp_position, vvp_position int
	window_width, window_height, window_x, window_y int
	font string
	tabsize int
}

func new_options() Options {
	return Options {
		show_search: true,
		show_error: true,
		ihp_position: 150,
		ohp_position: 670,
		vvp_position: 375,
		window_width: 800,
		window_height: 510,
		window_x: 0,
		window_y: 0,
		font: "Monospace Regular 10",
		tabsize: 2,
	}
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func load_options() {
	reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabbyoptions")
	defer file.Close()

	var str string
	for next_string_from_reader(reader, &str) {
		args := strings.Split(strings.TrimSpace(str), "\t")
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
			opt.window_width = atoi(args[1])
			opt.window_height = atoi(args[2])
			opt.window_x = atoi(args[3])
			opt.window_y = atoi(args[4])
		case "font":
			opt.font = args[1]
		case "tabsize":
			options_set_tabsize(atoi(args[1]))
		}
	}
}

func save_options() {
	file, err := os.OpenFile(os.Getenv("HOME")+"/.tabbyoptions", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		tabby_log("unable to save options")
		return
	}
	defer file.Close()

	file.Truncate(0)
	file.WriteString(strings.Join([]string{
		"show_search\t" + strconv.FormatBool(opt.show_search),
		"show_error\t" + strconv.FormatBool(opt.show_error),
		"space_not_tab\t" + strconv.FormatBool(opt.space_not_tab),
		"ihp_position\t" + strconv.Itoa(opt.ihp_position),
		"ohp_position\t" + strconv.Itoa(opt.ohp_position),
		"vvp_position\t" + strconv.Itoa(opt.vvp_position),
		"alloc_window\t" + strconv.Itoa(opt.window_width) + "\t" +
			strconv.Itoa(opt.window_height) + "\t" + strconv.Itoa(opt.window_x) + "\t" +
			strconv.Itoa(opt.window_y),
		"font\t" + opt.font,
		"tabsize\t" + strconv.Itoa(opt.tabsize),
	}, "\n"))
}

func compact_space(s string) string {
	for strings.Contains(s, "  ") || strings.Contains(s, "\t ") || strings.Contains(s, " \t") {
		s = strings.Replace(s, "  ", " ", -1)
		s = strings.Replace(s, "\t ", "\t", -1)
		s = strings.Replace(s, " \t", "\t", -1)
	}

	return strings.TrimSpace(s)
}

var opt = new_options()

func window_event_cb() {
	main_window.GetSize(&opt.window_width, &opt.window_height)
	main_window.GetPosition(&opt.window_x, &opt.window_y)

	source_view.ModifyFontEasy(opt.font)
	options_set_tabsize(opt.tabsize)
}

func ohp_cb(pos int) {
	opt.ohp_position = pos
}

func ihp_cb(pos int) {
	opt.ihp_position = pos
}

func vvp_cb(pos int) {
	opt.vvp_position = pos
}

func options_set_tabsize(s int) {
	opt.tabsize = s
	source_view.SetIndentWidth(s)
	source_view.SetTabWidth(uint(s))
}