package main

import (
	"glib"
	"gtk"
	"os"
)

func new_cb() {
	file_save_current()
	source_buf_set_content(nil)
	cur_file = ""
	source_buf.SetModified(true)
	refresh_title()
}

func open_cb() {
	file_save_current()
	file_dialog := gtk.FileChooserDialog2("", source_view.GetTopLevelAsWindow(),
		gtk.GTK_FILE_CHOOSER_ACTION_OPEN,
		gtk.GTK_STOCK_CANCEL, gtk.GTK_RESPONSE_CANCEL,
		gtk.GTK_STOCK_OPEN, gtk.GTK_RESPONSE_ACCEPT)
	file_dialog.SetCurrentFolder(prev_dir)
	res := file_dialog.Run()
	dialog_folder := file_dialog.GetCurrentFolder()
	dialog_file := file_dialog.GetFilename()
	file_dialog.Destroy()
	if gtk.GTK_RESPONSE_ACCEPT == res {
		prev_dir = dialog_folder
		file, _ := os.Open(dialog_file, os.O_RDONLY, 0700)
		if nil == file {
			bump_message("Unable to open file for reading: " + dialog_file)
			return
		}
		stat, _ := file.Stat()
		if nil == stat {
			bump_message("Unable to stat file: " + dialog_file)
			file.Close()
			return
		}
		buf := make([]byte, stat.Size)
		nread, _ := file.Read(buf)
		if nread != int(stat.Size) {
			bump_message("Unable to read whole file: " + dialog_file)
			file.Close()
			return
		}
		file.Close()
		cur_file = dialog_file
		if nread > 0 {
			if false == glib.Utf8Validate(buf, nread, nil) {
				bump_message("File " + cur_file + " is not correct utf8 text")
				close_cb()
				return
			}
		}

		source_buf_set_content(buf)
		add_file_record(cur_file, true)
		file_tree_store()
		refresh_title()
	}
}

func save_cb() {
	if "" == cur_file {
		save_as_cb()
	} else {
		file, _ := os.Open(cur_file, os.O_CREAT|os.O_WRONLY, 0700)
		if nil == file {
			bump_message("Unable to open file for writing: " + cur_file)
			return
		}
		var be, en gtk.GtkTextIter
		source_buf.GetStartIter(&be)
		source_buf.GetEndIter(&en)
		text_to_save := source_buf.GetText(&be, &en, true)
		nbytes, _ := file.WriteString(text_to_save)
		if nbytes != len(text_to_save) {
			bump_message("Error while writing to file: " + cur_file)
			return
		}
		source_buf.SetModified(false)
		file.Truncate(int64(nbytes))
		file.Close()
		refresh_title()
	}
}

func save_as_cb() {
	file_dialog := gtk.FileChooserDialog2("", source_view.GetTopLevelAsWindow(),
		gtk.GTK_FILE_CHOOSER_ACTION_SAVE,
		gtk.GTK_STOCK_CANCEL, gtk.GTK_RESPONSE_CANCEL,
		gtk.GTK_STOCK_SAVE, gtk.GTK_RESPONSE_ACCEPT)
	file_dialog.SetCurrentFolder(prev_dir)
	res := file_dialog.Run()
	dialog_folder := file_dialog.GetCurrentFolder()
	dialog_file := file_dialog.GetFilename()
	file_dialog.Destroy()
	if gtk.GTK_RESPONSE_ACCEPT == res {
		prev_dir = dialog_folder
		cur_file = dialog_file
		save_cb()
		add_file_record(cur_file, true)
		file_tree_store()
	}
}

func exit_cb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here
	gtk.MainQuit()
}

func close_cb() {
	if "" == cur_file {
		return
	}
	cur_file = ""
	delete_file_record(cur_file)
	file_tree_store()
	refresh_title()
	source_buf_set_content(nil)
}

func paste_done_cb() {
	var be, en gtk.GtkTextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	source_buf.RemoveTagByName("instance", &be, &en)
	selection_flag = false
}
