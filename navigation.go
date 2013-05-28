package main

import (
	"github.com/mattn/go-gtk/gtk"
	//	"github.com/mattn/go-gtk/gdk"
	"runtime"
)

const STACK_SIZE = 64
const MAX_SEL_LEN = 128

var selection_flag bool
var prev_selection string
var search_history []string

var file_stack [STACK_SIZE]string
var file_stack_top = 0
var file_stack_base = 0
var file_stack_max = 0

var prev_global bool = true
var prev_pattern string = ""

var accel_group *gtk.AccelGroup = nil

func find_entry_with_history() *gtk.ComboBoxEntry {
	entry := gtk.NewComboBoxEntryNewText()
	entry.SetVisible(true)
	selection := source_selection()
	if ("" != selection) && (len(selection) <= MAX_SEL_LEN) {
		entry.AppendText(selection)
	}
	for i := len(search_history) - 1; i >= 0; i-- {
		entry.AppendText(search_history[i])
	}
	entry.SetActive(0)
	return entry
}

func get_source() string {
	var be, en gtk.TextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	return source_buf.GetText(&be, &en, true)
}

func file_save_current() {
	var be, en gtk.TextIter
	rec, found := file_map[cur_file]
	if false == found {
		rec = new(FileRecord)
	}
	rec.buf = ([]byte)(get_source())
	rec.modified = source_buf.GetModified()
	source_buf.GetSelectionBounds(&be, &en)
	rec.sel_be = be.GetOffset()
	rec.sel_en = en.GetOffset()
	file_stack_push(cur_file)
	if found {
		runtime.GC()
	}
}

// Switches to another file. In most cases you want to call file_save_current 
// before this method. Otherwise current changes will be lost.
func file_switch_to(name string) {
	if "" == name {
		return
	}
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
	cur_file = name_to_set
	tree_view_set_cur_iter(true)
	source_buf.BeginNotUndoableAction()
	source_buf.SetText(text_to_set)
	source_buf.SetModified(modified_to_set)
	source_buf.EndNotUndoableAction()
	refresh_title()
	source_view.GrabFocus()
	var be_iter, en_iter gtk.TextIter
	source_buf.GetIterAtOffset(&be_iter, sel_be_to_set)
	source_buf.GetIterAtOffset(&en_iter, sel_en_to_set)
	move_focus_and_selection(&be_iter, &en_iter)

	prev_selection = ""
	mark_set_cb()
	lang_refresh()
}

func file_stack_push(name string) {
	if name == file_stack_at_top() {
		return
	}
	file_stack[file_stack_top] = name
	if file_stack_top == file_stack_max {
		stack_next(&file_stack_max)
	}
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
	var cur gtk.TextIter
	var be, en gtk.TextIter

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
	if (sel_len <= 1) || (sel_len >= MAX_SEL_LEN) {
		return
	} else {
		selection_flag = true
	}

	source_buf.GetStartIter(&cur)
	for cur.ForwardSearch(selection, 0, &be, &cur, nil) {
		source_buf.ApplyTagByName("instance", &be, &cur)
	}
}

func find_next_instance(start, be, en *gtk.TextIter, pattern string) bool {
	if start.ForwardSearch(pattern, 0, be, en, nil) {
		return true
	}
	source_buf.GetStartIter(be)
	return be.ForwardSearch(pattern, 0, be, en, nil)
}

func next_instance_cb() {
	var be, en gtk.TextIter
	source_buf.GetSelectionBounds(&be, &en)
	selection := source_buf.GetSlice(&be, &en, false)
	if "" == selection {
		return
	}
	// find_next_instance cannot return false because selection is not empty.
	find_next_instance(&en, &be, &en, selection)
	move_focus_and_selection(&be, &en)
}

func find_prev_instance(start, be, en *gtk.TextIter, pattern string) bool {
	if start.BackwardSearch(pattern, 0, be, en, nil) {
		return true
	}
	source_buf.GetEndIter(be)
	return be.BackwardSearch(pattern, 0, be, en, nil)
}

func prev_instance_cb() {
	var be, en gtk.TextIter
	source_buf.GetSelectionBounds(&be, &en)
	selection := source_buf.GetSlice(&be, &en, false)
	if "" == selection {
		return
	}
	// find_prev_instance cannot return false because selection is not empty.
	find_prev_instance(&be, &be, &en, selection)
	move_focus_and_selection(&be, &en)
}

func move_focus_and_selection(be *gtk.TextIter, en *gtk.TextIter) {
	source_buf.MoveMarkByName("insert", be)
	source_buf.MoveMarkByName("selection_bound", en)
	mark := source_buf.GetMark("insert")
	source_view.ScrollToMark(mark, 0, true, 1, 0.5)
}

func tree_view_scroll_to_cur_iter() {
	if "" == cur_file {
		return
	}
	//if false == tree_store.IterIsValid(&cur_iter) {
	//	return
	//}
	path := tree_model.GetPath(&cur_iter)
	tree_view.ScrollToCell(path, nil, true, 0.5, 0)
}

func source_selection() string {
	var be, en gtk.TextIter
	source_buf.GetSelectionBounds(&be, &en)
	return source_buf.GetSlice(&be, &en, false)
}

func next_file_cb() {
	if file_stack_top == file_stack_max {
		return
	}
	cur := file_stack_top
	for stack_next(&cur); ; stack_next(&cur) {
		if file_opened(file_stack[cur]) {
			file_save_current()
			file_switch_to(file_stack[cur])
			file_stack_top = cur
			return
		}
		if cur == file_stack_max {
			break
		}
	}
}

func prev_file_cb() {
	shift_flag := file_stack_top == file_stack_max
	file_save_current()
	if shift_flag {
		stack_prev(&file_stack_max)
	}
	// Popping out cur_file pushed in file_save_current. 
	// Wrong in case of "" is cur_file !!!
	file_stack_pop()
	file_switch_to(file_stack_pop())
}

func file_stack_at_top() string {
	t := file_stack_top
	stack_prev(&t)
	return file_stack[t]
}

func init_navigation() {
	accel_group = gtk.NewAccelGroup()
}
