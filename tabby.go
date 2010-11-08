package main

import (
	"os"
	"gtk"
	"gdk"
	"gdkpixbuf"
	"glib"
)

type FileRecord struct {
	name string
	iter *gtk.GtkTreeIter
}

var file_map map[string]FileRecord

var main_window *gtk.GtkWindow
var source_buf *gtk.GtkSourceBuffer
var tree_store *gtk.GtkTreeStore
var source_view *gtk.GtkSourceView
var selection_flag bool
var prev_selection string
var prev_dir string
var cur_file string

func delete_file_from_tree(name string) {
	if name == "" {
		return
	}
	file_rec, found := file_map[name]
	if false == found {
		return
	}
	if false == tree_store.IterIsValid(file_rec.iter) {
		bump_message("delete_file_from_tree: iterator is not valid!")
	}
	tree_store.Remove(file_rec.iter)
}

func add_file_to_tree(name string) {
	iter := new(gtk.GtkTreeIter)
	tree_store.Append(iter, nil)
	tree_store.Set(iter,
		gtk.Image().RenderIcon(gtk.GTK_STOCK_FILE, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf,
		name)
	file_map[name] = FileRecord{name, iter}
}

func buf_changed_cb() {
	if source_buf.GetModified() {
		main_window.SetTitle("* " + cur_file)
	} else {
		main_window.SetTitle(cur_file)
	}
}

func mark_set_cb() {
	//println("mark_set_cb called")
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

func bump_message(m string) {
	dialog := gtk.MessageDialog(
		main_window.GetTopLevelAsWindow(),
		gtk.GTK_DIALOG_MODAL,
		gtk.GTK_MESSAGE_INFO,
		gtk.GTK_BUTTONS_OK,
		m)
	dialog.Run()
	dialog.Destroy()
}

func open_cb() {
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
		if false == glib.Utf8Validate(buf, nread, nil) {
			bump_message("File " + cur_file + " is not correct utf8 text")
			close_cb()
			return
		}
		source_buf.BeginNotUndoableAction()
		source_buf.SetText(string(buf))
		source_buf.SetModified(false)
		source_buf.EndNotUndoableAction()

		add_file_to_tree(cur_file)
	}
}

func save_cb() {
	if "" == cur_file {
		save_as_cb()
	} else {
		file, _ := os.Open(cur_file, os.O_CREAT|os.O_WRONLY, 0700)
		if nil == file {
			// To be replaced with dialog.
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
		main_window.SetTitle(cur_file)
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
	}
}

func exit_cb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here
	gtk.MainQuit()
}

func close_cb() {
	delete_file_from_tree(cur_file)
	cur_file = ""
	main_window.SetTitle("")
	source_buf.BeginNotUndoableAction()
	source_buf.SetText("")
	source_buf.EndNotUndoableAction()
}

func paste_done_cb() {
	//println("paste_done_cb")
	var be, en gtk.GtkTextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	source_buf.RemoveTagByName("instance", &be, &en)
	selection_flag = false
}

func init_widgets() {
	lang_man := gtk.SourceLanguageManagerGetDefault()
	lang := lang_man.GetLanguage("go")
	if nil == lang.SourceLanguage {
		println("warning: no language specification")
	}
	source_buf = gtk.SourceBuffer()
	source_buf.SetLanguage(lang)
	source_buf.Connect("paste-done", paste_done_cb, nil)
	source_buf.Connect("mark-set", mark_set_cb, nil)
	source_buf.Connect("changed", buf_changed_cb, nil)

	source_buf.CreateTag("instance", map[string]string{"background": "#CCCC99"})

	tree_store = gtk.TreeStore(gdkpixbuf.GetGdkPixbufType(), gtk.TYPE_STRING)
	treeview := gtk.TreeView()
	treeview.ModifyFontEasy("Regular 8")
	treeview.SetModel(tree_store.ToTreeModel())
	treeview.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererPixbuf(), "pixbuf", 0))
	treeview.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererText(), "text", 1))
	treeview.SetHeadersVisible(false)

	source_view = gtk.SourceViewWithBuffer(source_buf)
	source_view.ModifyFontEasy("Monospace Regular 10")
	source_view.SetAutoIndent(true)
	source_view.SetHighlightCurrentLine(true)
	source_view.SetShowLineNumbers(true)
	source_view.SetRightMarginPosition(80)
	source_view.SetShowRightMargin(true)
	source_view.SetIndentWidth(2)
	source_view.SetInsertSpacesInsteadOfTabs(true)
	source_view.SetDrawSpaces(gtk.GTK_SOURCE_DRAW_SPACES_TAB)
	source_view.SetTabWidth(2)
	source_view.SetSmartHomeEnd(gtk.GTK_SOURCE_SMART_HOME_END_ALWAYS)

	vbox := gtk.VBox(false, 0)
	hpaned := gtk.HPaned()

	menubar := gtk.MenuBar()
	vbox.PackStart(menubar, false, false, 0)
	vbox.PackStart(hpaned, true, true, 0)

	file_item := gtk.MenuItemWithMnemonic("_File")
	menubar.Append(file_item)
	file_submenu := gtk.Menu()
	file_item.SetSubmenu(file_submenu)

	accel_group := gtk.AccelGroup()

	open_item := gtk.MenuItemWithMnemonic("_Open")
	file_submenu.Append(open_item)
	open_item.Connect("activate", open_cb, nil)
	open_item.AddAccelerator("activate", accel_group, gdk.GDK_o,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	save_item := gtk.MenuItemWithMnemonic("_Save")
	file_submenu.Append(save_item)
	save_item.Connect("activate", save_cb, nil)
	save_item.AddAccelerator("activate", accel_group, gdk.GDK_s,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	save_as_item := gtk.MenuItemWithMnemonic("Save _as")
	file_submenu.Append(save_as_item)
	save_as_item.Connect("activate", save_as_cb, nil)

	close_item := gtk.MenuItemWithMnemonic("_Close")
	file_submenu.Append(close_item)
	close_item.Connect("activate", close_cb, nil)
	close_item.AddAccelerator("activate", accel_group, gdk.GDK_w,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	exit_item := gtk.MenuItemWithMnemonic("E_xit")
	file_submenu.Append(exit_item)
	exit_item.Connect("activate", exit_cb, nil)

	tree_window := gtk.ScrolledWindow(nil, nil)
	tree_window.SetSizeRequest(300, 0)
	tree_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	hpaned.Add1(tree_window)
	tree_window.Add(treeview)

	text_window := gtk.ScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	hpaned.Add2(text_window)
	text_window.Add(source_view)

	main_window = gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.Maximize()
	main_window.SetTitle("tabby")
	main_window.Connect("destroy", exit_cb, "")
	main_window.Add(vbox)
	main_window.ShowAll()
	main_window.AddAccelGroup(accel_group)

	source_view.GrabFocus()
}

func init_vars() {
	file_map = make(map[string]FileRecord)
}

func main() {
	gtk.Init(nil)
	init_widgets()
	init_vars()
	gtk.Main()
}
