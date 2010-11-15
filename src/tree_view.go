package main

import (
	"gtk"
)

func tree_view_select_cb() {
	sel_file := tree_view_get_selected_path(tree_view, tree_model, 1)
	if name_is_dir(sel_file) {
		return
	}
	file_save_current()
	file_switch_to(sel_file)
}

func search_view_select_cb() {
	file := tree_view_get_selected_path(search_view, search_model, 0)
	file_save_current()
	file_switch_to(file)
	find_in_current_file(prev_pattern)
}

func tree_view_get_selected_path(tree_view *gtk.GtkTreeView, tree_model *gtk.GtkTreeModel, col int) string {
	var path *gtk.GtkTreePath
	var column *gtk.GtkTreeViewColumn
	tree_view.GetCursor(&path, &column)
	var iter gtk.GtkTreeIter
	tree_model.GetIterFromString(&iter, path.String())
	var ans string
	ans = ""
	for {
		var val gtk.GValue
		var next gtk.GtkTreeIter
		tree_model.GetValue(&iter, col, &val)
		ans = val.GetString() + ans
		if false == tree_model.IterParent(&next, &iter) {
			break
		}
		iter.Assign(&next)
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
