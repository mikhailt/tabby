package main

import (
	"os"
	"os/exec"
	"io"
	"bytes"
	"path/filepath"
	"github.com/mattn/go-gtk/gtk"
)

func gofmt_all() {
	for file, _ := range file_map {
		gofmt(file)
	}
	file_tree_store()
}

func gofmt(file string) {
	rec, _ := file_map[file]
	var buf []byte
	if file == cur_file {
		buf = []byte(get_source())
	} else {
		buf = rec.buf
	}
	bin := os.Getenv("GOBIN")
	if bin != "" {
		bin = filepath.Join(bin, "gofmt")
	} else {
		bin, _ = exec.LookPath("gofmt")
	}
	std, error, e := get_output([]string{bin}, buf)
	if e != nil {
		tabby_log(e.Error())
		return
	}
	if file == cur_file {
		error_buf.SetText(string(error))
	} else {
		rec.error = error
	}
	if 0 != len(error) {
		return
	}
	if file == cur_file {
		if string(buf) != string(std) {
			var be, en gtk.TextIter
			source_buf.GetSelectionBounds(&be, &en)
			sel_be := be.GetOffset()
			source_buf.SetText(string(std))
			source_buf.GetIterAtOffset(&be, sel_be)
			move_focus_and_selection(&be, &be)
		}
	} else if string(rec.buf) != string(std) {
		rec.buf = std
		rec.modified = true
	}
}

func get_output(args []string, input []byte) (std []byte, error []byte, e error) {
	inpr, inpw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	stdr, stdw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	errr, errw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	pid, err := os.StartProcess(args[0], args, &os.ProcAttr{
		Dir:   ".",
		Env:   os.Environ(),
		Files: []*os.File{inpr, stdw, errw},
	})

	if err != nil {
		return nil, nil, err
	}

	inpw.Write(input)
	inpw.Close()
	stdw.Close()
	errw.Close()

	var b bytes.Buffer
	io.Copy(&b, stdr)
	std = b.Bytes()
	b.Reset()
	io.Copy(&b, errr)
	error = b.Bytes()

	inpr.Close()
	stdr.Close()
	errr.Close()
	pid.Wait()
	return
}
