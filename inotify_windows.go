package main

// init_inotify initializes inotify
func init_inotify() {
}

// inotify_rm_watch removes a watch for a file/directory
func inotify_rm_watch(name string) {
}

// inotify_observe listens for events and sends them to a channel
func inotify_observe() {
}

// inotify_observe_collect collects the events from the channel and returns a map with event counts
func inotify_observe_collect(buf []byte) map[string]int {
	return make(map[string]int)
}

// inotify_dialog opens a dialog box to ask for user input on whether or not to continue watching a file/directory
func inotify_dialog(s map[string]int) bool {
	return false
}

// inotify_add_watch adds a watch for a file/directory
func inotify_add_watch(name string) {
}