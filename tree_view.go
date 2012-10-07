package main

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/glib"
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
// Requires properly set cur_file. As a side effect it also assigns correct 
// capitalization for first letters of strings kept in nodes according to @mark
// which denotes if current file is active or not.
func tree_view_set_cur_iter(mark bool) {
	if "" == cur_file {
		return
	}
	var parent gtk.GtkTreeIter
	cur_file_suffix := cur_file
	tree_model.GetIterFirst(&cur_iter)
	for {
		var gval glib.GValue
		tree_model.GetValue(&cur_iter, 0, &gval)
		gval_string := gval.GetString()
		icon := gval_string[0]
		node_path := gval_string[1:]
		if pos := slashed_prefix(cur_file_suffix, node_path); pos > 0 {
			if mark {
				tree_store.Set(&cur_iter, strings.ToUpper(string(icon)) + node_path)
			} else {
				tree_store.Set(&cur_iter, strings.ToLower(string(icon)) + node_path)
			}
			if pos == len(cur_file_suffix) {
				break
			}
			parent.Assign(&cur_iter)
			tree_model.IterChildren(&cur_iter, &parent)
			cur_file_suffix = cur_file_suffix[pos:]
		} else {
			tree_model.IterNext(&cur_iter)
		}
	}
}
