// Package main is the entry point of the application.
package main

type FileRecord struct {
	buf      []byte
	error    []byte
	modified bool
	sel_be   int
	sel_en   int
}

var file_map map[string]*FileRecord

// file_opened returns true if a file with the given name is already open.
func file_opened(name string) bool {
	_, found := file_map[name]
	return found
}

// delete_file_record removes the record of the file with the given name from the file_map.
// If the file tree node is present, it removes the node from the tree.
func delete_file_record(name string) {
	_, found := file_map[name]
	if false == found {
		return
	}
	inotify_rm_watch(name)
	file_tree_remove(&file_tree_root, name, true)
	delete(file_map, name)
}

// add_file_record adds a new record to the file_map with the given name and buffer.
// If bump_flag is true and the file is already open, it displays a message and returns false.
// It returns true if the file is successfully added to the file_map.
func add_file_record(name string, buf []byte, bump_flag bool) bool {
	_, found := file_map[name]
	if found {
		if bump_flag {
			bump_message("File " + name + " is already open")
		}
		return false
	}
	rec := new(FileRecord)
	file_map[name] = rec
	rec.modified = false
	rec.buf = buf
	file_tree_insert(name, rec)
	if file_is_saved(name) {
		inotify_add_watch(name)
	}
	return true
}