package main

import (
	"os"
	"gtk"
	"gdk"
	"gdkpixbuf"
	"glib"
	"runtime"
)

type FileRecord struct {
  buf []byte
  modified bool
}

type FileTreeNode struct {
  name string
  parent *FileTreeNode
  brother *FileTreeNode
  child *FileTreeNode
}

func NewFileTreeNode() *FileTreeNode {
  res := new(FileTreeNode)
  res.parent = nil
  res.brother = nil
  res.child = nil
  return res
}

var file_tree_root FileTreeNode

var file_map map[string]*FileRecord


var main_window *gtk.GtkWindow
var source_buf *gtk.GtkSourceBuffer
var tree_view *gtk.GtkTreeView
var tree_store *gtk.GtkTreeStore
var tree_model *gtk.GtkTreeModel
var source_view *gtk.GtkSourceView
var selection_flag bool
var prev_selection string
var prev_dir string
var cur_file string

func file_save_current() {
  if ("" == cur_file) {
    return
  }
  var be, en gtk.GtkTextIter
  source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	text_to_save := source_buf.GetText(&be, &en, true)
	rec := file_map[cur_file]
	rec.buf = ([]byte)(text_to_save[:])
	rec.modified = source_buf.GetModified()
  runtime.GC()
}

func file_switch_to(name string) {
  rec := file_map[name]
  source_buf.BeginNotUndoableAction()
	source_buf.SetText(string(rec.buf))
	source_buf.SetModified(rec.modified)
	source_buf.EndNotUndoableAction()
	cur_file = name
}

func is_dir_name(name string) bool {
  return ('/' == name[len(name) - 1])
}

// In case pos == 0 node means last smaller than this node;
// pos > 0 means that found node with common slashed prefix with name.
func file_tree_find_among_children(root *FileTreeNode, 
    name string) (node *FileTreeNode, position int, prev *FileTreeNode) {
  var pos int
  var cur_child, prev_child *FileTreeNode
  var last_smaller_node *FileTreeNode
  last_smaller_node = nil
  if (nil == root.child) {
    return nil, 0, nil
  }
  prev_child = nil
  for cur_child = root.child; nil != cur_child; cur_child = cur_child.brother {
    pos = slashed_prefix(cur_child.name, name)
    if (pos > 0) {
      return cur_child, pos, prev_child
    }
    if (cur_child.name < name) {
      last_smaller_node = cur_child
    }
    prev_child = cur_child
  }
  return last_smaller_node, 0, nil
}

func file_tree_insert(root *FileTreeNode, name string) {
  cur_child, pos, prev_child := file_tree_find_among_children(root, name)
  if (nil == cur_child) {
    // Inserting name in the beginning of the list of children of root.
    saved_child := root.child
    root.child = NewFileTreeNode()
    root.child.name = name
    root.child.brother = saved_child
    root.child.parent = root
    return
  }
  if (0 == pos) {
    // There is no child with common slashed prefix with name.
    saved_brother := cur_child.brother
    cur_child.brother = NewFileTreeNode()
    cur_child.brother.brother = saved_brother
    cur_child.brother.parent = cur_child.parent
    cur_child.brother.name = name
    return
  }
  child_name_len := len(cur_child.name)
  if (pos == child_name_len) {
    // cur_child is the directory containing current name.
    file_tree_insert(cur_child, name[child_name_len:])
    return
  }
  // cur_child is directory or file with common prefix with name.
  replacement := NewFileTreeNode()
  replacement.parent = cur_child.parent
  replacement.brother = cur_child.brother
  if (nil != prev_child) {
    prev_child.brother = replacement
  } else {
    root.child = replacement
  }
  replacement.name = name[:pos]
  replacement.child = cur_child
  cur_child.parent = replacement
  cur_child.name = cur_child.name[pos:]
  cur_child.brother = nil
  file_tree_insert(replacement, name[pos:])
  
}

