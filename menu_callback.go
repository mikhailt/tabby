package main

import (
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"os"
	"strconv"
)

var prev_dir string
var last_unsaved int = -1

func new_cb() {
	file_save_current()
	last_unsaved++
	file := "unsaved file " + strconv.Itoa(last_unsaved)
	add_file_record(file, []byte(""), true)
	file_map[file].modified = true
	file_tree_store()
	file_switch_to(file)
	tree_view_set_cur_iter(true)
}

func open_cb() {
	file_save_current()
	dialog_ok, dialog_file := file_chooser_dialog(OPEN_DIALOG)
	if false == dialog_ok {
		return
	}
	read_ok, buf := open_file_read_to_buf(dialog_file, true)
	if false == read_ok {
		return
	}
	if add_file_record(dialog_file, buf, true) {
		file_tree_store()
		file_switch_to(dialog_file)
	}
}

func open_rec_cb() {
	dialog_ok, dialog_dir := file_chooser_dialog(OPEN_DIR_DIALOG)
	if false == dialog_ok {
		return
	}
	dir, _ := os.OpenFile(dialog_dir, os.O_RDONLY, 0)
	if nil == dir {
		bump_message("Unable to open directory " + dialog_dir)
	}
	open_dir(dir, dialog_dir, true)
	dir.Close()
	file_tree_store()
}

func save_cb() {
	if !file_is_saved(cur_file) {
		save_as_cb()
		return
	}
	inotify_rm_watch(cur_file)
	defer inotify_add_watch(cur_file)
	file, _ := os.OpenFile(cur_file, os.O_CREATE|os.O_WRONLY, 0644)
	if nil == file {
		bump_message("Unable to open file for writing: " + cur_file)
		return
	}
	file_save_current()
	rec, _ := file_map[cur_file]
	nbytes, err := file.WriteString(string(rec.buf))
	if nbytes != len(rec.buf) {
		bump_message("Error while writing to file: " + cur_file)
		println("nbytes = ", nbytes, " errno = ", err)
		return
	}
	file.Truncate(int64(nbytes))
	file.Close()

	source_buf.SetModified(false)
	refresh_title()
}

func save_as_cb() {
	dialog_ok, dialog_file := file_chooser_dialog(SAVE_DIALOG)
	if false == dialog_ok {
		return
	}
	var be, en gtk.TextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	text_to_save := source_buf.GetText(&be, &en, true)
	add_file_record(dialog_file, []byte(text_to_save), true)
	file_tree_store()
	file_to_delete := cur_file
	file_switch_to(dialog_file)
	delete_file_record(file_to_delete)
	file_tree_store()
	save_cb()
	tree_view_set_cur_iter(true)
}

func exit_cb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here.
	session_save()
	if nil != listener {
		listener.Close()
	}
	gtk.MainQuit()
}

func close_cb() {
	if "" == cur_file {
		return
	}
	close_it := !source_buf.GetModified()
	if !close_it {
		close_it = bump_question("This file has been modified. Close it?")
	}
	if close_it {
		delete_file_record(cur_file)
		cur_file = file_stack_pop()
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
		file_switch_to(cur_file)
		file_tree_store()
	}
}

func paste_done_cb() {
	var be, en gtk.TextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	source_buf.RemoveTagByName("instance", &be, &en)
	selection_flag = false
}

// Reads file content to newly allocated buffer.
func open_file_read_to_buf(name string, verbose bool) (bool, []byte) {
	file, _ := os.OpenFile(name, os.O_RDONLY, 0644)
	if nil == file {
		bump_message("Unable to open file for reading: " + name)
		return false, nil
	}
	defer file.Close()
	stat, _ := file.Stat()
	if nil == stat {
		bump_message("Unable to stat file: " + name)
		return false, nil
	}
	buf := make([]byte, stat.Size())
	nread, _ := file.Read(buf)
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

func open_dir(dir *os.File, dir_name string, recursively bool) {
	names, _ := dir.Readdirnames(-1)
	for _, name := range names {
		abs_name := dir_name + "/" + name
		if name_is_ignored(abs_name) {
			continue
		}
		fi, _ := os.Lstat(abs_name)
		if nil == fi {
			continue
		}
		if fi.IsDir() {
			if recursively {
				child_dir, _ := os.OpenFile(abs_name, os.O_RDONLY, 0)
				if nil != child_dir {
					open_dir(child_dir, abs_name, true)
				}
				child_dir.Close()
			}
		} else {
			session_open_and_read_file(abs_name)
		}
	}
}

const (
	OPEN_DIALOG     = 0
	SAVE_DIALOG     = 1
	OPEN_DIR_DIALOG = 2
)

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

func error_chk_cb(current bool) {
	error_window.SetVisible(current)
	opt.show_error = current
}

func search_chk_cb(current bool) {
	search_view.window.SetVisible(current)
	opt.show_search = current
}

func notab_chk_cb(current bool) {
	opt.space_not_tab = current
	source_view.SetInsertSpacesInsteadOfTabs(opt.space_not_tab)
}

func gofmt_cb() {
	gofmt(cur_file)
}

func font_cb() {
	dialog := gtk.NewFontSelectionDialog("Pick a font")
	dialog.SetFontName(opt.font)
	if gtk.RESPONSE_OK == dialog.Run() {
		opt.font = dialog.GetFontName()
		source_view.ModifyFontEasy(opt.font)
	}
	dialog.Destroy()
}
