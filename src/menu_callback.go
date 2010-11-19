package main

import (
	"glib"
	"gtk"
	"os"
)

var prev_dir string

func new_cb() {
	file_save_current()
	file_switch_to("")
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
	dir, _ := os.Open(dialog_dir, os.O_RDONLY, 0)
	if nil == dir {
		bump_message("Unable to open directory " + dialog_dir)
	}
	open_dir(dir, dialog_dir, true)
	dir.Close()
	file_tree_store()
}

func save_cb() {
	if "" == cur_file {
		save_as_cb()
	} else {
		file, _ := os.Open(cur_file, os.O_CREAT|os.O_WRONLY, 0644)
		if nil == file {
			bump_message("Unable to open file for writing: " + cur_file)
			return
		}
		file_save_current()
		rec, _ := file_map[cur_file]
		nbytes, _ := file.WriteString(string(rec.buf))
		if nbytes != len(rec.buf) {
			bump_message("Error while writing to file: " + cur_file)
			return
		}
		file.Truncate(int64(nbytes))
		file.Close()

		source_buf.SetModified(false)
		refresh_title()
	}
}

func save_as_cb() {
	dialog_ok, dialog_file := file_chooser_dialog(SAVE_DIALOG)
	if false == dialog_ok {
		return
	}
	var be, en gtk.GtkTextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	text_to_save := source_buf.GetText(&be, &en, true)
	add_file_record(dialog_file, []byte(text_to_save), true)
	file_tree_store()
	cur_file = dialog_file
	save_cb()
	tree_view_set_cur_iter(true)
}

func exit_cb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here
	session_save()
	gtk.MainQuit()
}

func close_cb() {
	if "" == cur_file {
		return
	}
	delete_file_record(cur_file)
	file_switch_to(file_stack_pop())
	file_tree_store()
}

func paste_done_cb() {
	var be, en gtk.GtkTextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	source_buf.RemoveTagByName("instance", &be, &en)
	selection_flag = false
}

func open_file_read_to_buf(name string, verbose bool) (bool, []byte) {
	file, _ := os.Open(name, os.O_RDONLY, 0644)
	if nil == file {
		bump_message("Unable to open file for reading: " + name)
		return false, nil
	}
	stat, _ := file.Stat()
	if nil == stat {
		bump_message("Unable to stat file: " + name)
		file.Close()
		return false, nil
	}
	buf := make([]byte, stat.Size)
	nread, _ := file.Read(buf)
	if nread != int(stat.Size) {
		bump_message("Unable to read whole file: " + name)
		file.Close()
		return false, nil
	}
	file.Close()
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
		fi, _ := os.Lstat(abs_name)
		if nil == fi {
			continue
		}
		if fi.IsDirectory() {
			if recursively {
				child_dir, _ := os.Open(abs_name, os.O_RDONLY, 0)
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
	var action int
	var ok_stock string
	if OPEN_DIALOG == t {
		action = gtk.GTK_FILE_CHOOSER_ACTION_OPEN
		ok_stock = gtk.GTK_STOCK_OPEN
	} else if SAVE_DIALOG == t {
		action = gtk.GTK_FILE_CHOOSER_ACTION_SAVE
		ok_stock = gtk.GTK_STOCK_SAVE
	} else if OPEN_DIR_DIALOG == t {
		action = gtk.GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER
		ok_stock = gtk.GTK_STOCK_OPEN
	}
	dialog := gtk.FileChooserDialog2("", source_view.GetTopLevelAsWindow(),
		action,
		gtk.GTK_STOCK_CANCEL, gtk.GTK_RESPONSE_CANCEL,
		ok_stock, gtk.GTK_RESPONSE_ACCEPT)
	dialog.SetCurrentFolder(prev_dir)
	res := dialog.Run()
	dialog_folder := dialog.GetCurrentFolder()
	dialog_file := dialog.GetFilename()
	dialog.Destroy()
	if gtk.GTK_RESPONSE_ACCEPT == res {
		prev_dir = dialog_folder
		return true, dialog_file
	}
	return false, ""
}
