package main

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mikhailt/tabby/file_tree"
	"strconv"
)

type SearchView struct {
	cursor, size int
	view *gtk.TreeView
	store *gtk.TreeStore
	model *gtk.TreeModel
	window *gtk.ScrolledWindow
}

// Init initializes the SearchView
func (v *SearchView) Init() {
	v.store = gtk.NewTreeStore(gtk.TYPE_STRING)
	v.view = file_tree.NewSearchTree()
	v.view.ModifyFontEasy("Regular 8")
	v.model = v.store.ToTreeModel()
	v.view.SetModel(v.model)
	v.view.AppendColumn(gtk.NewTreeViewColumnWithAttributes("", 
		gtk.NewCellRendererText(), "text", 0))
	v.view.SetHeadersVisible(false)
	v.view.Connect("cursor-changed", func() {v.Select()}, nil)
	v.window = gtk.NewScrolledWindow(nil, nil)
	v.window.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	v.window.Add(v.view)
	v.window.SetVisible(opt.show_search)
}

// Select handles the selection of a file in the search view
func (v *SearchView) Select() {
	file := tree_view_get_selected_path(v.view, v.model, 0, false)
	if !file_opened(file) {
		return
	}
	file_save_current()
	file_switch_to(file)
	tree_view_scroll_to_cur_iter()
	if "" != prev_pattern {
		find_in_current_file(prev_pattern, false)
	}
}

// SetCursor sets the cursor for the search view
func (v *SearchView) SetCursor(pos int) {
	v.cursor = pos
	ppath := gtk.NewTreePathFromString(strconv.Itoa(pos))
	v.view.SetCursor(ppath, nil, false)
}

// AddFile adds a file to the search view
func (v *SearchView) AddFile(file string) {
	var iter gtk.TreeIter
	v.store.Append(&iter, nil)
	v.store.Set(&iter, file)
	v.size++
}

// NextResult selects the next result in the search view
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

// PrevResult selects the previous result in the search view
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

// PrepareToSearch prepares the search view for a new search
func (v *SearchView) PrepareToSearch() {
	v.size = 0
	v.cursor = -1
}