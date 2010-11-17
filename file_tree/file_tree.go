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
  int len;

  gtk_tree_model_get(model, iter, 0, &str, -1);
  len = strlen(str);
  if ('/' == str[len - 1]) {
    g_object_set(renderer, "foreground", "Blue", "foreground-set", TRUE, NULL);
  } else {
    g_object_set(renderer, "foreground", "Black", "foreground-set", TRUE, NULL);
  }
  g_object_set(renderer, "text", str, NULL);
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
*/
import "C"
import "gtk"

//g_object_set(renderer, "text", buf, NULL); 


func NewFileTree() *gtk.GtkTreeView {
  return &gtk.GtkTreeView{gtk.GtkContainer{gtk.GtkWidget{
    gtk.ToGtkWidget(C.create_tabby_file_tree())}}}
}