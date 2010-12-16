package main

import (
	"gtk"
	"gdk"
	"strings"
)

func find_global(pattern string, find_file bool) {
	var iter gtk.GtkTreeIter
	var pos int
	if find_file {
		prev_pattern = ""
	} else {
		prev_pattern = pattern
	}
	search_store.Clear()
	for name, rec := range file_map {
		if find_file {
			pos = strings.Index(name, pattern)
		} else {
			if name == cur_file {
				// find_in_current_file does required work for cur_file.
				continue
			}
			pos = strings.Index(string(rec.buf), pattern)
		}
		if -1 != pos {
			search_store.Append(&iter, nil)
			search_store.Set(&iter, name)
		}
	}
}

func find_cb() {
	dialog_ok, pattern, global, find_file := find_dialog()
	if false == dialog_ok {
		return
	}
	if find_file {
		find_global(pattern, true)
	} else {
		if global {
			find_global(pattern, false)
		}
		find_in_current_file(pattern, global)
	}
}

func find_in_current_file(pattern string, global bool) {
	var be, en gtk.GtkTextIter
	source_buf.GetSelectionBounds(&be, &en)
	if find_next_instance(&en, &be, &en, pattern) {
		move_focus_and_selection(&be, &en)
		mark_set_cb()
		if global {
			var iter gtk.GtkTreeIter
			search_store.Append(&iter, nil)
			search_store.Set(&iter, cur_file)
		}
	}
}

func find_dialog() (bool, string, bool, bool) {
	dialog := gtk.Dialog()
	defer dialog.Destroy()
	dialog.SetTitle("Find")
	dialog.AddButton("_Find", gtk.GTK_RESPONSE_ACCEPT)
	dialog.AddButton("_Cancel", gtk.GTK_RESPONSE_CANCEL)
	w := dialog.GetWidgetForResponse(gtk.GTK_RESPONSE_ACCEPT)
	dialog.AddAccelGroup(accel_group)
	w.AddAccelerator("clicked", accel_group, gdk.GDK_Return,
		0, gtk.GTK_ACCEL_VISIBLE)
	entry := find_entry_with_history()
	global_button := gtk.CheckButtonWithLabel("Global")
	global_button.SetVisible(true)
	global_button.SetActive(prev_global)
	file_button := gtk.CheckButtonWithLabel("Find file by name pattern")
	file_button.SetVisible(true)
	vbox := dialog.GetVBox()
	vbox.Add(entry)
	vbox.Add(global_button)
	vbox.Add(file_button)
	if gtk.GTK_RESPONSE_ACCEPT == dialog.Run() {
		entry_text := entry.GetActiveText()
		if nil == search_history {
			search_history = make([]string, 1)
			search_history[0] = entry_text
		} else {
			be := 0
			if 10 <= len(search_history) {
				be = 1
			}
			search_history = append(search_history[be:], entry_text)
		}
		prev_global = global_button.GetActive()
		return true, entry_text, prev_global, file_button.GetActive()
	}
	return false, "", false, false
}
