package main

import (
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"os"
	"strconv"
)

// Define global variables.
var prev_dir string
var last_unsaved int = -1

// Add commentary to new_cb function.
// Create new file with default name and add to file record.
func new_cb() {
	file_save_current() // Save current file if it's unsaved.
	last_unsaved++ // Increase last unsaved file counter.
	file := "unsaved file " + strconv.Itoa(last_unsaved) // Create new file name.
	add_file_record(file, []byte(""), true) // Add new file to file record.
	file_map[file].modified = true // Mark file as modified.
	file_tree_store() // Update file tree store.
	file_switch_to(file) // Switch to new file.
	tree_view_set_cur_iter(true) // Set iterator to current tree view.
}

// Add commentary to open_cb function.
// Open a file and add to file record.
func open_cb() {
	file_save_current() // Save current file if it's unsaved.
	dialog_ok, dialog_file := file_chooser_dialog(OPEN_DIALOG) // Show file chooser dialog.
	if false == dialog_ok {
		return
	}
	read_ok, buf := open_file_read_to_buf(dialog_file, true) // Read file content to buffer.
	if false == read_ok {
		return
	}
	if add_file_record(dialog_file, buf, true) { // Add new file to file record.
		file_tree_store() // Update file tree store.
		file_switch_to(dialog_file) // Switch to new file.
	}
}

// Add commentary to open_rec_cb function.
// Open a directory and its files recursively.
func open_rec_cb() {
	dialog_ok, dialog_dir := file_chooser_dialog(OPEN_DIR_DIALOG) // Show directory chooser dialog.
	if false == dialog_ok {
		return
	}
	dir, _ := os.OpenFile(dialog_dir, os.O_RDONLY, 0) // Open directory.
	if nil == dir {
		bump_message("Unable to open directory " + dialog_dir)
	}
	open_dir(dir, dialog_dir, true) // Open directory and its files recursively.
	dir.Close() // Close directory.
	file_tree_store() // Update file tree store.
}

// Add commentary to save_cb function.
// Save current file content to disk.
func save_cb() {
	if !file_is_saved(cur_file) { // Check if current file is not saved.
		save_as_cb() // Save current file as.
		return
	}
	inotify_rm_watch(cur_file) // Remove watch for inotify.
	defer inotify_add_watch(cur_file) // Add watch for inotify.
	file, _ := os.OpenFile(cur_file, os.O_CREATE|os.O_WRONLY, 0644) // Open file for writing.
	if nil == file {
		bump_message("Unable to open file for writing: " + cur_file)
		return
	}
	file_save_current() // Save current file.
	rec, _ := file_map[cur_file]
	nbytes, err := file.WriteString(string(rec.buf)) // Write file content to disk.
	if nbytes != len(rec.buf) {
		bump_message("Error while writing to file: " + cur_file)
		println("nbytes = ", nbytes, " errno = ", err)
		return
	}
	file.Truncate(int64(nbytes)) // Truncate file.
	file.Close() // Close file.

	source_buf.SetModified(false) // Set source buffer as unmodified.
	refresh_title() // Refresh title.
}

// Add commentary to save_as_cb function.
// Save current file with a new name.
func save_as_cb() {
	dialog_ok, dialog_file := file_chooser_dialog(SAVE_DIALOG) // Show file chooser dialog.
	if false == dialog_ok {
		return
	}
	var be, en gtk.TextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	text_to_save := source_buf.GetText(&be, &en, true) // Get content of source buffer.
	add_file_record(dialog_file, []byte(text_to_save), true) // Add new file to file record.
	file_tree_store() // Update file tree store.
	file_to_delete := cur_file // Save current file name to delete.
	file_switch_to(dialog_file) // Switch to new file.
	delete_file_record(file_to_delete) // Delete old file.
	file_tree_store() // Update file tree store.
	save_cb() // Save new file.
	tree_view_set_cur_iter(true) // Set iterator in tree view.
}

// Add commentary to exit_cb function.
// Exit application.
func exit_cb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here.
	session_save() // Save session.
	if nil != listener {
		listener.Close()
	}
	gtk.MainQuit() // Quit GTK main loop.
}

// Add commentary to close_cb function.
// Close current file.
func close_cb() {
	if "" == cur_file {
		return
	}
	close_it := !source_buf.GetModified() // Check if file is modified.
	if !close_it {
		close_it = bump_question("This file has been modified. Close it?")
	}
	if close_it {
		delete_file_record(cur_file) // Delete current file from file record.
		cur_file = file_stack_pop() // Pop file from file stack.
		if 0 == len(file_map) {
			new_cb()
		}
		if "" == cur_file {
			// Choose random open file then. Previous if implies that there are some 
			// opened files. At least unsaved.
			for cur_file, _ = range file_map {
				break
			}
		}
		file_switch_to(cur_file) // Switch to new file.
		file_tree_store() // Update file tree store.
	}
}

