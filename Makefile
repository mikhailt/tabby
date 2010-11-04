include $(GOROOT)/src/Make.inc

TARG=tabby
GOFILES=\
	tabby.go\

fix_style:
	cp ./go.lang /usr/share/gtksourceview-2.0/language-specs/
	cp ./classic.xml /usr/share/gtksourceview-2.0/styles/

include $(GOROOT)/src/Make.cmd
