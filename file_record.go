package main

type FileRecord struct {
	buf      []byte
	error    []byte
	modified bool
	sel_be   int
	sel_en   int
}

var file_map = make(map[string]*FileRecord)

func file_opened(name string) bool {
	_, found := file_map[name]
	return found
}

func delete_file_record(name string) {
	if _, found := file_map[name]; !found {
		return
	}
	inotify_rm_watch(name)
	file_tree_remove(&file_tree_root, name, true)
	delete(file_map, name)
}

func add_file_record(name string, buf []byte, bump_flag bool) bool {
	if _, found := file_map[name]; found {
		if bump_flag {
			bump_message("File " + name + " is already open")
		}
		return false
	}
	rec := &FileRecord{buf: buf, modified: false}
	file_map[name] = rec
	file_tree_insert(name, rec)
	if file_is_saved(name) {
		inotify_add_watch(name)
	}
	return true
}