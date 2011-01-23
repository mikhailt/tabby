package main

import (
	"gtk"
	"glib"
	"strings"
)

func tree_view_select_cb() {
	sel_file := tree_view_get_selected_path(tree_view, tree_model, 0, true)
	if "" == sel_file {
		return
	}
	if name_is_dir(sel_file) {
		return
	}
	file_save_current()
	file_switch_to(sel_file)
}

func search_view_select_cb() {
	file := tree_view_get_selected_path(search_view, search_model, 0, false)
	file_save_current()
	file_switch_to(file)
	tree_view_scroll_to_cur_iter()
	if "" != prev_pattern {
		find_in_current_file(prev_pattern, false)
	}
}

func tree_view_get_selected_path(tree_view *gtk.GtkTreeView, tree_model *gtk.GtkTreeModel, col int, shift bool) string {
	var path *gtk.GtkTreePath
	var column *gtk.GtkTreeViewColumn
	tree_view.GetCursor(&path, &column)
	if nil == path.TreePath {
		return ""
	}
	var iter gtk.GtkTreeIter
	tree_model.GetIterFromString(&iter, path.String())
	var ans string
	ans = ""
	for {
		var val glib.GValue
		var next gtk.GtkTreeIter
		tree_model.GetValue(&iter, col, &val)
		if shift {
			ans = val.GetString()[1:] + ans
		} else {
			ans = val.GetString() + ans
		}
		if false == tree_model.IterParent(&next, &iter) {
			break
		}
		iter.Assign(&next)
	}
	return ans
}

// Sets cur_iter pointing to tree_store node corresponding to current file.
// Requires properly set cur_file.
func tree_view_set_cur_iter(mark bool) {
	if "" == cur_file {
		return
	}
	var parent gtk.GtkTreeIter
	name := cur_file
	tree_model.GetIterFirst(&cur_iter)
	for {
		var val glib.GValue
		tree_model.GetValue(&cur_iter, 0, &val)
		whole_string := val.GetString()
		cur_str := whole_string[1:]
		pos := slashed_prefix(name, cur_str)
		if pos > 0 {
			if mark {
				tree_store.Set(&cur_iter, strings.ToUpper(whole_string[:1])+cur_str)
			} else {
				tree_store.Set(&cur_iter, strings.ToLower(whole_string[:1])+cur_str)
			}
			if pos == len(name) {
				break
			}
			parent.Assign(&cur_iter)
			tree_model.IterChildren(&cur_iter, &parent)
			name = name[pos:]
		} else {
			tree_model.IterNext(&cur_iter)
		}
	}
}
