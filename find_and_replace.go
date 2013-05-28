package main

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/gdk"
	"strings"
	"strconv"
)

func fnr_cb() {
	fnr_dialog()
}

func fnr_dialog() {
	var fnr_cnt int = 0
	var scope_be, scope_en gtk.TextIter
	if MAX_SEL_LEN < len(source_selection()) {
		source_buf.GetSelectionBounds(&scope_be, &scope_en)
	} else {
		source_buf.GetStartIter(&scope_be)
		source_buf.GetEndIter(&scope_en)
	}
	source_buf.CreateMark("fnr_be", &scope_be, true)
	source_buf.CreateMark("fnr_en", &scope_en, false)
	var map_filled bool = false
	var global_map map[string]int
	var insert_set bool = false

	dialog := gtk.NewDialog()
	dialog.SetTitle("Find and Replace")
	dialog.AddButton("_Find Next", gtk.RESPONSE_OK)
	dialog.AddButton("_Replace", gtk.RESPONSE_YES)
	dialog.AddButton("Replace _All", gtk.RESPONSE_APPLY)
	dialog.AddButton("_Close", gtk.RESPONSE_CLOSE)
	dialog.AddAccelGroup(accel_group)

	entry := find_entry_with_history()
	replacement := find_entry_with_history()

	global_button := gtk.NewCheckButtonWithLabel("Global")
	global_button.SetVisible(true)
	global_button.SetActive(prev_global)

	vbox := dialog.GetVBox()
	vbox.Add(entry)
	vbox.Add(replacement)
	vbox.Add(global_button)

	find_next_button := dialog.GetWidgetForResponse(int(gtk.RESPONSE_OK))
	replace_button := dialog.GetWidgetForResponse(int(gtk.RESPONSE_YES))
	replace_all_button := dialog.GetWidgetForResponse(int(gtk.RESPONSE_APPLY))
	close_button := dialog.GetWidgetForResponse(int(gtk.RESPONSE_CLOSE))

	find_next_button.Connect("clicked", func() {
		fnr_pre_cb(global_button, &insert_set)
		if !fnr_find_next(entry.GetActiveText(), prev_global, &map_filled, &global_map) {
			fnr_close_and_report(dialog, fnr_cnt)
		}
	},
		nil)
	find_next_button.AddAccelerator("clicked", accel_group, gdk.KEY_Return,
		0, gtk.ACCEL_VISIBLE)

	replace_button.Connect("clicked", func() {
		fnr_pre_cb(global_button, &insert_set)
		done, next_found := fnr_replace(entry.GetActiveText(), replacement.GetActiveText(),
			prev_global, &map_filled, &global_map)
		fnr_cnt += done
		if !next_found {
			fnr_close_and_report(dialog, fnr_cnt)
		}
	},
		nil)

	replace_all_button.Connect("clicked", func() {
		insert_set = false
		fnr_pre_cb(global_button, &insert_set)
		fnr_cnt += fnr_replace_all_local(entry.GetActiveText(), replacement.GetActiveText())
		if prev_global {
			fnr_cnt += fnr_replace_all_global(entry.GetActiveText(), replacement.GetActiveText())
			file_tree_store()
		}
		fnr_close_and_report(dialog, fnr_cnt)
	},
		nil)

	close_button.Connect("clicked", func() { dialog.Destroy() }, nil)

	dialog.Run()
}

func fnr_replace_all_local(entry string, replacement string) int {
	cnt := 0
	var t bool = true
	if !fnr_find_next(entry, false, &t, nil) {
		return 0
	}
	for {
		done, next_found := fnr_replace(entry, replacement, false, &t, nil)
		cnt += done
		if !next_found {
			break
		}
	}
	return cnt
}

func fnr_replace_all_global(entry, replacement string) int {
	total_cnt := 0
	lent := len(entry)
	lrep := len(replacement)
	inds := make(map[int]int)
	for file, rec := range file_map {
		if file == cur_file {
			continue
		}
		cnt := 0
		scope := rec.buf[:]
		for {
			pos := strings.Index(string(scope), entry)
			if -1 == pos {
				break
			}
			inds[cnt] = pos
			cnt++
			scope = scope[pos+lent:]
		}
		if 0 == cnt {
			continue
		}
		buf := make([]byte, len(rec.buf)+cnt*(lrep-lent))
		scope = rec.buf[:]
		dest_scope := buf[:]
		for y := 0; y < cnt; y++ {
			shift := inds[y]
			copy(dest_scope, scope[:shift])
			dest_scope = dest_scope[shift:]
			copy(dest_scope, replacement)
			dest_scope = dest_scope[lrep:]
			scope = scope[shift+lent:]
		}
		copy(dest_scope, scope)
		rec.buf = buf
		rec.modified = true
		total_cnt += cnt
	}
	return total_cnt
}

