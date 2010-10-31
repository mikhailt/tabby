package main

import (
	//"os"
	"gtk"
	//"gdkpixbuf"
	//"path"
	"fmt"
	//"reflect"
	"strings"
	"sync"
	//"unsafe"
  "time"
  "runtime"
)

var ncpu int

func min(a, b int) int {
  if (a < b) {
    return a
  }
  return b
}

func search(ch chan int, mu *sync.Mutex, buf *gtk.GtkTextBuffer, text string,
            shift int, selection string, sel_len int) {
  var be, en gtk.GtkTextIter

  for be_ind := 0; ; {
    be_ind = strings.Index(text, selection)
    if -1 == be_ind {
      ch <- 0
      break
    }
    shift += be_ind
    text = text[be_ind + sel_len :]

    mu.Lock()
    buf.GetIterAtOffset(&be, shift)
    buf.GetIterAtOffset(&en, shift+sel_len)
    buf.ApplyTagByName("instance", &be, &en)
    mu.Unlock()

    shift += sel_len
  }
}

var prev_selection string

func highlight_instances(buf *gtk.GtkTextBuffer) {
	var start, end gtk.GtkTextIter

	buf.GetStartIter(&start)
	buf.GetEndIter(&end)

	if buf.GetHasSelection() {
		var be, en gtk.GtkTextIter
    buf.GetSelectionBounds(&be, &en)
		selection := buf.GetSlice(&be, &en, false)
    if prev_selection != selection {
      sel_len := len(selection)
      if (20 < sel_len) || (2 > sel_len) {
				return
			}

      time_start := time.Nanoseconds()

    	buf.RemoveTagByName("instance", &start, &end)
    	prev_selection = selection
			fmt.Printf("selection = %s\n", selection)

			text := buf.GetSlice(&start, &end, false)
			text_len := len(text)
			if (text_len < 10000000) {
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
			} else {
        ch := make(chan int)
			  var mu sync.Mutex
			  cores_left := ncpu - 2
			  q := text_len / (ncpu - cores_left)
        for y := 0; y < (ncpu - cores_left); y++ {
          go search(ch, &mu, buf, text[y * q : min((y + 1)*q + 30, text_len)], y * q, selection, sel_len)
        }
        for cnt := (ncpu - cores_left); cnt > 0; {
          _ = <- ch
          cnt--
          fmt.Printf("cnt--\n");
        }
			}

			time_end := time.Nanoseconds()
			total_time := time_end - time_start
			fmt.Printf("elapsed time = %d.%d ms\n", total_time / 1000000, total_time % 1000000)

    }
	} else {
		if prev_selection != "" {
			prev_selection = ""
      buf.RemoveTagByName("instance", &start, &end)
    }
  }
}

func main() {
  ncpu = runtime.GOMAXPROCS(0)
	gtk.Init(nil)
	window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	window.SetTitle("tabby")
	window.Connect("destroy", func(w *gtk.GtkWidget, user_data string) {
		gtk.MainQuit()
	},
		"")

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
	textview.ModifyFontEasy("Monospace Regular 10")
	var start, end gtk.GtkTextIter
	buffer := textview.GetBuffer()

	buffer.Connect("mark-set", func() {
		highlight_instances(buffer)
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

	window.Add(vbox)
	window.SetSizeRequest(1280, 974)
	window.ShowAll()
	gtk.Main()
}
