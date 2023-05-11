package main

// initialize inotify
func init_inotify() {
}

// remove watch on file
func inotify_rm_watch(name string) {
}

// observe changes on files
func inotify_observe() {
}

// collect and return changes as map
func inotify_observe_collect(buf []byte) map[string]int {
	return make(map[string]int)
}

// display a dialog
func inotify_dialog(s map[string]int) bool {
	return false
}

// add watch on file
func inotify_add_watch(name string) {
}