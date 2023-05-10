package main

// Initializes the inotify functionality
func init_inotify() {
}

// Removes the watch for a given name
func inotify_rm_watch(name string) {
}

// Starts observing for file changes
func inotify_observe() {
}

// Collects observed data into a map and returns it
func inotify_observe_collect(buf []byte) map[string]int {
	return make(map[string]int)
}

// Displays a dialog box based on observed data and returns the user response
func inotify_dialog(s map[string]int) bool {
	return false
}

// Adds a watch for a given name
func inotify_add_watch(name string) {
}