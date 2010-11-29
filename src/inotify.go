package main

import (
  "syscall"
  "unsafe"
  "gdk"
)

var name_by_fd map[int32]string

var inotify_fd int
var buf_size int

func inotify_init() {
  name_by_fd = make(map[int32]string)
  var event syscall.InotifyEvent
  buf_size = unsafe.Sizeof(event)
  buf_size *= 128
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
  buf := make([]byte, buf_size)
  for ; ; {
    nread, _ := syscall.Read(inotify_fd, buf)
    //if nread != buf_size {
    //  println("tabby: Read from inotify fd failed, nread = ", nread, 
    //    " errno = ", err)
    //  continue
   // }
    println("nread = ", nread)
    event := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[0]))
    gdk.ThreadsEnter()
    bump_message("file " + name_by_fd[event.Wd] + " chaned, mask = ")
    gdk.ThreadsLeave()
  }
}