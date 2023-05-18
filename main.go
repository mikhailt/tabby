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

// main_window is the main window of the application.
var main_window *gtk.Window

// source_buf is the buffer of the source view.
var source_buf *gtksourceview.SourceBuffer

// source_view is the main view for editing files.
var source_view *gtksourceview.SourceView

// error_buf is the buffer for the error view.
var error_buf *gtk.TextBuffer

// error_view is the view for displaying errors.
var error_view *gtk.TextView

// error_window is the window that holds the error_view.
var error_window *gtk.ScrolledWindow

// tree_view is the view for displaying files in a tree structure.
var tree_view *gtk.TreeView

// tree_store is the data store for the tree view.
var tree_store *gtk.TreeStore

// tree_model is the model for the tree view.
var tree_model *gtk.TreeModel

// search_view is the view for displaying search results.
var search_view SearchView

// cur_file is the name of the currently open file.
var cur_file string = ""

// cur_iter is the current iterator for the tree view.
var cur_iter gtk.TreeIter

// refresh_title updates the main window title based on the currently open file and its modified state.
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

// buf_changed_cb is called when the buffer has been modified.
func buf_changed_cb() {
	refresh_title()
}

// bump_message displays a message dialog with an "OK" button.
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

// bump_question displays a message dialog with "Yes" and "No" buttons.
// It returns true if "Yes" was clicked, false otherwise.
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

// init_tabby initializes the application.
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
	inner_h