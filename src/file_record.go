package main

import (
	"gtk"
	"runtime"
)

type FileRecord struct {
	buf      []byte
	modified bool
	shift    int
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
	rec.buf = ([]byte)(text_to_save)
	rec.modified = source_buf.GetModified()
	source_buf.GetSelectionBounds(&be, &en)
	rec.shift = be.GetOffset()
	runtime.GC()
}

func file_switch_to(name string) {
	rec, found := file_map[name]
	var text_to_set string
	var modified_to_set bool
	var name_to_set string
	var shift_to_set int
	if found {
		text_to_set = string(rec.buf)
		modified_to_set = rec.modified
		name_to_set = name
		shift_to_set = rec.shift
	} else {
		text_to_set = ""
		modified_to_set = true
		name_to_set = ""
		shift_to_set = 0
	}
	source_buf.BeginNotUndoableAction()
	source_buf.SetText(text_to_set)
	source_buf.SetModified(modified_to_set)
	source_buf.EndNotUndoableAction()
	cur_file = name_to_set
	tree_view_set_cur_iter()
	refresh_title()
	source_view.GrabFocus()
	var iter gtk.GtkTextIter
	source_buf.GetIterAtOffset(&iter, shift_to_set)
	mark := source_buf.GetMark("focus_mark")
	source_buf.MoveMark(mark, &iter)
	source_view.ScrollToMark(mark, 0, true, 0, 0.5)
	source_buf.MoveMarkByName("insert", &iter)
	source_buf.MoveMarkByName("selection_bound", &iter)
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

func add_file_record(name string, buf []byte, bump_flag bool) bool {
	_, found := file_map[name]
	if found {
		if bump_flag {
			bump_message("File " + name + " is already open")
		}
		return false
	}
	rec := new(FileRecord)
	file_map[name] = rec
	rec.modified = false
	rec.buf = buf
	file_tree_insert(name, rec)
	return true
}
