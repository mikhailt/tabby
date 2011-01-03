include $(GOROOT)/src/Make.inc

.DEFAULT_GOAL=all

TARG=tabby
GOFILES=\
	src/main.go\
	src/args.go\
	src/file_tree.go\
	src/file_record.go\
	src/menu_callback.go\
	src/tree_view.go\
	src/session.go\
	src/navigation.go\
	src/inotify.go\
	src/options.go\
	src/tools.go\
	src/find_and_replace.go\
	src/find.go

fix_style:
	sudo cp ./go.lang /usr/share/gtksourceview-2.0/language-specs/
	sudo cp ./classic.xml /usr/share/gtksourceview-2.0/styles/

build_file_tree:
	cd ./file_tree && gomake install

all:
	cp ./.tabbyignore ~/
	make build_file_tree
	make tabby

c:
	gomake clean
	cd ./file_tree && gomake clean

include $(GOROOT)/src/Make.cmd
