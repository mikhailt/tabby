package main

import (
	"gtk"
	"gdk"
	"runtime"
	"strings"
)

const STACK_SIZE = 64

var selection_flag bool
var prev_selection string

var file_stack [STACK_SIZE]string
var file_stack_top = 0
var file_stack_base = 0

var prev_global bool = false
var prev_pattern string = ""

var accel_group *gtk.GtkAccelGroup = nil

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
	rec.sel_be = be.GetOffset()
	rec.sel_en = en.GetOffset()
	file_stack_push(cur_file)
	runtime.GC()
}

func file_switch_to(name string) {
	tree_view_set_cur_iter(false)
	rec, found := file_map[name]
	var text_to_set string
	var modified_to_set bool
	var name_to_set string
	var sel_be_to_set, sel_en_to_set int
	if found {
		text_to_set = string(rec.buf)
		modified_to_set = rec.modified
		name_to_set = name
		sel_be_to_set = rec.sel_be
		sel_en_to_set = rec.sel_en
	} else {
		text_to_set = ""
		modified_to_set = true
		name_to_set = ""
		sel_be_to_set = 0
		sel_en_to_set = 0
	}
	source_buf.BeginNotUndoableAction()
	source_buf.SetText(text_to_set)
	source_buf.SetModified(modified_to_set)
	source_buf.EndNotUndoableAction()
	cur_file = name_to_set
	tree_view_set_cur_iter(true)
	refresh_title()
	source_view.GrabFocus()
	var be_iter, en_iter gtk.GtkTextIter
	source_buf.GetIterAtOffset(&be_iter, sel_be_to_set)
	source_buf.GetIterAtOffset(&en_iter, sel_en_to_set)
	move_focus_and_selection(&be_iter, &en_iter)

	prev_selection = ""
	mark_set_cb()
}

func file_stack_push(name string) {
	file_stack[file_stack_top] = name
	stack_next(&file_stack_top)
	if file_stack_top == file_stack_base {
		stack_next(&file_stack_base)
	}
}

func file_stack_pop() string {
	for {
		if file_stack_base == file_stack_top {
			return ""
		}
		stack_prev(&file_stack_top)
		res := file_stack[file_stack_top]
		if file_opened(res) {
			return res
		}
	}
	return ""
}

func stack_next(a *int) {
	*a++
	if STACK_SIZE == *a {
		*a = 0
	}
}

func stack_prev(a *int) {
	*a--
	if -1 == *a {
		*a = STACK_SIZE - 1
	}
}

func mark_set_cb() {
	var cur gtk.GtkTextIter
	var be, en gtk.GtkTextIter

	source_buf.GetSelectionBounds(&be, &en)
	selection := source_buf.GetSlice(&be, &en, false)
	if prev_selection == selection {
		return
	}
	prev_selection = selection

	if selection_flag {
		source_buf.GetStartIter(&be)
		source_buf.GetEndIter(&en)
		source_buf.RemoveTagByName("instance", &be, &en)
		selection_flag = false
	}

	sel_len := len(selection)
	if (sel_len <= 1) || (sel_len >= 100) {
		return
	} else {
		selection_flag = true
	}

	source_buf.GetStartIter(&cur)
	for cur.ForwardSearch(selection, 0, &be, &cur, nil) {
		source_buf.ApplyTagByName("instance", &be, &cur)
	}
}

func find_next_instance(start, be, en *gtk.GtkTextIter, pattern string) bool {
	if start.ForwardSearch(pattern, 0, be, en, nil) {
		return true
	}
	source_buf.GetStartIter(be)
	return be.ForwardSearch(pattern, 0, be, en, nil)
}

func next_instance_cb() {
	var be, en gtk.GtkTextIter
	source_buf.GetSelectionBounds(&be, &en)
	selection := source_buf.GetSlice(&be, &en, false)
	if "" == selection {
		return
	}
	// find_next_instance cannot return false because selection is not empty.
	find_next_instance(&en, &be, &en, selection)
	move_focus_and_selection(&be, &en)
}

func find_global(pattern string) {
	var iter gtk.GtkTreeIter
	prev_pattern = pattern
	search_store.Clear()
	for name, rec := range file_map {
		if -1 != strings.Index(string(rec.buf), pattern) {
			search_store.Append(&iter, nil)
			search_store.Set(&iter, name)
		}
	}
}

func find_cb() {
	dialog_ok, pattern, global := find_dialog()
	if false == dialog_ok {
		return
	}
	if global {
		find_global(pattern)
	}
	find_in_current_file(pattern)
}

func find_in_current_file(pattern string) {
	var be, en gtk.GtkTextIter
	source_buf.GetSelectionBounds(&be, &en)
	if find_next_instance(&en, &be, &en, pattern) {
		move_focus_and_selection(&be, &en)
		mark_set_cb()
	}
}

func find_dialog() (bool, string, bool) {
	if nil == accel_group {
		accel_group = gtk.AccelGroup()
	}
	dialog := gtk.Dialog()
	defer dialog.Destroy()
	dialog.SetTitle("Find")
	dialog.AddButton("_Find", gtk.GTK_RESPONSE_ACCEPT)
	dialog.AddButton("_Cancel", gtk.GTK_RESPONSE_CANCEL)
	w := dialog.GetWidgetForResponse(gtk.GTK_RESPONSE_ACCEPT)
	dialog.AddAccelGroup(accel_group)
	w.AddAccelerator("clicked", accel_group, gdk.GDK_Return,
		0, gtk.GTK_ACCEL_VISIBLE)
	vbox := dialog.GetVBox()
	entry := gtk.Entry()
	entry.SetVisible(true)
	vbox.Add(entry)
	global_button := gtk.CheckButtonWithLabel("global")
	global_button.SetVisible(true)
	global_button.SetActive(prev_global)
	vbox.Add(global_button)
	if gtk.GTK_RESPONSE_ACCEPT == dialog.Run() {
		prev_global = global_button.GetActive()
		return true, entry.GetText(), prev_global
	}
	return false, "", false
}

func move_focus_and_selection(be *gtk.GtkTextIter, en *gtk.GtkTextIter) {
	source_buf.MoveMarkByName("insert", be)
	source_buf.MoveMarkByName("selection_bound", en)
	mark := source_buf.GetMark("insert")
	source_view.ScrollToMark(mark, 0, true, 1, 0.5)
}

func tree_view_scroll_to_cur_iter() {
	if "" == cur_file {
		return
	}
	if false == tree_store.IterIsValid(&cur_iter) {
		return
	}
	path := tree_model.GetPath(&cur_iter)
	tree_view.ScrollToCell(path, nil, true, 0.5, 0)
}
