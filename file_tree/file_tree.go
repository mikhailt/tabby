package file_tree

/*
#include <gtk/gtk.h>
#include <stdlib.h>
#include <string.h>
*/

//tabby_renderer function renders the text and background color for a tree view column
//based on the first character of the string
//It takes GtkTreeViewColumn, GtkCellRenderer, GtkTreeModel, GtkTreeIter and a pointer as arguments
//It returns nothing
func tabby_renderer(col *C.GtkTreeViewColumn, renderer *C.GtkCellRenderer, model *C.GtkTreeModel, iter *C.GtkTreeIter, user_data C.gpointer) {
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

//search_renderer function renders the text for a search tree view column
//It takes GtkTreeViewColumn, GtkCellRenderer, GtkTreeModel, GtkTreeIter and a pointer as arguments
//It returns nothing
func search_renderer(col *C.GtkTreeViewColumn, renderer *C.GtkCellRenderer, model *C.GtkTreeModel, iter *C.GtkTreeIter, user_data C.gpointer) {
  gchar* str;
  gchar* p;

  gtk_tree_model_get(model, iter, 0, &str, -1);
  for (p = &str[strlen(str) - 1]; '/' != *p; --p) {
    ;
  }
  g_object_set(renderer, "text", p + 1, NULL);
  free(str);
}

//create_tabby_file_tree function creates a file tree view widget with a single column
//It takes no arguments
//It returns a void pointer to the widget
func create_tabby_file_tree() C.voidptr {
  GtkTreeViewColumn   *col;
  GtkCellRenderer     *renderer;
  GtkWidget           *view;

  view = gtk_tree_view_new();
  col = gtk_tree_view_column_new();
  gtk_tree_view_append_column(GTK_TREE_VIEW(view), col);
  renderer = gtk_cell_renderer_text_new();
  gtk_tree_view_column_pack_start(col, renderer, TRUE);
  gtk_tree_view_column_set_cell_data_func(col, renderer, tabby_renderer, NULL,
                                          NULL);
  return view;
}

//create_tabby_search_tree function creates a search tree view widget with a single column
//It takes no arguments
//It returns a void pointer to the widget
func create_tabby_search_tree() C.voidptr {
  GtkTreeViewColumn   *col;
  GtkCellRenderer     *renderer;
  GtkWidget           *view;

  view = gtk_tree_view_new();
  col = gtk_tree_view_column_new();
  gtk_tree_view_append_column(GTK_TREE_VIEW(view), col);
  renderer = gtk_cell_renderer_text_new();
  gtk_tree_view_column_pack_start(col, renderer, TRUE);
  gtk_tree_view_column_set_cell_data_func(col, renderer, search_renderer, NULL,
                                          NULL);
  return view;
}

//import "C" statement imports C code
//It returns nothing
// #cgo pkg-config: gtk+-2.0
import "C"

//NewFileTree function creates and returns a new file tree widget
//It takes no arguments
//It returns a pointer to the widget
func NewFileTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(C.create_tabby_file_tree())}}
}

//NewSearchTree function creates and returns a new search tree widget
//It takes no arguments
//It returns a pointer to the widget
func NewSearchTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(C.create_tabby_search_tree())}}
}