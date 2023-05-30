package main

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/gdk"
	"strings"
)

var prev_pattern string
var search_history []string
var prev_global bool

func find_global(pattern string, find_file bool) {
	if find_file {
		prev_pattern = ""
	} else {
		prev_pattern = pattern
	}
	search_view.store.Clear()
	for name, rec := range file_map {
		if find_file {
			if strings.Contains(name, pattern) {
				search_view.AddFile(name)
			}
		} else {
			if name == cur_file {
				continue
			}
			if strings.Contains(string(rec.buf), pattern) {
				search_view.AddFile(name)
			}
		}
	}
}

func find_cb() {
	find_common(false)
}

func find_file_cb() {
	find_common(true)
}

func find_common(find_file bool) {
	found_in_cur_file := false
	if dialog_ok, pattern, global, find_file := find_dialog(find_file); !dialog_ok {
		return
	} else {
		if global {
			search_view.PrepareToSearch()
		}
		if find_file {
			find_global(pattern, true)
		} else {
			if global {
				find_global(pattern, false)
			}
			found_in_cur_file = find_in_current_file(pattern, global)
		}
		if global && !found_in_cur_file {
			search_view.SetCursor(0)
		}
	}
}

// Returns true if pattern was found in current file, false o/w.
func find_in_current_file(pattern string, global bool) bool {
	var be, en gtk.TextIter
	source_buf.GetSelectionBounds(&be, &en)
	if find_next_instance(&en, &be, &en, pattern) {
		move_focus_and_selection(&be, &en)
		mark_set_cb()
		if global {
			search_view.AddFile(cur_file)
		}
		return true
	}
	return false
}

func find_dialog(find_file bool) (bool, string, bool, bool) {
	dialog := gtk.NewDialog()
	defer dialog.Destroy()
	dialog.SetTitle("Find")
	dialog.AddButton("_Find", gtk.RESPONSE_ACCEPT)
	dialog.AddButton("_Cancel", gtk.RESPONSE_CANCEL)
	w := dialog.GetWidgetForResponse(int(gtk.RESPONSE_ACCEPT))
	dialog.AddAccelGroup(accel_group)
	w.AddAccelerator("clicked", accel_group, gdk.KEY_Return, 0, gtk.ACCEL_VISIBLE)
	entry := find_entry_with_history()
	global_button := gtk.NewCheckButtonWithLabel("Global")
	global_button.SetVisible(true)
	global_button.SetActive(prev_global)
	file_button := gtk.NewCheckButtonWithLabel("Find file by name pattern")
	file_button.SetVisible(true)
	file_button.SetActive(find_file)
	vbox := dialog.GetVBox()
	vbox.Add(entry)
	vbox.Add(global_button)
	vbox.Add(file_button)
	if gtk.RESPONSE_ACCEPT == dialog.Run() {
		entry_text := entry.GetActiveText()
		if search_history == nil {
			search_history = []string{entry_text}
		} else {
			if len(search_history) >= 10 {
				search_history = search_history[1:]
			}
			search_history = append(search_history, entry_text)
		}
		prev_global = global_button.GetActive()
		return true, entry_text, prev_global, file_button.GetActive()
	}
	return false, "", false, false
}