// Dumps root subtree to tree_store at iter. Flag is false for dumping files and
// true for directories.
func file_tree_store_rec(root *FileTreeNode, iter *gtk.GtkTreeIter, flag bool) {
  var child_iter gtk.GtkTreeIter
  var gtk_icon string
  for cur_child := root.child; nil != cur_child; cur_child = cur_child.brother {
    if (flag != is_dir_name(cur_child.name)) {
      continue
    }
    tree_store.Append(&child_iter, iter)
    if ('/' == cur_child.name[len(cur_child.name) - 1]) {
      gtk_icon = gtk.GTK_STOCK_DIRECTORY
    } else {
      gtk_icon = gtk.GTK_STOCK_FILE
    }
    tree_store.Set(&child_iter, 
      gtk.Image().RenderIcon(gtk_icon, gtk.GTK_ICON_SIZE_MENU, "").Pixbuf, 
      cur_child.name);
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


func slashed_prefix(a string, b string) int {
  bar := len(a)
  l := len(b)
  if (l < bar) {
    bar = l
  }
  last_slash := 0
  for y := 0; y < bar; y++ {
    if (a[y] != b[y]) {
      return last_slash
    }
    if ('/' == a[y]) {
      last_slash = y + 1
    }
  }
  return bar
}

func file_tree_merge_parent_and_child(child *FileTreeNode) {
  parent := child.parent
  if (parent != &file_tree_root) {
    grand_parent := parent.parent
    new_name := parent.name + child.name
    if is_dir_name(new_name) {
      parent.child = child.child
      parent.name = new_name 
    } else {
      file_tree_remove(grand_parent, parent.name, false)
      file_tree_insert(grand_parent, new_name)
    }
  }
}

func file_tree_remove_node(cur *FileTreeNode, prev *FileTreeNode) {
  if (nil != prev) {
    prev.brother = cur.brother
  } else {
    cur.parent.child = cur.brother
  }
}

func file_tree_remove(root *FileTreeNode, name string, merge_flag bool) {
  cur_child, pos, prev_child := file_tree_find_among_children(root, name)
  name_len := len(name)
  if (pos < name_len) {
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
    if (false == merge_flag) {
      return
    }
    if (nil != prev_child) {
      if (nil == cur_child.brother) && (prev_child == cur_child.parent.child) {
        // There are only two children of root: prev & cur 
        file_tree_merge_parent_and_child(prev_child)
      } 
    } else {
      // prev_child == nil
      if (nil == cur_child.brother.brother) {
        // Only two children of root: cur and his brother
        file_tree_merge_parent_and_child(cur_child.brother)
      }
    }
    return 
  } 
  bump_message("file_tree_remove: unexpected case: name = " + name)
}

func delete_file_from_tree(name string) {
  _, found := file_map[name]
  if (false == found) {
    return
  }
  file_tree_remove(&file_tree_root, name, true)
  file_map[name] = nil, false
}

func add_file_to_tree(name string, bump_flag bool) {
  _, found := file_map[name]
  if (found) {
    if (bump_flag) {
      bump_message("File " + name + " is already open")
    }
    return
  }
  file_tree_insert(&file_tree_root, name)
}

func buf_changed_cb() {
	if source_buf.GetModified() {
		main_window.SetTitle("* " + cur_file)
	} else {
		main_window.SetTitle(cur_file)
	}
}

func mark_set_cb() {
	var cur gtk.GtkTextIter
	var be, en gtk.GtkTextIter

	source_buf.GetSelectionBounds(&be, &en)
	selection := source_buf.GetSlice(&be, &en, false)
	if prev_selection == selection {
		return
	}
	prev_selection = selection

	if selection_flag {
		source_buf.GetStartIter(&be)
		source_buf.GetEndIter(&en)
		source_buf.RemoveTagByName("instance", &be, &en)
		selection_flag = false
	}

	sel_len := len(selection)
	if (sel_len <= 1) || (sel_len >= 100) {
		return
	} else {
		selection_flag = true
	}

	source_buf.GetStartIter(&cur)
	for cur.ForwardSearch(selection, 0, &be, &cur, nil) {
		source_buf.ApplyTagByName("instance", &be, &cur)
	}
}

func bump_message(m string) {
	dialog := gtk.MessageDialog(
		main_window.GetTopLevelAsWindow(),
		gtk.GTK_DIALOG_MODAL,
		gtk.GTK_MESSAGE_INFO,
		gtk.GTK_BUTTONS_OK,
		m)
	dialog.Run()
	dialog.Destroy()
}

func open_cb() {
  file_save_current()
	file_dialog := gtk.FileChooserDialog2("", source_view.GetTopLevelAsWindow(),
		gtk.GTK_FILE_CHOOSER_ACTION_OPEN,
		gtk.GTK_STOCK_CANCEL, gtk.GTK_RESPONSE_CANCEL,
		gtk.GTK_STOCK_OPEN, gtk.GTK_RESPONSE_ACCEPT)
	file_dialog.SetCurrentFolder(prev_dir)
	res := file_dialog.Run()
	dialog_folder := file_dialog.GetCurrentFolder()
	dialog_file := file_dialog.GetFilename()
	file_dialog.Destroy()
	if gtk.GTK_RESPONSE_ACCEPT == res {
		prev_dir = dialog_folder
		file, _ := os.Open(dialog_file, os.O_RDONLY, 0700)
		if nil == file {
			bump_message("Unable to open file for reading: " + dialog_file)
			return
		}
		stat, _ := file.Stat()
		if nil == stat {
			bump_message("Unable to stat file: " + dialog_file)
			file.Close()
			return
		}
		buf := make([]byte, stat.Size)
		nread, _ := file.Read(buf)
		if nread != int(stat.Size) {
			bump_message("Unable to read whole file: " + dialog_file)
			file.Close()
			return
		}
		file.Close()
		cur_file = dialog_file
		if false == glib.Utf8Validate(buf, nread, nil) {
			bump_message("File " + cur_file + " is not correct utf8 text")
			close_cb()
			return
		}
		
		source_buf.BeginNotUndoableAction()
		source_buf.SetText(string(buf))
		source_buf.SetModified(false)
		source_buf.EndNotUndoableAction()

		add_file_to_tree(cur_file, true)
		file_map[cur_file] = new(FileRecord)
		file_tree_store()
	}
}

func save_cb() {
	if "" == cur_file {
		save_as_cb()
	} else {
		file, _ := os.Open(cur_file, os.O_CREAT|os.O_WRONLY, 0700)
		if nil == file {
			bump_message("Unable to open file for writing: " + cur_file)
			return
		}
		var be, en gtk.GtkTextIter
		source_buf.GetStartIter(&be)
		source_buf.GetEndIter(&en)
		text_to_save := source_buf.GetText(&be, &en, true)
		nbytes, _ := file.WriteString(text_to_save)
		if nbytes != len(text_to_save) {
			bump_message("Error while writing to file: " + cur_file)
			return
		}
		source_buf.SetModified(false)
		file.Truncate(int64(nbytes))
		file.Close()
		main_window.SetTitle(cur_file)
	}
}

func save_as_cb() {
	file_dialog := gtk.FileChooserDialog2("", source_view.GetTopLevelAsWindow(),
		gtk.GTK_FILE_CHOOSER_ACTION_SAVE,
		gtk.GTK_STOCK_CANCEL, gtk.GTK_RESPONSE_CANCEL,
		gtk.GTK_STOCK_SAVE, gtk.GTK_RESPONSE_ACCEPT)
	file_dialog.SetCurrentFolder(prev_dir)
	res := file_dialog.Run()
	dialog_folder := file_dialog.GetCurrentFolder()
	dialog_file := file_dialog.GetFilename()
	file_dialog.Destroy()
	if gtk.GTK_RESPONSE_ACCEPT == res {
		prev_dir = dialog_folder
		cur_file = dialog_file
		save_cb()
	}
}

func exit_cb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here
	gtk.MainQuit()
}

func close_cb() {
  if ("" == cur_file) {
    return
  }
	delete_file_from_tree(cur_file)
	file_tree_store()
	cur_file = ""
	main_window.SetTitle("")
	source_buf.BeginNotUndoableAction()
	source_buf.SetText("")
	source_buf.EndNotUndoableAction()
}

func paste_done_cb() {
	var be, en gtk.GtkTextIter
	source_buf.GetStartIter(&be)
	source_buf.GetEndIter(&en)
	source_buf.RemoveTagByName("instance", &be, &en)
	selection_flag = false
}

func tree_view_path(iter *gtk.GtkTreeIter) string {
  var ans string
  ans = ""
  for ; ; {
    var val gtk.GValue
    var next gtk.GtkTreeIter
    tree_model.GetValue(iter, 1, &val)
    ans = val.GetString() + ans
    if (false == tree_model.IterParent(&next, iter)) {
      break
    }
    iter = &next
  }
  return ans
}

func tree_view_select_cb() {
  var path *gtk.GtkTreePath;
  var column *gtk.GtkTreeViewColumn;
  tree_view.GetCursor(&path, &column);
  var iter gtk.GtkTreeIter
  tree_model.GetIterFromString(&iter, path.String())
  sel_file := tree_view_path(&iter)
  file_save_current()
  file_switch_to(sel_file)
}

func init_widgets() {
	lang_man := gtk.SourceLanguageManagerGetDefault()
	lang := lang_man.GetLanguage("go")
	if nil == lang.SourceLanguage {
		println("warning: no language specification")
	}
	source_buf = gtk.SourceBuffer()
	source_buf.SetLanguage(lang)
	source_buf.Connect("paste-done", paste_done_cb, nil)
	source_buf.Connect("mark-set", mark_set_cb, nil)
	source_buf.Connect("changed", buf_changed_cb, nil)

	source_buf.CreateTag("instance", map[string]string{"background": "#FF8080"})

	tree_store = gtk.TreeStore(gdkpixbuf.GetGdkPixbufType(), gtk.TYPE_STRING)
	tree_view = gtk.TreeView()
	tree_view.ModifyFontEasy("Regular 8")
	tree_model = tree_store.ToTreeModel()
	tree_view.SetModel(tree_model)
	tree_view.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererPixbuf(), "pixbuf", 0))
	tree_view.AppendColumn(gtk.TreeViewColumnWithAttributes(
		"", gtk.CellRendererText(), "text", 1))
	tree_view.SetHeadersVisible(false)
	tree_view.Connect("cursor-changed", tree_view_select_cb, nil)

	source_view = gtk.SourceViewWithBuffer(source_buf)
	source_view.ModifyFontEasy("Monospace Regular 10")
	source_view.SetAutoIndent(true)
	source_view.SetHighlightCurrentLine(true)
	source_view.SetShowLineNumbers(true)
	source_view.SetRightMarginPosition(80)
	source_view.SetShowRightMargin(true)
	source_view.SetIndentWidth(2)
	source_view.SetInsertSpacesInsteadOfTabs(true)
	source_view.SetDrawSpaces(gtk.GTK_SOURCE_DRAW_SPACES_TAB)
	source_view.SetTabWidth(2)
	source_view.SetSmartHomeEnd(gtk.GTK_SOURCE_SMART_HOME_END_ALWAYS)

	vbox := gtk.VBox(false, 0)
	hpaned := gtk.HPaned()

	menubar := gtk.MenuBar()
	vbox.PackStart(menubar, false, false, 0)
	vbox.PackStart(hpaned, true, true, 0)

	file_item := gtk.MenuItemWithMnemonic("_File")
	menubar.Append(file_item)
	file_submenu := gtk.Menu()
	file_item.SetSubmenu(file_submenu)

	accel_group := gtk.AccelGroup()

	open_item := gtk.MenuItemWithMnemonic("_Open")
	file_submenu.Append(open_item)
	open_item.Connect("activate", open_cb, nil)
	open_item.AddAccelerator("activate", accel_group, gdk.GDK_o,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	save_item := gtk.MenuItemWithMnemonic("_Save")
	file_submenu.Append(save_item)
	save_item.Connect("activate", save_cb, nil)
	save_item.AddAccelerator("activate", accel_group, gdk.GDK_s,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	save_as_item := gtk.MenuItemWithMnemonic("Save _as")
	file_submenu.Append(save_as_item)
	save_as_item.Connect("activate", save_as_cb, nil)

	close_item := gtk.MenuItemWithMnemonic("_Close")
	file_submenu.Append(close_item)
	close_item.Connect("activate", close_cb, nil)
	close_item.AddAccelerator("activate", accel_group, gdk.GDK_w,
		gdk.GDK_CONTROL_MASK, gtk.GTK_ACCEL_VISIBLE)

	exit_item := gtk.MenuItemWithMnemonic("E_xit")
	file_submenu.Append(exit_item)
	exit_item.Connect("activate", exit_cb, nil)

	tree_window := gtk.ScrolledWindow(nil, nil)
	tree_window.SetSizeRequest(300, 0)
	tree_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_AUTOMATIC)
	hpaned.Add1(tree_window)
	tree_window.Add(tree_view)

	text_window := gtk.ScrolledWindow(nil, nil)
	text_window.SetPolicy(gtk.GTK_POLICY_AUTOMATIC, gtk.GTK_POLICY_ALWAYS)
	hpaned.Add2(text_window)
	text_window.Add(source_view)

	main_window = gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	main_window.Maximize()
	main_window.SetTitle("tabby")
	main_window.Connect("destroy", exit_cb, "")
	main_window.Add(vbox)
	main_window.ShowAll()
	main_window.AddAccelGroup(accel_group)

	source_view.GrabFocus()
}

func init_vars() {
	file_map = make(map[string]*FileRecord)
  cur_file = ""
}

func main() {
	gtk.Init(nil)
	init_widgets()
	init_vars()
	gtk.Main()
}
