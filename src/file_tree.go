package main

import (
	"gtk"
)

type FileTreeNode struct {
	name    string
	parent  *FileTreeNode
	brother *FileTreeNode
	child   *FileTreeNode
	rec     *FileRecord
}

func NewFileTreeNode(rec *FileRecord) *FileTreeNode {
	res := new(FileTreeNode)
	res.parent = nil
	res.brother = nil
	res.child = nil
	res.rec = rec
	return res
}

var file_tree_root FileTreeNode

func name_is_dir(name string) bool {
	return ('/' == name[len(name)-1])
}

func slashed_prefix(a string, b string) int {
	la := len(a)
	lb := len(b)
	var bar int
	if la < lb {
		bar = la
	} else {
		bar = lb
	}
	last_slash := 0
	for y := 0; y < bar; y++ {
		if a[y] != b[y] {
			return last_slash
		}
		if '/' == a[y] {
			last_slash = y + 1
		}
	}
	if la == lb {
		return bar
	}
	return last_slash
}

// In case pos == 0 node means last smaller than this node;
// pos > 0 means that found node with common slashed prefix with name.
func file_tree_find_among_children(root *FileTreeNode, name string) (node *FileTreeNode, position int, prev *FileTreeNode) {
	var pos int
	var cur_child, prev_child *FileTreeNode
	var last_smaller_node *FileTreeNode
	last_smaller_node = nil
	if nil == root.child {
		return nil, 0, nil
	}
	prev_child = nil
	for cur_child = root.child; nil != cur_child; cur_child = cur_child.brother {
		pos = slashed_prefix(cur_child.name, name)
		if pos > 0 {
			return cur_child, pos, prev_child
		}
		if cur_child.name < name {
			last_smaller_node = cur_child
		}
		prev_child = cur_child
	}
	return last_smaller_node, 0, prev_child
}

func file_tree_insert(name string, rec *FileRecord) {
	file_tree_insert_rec(&file_tree_root, name, rec)
}

func file_tree_insert_rec(root *FileTreeNode, name string, rec *FileRecord) {
	cur_child, pos, prev_child := file_tree_find_among_children(root, name)
	if nil == cur_child {
		// Inserting name in the beginning of the list of children of root.
		saved_child := root.child
		root.child = NewFileTreeNode(rec)
		root.child.name = name
		root.child.brother = saved_child
		root.child.parent = root
		return
	}
	if 0 == pos {
		// There is no child with common slashed prefix with name.
		saved_brother := cur_child.brother
		cur_child.brother = NewFileTreeNode(rec)
		cur_child.brother.brother = saved_brother
		cur_child.brother.parent = cur_child.parent
		cur_child.brother.name = name
		return
	}
	child_name_len := len(cur_child.name)
	if pos == child_name_len {
		// cur_child is the directory containing current name.
		file_tree_insert_rec(cur_child, name[child_name_len:], rec)
		return
	}
	// cur_child is directory or file with common prefix with name.
	replacement := NewFileTreeNode(nil)
	replacement.parent = cur_child.parent
	replacement.brother = cur_child.brother
	if nil != prev_child {
		prev_child.brother = replacement
	} else {
		root.child = replacement
	}
	replacement.name = name[:pos]
	replacement.child = cur_child
	cur_child.parent = replacement
	cur_child.name = cur_child.name[pos:]
	cur_child.brother = nil
	file_tree_insert_rec(replacement, name[pos:], rec)
}

// Dumps root subtree to tree_store at iter. Flag is false for dumping files and
// true for directories.
func file_tree_store_rec(root *FileTreeNode, iter *gtk.GtkTreeIter, flag bool) {
	var child_iter gtk.GtkTreeIter
	var gtk_icon string
	for cur_child := root.child; nil != cur_child; cur_child = cur_child.brother {
		if flag != name_is_dir(cur_child.name) {
			continue
		}
		tree_store.Append(&child_iter, iter)
		if '/' == cur_child.name[len(cur_child.name)-1] {
			gtk_icon = gtk.GTK_STOCK_DIRECTORY
		} else {
			if cur_child.rec.modified {
				gtk_icon = gtk.GTK_STOCK_DELETE
			} else {
				gtk_icon = gtk.GTK_STOCK_FILE
			}
		}
		tree_store.Set(&child_iter,
			gtk.Image().RenderIcon(gtk_icon, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf,
			cur_child.name)
		file_tree_store_rec(cur_child, &child_iter, false)
		file_tree_store_rec(cur_child, &child_iter, true)
	}
}

func file_tree_store() {
	tree_store.Clear()
	file_tree_store_rec(&file_tree_root, nil, false)
	file_tree_store_rec(&file_tree_root, nil, true)
	tree_view.ExpandAll()
}

func file_tree_remove(root *FileTreeNode, name string, merge_flag bool) {
	cur_child, pos, prev_child := file_tree_find_among_children(root, name)
	name_len := len(name)
	if pos < name_len {
		// name lies inside cur_child directory.
		file_tree_remove(cur_child, name[pos:], true)
		return
	} else {
		// name lies inside root directory.
		if (nil == prev_child) && (nil == cur_child.brother) {
			// Only one child in current dir. It means that root is file_tree_root.
			file_tree_root.child = nil
			return
		}
		file_tree_remove_node(cur_child, prev_child)
		if false == merge_flag {
			return
		}
		if nil != prev_child {
			if (nil == cur_child.brother) && (prev_child == cur_child.parent.child) {
				// There are only two children of root: prev & cur
				file_tree_merge_parent_and_child(prev_child)
			}
		} else {
			// prev_child == nil
			if nil == cur_child.brother.brother {
				// Only two children of root: cur and his brother
				file_tree_merge_parent_and_child(cur_child.brother)
			}
		}
		
		return
	}
	bump_message("file_tree_remove: unexpected case: name = " + name)
}

func file_tree_remove_node(cur *FileTreeNode, prev *FileTreeNode) {
	if nil != prev {
		prev.brother = cur.brother
	} else {
		cur.parent.child = cur.brother
	}
}

func file_tree_merge_parent_and_child(child *FileTreeNode) {
	parent := child.parent
	if parent != &file_tree_root {
		grand_parent := parent.parent
		new_name := parent.name + child.name
		if name_is_dir(new_name) {
			grand_parent.child = child
			child.parent = grand_parent
			child.name = new_name
		} else {
			file_tree_remove(grand_parent, parent.name, false)
			file_tree_insert_rec(grand_parent, new_name, child.rec)
		}
	}
}
