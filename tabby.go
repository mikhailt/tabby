package main

import (
	//"os"
	"gtk"
	//"gdkpixbuf"
	//"path"
	"fmt"
	//"reflect"
	"strings"
	//"runtime"
	//"unsafe"
)

func buffer_changed(buf *gtk.GtkTextBuffer) {
	fmt.Printf("changed\n")
}


func highlight_instances(buf *gtk.GtkTextBuffer) {
	var be, en gtk.GtkTextIter
	var start, end gtk.GtkTextIter

	buf.GetStartIter(&start)
	buf.GetEndIter(&end)

	buf.RemoveTagByName("global_font", &start, &end)
	buf.ApplyTagByName("global_font", &start, &end)
	buf.RemoveTagByName("instance", &start, &end)

	if buf.GetHasSelection() {
		fmt.Printf("beep\n")
		buf.GetIterAtMark(&be, buf.GetMark("selection_bound"))
		buf.GetIterAtMark(&en, buf.GetMark("insert"))
		selection := buf.GetSlice(&be, &en, false)
		sel_len := len(selection)
		if (20 < sel_len) || (2 > sel_len ) {
			return
		}

		text := buf.GetSlice(&start, &end, false)
		shift := 0
		for be_ind := 0; ; {
			be_ind = strings.Index(text, selection)
			if -1 == be_ind {
				break
			}
			shift += be_ind
			buf.GetIterAtOffset(&be, shift)
			buf.GetIterAtOffset(&en, shift+sel_len)
			buf.ApplyTagByName("instance", &be, &en)
			text = text[be_ind+sel_len:]
			shift += sel_len
		}
	}
}

func main() {
	gtk.Init(nil)
	window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	window.SetTitle("tabby")
	window.Connect("destroy", func(w *gtk.GtkWidget, user_data string) {
		println("got destroy!", user_data)
		gtk.MainQuit()
	},
		"foo")

	//--------------------------------------------------------
	// GtkVBox
	//--------------------------------------------------------
	vbox := gtk.VBox(false, 1)

	//--------------------------------------------------------
	// GtkMenuBar
	//--------------------------------------------------------
	menubar := gtk.MenuBar()
	vbox.PackStart(menubar, false, false, 0)

	//--------------------------------------------------------
	// GtkMenuItem
	//--------------------------------------------------------
	filemenu := gtk.MenuItemWithMnemonic("_File")
	menubar.Append(filemenu)
	filesubmenu := gtk.Menu()
	filemenu.SetSubmenu(filesubmenu)

	exitmenuitem := gtk.MenuItemWithMnemonic("E_xit")
	exitmenuitem.Connect("activate", func() {
		gtk.MainQuit()
	},
		nil)
	filesubmenu.Append(exitmenuitem)

	filemenu = gtk.MenuItemWithMnemonic("_Help")
	menubar.Append(filemenu)
	filesubmenu = gtk.Menu()
	filemenu.SetSubmenu(filesubmenu)

	//--------------------------------------------------------
	// GtkVPaned
	//--------------------------------------------------------
	vpaned := gtk.VPaned()
	vbox.Add(vpaned)

	// GtkFrame
	frame := gtk.Frame("File name")
	framebox := gtk.VBox(false, 1)
	frame.Add(framebox)
	vpaned.Add2(frame)

	//--------------------------------------------------------
	// GtkTextView
	//--------------------------------------------------------
	swin := gtk.ScrolledWindow(nil, nil)
	swin.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	swin.SetShadowType(gtk.GTK_SHADOW_IN)
	textview := gtk.TextView()
	var start, end gtk.GtkTextIter
	buffer := textview.GetBuffer()

	buffer.Connect("mark-set", func() {
		highlight_instances(buffer)
	},
		nil)
	buffer.Connect("changed", func() {
		buffer_changed(buffer)
	},
		nil)
	buffer.GetStartIter(&start)
	buffer.Insert(&start, "Hello\nSome more words\n#include \"iostream\"\nHello")
	buffer.GetEndIter(&end)
	buffer.Insert(&end, "World!")
	buffer.CreateTag("instance", map[string]string{
		"background": "#CCCC99"})
	buffer.CreateTag("global_font", map[string]string{
		"font": "Monospace Regular 10"})
	buffer.GetStartIter(&start)
	buffer.GetEndIter(&end)
	buffer.ApplyTagByName("global_font", &start, &end)
	swin.Add(textview)
	framebox.Add(swin)

	// Event
	window.Add(vbox)
	window.SetSizeRequest(1280, 974)
	window.ShowAll()
	gtk.Main()
}
