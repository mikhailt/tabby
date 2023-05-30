package main

import (
	"github.com/mattn/go-gtk/gtk"
)

type FileTreeNode struct {
	name    string
	parent  *FileTreeNode
	brother *FileTreeNode
	child   *FileTreeNode
	rec     *FileRecord
}

func NewFileTreeNode(rec *FileRecord) *FileTreeNode {
	return &FileTreeNode{rec: rec}
}

var file_tree_root FileTreeNode

func name_is_dir(name string) bool {
	return name[len(name)-1] == '/'
}

func slashed_prefix(a, b string) int {
	min := len(a)
	if len(b) < min {
		min = len(b)
	}
	for i := 0; i < min; i++ {
		if a[i] != b[i] {
			return i
		}
		if a[i] == '/' {
			return i + 1
		}
	}
	return min
}

func file_tree_find_among_children(root *FileTreeNode, name string) (*FileTreeNode, int, *FileTreeNode) {
	var pos int
	last_smaller_node := &FileTreeNode{}
	for cur_child, prev_child := root.child, (*FileTreeNode)(nil); cur_child != nil; prev_child, cur_child = cur_child, cur_child.brother {
		pos = slashed_prefix(cur_child.name, name)
		if pos > 0 {
			return cur_child, pos, prev_child
		}
		if cur_child.name < name {
			last_smaller_node = cur_child
		}
	}
	return last_smaller_node, 0, prev_child
}

func file_tree_insert(name string, rec *FileRecord) {
	file_tree_insert_rec(&file_tree_root, name, rec)
}

func file_tree_insert_rec(root *FileTreeNode, name string, rec *FileRecord) {
	cur_child, pos, prev_child := file_tree_find_among_children(root, name)
	if cur_child == nil {
		root.child = &FileTreeNode{
			name:   name,
			child:  nil,
			brother: root.child,
			parent: root,
			rec:    rec,
		}
		return
	}
	if pos == 0 {
		cur_child.brother = &FileTreeNode{
			name:   name,
			child:  nil,
			brother: cur_child.brother,
			parent: cur_child.parent,
			rec:    rec,
		}
		return
	}
	child_name_len := len(cur_child.name)
	if pos == child_name_len {
		file_tree_insert_rec(cur_child, name[child_name_len:], rec)
		return
	}
	replacement := &FileTreeNode{
		name:   name[:pos],
		parent: cur_child.parent,
		brother: cur_child.brother,
		child:  cur_child,
		rec:    nil,
	}
	cur_child.parent = replacement
	cur_child.name = cur_child.name[pos:]
	if prev_child != nil {
		prev_child.brother = replacement
	} else {
		root.child = replacement
	}
	file_tree_insert_rec(replacement, name[pos:], rec)
}

func file_tree_store_rec(root *FileTreeNode, iter *gtk.TreeIter, flag bool) {
	for cur_child := root.child; cur_child != nil; cur_child = cur_child.brother {
		if name_is_dir(cur_child.name) != flag {
			continue
		}
		var icon byte
		if flag {
			icon = 'd'
		} else if cur_child.rec.modified {
			icon = 'c'
		} else {
			icon = 'b'
		}
		child_iter := gtk.TreeIter{}
		tree_store.Append(&child_iter, iter)
		tree_store.Set(&child_iter, string(icon)+cur_child.name)
		file_tree_store_rec(cur_child, &child_iter, false)
		file_tree_store_rec(cur_child, &child_iter, true)
	}
}

func file_tree_store() {
	tree_store.Clear()
	file_tree_store_rec(&file_tree_root, nil, false)
	file_tree_store_rec(&file_tree_root, nil, true)
	tree_view.ExpandAll()
	tree_view_set_cur_iter(true)
}

func file_tree_remove(root *FileTreeNode, name string, merge_flag bool) {
	cur_child, pos, prev_child := file_tree_find_among_children(root, name)
	name_len := len(name)
	if pos < name_len {
		file_tree_remove(cur_child, name[pos:], true)
		return
	}
	if prev_child == nil && cur_child.brother == nil {
		file_tree_root.child = nil
		return
	}
	file_tree_remove_node(cur_child, prev_child)
	if !merge_flag {
		return
	}
	if prev_child != nil && cur_child.brother == nil && prev_child == cur_child.parent.child {
		file_tree_merge_parent_and_child(prev_child)
	} else if prev_child == nil && cur_child.brother != nil && cur_child.brother.brother == nil {
		file_tree_merge_parent_and_child(cur_child.brother)
	}
}

func file_tree_remove_node(cur *FileTreeNode, prev *FileTreeNode) {
	if prev != nil {
		prev.brother = cur.brother
	} else {
		cur.parent.child = cur.brother
	}
}

func file_tree_merge_parent_and_child(child *FileTreeNode) {
	parent := child.parent
	if parent == &file_tree_root {
		return
	}
	_, _, parent_prev := file_tree_find_among_children(parent.parent, parent.name)
	child.parent = parent.parent
	child.brother = parent.brother
	child.name = parent.name + child.name
	if parent_prev != nil {
		parent_prev.brother = child
	} else {
		parent.parent.child = child
	}
}