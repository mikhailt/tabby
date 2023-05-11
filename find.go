package main

import (
    "github.com/mattn/go-gtk/gtk"
    "github.com/mattn/go-gtk/gdk"
    "strings"
)

// find_global function searches for the given pattern globally or for the file name
// and updates the search_view accordingly.
func find_global(pattern string, find_file bool) {
    var pos int
    if find_file {
        prev_pattern = ""
    } else {
        prev_pattern = pattern
    }
    search_view.store.Clear()
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
            search_view.AddFile(name)
        }
    }
}

// find_cb function is called when "Find" button is clicked and it calls find_common function
// with false value.
func find_cb() {
    find_common(false)
}

// find_file_cb function is called when "Find file" button is clicked and it calls find_common 
// function with true value.
func find_file_cb() {
    find_common(true)
}

// find_common function is called by find_cb and find_file_cb functions. It displays the find dialogue,
// sets up the search_view and calls find_global function or find_in_current_file function based on 
// the user input.
func find_common(find_file bool) {
    found_in_cur_file := false
    dialog_ok, pattern, global, find_file := find_dialog(find_file)
    if false == dialog_ok {
        return
    }
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

// find_in_current_file function searches for the given pattern in the current file and
// returns true if the pattern is found, false otherwise.
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

// find_dialog function displays the find dialogue and returns the user input.
func find_dialog(find_file bool) (bool, string, bool, bool) {
    dialog := gtk.NewDialog()
    defer dialog.Destroy()
    dialog.SetTitle("Find")
    dialog.AddButton("_Find", gtk.RESPONSE_ACCEPT)
    dialog.AddButton("_Cancel", gtk.RESPONSE_CANCEL)
    w := dialog.GetWidgetForResponse(int(gtk.RESPONSE_ACCEPT))
    dialog.AddAccelGroup(accel_group)
    w.AddAccelerator("clicked", accel_group, gdk.KEY_Return,
        0, gtk.ACCEL_VISIBLE)
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