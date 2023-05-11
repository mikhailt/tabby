// Initializes the tabby editor
func init_tabby() {
	init_navigation()
	init_inotify()

	search_view.Init()

	source_buf = gtksourceview.NewSourceBuffer()
	source_buf.Connect("paste-done", paste_done_cb, nil)
	source_buf.Connect("mark-set", mark_set_cb, nil)
	source_buf.Connect("modified-changed", buf_changed_cb, nil)

	init_lang()

	source_buf.CreateTag("instance", map[string]string{"background": "#FF8080"})

	tree_store = gtk.NewTreeStore(gtk.TYPE_STRING)
	tree_view = file_tree.NewFileTree()
	tree_view.ModifyFontEasy("Regular 8")
	tree_model = tree_store.ToTreeModel()
	tree_view.SetModel(tree_model)
	tree_view.SetHeadersVisible(false)
	tree_view.Connect("cursor-changed", tree_view_select_cb, nil)

	error_view = gtk.NewTextView()
	error_view.ModifyFontEasy("Monospace Regular 8")
	error_view.SetEditable(false)
	error_buf = error_view.GetBuffer()

	source_view = gtksourceview.NewSourceViewWithBuffer(source_buf)
	source_view.ModifyFontEasy("Monospace Regular 10")
	source_view.SetAutoIndent(true)
	source_view.SetHighlightCurrentLine(true)
	source_view.SetShowLineNumbers(true)
	source_view.SetRightMarginPosition(80)
	source_view.SetShowRightMargin(true)
	source_view.SetIndentWidth(2)
	source_view.SetTabWidth(2)
	source_view.SetInsertSpacesInsteadOfTabs(opt.space_not_tab)
	source_view.SetDrawSpaces(gtksourceview.SOURCE_DRAW_SPACES_TAB)
	source_view.SetSmartHomeEnd(gtksourceview.SOURCE_SMART_HOME_END_ALWAYS)
	source_view.SetWrapMode(gtk.WRAP_WORD)

	vbox := gtk.NewVBox(false, 0)
	inner_hpaned := gtk.NewHPaned()
	view_vpaned := gtk.NewVPaned()
	outer_hpaned := gtk.NewHPaned()
	outer_hpaned.Add1(inner_hpaned)
	inner_hpaned.Add2(view_vpaned)

	menubar := gtk.NewMenuBar()
	vbox.PackStart(menubar, false, false, 0)
	vbox.PackStart(outer_hpaned, true, true, 0)

	file_item := gtk.NewMenuItemWithMnemonic("_File")
	menubar.Append(file_item)
	file_submenu := gtk.NewMenu()
	file_item.SetSubmenu(file_submenu)

	accel_group := gtk.NewAccelGroup()

	// New function added to comment this
	new_item := gtk.NewMenuItemWithMnemonic("_New")
	file_submenu.Append(new_item)
	new_item.Connect("activate", new_cb, nil)
	new_item.AddAccelerator("activate", accel_group, gdk.KEY_n, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	open_item := gtk.NewMenuItemWithMnemonic("_Open")
	file_submenu.Append(open_item)
	open_item.Connect("activate", open_cb, nil)
	open_item.AddAccelerator("activate", accel_group, gdk.KEY_o, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	open_rec_item := gtk.NewMenuItemWithMnemonic("Open _Recursively")
	file_submenu.Append(open_rec_item)
	open_rec_item.Connect("activate", open_rec_cb, nil)

	save_item := gtk.NewMenuItemWithMnemonic("_Save")
	file_submenu.Append(save_item)
	save_item.Connect("activate", save_cb, nil)
	save_item.AddAccelerator("activate", accel_group, gdk.KEY_s, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	save_as_item := gtk.NewMenuItemWithMnemonic("Save _as")
	file_submenu.Append(save_as_item)
	save_as_item.Connect("activate", save_as_cb, nil)

	close_item := gtk.NewMenuItemWithMnemonic("_Close")
	file_submenu.Append(close_item)
	close_item.Connect("activate", close_cb, nil)
	close_item.AddAccelerator("activate", accel_group, gdk.KEY_w, gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE)

	exit_item := gtk.NewMenuItemWithMnemonic("E_xit")
	file_submenu.Append(exit_item)
	exit_item.Connect("activate", exit_cb, nil)

	navigation_item := gtk.NewMenuItemWithMnemonic("_Navigation")
	menubar.Append(navigation_item)
	navigation_submenu := gtk.NewMenu()
	navigation_item.SetSubmenu(navigation_submenu)

	prev_instance_item := gtk.NewMenuItemWithMnemonic("_Previous Instance")
	navigation_submenu.Append(prev_instance_item)
	prev_instance_item.Connect("activate", prev_instance_cb, nil)
	prev_instance_item.AddAccelerator("activate", accel_group, gdk.KEY_F2,
		0, gtk.ACCEL_VISIBLE)

	next_instance_item := gtk.NewMenuItemWithMnemonic("_Next Instance")
	navigation_submenu.Append(next_instance_item)
	next_instance_item.Connect("activate", next_instance_cb, nil)
	next_instance_item.AddAccelerator("activate", accel_group, gdk.KEY_F3,
		0, gtk.ACCEL_VISIBLE)

	prev_result_item := gtk.NewMenuItemWithMnemonic("Prev search result")
	navigation_submenu.Append(prev_result_item)
	prev_result_item.Connect("activate",