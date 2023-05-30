package main

import (
    "github.com/mattn/go-gtk/gtk"
    "github.com/mattn/go-gtk/glib"
    "strings"
)

func tree_view_select_cb() {
    sel_file := tree_view_get_selected_path(tree_view, tree_model, 0, true)
    if sel_file == "" || name_is_dir(sel_file) {
        return
    }
    file_save_current()
    file_switch_to(sel_file)
}

func tree_view_get_selected_path(tree_view *gtk.TreeView, tree_model *gtk.TreeModel, col int, shift bool) string {
    path, column := tree_view.GetCursor()
    if path == nil {
        return ""
    }
    iter := tree_model.GetIter(path)
    var ans string
    for {
        val := new(glib.Value)
        tree_model.GetValue(iter, col, val)
        if shift {
            ans = val.GetString()[1:] + ans
        } else {
            ans = val.GetString() + ans
        }
        if !tree_model.IterParent(&iter, iter) {
            break
        }
    }
    return ans
}

func tree_view_set_cur_iter(mark bool) {
    if cur_file == "" {
        return
    }
    cur_file_suffix := cur_file
    tree_model.GetIterFirst(&cur_iter)
    for {
        gval := new(glib.Value)
        tree_model.GetValue(cur_iter, 0, gval)
        icon, node_path := gval.GetString()[0], gval.GetString()[1:]
        pos := slashed_prefix(cur_file_suffix, node_path)
        if pos > 0 {
            if mark {
                tree_store.Set(cur_iter, strings.ToUpper(string(icon))+node_path)
            } else {
                tree_store.Set(cur_iter, strings.ToLower(string(icon))+node_path)
            }
            if pos == len(cur_file_suffix) {
                break
            }
            parent := &cur_iter
            tree_model.IterChildren(&cur_iter, parent)
            cur_file_suffix = cur_file_suffix[pos:]
        } else {
            tree_model.IterNext(&cur_iter)
        }
    }
}