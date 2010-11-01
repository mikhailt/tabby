package main

import (
	//"os"
	"gtk"
	//"gdkpixbuf"
	//"path"
	//"fmt"
	//"reflect"
	//"strings"
	//"sync"
	//"unsafe"
	//"time"
	//"runtime"
)

func main() {
	gtk.Init(nil)
	main_window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.SetSizeRequest(1024, 800)
	main_window.SetTitle("tabby")
	main_window.Connect("destroy",
		func(w *gtk.GtkWidget, user_data string) { gtk.MainQuit() },
		"")

	vbox := gtk.VBox(false, 0)
	main_window.Add(vbox)

	menubar := gtk.MenuBar()
	vbox.PackStart(menubar, false, false, 0)

	file_item := gtk.MenuItemWithMnemonic("_File")
	menubar.Append(file_item)
	file_submenu := gtk.Menu()
	file_item.SetSubmenu(file_submenu)

	exit_item := gtk.MenuItemWithMnemonic("E_xit")
	file_submenu.Append(exit_item)
	exit_item.Connect("activate", func() { gtk.MainQuit() }, nil)

	scrolled_window := gtk.ScrolledWindow(nil, nil)
	scrolled_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	vbox.PackStart(scrolled_window, true, true, 0)

	source_view := gtk.SourceView()
	source_view.ModifyFontEasy("Monospace Regular 10")
	source_view.SetAutoIndent(true)
	source_view.SetHighlightCurrentLine(true)
	source_view.SetShowLineNumbers(true)
	source_view.SetRightMarginPosition(80)
	source_view.SetShowRightMargin(true)
	source_view.SetIndentWidth(2)
	source_view.SetInsertSpacesInsteadOfTabs(true)
	source_view.SetDrawSpaces(gtk.GTK_SOURCE_DRAW_SPACES_SPACE |
		gtk.GTK_SOURCE_DRAW_SPACES_TAB)
	source_view.SetTabWidth(2)
	source_view.SetSmartHomeEnd(gtk.GTK_SOURCE_SMART_HOME_END_ALWAYS)

	scrolled_window.Add(source_view)

	main_window.ShowAll()
	gtk.Main()
}
