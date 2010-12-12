package main

import (
	"syscall"
	"unsafe"
	"gdk"
	"gtk"
)

var name_by_wd map[int32]string
var wd_by_name map[string]int32

var inotify_fd int
var event_size int

var epoll_fd int

const NEVENTS int = 1024

func inotify_init() {
	name_by_wd = make(map[int32]string)
	wd_by_name = make(map[string]int32)
	var event syscall.InotifyEvent
	event_size = unsafe.Sizeof(event)
	inotify_fd, _ = syscall.InotifyInit()
	if -1 == inotify_fd {
		bump_message("InotifyInit failed, file changes outside of tabby " +
			"will remain unnoticed")
		return
	}
	epoll_fd, _ = syscall.EpollCreate(1)
	if -1 == epoll_fd {
		println("tabby: inotify_init: EpollCreate failed")
	}
	var epoll_event syscall.EpollEvent
	epoll_event.Events = syscall.EPOLLIN
	syscall.EpollCtl(epoll_fd, syscall.EPOLL_CTL_ADD, inotify_fd, &epoll_event)
	go inotify_observe()
}

func inotify_add_watch(name string) {
	wd, err := syscall.InotifyAddWatch(inotify_fd, name,
		syscall.IN_MODIFY|syscall.IN_DELETE_SELF|syscall.IN_MOVE_SELF)
	if -1 == wd {
		if err == syscall.ENOENT {
			// Dirty hack.
			return
		}
		println("tabby: InotifyAddWatch failed, changes of file ", name,
			" outside of tabby will remain unnoticed, errno = ", err)
		return
	}
	name_by_wd[int32(wd)] = name
	wd_by_name[name] = int32(wd)
}

func inotify_rm_watch(name string) {
	wd, found := wd_by_name[name]
	if false == found {
		return
	}
	retval, _ /*err*/ := syscall.InotifyRmWatch(inotify_fd, uint32(wd))
	if -1 == retval {
		//println("tabby: InotifyRmWatch failed, errno = ", err)
		return
	}
	name_by_wd[wd] = "", false
	wd_by_name[name] = 0, false
}

func inotify_observe() {
	buf := make([]byte, event_size*NEVENTS)
	for {
		collect := inotify_observe_collect(buf)
		if 0 == len(collect) {
			continue
		}
		gdk.ThreadsEnter()
		file_save_current()
		reload := inotify_dialog(collect)
		for name, _ := range collect {
			rec, rec_found := file_map[name]
			if false == rec_found {
				println("tabby: inotify_observe: ", name, " not found in file_map")
				continue
			}
			if reload {
				// Reload file content.
				read_ok, buf := open_file_read_to_buf(name, true)
				if read_ok {
					rec.buf = buf
					rec.modified = false
					inotify_rm_watch(name)
					inotify_add_watch(name)
				} else {
					rec.modified = true
				}
			} else {
				// Keep file as is.
				rec.modified = true
			}
		}
		file_tree_store()
		// So as to renew current TextBuffer it is required to switch to cur_file.
		file_switch_to(cur_file)
		gdk.ThreadsLeave()
	}
}

func inotify_observe_collect(buf []byte) map[string]int {
	epoll_buf := make([]syscall.EpollEvent, 1)
	collect := make(map[string]int)
	for {
		nread, _ := syscall.Read(inotify_fd, buf)
		for offset := 0; offset < nread; offset += event_size {
			event := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			if syscall.IN_IGNORED == event.Mask {
				continue
			}
			collect[name_by_wd[event.Wd]] = 1
		}
		nevents, err := syscall.EpollWait(epoll_fd, epoll_buf, 500)
		if 0 >= nevents {
			if -1 == nevents {
				println("tabby: inotify_observe_collect: EpollWait failed, errno = ",
					err)
			}
			break
		}
	}
	return collect
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
	for name, _ := range s {
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
