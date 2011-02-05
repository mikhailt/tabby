package main

import (
	"gtk"
	"file_tree"
	"strconv"
)

type SearchView struct {
	cursor, size int
	view *gtk.GtkTreeView
	store *gtk.GtkTreeStore
	model *gtk.GtkTreeModel
	window *gtk.GtkScrolledWindow
}

func (v *SearchView) Init() {
	v.store = gtk.TreeStore(gtk.TYPE_STRING)
	v.view = file_tree.NewSearchTree()
	v.view.ModifyFontEasy("Regular 8")
	v.model = v.store.ToTreeModel()
	v.view.SetModel(v.model)
	v.view.AppendColumn(gtk.TreeViewColumnWithAttributes("", 
		gtk.CellRendererText(), "text", 0))
	v.view.SetHeadersVisible(false)
	v.view.Connect("cursor-changed", func() {v.Select()}, nil)
	v.window = gtk.ScrolledWindow(nil, nil)
	v.window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	v.window.Add(v.view)
	v.window.SetVisible(opt.show_search)
}

func (v *SearchView) Select() {
	file := tree_view_get_selected_path(v.view, v.model, 0, false)
	file_save_current()
	file_switch_to(file)
	tree_view_scroll_to_cur_iter()
	if "" != prev_pattern {
		find_in_current_file(prev_pattern, false)
	}
}

func (v *SearchView) SetCursor(pos int) {
	v.cursor = pos
	ppath := gtk.TreePathFromString(strconv.Itoa(pos))
	v.view.SetCursor(ppath, nil, false)
}

func (v *SearchView) AddFile(file string) {
	var iter gtk.GtkTreeIter
	v.store.Append(&iter, nil)
	v.store.Set(&iter, file)
	v.size++
}

func (v *SearchView) NextResult() {
	if 0 == v.size {
		return
	}
	v.cursor++
	if v.cursor == v.size {
		v.cursor = 0
	}
	v.SetCursor(v.cursor)
}

func (v *SearchView) PrevResult() {
	if 0 == v.size {
		return
	}
	v.cursor--
	if v.cursor < 0 {
		v.cursor = v.size - 1
	}
	v.SetCursor(v.cursor)
}

func (v *SearchView) PrepareToSearch() {
	v.size = 0
	v.cursor = -1
}