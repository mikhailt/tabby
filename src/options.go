package main

import (
	"os"
	"strings"
	"strconv"
)

type Options struct {
	showError, showSearch, spaceNotTab          bool
	ohpPosition, ihpPosition, vvpPosition       int
	windowWidth, windowHeight, windowX, windowY int
}

func new_options() (o Options) {
	o.showSearch = true
	o.showError = true
	o.ihpPosition = 150
	o.ohpPosition = 670
	o.vvpPosition = 375
	o.windowWidth, o.windowHeight = 800, 510
	o.windowX, o.windowY = 0, 0
	return
}

func atoi(s string) (i int) {
	i, _ = strconv.Atoi(s)
	return
}

func load_options() {
	reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabbyoptions")
	defer file.Close()
	var str string
	for next_string_from_reader(reader, &str) {
		args := strings.Split(compact_space(str), "\t", -1)
		switch args[0] {
		case "spaceNotTab":
			opt.spaceNotTab, _ = strconv.Atob(args[1])
		case "showSearch":
			opt.showSearch, _ = strconv.Atob(args[1])
		case "showError":
			opt.showError, _ = strconv.Atob(args[1])
		case "ihpPosition":
			opt.ihpPosition = atoi(args[1])
		case "ohpPosition":
			opt.ohpPosition = atoi(args[1])
		case "vvpPosition":
			opt.vvpPosition = atoi(args[1])
		case "allocWindow":
			opt.windowWidth, opt.windowHeight, opt.windowX, opt.windowY = atoi(args[1]),
				atoi(args[2]), atoi(args[3]), atoi(args[4])
		}
	}
}
func save_options() {
	file, _ := os.Open(os.Getenv("HOME")+"/.tabbyoptions", os.O_CREAT|os.O_WRONLY, 0644)
	if nil == file {
		println("tabby: unable to save options")
		return
	}
	file.Truncate(0)
	file.WriteString("showSearch\t" + strconv.Btoa(opt.showSearch) + "\n")
	file.WriteString("showError\t" + strconv.Btoa(opt.showError) + "\n")
	file.WriteString("spaceNotTab\t" + strconv.Btoa(opt.spaceNotTab) + "\n")
	file.WriteString("ihpPosition\t" + strconv.Itoa(opt.ihpPosition) + "\n")
	file.WriteString("ohpPosition\t" + strconv.Itoa(opt.ohpPosition) + "\n")
	file.WriteString("vvpPosition\t" + strconv.Itoa(opt.vvpPosition) + "\n")
	file.WriteString("allocWindow\t" + strconv.Itoa(opt.windowWidth) + "\t" +
		strconv.Itoa(opt.windowHeight) + "\t" + strconv.Itoa(opt.windowX) + "\t" +
		strconv.Itoa(opt.windowY) + "\n")
	file.Close()
}

func compact_space(s string) string {
	s = strings.TrimSpace(s)
	n := replace_space(s)
	for n != s {
		s = n
		n = replace_space(s)
	}
	return s
}

func replace_space(s string) string {
	return strings.Replace(strings.Replace(strings.Replace(s, "  ", " ", -1),
		"\t ", "\t", -1),
		" \t", "\t", -1)
}

var opt Options = new_options()

func window_event_cb() {
	main_window.GetSize(&opt.windowWidth, &opt.windowHeight)
	main_window.GetPosition(&opt.windowX, &opt.windowY)
}

func ohp_cb(pos int) {
	opt.ohpPosition = pos
}

func ihp_cb(pos int) {
	opt.ihpPosition = pos
}

func vvp_cb(pos int) {
	opt.vvpPosition = pos
}
