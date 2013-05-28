package file_tree

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
// #cgo pkg-config: gtk+-2.0
import "C"
import "github.com/mattn/go-gtk/gtk"

func NewFileTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(C.create_tabby_file_tree())}}
}

func NewSearchTree() *gtk.TreeView {
	return &gtk.TreeView{gtk.Container{
		*gtk.WidgetFromNative(C.create_tabby_search_tree())}}
}
