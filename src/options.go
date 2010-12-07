package main

import (
"gtk"
"os"
"strings"
"strconv"
)
type Options struct{
  showSearch  bool
}
func newOptions()(o Options){
  o.showSearch=true
  return
}
func loadOptions() {
	reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabbyoptions")
	defer file.Close()
	var str string
	for next_string_from_reader(reader, &str) {
		args:=strings.Split(compactSpace(str),"\t",-1)
		switch args[0]{
		case "showSearch":
		  opt.showSearch,_=strconv.Atob(args[1])
		}
	}
}
func saveOptions() {
	file, _ := os.Open(os.Getenv("HOME")+"/.tabbyoptions", os.O_CREAT|os.O_WRONLY, 0644)
	if nil == file {
		println("tabby: unable to save options")
		return
	}
	file.Truncate(0)
	file.WriteString("showSearch\t"+strconv.Btoa(opt.showSearch) + "\n")
	file.Close()
}
func compactSpace(s string)string{
  s=strings.TrimSpace(s)
  n:=replaceSpace(s)
  for n!=s{
    s=n
    n=replaceSpace(s)
  }
  return s
}
func replaceSpace(s string)string{
  return strings.Replace(strings.Replace(strings.Replace(s,"  "," ",-1),"\t ","\t",-1)," \t","\t",-1)
}
var opt Options=newOptions()
var search_window *gtk.GtkScrolledWindow