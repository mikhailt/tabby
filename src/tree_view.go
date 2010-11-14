package main

import (
	"gtk"
)

func tree_view_select_cb() {
	var path *gtk.GtkTreePath
	var column *gtk.GtkTreeViewColumn
	tree_view.GetCursor(&path, &column)
	var iter gtk.GtkTreeIter
	tree_model.GetIterFromString(&iter, path.String())
	sel_file := tree_view_path(&iter)
	if name_is_dir(sel_file) {
		return
	}
	file_save_current()
	file_switch_to(sel_file)
}

func tree_view_path(iter *gtk.GtkTreeIter) string {
	var ans string
	ans = ""
	for {
		var val gtk.GValue
		var next gtk.GtkTreeIter
		tree_model.GetValue(iter, 1, &val)
		ans = val.GetString() + ans
		if false == tree_model.IterParent(&next, iter) {
			break
		}
		iter = &next
	}
	return ans
}

// Sets cur_iter pointing to tree_store node corresponding to current file.
// Requires properly set cur_file.
func tree_view_set_cur_iter() {
	if "" == cur_file {
		return
	}
	var parent gtk.GtkTreeIter
	name := cur_file
	tree_model.GetIterFirst(&cur_iter)
	for {
		var val gtk.GValue
		tree_model.GetValue(&cur_iter, 1, &val)
		cur_str := val.GetString()
		pos := slashed_prefix(name, cur_str)
		if pos == len(name) {
			break
		} else if pos > 0 {
			parent.Assign(&cur_iter)
			tree_model.IterChildren(&cur_iter, &parent)
			name = name[pos:]
		} else {
			tree_model.IterNext(&cur_iter)
		}
	}
}
