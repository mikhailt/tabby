package file_tree

// tabby_renderer is a function used as a callback to process the data in a GtkTreeViewColumn.
// It receives GtkTreeViewColumn, GtkCellRenderer, GtkTreeModel, GtkTreeIter and gpointer as parameters.
// It sets the foreground and background color of the renderer based on the first character of the string data.
// It also sets the text of the renderer to the rest of the string data.
// It frees the memory allocated for the string data.
func tabby_renderer(col *C.GtkTreeViewColumn, renderer *C.GtkCellRenderer,
	model *C.GtkTreeModel, iter *C.GtkTreeIter, user_data C.gpointer) {
	gchar* str;
	unsigned char c;

	gtk_tree_model_get(model, iter, 0, &str, -1);
	c = str[0];
	if ('d' == c) {
		g_object_set(renderer, "foreground", "Blue", "foreground-set", TRUE, 
			"background", "White", "background-set", TRUE, NULL);
	} else if ('D' == c) {
		g_object_set(renderer, "foreground", "Blue", "foreground-set", TRUE,
			"background", "#DBEDFF", "background-set", TRUE, NULL);
	} else {
		if (('C' == c) || ('B' == c)) {
			g_object_set(renderer, "background", "#B3FFC2", "background-set", TRUE, NULL);
		} else {
			g_object_set(renderer, "background", "#FFFAFA", "background-set", TRUE, NULL);
		}
		if (('c' == c) || ('C' == c)) {
			g_object_set(renderer, "foreground", "Red", "foreground-set", TRUE, NULL);  
		} else {
			g_object_set(renderer, "foreground", "Black", "foreground-set", TRUE, NULL);
		}
	}
	g_object_set(renderer, "text", str + 1, NULL);
	free(str);
}

// search_renderer is a function used as a callback to process the data in a GtkTreeViewColumn.
// It receives GtkTreeViewColumn, GtkCellRenderer, GtkTreeModel, GtkTreeIter and gpointer as parameters.
// It sets the text of the renderer to the last element of the string data.
// It frees the memory allocated for the string data.
func search_renderer(col *C.GtkTreeViewColumn, renderer *C.GtkCellRenderer,
	model *C.GtkTreeModel, iter *C.GtkTreeIter, user_data C.gpointer) {
	gchar* str;
	gchar* p;

	gtk_tree_model_get(model, iter, 0, &str, -1);
	for (p = &str[strlen(str) - 1]; '/' != *p; --p) {
		;
	}
	g_object_set(renderer, "text", p + 1, NULL);
	free(str);
}

// create_tabby_file_tree is a function that creates and returns a new GtkTreeView for displaying file tree data.
// It creates a new GtkTreeViewColumn and sets the cell data function to tabby_renderer.
func create_tabby_file_tree() unsafe.Pointer {
	var col *C.GtkTreeViewColumn
	var renderer *C.GtkCellRenderer
	var view *C.GtkWidget

	view = gtk_tree_view_new()
	col = gtk_tree_view_column_new()
	gtk_tree_view_append_column(GTK_TREE_VIEW(view), col)
	renderer = gtk_cell_renderer_text_new()
	gtk_tree_view_column_pack_start(col, renderer, TRUE)
	gtk_tree_view_column_set_cell_data_func(col, renderer, tabby_renderer, nil,
		nil)
	return unsafe.Pointer(view)
}

// create_tabby_search_tree is a function that creates and returns a new GtkTreeView for displaying search results data.
// It creates a new GtkTreeViewColumn and sets the cell data function to search_renderer.
func create_tabby_search_tree() unsafe.Pointer {
	var col *C.GtkTreeViewColumn
	var renderer *C.GtkCellRenderer
	var view *C.GtkWidget

	view = gtk_tree_view_new()
	col = gtk_tree_view_column_new()
	gtk_tree_view_append_column(GTK_TREE_VIEW(view), col)
	renderer = gtk_cell_renderer_text_new()
	gtk_tree_view_column_pack_start(col, renderer, TRUE)
	gtk_tree_view_column_set_cell_data_func(col, renderer, search_renderer, nil,
		nil)
	return unsafe.Pointer(view)
}

// NewFileTree is a function that creates and returns a new *gtk.TreeView for displaying file tree data.
func NewFileTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(create_tabby_file_tree())}}
}

// NewSearchTree is a function that creates and returns a new *gtk.TreeView for displaying search results data.
func NewSearchTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(create_tabby_search_tree())}}
}