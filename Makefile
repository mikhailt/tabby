include $(GOROOT)/src/Make.inc

TARG=tabby
GOFILES=\
	src/main.go\
	src/file_tree.go\
	src/file_record.go\
	src/menu_callback.go

fix_style:
	cp ./go.lang /usr/share/gtksourceview-2.0/language-specs/
	cp ./classic.xml /usr/share/gtksourceview-2.0/styles/

include $(GOROOT)/src/Make.cmd