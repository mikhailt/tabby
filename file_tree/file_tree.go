package file_tree

// tabby_renderer is a function that sets the cell renderer properties based on the value in the model
/*
#include <gtk/gtk.h>
#include <stdlib.h>
#include <string.h>

void tabby_renderer(GtkTreeViewColumn *col,
                    GtkCellRenderer   *renderer,
                    GtkTreeModel      *model,
                    GtkTreeIter       *iter,
                    gpointer           user_data) {
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
*/

// search_renderer is a function that sets the cell renderer properties to display the name of the file being searched
/*
void search_renderer(GtkTreeViewColumn *col,
                     GtkCellRenderer   *renderer,
                     GtkTreeModel      *model,
                     GtkTreeIter       *iter,
                     gpointer           user_data) {
  gchar* str;
  gchar* p;

  gtk_tree_model_get(model, iter, 0, &str, -1);
  for (p = &str[strlen(str) - 1]; '/' != *p; --p) {
    ;
  }
  g_object_set(renderer, "text", p + 1, NULL);
  free(str);
}
*/

// create_tabby_file_tree is a function that creates and returns a new tabby file tree structure
/*
static void* create_tabby_file_tree() {
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
*/

// create_tabby_search_tree is a function that creates and returns a new tabby search tree structure
/*
static void* create_tabby_search_tree() {
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
*/

// NewFileTree is a function that creates and returns a new file tree widget
func NewFileTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(C.create_tabby_file_tree())}}
}

// NewSearchTree is a function that creates and returns a new search tree widget
func NewSearchTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(C.create_tabby_search_tree())}}
}