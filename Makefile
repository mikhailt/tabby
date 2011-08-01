include $(GOROOT)/src/Make.inc

ifeq ($(GOARCH), amd64)
	C = 6g
	L = 6l
else ifeq ($(GOARCH), 386)
	C = 8g
	L = 8l
else 
	C = 5g
	L = 5l
endif

.DEFAULT_GOAL=all

tabby: src/*
	$C -o tabby.6 ./src/*
	$L -o tabby tabby.6

fix_style:
	sudo cp ./go.lang /usr/share/gtksourceview-2.0/language-specs/
	sudo chmod 644 /usr/share/gtksourceview-2.0/language-specs/go.lang
	sudo cp ./classic.xml /usr/share/gtksourceview-2.0/styles/
	sudo chmod 644 /usr/share/gtksourceview-2.0/styles/classic.xml

build_file_tree: file_tree/*
	cd ./file_tree && gomake install

all: build_file_tree tabby
	cp ./.tabbyignore ~/

c:
	rm -f tabby *.6
	cd ./file_tree && gomake clean

install: all
	install -m 755 ./tabby ${GOBIN}
