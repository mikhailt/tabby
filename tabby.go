package main

import (
	//"os"
	"gtk"
	//"gdkpixbuf"
	//"path"
	"fmt"
	//"reflect"
	//"strings"
	//"sync"
	//"unsafe"
	//"time"
	//"runtime"
	"strconv"
	"gdkpixbuf"
)

var lang_man *gtk.GtkSourceLanguageManager
var source_buf *gtk.GtkSourceBuffer
var source_view *gtk.GtkSourceView
var selection_flag bool


var prev_selection string

func highlight_instances() {
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

func main() {
	gtk.Init(nil)
	main_window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.SetSizeRequest(1024, 800)
	main_window.SetTitle("tabby")
	main_window.Connect("destroy",
		func(w *gtk.GtkWidget, user_data string) { gtk.MainQuit() },
		"")

	lang_man = gtk.GtkSourceLanguageManagerGetDefault()
	lang := lang_man.GetLanguage("go")
	if nil == lang.SourceLanguage {
		fmt.Printf("warning: no language specification\n")
	}
	source_buf = gtk.GtkSourceBufferNew()
	source_buf.SetLanguage(lang)
	source_buf.Connect("mark-set", func() { highlight_instances() }, nil)
	source_buf.CreateTag("instance", map[string]string{"background": "#CCCC99"})

	source_view = gtk.GtkSourceViewNewWithBuffer(source_buf)
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

	store := gtk.TreeStore(gdkpixbuf.GetGdkPixbufType(), gtk.TYPE_STRING)
	treeview := gtk.TreeView()
	treeview.ModifyFontEasy("Regular 8")
	treeview.SetModel(store.ToTreeModel())
	treeview.AppendColumn(gtk.TreeViewColumnWithAttributes("", gtk.CellRendererPixbuf(), "pixbuf", 0))
	treeview.AppendColumn(gtk.TreeViewColumnWithAttributes("", gtk.CellRendererText(), "text", 1))
	treeview.SetHeadersVisible(false)

	for n := 1; n <= 10; n++ {
		var iter1, iter2, iter3 gtk.GtkTreeIter
		store.Append(&iter1, nil)
		store.Set(&iter1, gtk.Image().RenderIcon(gtk.GTK_STOCK_DIRECTORY, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf, "Folder"+strconv.Itoa(n))
		store.Append(&iter2, &iter1)
		store.Set(&iter2, gtk.Image().RenderIcon(gtk.GTK_STOCK_DIRECTORY, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf, "SubFolder"+strconv.Itoa(n))
		store.Append(&iter3, &iter2)
		store.Set(&iter3, gtk.Image().RenderIcon(gtk.GTK_STOCK_FILE, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf, "File"+strconv.Itoa(n))
	}

	vbox := gtk.VBox(false, 0)
	hpaned := gtk.HPaned()
	main_window.Add(vbox)

	menubar := gtk.MenuBar()
	vbox.PackStart(menubar, false, false, 0)
	vbox.PackStart(hpaned, true, true, 0)

	file_item := gtk.MenuItemWithMnemonic("_File")
	menubar.Append(file_item)
	file_submenu := gtk.Menu()
	file_item.SetSubmenu(file_submenu)

	exit_item := gtk.MenuItemWithMnemonic("E_xit")
	file_submenu.Append(exit_item)
	exit_item.Connect("activate", func() { gtk.MainQuit() }, nil)

	tree_window := gtk.ScrolledWindow(nil, nil)
	tree_window.SetSizeRequest(300, 0)
	tree_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	hpaned.Add1(tree_window)
	tree_window.Add(treeview)

	text_window := gtk.ScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	hpaned.Add2(text_window)
	text_window.Add(source_view)

	main_window.ShowAll()
	gtk.Main()
}
