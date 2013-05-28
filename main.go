package main

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtksourceview"
	"github.com/mikhailt/tabby/file_tree"
	"strconv"
	"fmt"
)

var main_window *gtk.Window
var source_buf *gtksourceview.SourceBuffer
var source_view *gtksourceview.SourceView

var error_buf *gtk.TextBuffer
var error_view *gtk.TextView
var error_window *gtk.ScrolledWindow

var tree_view *gtk.TreeView
var tree_store *gtk.TreeStore
var tree_model *gtk.TreeModel

var search_view SearchView

var cur_file string = ""
var cur_iter gtk.TreeIter

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
	var val glib.GValue
	tree_model.GetValue(&cur_iter, 0, &val)
	tree_store.Set(&cur_iter, string(icon)+val.GetString()[1:])
}

func buf_changed_cb() {
	refresh_title()
}

func bump_message(m string) {
	dialog := gtk.NewMessageDialog(
		main_window.GetTopLevelAsWindow(),
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_OK,
		m)
	dialog.Run()
	dialog.Destroy()
}

func bump_question(m string) (b bool) {
	dialog := gtk.NewMessageDialog(
		main_window.GetTopLevelAsWindow(),
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_WARNING,
		gtk.BUTTONS_YES_NO,
		m)
	b = dialog.Run() == gtk.RESPONSE_YES
	dialog.Destroy()
	return
}

