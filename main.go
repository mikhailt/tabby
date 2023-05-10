package main

import (
    "fmt"
    "github.com/mattn/go-gtk/gdk"
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gtk"
    "github.com/mattn/go-gtk/gtksourceview"
    file_tree "github.com/mikhailt/tabby/file_tree"
    "strconv"
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

// Add a commentary in front of refresh_title function
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

// Add a commentary in front of buf_changed_cb function
func buf_changed_cb() {
    refresh_title()
}

// Add a commentary in front of bump_message function
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

// Add a commentary in front of bump_question function
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

// Add a commentary in front of init_tabby function
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

    new_item := gtk.NewMenuItemWithMnemonic("_New