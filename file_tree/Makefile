include $(GOROOT)/src/Make.inc

TARG     = file_tree
CGOFILES = file_tree.go

CGO_CFLAGS  = `pkg-config --cflags gtk+-2.0`
CGO_LDFLAGS = `pkg-config --libs gtk+-2.0`

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O
