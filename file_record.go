package main

type FileRecord struct {
	buf      []byte
	error    []byte
	modified bool
	sel_be   int
	sel_en   int
}

var file_map map[string]*FileRecord

// file_opened checks if a file is already open
func file_opened(name string) bool {
	_, found := file_map[name]
	return found
}

// delete_file_record deletes a file record for a given file name
func delete_file_record(name string) {
	_, found := file_map[name]
	if false == found {
		return
	}
	inotify_rm_watch(name)
	file_tree_remove(&file_tree_root, name, true)
	delete(file_map, name)
}

// add_file_record adds a file record for a given file name and its contents
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