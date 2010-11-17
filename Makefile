include $(GOROOT)/src/Make.inc

TARG=tabby
GOFILES=\
	src/main.go\
	src/file_tree.go\
	src/file_record.go\
	src/menu_callback.go\
	src/tree_view.go\
	src/session.go\
	src/navigation.go

fix_style:
	cp ./go.lang /usr/share/gtksourceview-2.0/language-specs/
	cp ./classic.xml /usr/share/gtksourceview-2.0/styles/

build_file_tree:
	cd ./file_tree && gomake install

all: ./src/* ./file_tree/*
	make build_file_tree
	make fix_style
	make tabby

include $(GOROOT)/src/Make.cmd