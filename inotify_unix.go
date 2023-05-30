package main

import (
	"syscall"
	"unsafe"
	"time"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
)

var (
	nameByWD   = make(map[int32]string)
	wdByName   = make(map[string]int32)
	inotifyFD  int
	eventSize  int
	epollFD    int
	nergens    = 1024
	accelGroup *gtk.AccelGroup
)

func initInotify() {
	var err error

	eventSize = int(unsafe.Sizeof(syscall.InotifyEvent{}))
	inotifyFD, err = syscall.InotifyInit()
	if err != nil {
		println("InotifyInit failed. File changes outside of tabby will remain unnoticed.")
		return
	}

	epollFD, err = syscall.EpollCreate(1)
	if err != nil {
		tabbyLog("initInotify: " + err.Error())
	}

	var epollev syscall.EpollEvent
	epollev.Events = syscall.EPOLLIN
	syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, inotifyFD, &epollev)

	go inotifyObserve()
}

func inotifyAddWatch(name string) {
	wd, err := syscall.InotifyAddWatch(inotifyFD, name, syscall.IN_MODIFY|syscall.IN_DELETE_SELF|syscall.IN_MOVE_SELF)
	if err != nil {
		if err == syscall.ENOENT {
			// Dirty hack.
			return
		}
		println("InotifyAddWatch failed. Changes of file " + name + " outside of tabby will remain unnoticed. " +
			"Errno: " + err.Error())
		return
	}

	nameByWD[int32(wd)] = name
	wdByName[name] = int32(wd)
}

func inotifyRmWatch(name string) {
	wd, found := wdByName[name]
	if !found {
		return
	}

	retval, _ := syscall.InotifyRmWatch(inotifyFD, uint32(wd))
	if retval == -1 {
		println("InotifyRmWatch failed, errno = ", err)
	}

	delete(nameByWD, wd)
	delete(wdByName, name)
}

func inotifyObserve() {
	buf := make([]byte, eventSize*nergens)
	for {
		collect := inotifyObserveCollect(buf)
		if len(collect) == 0 {
			continue
		}

		gdk.ThreadsEnter()
		fileSaveCurrent()

		reload := inotifyDialog(collect)

		for name := range collect {
			rec, recFound := fileMap[name]
			if !recFound {
				tabbyLog("InotifyObserve: " + name + " not found in fileMap.")
				continue
			}

			if reload {
				if readOK, buf := openFileReadToBuf(name, true); readOK {
					rec.buf = buf
					rec.modified = false
					inotifyRmWatch(name)
					inotifyAddWatch(name)
				} else {
					rec.modified = true
				}
			} else {
				rec.modified = true
			}
		}

		fileTreeStore()
		fileSwitchTo(curFile)
		gdk.ThreadsLeave()
	}
}

func inotifyObserveCollect(buf []byte) map[string]int {
	epollBuf := make([]syscall.EpollEvent, 1)
	collect := make(map[string]int)
	for {
		nread, _ := syscall.Read(inotifyFD, buf)
		for offset := 0; offset < nread; offset += eventSize {
			event := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			if event.Mask == syscall.IN_IGNORED {
				continue
			}

			collect[nameByWD[event.Wd]] = 1
		}

		nevents, err := syscall.EpollWait(epollFD, epollBuf, 500*time.Millisecond)
		if nevents <= 0 {
			if err != nil {
				tabbyLog("InotifyObserveCollect: " + err.Error())
			}

			break
		}
	}

	return collect
}

func inotifyDialog(s map[string]int) bool {
	if accelGroup == nil {
		accelGroup = gtk.NewAccelGroup()
	}

	inotifyDlg := gtk.NewDialog()
	defer inotifyDlg.Destroy()

	inotifyDlg.SetTitle("Some files have been modified outside of tabby")
	inotifyDlg.AddButton("_Reload all", gtk.RESPONSE_ACCEPT)
	inotifyDlg.AddButton("_Keep all as is", gtk.RESPONSE_CANCEL)

	w := inotifyDlg.GetWidgetForResponse(int(gtk.RESPONSE_ACCEPT))
	inotifyDlg.AddAccelGroup(accelGroup)
	w.AddAccelerator("clicked", accelGroup, gdk.KEY_Return, 0, gtk.ACCEL_VISIBLE)

	inotifyDlg.SetSizeRequest(800, 350)

	inotifyStore := gtk.NewTreeStore(gtk.TYPE_STRING)
	inotifyView := gtk.NewTreeView()
	inotifyView.AppendColumn(gtk.NewTreeViewColumnWithAttributes("text", gtk.NewCellRendererText(), "text", 0))
	inotifyView.ModifyFontEasy("Regular 8")

	inotifyModel := inotifyStore.ToTreeModel()
	inotifyView.SetModel(inotifyModel)
	inotifyView.SetHeadersVisible(false)

	var iter gtk.TreeIter
	for name := range s {
		inotifyStore.Append(&iter, nil)
		inotifyStore.Set(&iter, name)
	}

	inotifyView.SetVisible(true)

	viewWindow := gtk.NewScrolledWindow(nil, nil)
	viewWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	viewWindow.SetVisible(true)
	viewWindow.Add(inotifyView)

	vbox := inotifyDlg.GetVBox()
	vbox.Add(viewWindow)

	if inotifyDlg.Run() == int(gtk.RESPONSE_ACCEPT) {
		return true
	}

	return false
}