// Add commentary to paste_done_cb function.
// Remove tag from pasted text.
func paste_done_cb() {
	var be, en gtk.TextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	source_buf.RemoveTagByName("instance", &be, &en)
	selection_flag = false
}

// Add commentary to open_file_read_to_buf function.
// Read file content to buffer.
func open_file_read_to_buf(name string, verbose bool) (bool, []byte) {
	file, _ := os.OpenFile(name, os.O_RDONLY, 0644) // Open file for reading.
	if nil == file {
		bump_message("Unable to open file for reading: " + name)
		return false, nil
	}
	defer file.Close()
	stat, _ := file.Stat() // Get file status.
	if nil == stat {
		bump_message("Unable to stat file: " + name)
		return false, nil
	}
	buf := make([]byte, stat.Size()) // Create buffer.
	nread, _ := file.Read(buf) // Read file content to buffer.
	if nread != int(stat.Size()) {
		bump_message("Unable to read whole file: " + name)
		return false, nil
	}
	if nread > 0 {
		if false == glib.Utf8Validate(buf, nread, nil) {
			if verbose {
				bump_message("File " + name + " is not correct utf8 text")
			}
			return false, nil
		}
	}
	return true, buf
}

// Add commentary to open_dir function.
// Open a directory and its files.
func open_dir(dir *os.File, dir_name string, recursively bool) {
	names, _ := dir.Readdirnames(-1) // Get directory names.
	for _, name := range names {
		abs_name := dir_name + "/" + name
		if name_is_ignored(abs_name) {
			continue
		}
		fi, _ := os.Lstat(abs_name) // Get file status.
		if nil == fi {
			continue
		}
		if fi.IsDir() {
			if recursively {
				child_dir, _ := os.OpenFile(abs_name, os.O_RDONLY, 0) // Open child directory.
				if nil != child_dir {
					open_dir(child_dir, abs_name, true) // Open directory recursively.
				}
				child_dir.Close() // Close child directory.
			}
		} else {
			session_open_and_read_file(abs_name) // Open and read file to session.
		}
	}
}

// Define constants for file chooser dialog.
const (
	OPEN_DIALOG     = 0
	SAVE_DIALOG     = 1
	OPEN_DIR_DIALOG = 2
)

// Display file chooser dialog.
func file_chooser_dialog(t int) (bool, string) {
	var action gtk.FileChooserAction
	var ok_stock string
	if OPEN_DIALOG == t {
		action = gtk.FILE_CHOOSER_ACTION_OPEN
		ok_stock = gtk.STOCK_OPEN
	} else if SAVE_DIALOG == t {
		action = gtk.FILE_CHOOSER_ACTION_SAVE
		ok_stock = gtk.STOCK_SAVE
	} else if OPEN_DIR_DIALOG == t {
		action = gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER
		ok_stock = gtk.STOCK_OPEN
	}
	dialog := gtk.NewFileChooserDialog("", source_view.GetTopLevelAsWindow(),
		action,
		gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL,
		ok_stock, gtk.RESPONSE_ACCEPT)
	dialog.SetCurrentFolder(prev_dir)
	res := dialog.Run()
	dialog_folder := dialog.GetCurrentFolder()
	dialog_file := dialog.GetFilename()
	dialog.Destroy()
	if gtk.RESPONSE_ACCEPT == res {
		prev_dir = dialog_folder
		return true, dialog_file
	}
	return false, ""
}

// Add commentary to error_chk_cb function.
// Show/hide error window.
func error_chk_cb(current bool) {
	error_window.SetVisible(current)
	opt.show_error = current
}

// Add commentary to search_chk_cb function.
// Show/hide search window.
func search_chk_cb(current bool) {
	search_view.window.SetVisible(current)
	opt.show_search = current
}

// Add commentary to notab_chk_cb function.
// Set insertion of spaces instead of tabs.
func notab_chk_cb(current bool) {
	opt.space_not_tab = current
	source_view.SetInsertSpacesInsteadOfTabs(opt.space_not_tab)
}

// Add commentary to gofmt_cb function.
// Call gofmt to format current file.
func gofmt_cb() {
	gofmt(cur_file)
}

// Add commentary to font_cb function.
// Choose and set font for source view.
func font_cb() {
	dialog := gtk.NewFontSelectionDialog("Pick a font")
	dialog.SetFontName(opt.font)
	if gtk.RESPONSE_OK == dialog.Run() {
		opt.font = dialog.GetFontName()
		source_view.ModifyFontEasy(opt.font)
	}
	dialog.Destroy()
}