func init_tabby() {
	init_navigation()
	init_inotify()

	search_view.Init()

	source_buf = gtksourceview.NewSourceBuffer()
	source_buf.Connect("paste-done", paste_done_cb, nil)
	source_buf.Connect("mark-set", mark_set_cb, nil)
	source_buf.Connect("modified-changed", buf_changed_cb, nil)

	init_lang()

	source_buf.CreateTag("instance", map[string]string{"background": "#FF8080"})

	tree_store = gtk.NewTreeStore(gtk.TYPE_STRING)
	tree_view = file_tree.NewFileTree()
	tree_view.ModifyFontEasy("Regular 8")
	tree_model = tree_store.ToTreeModel()
	tree_view.SetModel(tree_model)
	tree_view.SetHeadersVisible(false)
	tree_view.Connect("cursor-changed", tree_view_select_cb, nil)

	error_view = gtk.NewTextView()
	error_view.ModifyFontEasy("Monospace Regular 8")
	error_view.SetEditable(false)
	error_buf = error_view.GetBuffer()

	source_view = gtksourceview.NewSourceViewWithBuffer(source_buf)
	source_view.ModifyFontEasy("Monospace Regular 10")
	source_view.SetAutoIndent(true)
	source_view.SetHighlightCurrentLine(true)
	source_view.SetShowLineNumbers(true)
	source_view.SetRightMarginPosition(80)
	source_view.SetShowRightMargin(true)
	source_view.SetIndentWidth(2)
	source_view.SetTabWidth(2)
	source_view.SetInsertSpacesInsteadOfTabs(opt.space_not_tab)
	source_view.SetDrawSpaces(gtksourceview.SOURCE_DRAW_SPACES_TAB)
	source_view.SetSmartHomeEnd(gtksourceview.SOURCE_SMART_HOME_END_ALWAYS)
	source_view.SetWrapMode(gtk.WRAP_WORD)

	vbox := gtk.NewVBox(false, 0)
	inner_hpaned := gtk.NewHPaned()
	view_vpaned := gtk.NewVPaned()
	outer_hpaned := gtk.NewHPaned()
	outer_hpaned.Add1(inner_hpaned)
	inner_hpaned.Add2(view_vpaned)

	menubar := gtk.NewMenuBar()
	vbox.PackStart(menubar, false, false, 0)
	vbox.PackStart(outer_hpaned, true, true, 0)

	file_item := gtk.NewMenuItemWithMnemonic("_File")
	menubar.Append(file_item)
	file_submenu := gtk.NewMenu()
	file_item.SetSubmenu(file_submenu)

	accel_group := gtk.NewAccelGroup()

	new_item := gtk.NewMenuItemWithMnemonic("_New")
	file_submenu.Append(new_item)
	new_item.Connect("activate", new_cb, nil)
	new_item.AddAccelerator("activate", accel_group, gdk.KEY_n, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	open_item := gtk.NewMenuItemWithMnemonic("_Open")
	file_submenu.Append(open_item)
	open_item.Connect("activate", open_cb, nil)
	open_item.AddAccelerator("activate", accel_group, gdk.KEY_o, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	open_rec_item := gtk.NewMenuItemWithMnemonic("Open _Recursively")
	file_submenu.Append(open_rec_item)
	open_rec_item.Connect("activate", open_rec_cb, nil)

	save_item := gtk.NewMenuItemWithMnemonic("_Save")
	file_submenu.Append(save_item)
	save_item.Connect("activate", save_cb, nil)
	save_item.AddAccelerator("activate", accel_group, gdk.KEY_s, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	save_as_item := gtk.NewMenuItemWithMnemonic("Save _as")
	file_submenu.Append(save_as_item)
	save_as_item.Connect("activate", save_as_cb, nil)

	close_item := gtk.NewMenuItemWithMnemonic("_Close")
	file_submenu.Append(close_item)
	close_item.Connect("activate", close_cb, nil)
	close_item.AddAccelerator("activate", accel_group, gdk.KEY_w, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	exit_item := gtk.NewMenuItemWithMnemonic("E_xit")
	file_submenu.Append(exit_item)
	exit_item.Connect("activate", exit_cb, nil)

	navigation_item := gtk.NewMenuItemWithMnemonic("_Navigation")
	menubar.Append(navigation_item)
	navigation_submenu := gtk.NewMenu()
	navigation_item.SetSubmenu(navigation_submenu)

	prev_instance_item := gtk.NewMenuItemWithMnemonic("_Previous Instance")
	navigation_submenu.Append(prev_instance_item)
	prev_instance_item.Connect("activate", prev_instance_cb, nil)
	prev_instance_item.AddAccelerator("activate", accel_group, gdk.KEY_F2,
		0, gtk.ACCEL_VISIBLE)

	next_instance_item := gtk.NewMenuItemWithMnemonic("_Next Instance")
	navigation_submenu.Append(next_instance_item)
	next_instance_item.Connect("activate", next_instance_cb, nil)
	next_instance_item.AddAccelerator("activate", accel_group, gdk.KEY_F3,
		0, gtk.ACCEL_VISIBLE)

	prev_result_item := gtk.NewMenuItemWithMnemonic("Prev search result")
	navigation_submenu.Append(prev_result_item)
	prev_result_item.Connect("activate", func() {search_view.PrevResult()}, nil)
	prev_result_item.AddAccelerator("activate", accel_group, gdk.KEY_F4,
		0, gtk.ACCEL_VISIBLE)

	next_result_item := gtk.NewMenuItemWithMnemonic("Next search result")
	navigation_submenu.Append(next_result_item)
	next_result_item.Connect("activate", func() {search_view.NextResult()}, nil)
	next_result_item.AddAccelerator("activate", accel_group, gdk.KEY_F5,
		0, gtk.ACCEL_VISIBLE)

	find_item := gtk.NewMenuItemWithMnemonic("_Find")
	navigation_submenu.Append(find_item)
	find_item.Connect("activate", find_cb, nil)
	find_item.AddAccelerator("activate", accel_group, gdk.KEY_f, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	find_file_item := gtk.NewMenuItemWithMnemonic("_Find file")
	navigation_submenu.Append(find_file_item)
	find_file_item.Connect("activate", find_file_cb, nil)
	find_file_item.AddAccelerator("activate", accel_group, gdk.KEY_d, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	fnr_item := gtk.NewMenuItemWithMnemonic("Find and Replace")
	navigation_submenu.Append(fnr_item)
	fnr_item.Connect("activate", fnr_cb, nil)
	fnr_item.AddAccelerator("activate", accel_group, gdk.KEY_r, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	prev_file_item := gtk.NewMenuItemWithMnemonic("Prev File")
	navigation_submenu.Append(prev_file_item)
	prev_file_item.Connect("activate", prev_file_cb, nil)
	prev_file_item.AddAccelerator("activate", accel_group, gdk.KEY_F7,
		0, gtk.ACCEL_VISIBLE)

	next_file_item := gtk.NewMenuItemWithMnemonic("Next File")
	navigation_submenu.Append(next_file_item)
	next_file_item.Connect("activate", next_file_cb, nil)
	next_file_item.AddAccelerator("activate", accel_group, gdk.KEY_F8,
		0, gtk.ACCEL_VISIBLE)

	tools_item := gtk.NewMenuItemWithMnemonic("_Tools")
	menubar.Append(tools_item)
	tools_submenu := gtk.NewMenu()
	tools_item.SetSubmenu(tools_submenu)

	gofmt_item := gtk.NewMenuItemWithMnemonic("_Gofmt")
	tools_submenu.Append(gofmt_item)
	gofmt_item.Connect("activate", gofmt_cb, nil)
	gofmt_item.AddAccelerator("activate", accel_group, gdk.KEY_F9,
		0, gtk.ACCEL_VISIBLE)

	gofmtAll_item := gtk.NewMenuItemWithMnemonic("Gofmt _All")
	tools_submenu.Append(gofmtAll_item)
	gofmtAll_item.Connect("activate", gofmt_all, nil)
	gofmtAll_item.AddAccelerator("activate", accel_group, gdk.KEY_F9, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	options_item := gtk.NewMenuItemWithMnemonic("_Options")
	menubar.Append(options_item)
	options_submenu := gtk.NewMenu()
	options_item.SetSubmenu(options_submenu)

	search_chkitem := gtk.NewCheckMenuItemWithMnemonic("Show _Searchview")
	options_submenu.Append(search_chkitem)
	search_chkitem.SetActive(opt.show_search)
	search_chkitem.Connect("toggled", func() { search_chk_cb(search_chkitem.GetActive()) }, nil)

	error_chkitem := gtk.NewCheckMenuItemWithMnemonic("Show _Errorview")
	options_submenu.Append(error_chkitem)
	error_chkitem.SetActive(opt.show_error)
	error_chkitem.Connect("toggled", func() { error_chk_cb(error_chkitem.GetActive()) }, nil)

	notab_chkitem := gtk.NewCheckMenuItemWithMnemonic("Spaces for _Tabs")
	options_submenu.Append(notab_chkitem)
	notab_chkitem.SetActive(opt.space_not_tab)
	notab_chkitem.Connect("toggled", func() { notab_chk_cb(notab_chkitem.GetActive()) }, nil)

	font_item := gtk.NewMenuItemWithMnemonic("_Font")
	options_submenu.Append(font_item)
	font_item.Connect("activate", font_cb, nil)

	tabsize_item := gtk.NewMenuItemWithMnemonic("_Tab size")
	options_submenu.Append(tabsize_item)
	tabsize_submenu := gtk.NewMenu()
	tabsize_item.SetSubmenu(tabsize_submenu)
	const tabsize_cnt = 8
	tabsize_chk := make([]*gtk.CheckMenuItem, tabsize_cnt)
	for y := 0; y < tabsize_cnt; y++ {
		tabsize_chk[y] = gtk.NewCheckMenuItemWithMnemonic(strconv.Itoa(y + 1))
		tabsize_submenu.Append(tabsize_chk[y])
		cur_ind := y
		tabsize_chk[y].Connect("activate", func() {
			if false == tabsize_chk[cur_ind].GetActive() {
				active_cnt := 0
				for j := 0; j < tabsize_cnt; j++ {
					if tabsize_chk[j].GetActive() {
						active_cnt++
					}
				}
				if 0 == active_cnt {
					tabsize_chk[cur_ind].SetActive(true)
				}
				return
			}
			for j := 0; j < tabsize_cnt; j++ {
				if j != cur_ind {
					tabsize_chk[j].SetActive(false)
				}
			}
			options_set_tabsize(cur_ind + 1)
		},
			nil)
	}

	tree_window := gtk.NewScrolledWindow(nil, nil)
	tree_window.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	inner_hpaned.Add1(tree_window)
	tree_window.Add(tree_view)

	outer_hpaned.Add2(search_view.window)

	text_window := gtk.NewScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_ALWAYS)
	view_vpaned.Add1(text_window)
	text_window.Add(source_view)

	error_window = gtk.NewScrolledWindow(nil, nil)
	error_window.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_ALWAYS)
	view_vpaned.Add2(error_window)
	error_window.Add(error_view)

	inner_hpaned.Connect("size_request", func() { ohp_cb(outer_hpaned.GetPosition()) }, nil)
	view_vpaned.Connect("size_request", func() { ihp_cb(inner_hpaned.GetPosition()) }, nil)
	source_view.Connect("size_request", func() { vvp_cb(view_vpaned.GetPosition()) }, nil)
	outer_hpaned.SetPosition(opt.ohp_position)
	inner_hpaned.SetPosition(opt.ihp_position)
	view_vpaned.SetPosition(opt.vvp_position)

	main_window = gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	main_window.AddAccelGroup(accel_group)
	main_window.SetSizeRequest(400, 200) //minimum size
	main_window.Resize(opt.window_width, opt.window_height)
	main_window.Move(opt.window_x, opt.window_y)
	main_window.Connect("destroy", exit_cb, "")
	main_window.Connect("configure-event", window_event_cb, "")
	main_window.Add(vbox)
	// init_tabby blocks for some reason when is called after ShowAll.
	init_vars()
	main_window.ShowAll()
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
	fmt.Println("tabby: " + m)
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
	gdk.ThreadsInit()
	if false == init_args() {
		return
	}
	load_options()
	defer save_options()
	
	gdk.ThreadsEnter()
	gtk.Init(nil)
	init_tabby()
	gtk.Main()
}
