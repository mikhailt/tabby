package main

// Initializes inotify
func init_inotify() {
}

// Remove watch with given name from inotify
func inotify_rm_watch(name string) {
}

// Observe inotify events
func inotify_observe() {
}

// Collect inotify buffer events in a map
func inotify_observe_collect(buf []byte) map[string]int {
	return make(map[string]int)
}

// Display dialog for inotify events
func inotify_dialog(s map[string]int) bool {
	return false
}

// Add watch with given name to inotify
func inotify_add_watch(name string) {
}