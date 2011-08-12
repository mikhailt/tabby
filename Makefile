include $(GOROOT)/src/Make.inc

ifeq ($(GOARCH), amd64)
	C = 6g
	L = 6l
	O = 6
else ifeq ($(GOARCH), 386)
	C = 8g
	L = 8l
	O = 8
else 
	C = 5g
	L = 5l
	O = 5
endif


ifeq ($(GOOS),windows)
PREFIX=c:/gtk
TARG=tabby.exe
SRCS=\
	src/args.go \
	src/file_record.go \
	src/file_tree.go \
	src/find.go \
	src/find_and_replace.go \
	src/inotify_windows.go \
	src/lang.go \
	src/main.go \
	src/menu_callback.go \
	src/navigation.go \
	src/options.go \
	src/search_view.go \
	src/session.go \
	src/tools.go \
	src/tree_view.go \

else
PREFIX=/usr
TARG=tabby
SRCS=\
	src/args.go \
	src/file_record.go \
	src/file_tree.go \
	src/find.go \
	src/find_and_replace.go \
	src/inotify_unix.go \
	src/lang.go \
	src/main.go \
	src/menu_callback.go \
	src/navigation.go \
	src/options.go \
	src/search_view.go \
	src/session.go \
	src/tools.go \
	src/tree_view.go \

endif

.DEFAULT_GOAL=all

${TARG}: ${SRCS}
	$C -o tabby.${O} ${SRCS}
	$L -o ${TARG} tabby.${O}

fix_style:
	sudo cp ./go.lang ${PREFIX}/share/gtksourceview-2.0/language-specs/
	sudo chmod 644 ${PREFIX}/share/gtksourceview-2.0/language-specs/go.lang
	sudo cp ./classic.xml ${PREFIX}/share/gtksourceview-2.0/styles/
	sudo chmod 644 ${PREFIX}/share/gtksourceview-2.0/styles/classic.xml

build_file_tree: file_tree/*
	cd ./file_tree && gomake install

all: build_file_tree ${TARG}
	cp ./.tabbyignore ~/

c:
	rm -f ${TARG} *.${O}
	cd ./file_tree && gomake clean

install: all
	install -m 755 ./${TARG} ${GOBIN}
