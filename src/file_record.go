package main

import (
	"gtk"
	"runtime"
)

type FileRecord struct {
	buf      []byte
	modified bool
}

var file_map map[string]*FileRecord

func file_save_current() {
	if "" == cur_file {
		return
	}
	var be, en gtk.GtkTextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	text_to_save := source_buf.GetText(&be, &en, true)
	rec, found := file_map[cur_file]
	if false == found {
		rec = new(FileRecord)
	}
	rec.buf = ([]byte)(text_to_save[:])
	rec.modified = source_buf.GetModified()
	runtime.GC()
}

func file_switch_to(name string) {
	rec := file_map[name]
	source_buf.BeginNotUndoableAction()
	if nil == rec.buf {
		source_buf.SetText("")
	} else {
		source_buf.SetText(string(rec.buf))
	}
	source_buf.SetModified(rec.modified)
	source_buf.EndNotUndoableAction()
	cur_file = name
	refresh_title()
}

func file_opened(name string) bool {
	_, found := file_map[name]
	return found
}


func delete_file_record(name string) {
	_, found := file_map[name]
	if false == found {
		return
	}
	file_tree_remove(&file_tree_root, name, true)
	file_map[name] = nil, false
}

func add_file_record(name string, bump_flag bool) {
	_, found := file_map[name]
	if found {
		if bump_flag {
			bump_message("File " + name + " is already open")
		}
		return
	}
	rec := new(FileRecord)
	file_map[name] = rec
	rec.modified = false
	rec.buf = nil
	file_tree_insert(name)
}
