package main

import (
	"gtk"
	"gdk"
	"gdkpixbuf"
)

var main_window *gtk.GtkWindow
var source_buf *gtk.GtkSourceBuffer
var tree_view *gtk.GtkTreeView
var tree_store *gtk.GtkTreeStore
var tree_model *gtk.GtkTreeModel
var source_view *gtk.GtkSourceView
var selection_flag bool
var prev_selection string
var prev_dir string
var cur_file string

func refresh_title() {
	if "" == cur_file {
		main_window.SetTitle("*")
		return
	}
	if source_buf.GetModified() {
		main_window.SetTitle("* " + cur_file)
	} else {
		main_window.SetTitle(cur_file)
	}
}

func buf_changed_cb() {
	refresh_title()
}

func mark_set_cb() {
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

func source_buf_set_content(buf []byte) {
	source_buf.BeginNotUndoableAction()
	if nil == buf {
		source_buf.SetText("")
	} else {
		source_buf.SetText(string(buf))
	}
	source_buf.SetModified(false)
	source_buf.EndNotUndoableAction()
}

func tree_view_select_cb() {
	var path *gtk.GtkTreePath
	var column *gtk.GtkTreeViewColumn
	tree_view.GetCursor(&path, &column)
	var iter gtk.GtkTreeIter
	tree_model.GetIterFromString(&iter, path.String())
	sel_file := tree_view_path(&iter)
	if name_is_dir(sel_file) {
		return
	}
	file_save_current()
	file_switch_to(sel_file)
}

func tree_view_path(iter *gtk.GtkTreeIter) string {
	var ans string
	ans = ""
	for {
		var val gtk.GValue
		var next gtk.GtkTreeIter
		tree_model.GetValue(iter, 1, &val)
		ans = val.GetString() + ans
		if false == tree_model.IterParent(&next, iter) {
			break
		}
		iter = &next
	}
	return ans
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

	source_buf.CreateTag("instance", map[string]string{"background": "#FF8080"})

	tree_store = gtk.TreeStore(gdkpixbuf.GetGdkPixbufType(), gtk.TYPE_STRING)
	tree_view = gtk.TreeView()
	tree_view.ModifyFontEasy("Regular 8")
	tree_model = tree_store.ToTreeModel()
	tree_view.SetModel(tree_model)
	tree_view.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererPixbuf(), "pixbuf", 0))
	tree_view.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererText(), "text", 1))
	tree_view.SetHeadersVisible(false)
	tree_view.Connect("cursor-changed", tree_view_select_cb, nil)

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

	new_item := gtk.MenuItemWithMnemonic("_New")
	file_submenu.Append(new_item)
	new_item.Connect("activate", new_cb, nil)
	new_item.AddAccelerator("activate", accel_group, gdk.GDK_n,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

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
	tree_window.Add(tree_view)

	text_window := gtk.ScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	hpaned.Add2(text_window)
	text_window.Add(source_view)

	main_window = gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.Maximize()
	main_window.Connect("destroy", exit_cb, "")
	main_window.Add(vbox)
	main_window.ShowAll()
	main_window.AddAccelGroup(accel_group)

	source_view.GrabFocus()
}

func init_vars() {
	file_map = make(map[string]*FileRecord)
	cur_file = ""
	refresh_title()
}

func main() {
	gtk.Init(nil)
	init_widgets()
	init_vars()
	gtk.Main()
}
