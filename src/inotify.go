package main

import (
  "syscall"
  "unsafe"
  "gdk"
  "gtk"
)

var name_by_fd map[int32]string

var inotify_fd int
var event_size int

const NEVENTS int = 1024

func inotify_init() {
  name_by_fd = make(map[int32]string)
  var event syscall.InotifyEvent
  event_size = unsafe.Sizeof(event)
  event_size *= 128
  inotify_fd, _ = syscall.InotifyInit()
  if -1 == inotify_fd {
    bump_message("InotifyInit failed, file changes outside of tabby " + 
      "will remain unnoticed")
    return
  }
  go inotify_observe()
}

func inotify_add_watch(name string) {
  wd, err := syscall.InotifyAddWatch(inotify_fd, name, syscall.IN_MODIFY)
  if -1 == wd {
    println("tabby: InotifyAddWatch failed, changes of file ", name, 
      " outside of tabby will remain unnoticed, errno = ", err)
    return
  }
  name_by_fd[int32(wd)] = name
}

func inotify_observe() {
  buf := make([]byte, event_size * NEVENTS)
  for ; ; {
    nread, _ := syscall.Read(inotify_fd, buf)
    collect := make(map[string]int)
    for offset := 0; offset < nread; offset += event_size {
      event := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
      collect[name_by_fd[event.Wd]] = 1
    }
    gdk.ThreadsEnter()
    reaload := inotify_dialog(collect)
    println(reaload)
    //bump_message("file " + name_by_fd[event.Wd] + " chaned, mask = ")
    gdk.ThreadsLeave()
  }
}

// Returns true in case of reloading files, and false in case of keeping as is.
func inotify_dialog(s map[string]int) bool {
	if nil == accel_group {
		accel_group = gtk.AccelGroup()
	}
  inotify_dlg := gtk.Dialog()
  defer inotify_dlg.Destroy()
  inotify_dlg.SetTitle("Some files have beed modified outside of tabby")
  inotify_dlg.AddButton("_Reload all", gtk.GTK_RESPONSE_ACCEPT)
  inotify_dlg.AddButton("_Keep all as is", gtk.GTK_RESPONSE_CANCEL)
  w := inotify_dlg.GetWidgetForResponse(gtk.GTK_RESPONSE_ACCEPT)
  inotify_dlg.AddAccelGroup(accel_group)
  w.AddAccelerator("clicked", accel_group, gdk.GDK_Return,
	  0, gtk.GTK_ACCEL_VISIBLE)
	inotify_dlg.SetSizeRequest(800, 350)
  inotify_store := gtk.TreeStore(gtk.TYPE_STRING)
  inotify_view := gtk.TreeView()
  inotify_view.AppendColumn(
    gtk.TreeViewColumnWithAttributes("text", gtk.CellRendererText(), "text", 0))
  inotify_view.ModifyFontEasy("Regular 8")
  inotify_model := inotify_store.ToTreeModel()
  inotify_view.SetModel(inotify_model)
  inotify_view.SetHeadersVisible(false)
  var iter gtk.GtkTreeIter
  for name, _ := range(s) {
	  inotify_store.Append(&iter, nil)
	  inotify_store.Set(&iter, name)
	}
	inotify_view.SetVisible(true)
  view_window := gtk.ScrolledWindow(nil, nil)
  view_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
  view_window.SetVisible(true)
  view_window.Add(inotify_view)
	vbox := inotify_dlg.GetVBox()
	vbox.Add(view_window)
	if gtk.GTK_RESPONSE_ACCEPT == inotify_dlg.Run() {
		return true
	}
	return false
}