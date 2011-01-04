package main

import (
	"gtk"
	"gdk"
	"file_tree"
)

var main_window *gtk.GtkWindow
var source_buf *gtk.GtkSourceBuffer
var source_view *gtk.GtkSourceView

var error_buf *gtk.GtkTextBuffer
var error_view *gtk.GtkTextView
var error_window *gtk.GtkScrolledWindow

var tree_view *gtk.GtkTreeView
var tree_store *gtk.GtkTreeStore
var tree_model *gtk.GtkTreeModel

var search_view *gtk.GtkTreeView
var search_store *gtk.GtkTreeStore
var search_model *gtk.GtkTreeModel
var search_window *gtk.GtkScrolledWindow

var cur_file string = ""
var cur_iter gtk.GtkTreeIter

func refresh_title() {
	if "" == cur_file {
		main_window.SetTitle("*")
		return
	}
	var icon byte
	if source_buf.GetModified() {
		main_window.SetTitle("* " + cur_file)
		icon = 'C'
	} else {
		main_window.SetTitle(cur_file)
		icon = 'B'
	}
	if tree_store.IterIsValid(&cur_iter) {
		var val gtk.GValue
		tree_model.GetValue(&cur_iter, 0, &val)
		tree_store.Set(&cur_iter, string(icon)+val.GetString()[1:])
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

func bump_question(m string) (b bool) {
	dialog := gtk.MessageDialog(
		main_window.GetTopLevelAsWindow(),
		gtk.GTK_DIALOG_MODAL,
		gtk.GTK_MESSAGE_WARNING,
		gtk.GTK_BUTTONS_YES_NO,
		m)
	b = dialog.Run() == gtk.GTK_RESPONSE_YES
	dialog.Destroy()
	return
}

func init_tabby() {
	init_navigation()
	gdk.ThreadsInit()
	inotify_init()

	lang_man := gtk.SourceLanguageManagerGetDefault()
	lang := lang_man.GetLanguage("go")
	if nil == lang.SourceLanguage {
		tabby_log("warning: no language specification")
	}
	source_buf = gtk.SourceBuffer()
	source_buf.SetLanguage(lang)
	source_buf.Connect("paste-done", paste_done_cb, nil)
	source_buf.Connect("mark-set", mark_set_cb, nil)
	source_buf.Connect("changed", buf_changed_cb, nil)

	source_buf.CreateTag("instance", map[string]string{"background": "#FF8080"})

	tree_store = gtk.TreeStore(gtk.TYPE_STRING)
	tree_view = file_tree.NewFileTree()
	tree_view.ModifyFontEasy("Regular 8")
	tree_model = tree_store.ToTreeModel()
	tree_view.SetModel(tree_model)
	tree_view.SetHeadersVisible(false)
	tree_view.Connect("cursor-changed", tree_view_select_cb, nil)

	search_store = gtk.TreeStore(gtk.TYPE_STRING)
	search_view = file_tree.NewSearchTree()
	search_view.ModifyFontEasy("Regular 8")
	search_model = search_store.ToTreeModel()
	search_view.SetModel(search_model)
	search_view.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererText(), "text", 0))
	search_view.SetHeadersVisible(false)
	search_view.Connect("cursor-changed", search_view_select_cb, nil)

	error_view = gtk.TextView()
	error_view.ModifyFontEasy("Monospace Regular 8")
	error_view.SetEditable(false)
	error_buf = error_view.GetBuffer()

	source_view = gtk.SourceViewWithBuffer(source_buf)
	source_view.ModifyFontEasy("Monospace Regular 10")
	source_view.SetAutoIndent(true)
	source_view.SetHighlightCurrentLine(true)
	source_view.SetShowLineNumbers(true)
	source_view.SetRightMarginPosition(80)
	source_view.SetShowRightMargin(true)
	source_view.SetIndentWidth(2)
	source_view.SetInsertSpacesInsteadOfTabs(opt.space_not_tab)
	source_view.SetDrawSpaces(gtk.GTK_SOURCE_DRAW_SPACES_TAB)
	source_view.SetTabWidth(2)
	source_view.SetSmartHomeEnd(gtk.GTK_SOURCE_SMART_HOME_END_ALWAYS)
	source_view.SetWrapMode(gtk.GTK_WRAP_WORD)

	vbox := gtk.VBox(false, 0)
	inner_hpaned := gtk.HPaned()
	view_vpaned := gtk.VPaned()
	outer_hpaned := gtk.HPaned()
	outer_hpaned.Add1(inner_hpaned)
	inner_hpaned.Add2(view_vpaned)

	menubar := gtk.MenuBar()
	vbox.PackStart(menubar, false, false, 0)
	vbox.PackStart(outer_hpaned, true, true, 0)

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

	open_rec_item := gtk.MenuItemWithMnemonic("Open _Recursively")
	file_submenu.Append(open_rec_item)
	open_rec_item.Connect("activate", open_rec_cb, nil)

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

	prev_instance_item := gtk.MenuItemWithMnemonic("_Previous Instance")
	navigation_submenu.Append(prev_instance_item)
	prev_instance_item.Connect("activate", prev_instance_cb, nil)
	prev_instance_item.AddAccelerator("activate", accel_group, gdk.GDK_F2,
		0, gtk.GTK_ACCEL_VISIBLE)

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

	fnr_item := gtk.MenuItemWithMnemonic("Find and Replace")
	navigation_submenu.Append(fnr_item)
	fnr_item.Connect("activate", fnr_cb, nil)
	fnr_item.AddAccelerator("activate", accel_group, gdk.GDK_r,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	prev_file_item := gtk.MenuItemWithMnemonic("Prev File")
	navigation_submenu.Append(prev_file_item)
	prev_file_item.Connect("activate", prev_file_cb, nil)
	prev_file_item.AddAccelerator("activate", accel_group, gdk.GDK_F7,
		0, gtk.GTK_ACCEL_VISIBLE)

	next_file_item := gtk.MenuItemWithMnemonic("Next File")
	navigation_submenu.Append(next_file_item)
	next_file_item.Connect("activate", next_file_cb, nil)
	next_file_item.AddAccelerator("activate", accel_group, gdk.GDK_F8,
		0, gtk.GTK_ACCEL_VISIBLE)

	tools_item := gtk.MenuItemWithMnemonic("_Tools")
	menubar.Append(tools_item)
	tools_submenu := gtk.Menu()
	tools_item.SetSubmenu(tools_submenu)

	gofmt_item := gtk.MenuItemWithMnemonic("_Gofmt")
	tools_submenu.Append(gofmt_item)
	gofmt_item.Connect("activate", gofmt_cb, nil)
	gofmt_item.AddAccelerator("activate", accel_group, gdk.GDK_F9,
		0, gtk.GTK_ACCEL_VISIBLE)

	gofmtAll_item := gtk.MenuItemWithMnemonic("Gofmt _All")
	tools_submenu.Append(gofmtAll_item)
	gofmtAll_item.Connect("activate", gofmt_all, nil)
	gofmtAll_item.AddAccelerator("activate", accel_group, gdk.GDK_F9,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	view_item := gtk.MenuItemWithMnemonic("_View")
	menubar.Append(view_item)
	view_submenu := gtk.Menu()
	view_item.SetSubmenu(view_submenu)

	search_chkitem := gtk.CheckMenuItemWithMnemonic("Show _Searchview")
	view_submenu.Append(search_chkitem)
	search_chkitem.SetActive(opt.show_search)
	search_chkitem.Connect("toggled", func() { search_chk_cb(search_chkitem.GetActive()) }, nil)

	error_chkitem := gtk.CheckMenuItemWithMnemonic("Show _Errorview")
	view_submenu.Append(error_chkitem)
	error_chkitem.SetActive(opt.show_error)
	error_chkitem.Connect("toggled", func() { error_chk_cb(error_chkitem.GetActive()) }, nil)

	notab_chkitem := gtk.CheckMenuItemWithMnemonic("Spaces for _Tabs")
	view_submenu.Append(notab_chkitem)
	notab_chkitem.SetActive(opt.space_not_tab)
	notab_chkitem.Connect("toggled", func() { notab_chk_cb(notab_chkitem.GetActive()) }, nil)

	tree_window := gtk.ScrolledWindow(nil, nil)
	tree_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	inner_hpaned.Add1(tree_window)
	tree_window.Add(tree_view)

	search_window = gtk.ScrolledWindow(nil, nil)
	search_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	outer_hpaned.Add2(search_window)
	search_window.Add(search_view)

	text_window := gtk.ScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	view_vpaned.Add1(text_window)
	text_window.Add(source_view)

	error_window = gtk.ScrolledWindow(nil, nil)
	error_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	view_vpaned.Add2(error_window)
	error_window.Add(error_view)

	inner_hpaned.Connect("size_request", func() { ohp_cb(outer_hpaned.GetPosition()) }, nil)
	view_vpaned.Connect("size_request", func() { ihp_cb(inner_hpaned.GetPosition()) }, nil)
	source_view.Connect("size_request", func() { vvp_cb(view_vpaned.GetPosition()) }, nil)
	outer_hpaned.SetPosition(opt.ohp_position)
	inner_hpaned.SetPosition(opt.ihp_position)
	view_vpaned.SetPosition(opt.vvp_position)

	main_window = gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.AddAccelGroup(accel_group)
	main_window.SetSizeRequest(400, 200) //minimum size
	main_window.Resize(opt.window_width, opt.window_height)
	main_window.Move(opt.window_x, opt.window_y)
	main_window.Connect("destroy", exit_cb, "")
	main_window.Connect("configure-event", window_event_cb, "")
	main_window.Add(vbox)
	// init_tabby blocks for some reason if called after ShowAll.
	init_vars()
	main_window.ShowAll()
	search_window.SetVisible(opt.show_search)
	error_window.SetVisible(opt.show_error)
	// Cannot be called before ShowAll. This is also not clear.
	file_switch_to(file_stack_pop())
	stack_prev(&file_stack_max)
	if "" == cur_file {
		new_cb()
	}
	source_view.GrabFocus()
}

func tabby_log(m string) {
	println("tabby: " + m)
}

func init_vars() {
	file_map = make(map[string]*FileRecord)
	refresh_title()
	if 0 == len(tabby_args) {
		session_restore()
	}
	open_files_from_args()
	file_tree_store()
}

func main() {
	if false == init_args() {
		return
	}
	load_options()
	defer save_options()
	gtk.Init(nil)
	init_tabby()
	gtk.Main()
}