func fnr_pre_cb(global_button *gtk.CheckButton, insert_set *bool) {
	prev_global = global_button.GetActive()
	fnr_refresh_scope(prev_global)
	fnr_set_insert(insert_set)
}

func fnr_close_and_report(dialog *gtk.Dialog, fnr_cnt int) {
	dialog.Destroy()
	bump_message(strconv.Itoa(fnr_cnt) + " replacements were done.")
}

func fnr_set_insert(insert_set *bool) {
	if false == *insert_set {
		*insert_set = true
		var scope_be gtk.TextIter
		get_iter_at_mark_by_name("fnr_be", &scope_be)
		source_buf.MoveMarkByName("insert", &scope_be)
		source_buf.MoveMarkByName("selection_bound", &scope_be)
	}
}

func fnr_refresh_scope(global bool) {
	var be, en gtk.TextIter
	if global {
		source_buf.GetStartIter(&be)
		source_buf.GetEndIter(&en)
		source_buf.CreateMark("fnr_be", &be, true)
		source_buf.CreateMark("fnr_en", &en, false)
	}
}

func fnr_find_next(pattern string, global bool, map_filled *bool, m *map[string]int) bool {
	var be, en, scope_en gtk.TextIter
	get_iter_at_mark_by_name("fnr_en", &scope_en)
	get_iter_at_mark_by_name("selection_bound", &en)
	if en.ForwardSearch(pattern, 0, &be, &en, &scope_en) {
		move_focus_and_selection(&be, &en)
		return true
	}
	// Have to switch to next file or to beginning of current depending on <global>.
	if global {
		// Switch to next file.
		fnr_find_next_fill_global_map(pattern, m, map_filled)
		next_file := pop_string_from_map(m)
		if "" == next_file {
			return false
		}
		file_save_current()
		file_switch_to(next_file)
		fnr_refresh_scope(true)
		source_buf.GetStartIter(&be)
		source_buf.MoveMarkByName("insert", &be)
		source_buf.MoveMarkByName("selection_bound", &be)
		return fnr_find_next(pattern, global, map_filled, m)
	} else {
		// Temporary fix. Is there necessity to search the document all over again?
		return false
		// Start search from beginning of scope.
		// get_iter_at_mark_by_name("fnr_be", &be)
		// if be.ForwardSearch(pattern, 0, &be, &en, &scope_en) {
		//	move_focus_and_selection(&be, &en)
		//	return true
		//} else {
		//	return false
		//}
	}
	return false
}

func fnr_find_next_fill_global_map(pattern string, m *map[string]int, map_filled *bool) {
	if *map_filled {
		return
	}
	*map_filled = true
	*m = make(map[string]int)
	for file, rec := range file_map {
		if cur_file == file {
			continue
		}
		if -1 != strings.Index(string(rec.buf), pattern) {
			(*m)[file] = 1
		}
	}
}


// Returns (done, next_found)
func fnr_replace(entry string, replacement string, global bool, map_filled *bool, global_map *map[string]int) (int, bool) {
	if entry != source_selection() {
		return 0, true
	}
	source_buf.DeleteSelection(false, true)
	source_buf.InsertAtCursor(replacement)
	var be, en gtk.TextIter
	source_buf.GetSelectionBounds(&be, &en)
	source_buf.MoveMarkByName("insert", &en)
	return 1, fnr_find_next(entry, global, map_filled, global_map)
}

func pop_string_from_map(m *map[string]int) string {
	if 0 == len(*m) {
		return ""
	}
	for s, _ := range *m {
	  delete(*m, s)
		return s
	}
	return ""
}

func get_iter_at_mark_by_name(mark_name string, iter *gtk.TextIter) {
	mark := source_buf.GetMark(mark_name)
	source_buf.GetIterAtMark(iter, mark)
}
