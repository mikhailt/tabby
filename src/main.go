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

var cur_file string
var cur_iter gtk.GtkTreeIter

func refresh_title() {
	if "" == cur_file {
		main_window.SetTitle("*")
		return
	}
	var gtk_icon string
	if source_buf.GetModified() {
		main_window.SetTitle("* " + cur_file)
		gtk_icon = gtk.GTK_STOCK_DELETE
	} else {
		main_window.SetTitle(cur_file)
		gtk_icon = gtk.GTK_STOCK_FILE
	}
	if tree_store.IterIsValid(&cur_iter) {
		tree_store.Set(&cur_iter,
			gtk.Image().RenderIcon(gtk_icon, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf)
	}
}

func buf_changed_cb() {
	refresh_title()
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

func init_tabby() {
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

  navigation_item := gtk.MenuItemWithMnemonic("_Navigation")
	menubar.Append(navigation_item)
	navigation_submenu := gtk.Menu()
	navigation_item.SetSubmenu(navigation_submenu)
	
	next_instance_item := gtk.MenuItemWithMnemonic("_Next Instance")
	navigation_submenu.Append(next_instance_item)
	next_instance_item.Connect("activate", next_instance_cb, nil)
	next_instance_item.AddAccelerator("activate", accel_group, gdk.GDK_F3,
		0, gtk.GTK_ACCEL_VISIBLE)
		
	find_item := gtk.MenuItemWithMnemonic("_Find")
	navigation_submenu.Append(find_item)
	find_item.Connect("activate", find_cb, nil)
	find_item.AddAccelerator("activate", accel_group, gdk.GDK_f,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	tree_window := gtk.ScrolledWindow(nil, nil)
	tree_window.SetSizeRequest(330, 0)
	tree_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	hpaned.Add1(tree_window)
	tree_window.Add(tree_view)

	text_window := gtk.ScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	hpaned.Add2(text_window)
	text_window.Add(source_view)

	main_window = gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.AddAccelGroup(accel_group)
	main_window.Maximize()
	main_window.Connect("destroy", exit_cb, "")
	main_window.Add(vbox)
	// init_tabby blocks for some reason if called after ShowAll.
	init_vars()
	main_window.ShowAll()
	// Cannot be called before ShowAll. This is also not clear.
	file_switch_to(file_stack_pop())
	source_view.GrabFocus()
}

func init_vars() {
	file_map = make(map[string]*FileRecord)
	cur_file = ""
  refresh_title()
  session_restore()
  file_tree_store()
}

func main() {
	gtk.Init(nil)
	init_tabby()
	gtk.Main()
}
