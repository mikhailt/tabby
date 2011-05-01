tabby
======

  Source code editor written in Go using go-gtk bindings. It aims to handle 
  navigation effectively among large number of files.

SCREENSHOT:
-----------
Saturday, April 30, 2011 16:04
Forked from github.com/mikhailt/tabby.
![tabby!](https://github.com/mikhailt/tabby/raw/gh-pages/tabby.png "tabby!")

DEPENDENCIES:
--------
  go
  go-gtk
  libgtk2.0-dev
  libgtksourceview2.0-dev

BUILD:
--------
  Compile & run:
    make
    ./tabby 
    
  Put style sheets to corresponding places:
    make fix_style
    
  Install:
    make install
