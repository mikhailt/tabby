package main

import (
	"os"
	"io"
	"bytes"
	"gtk"
)

func gofmt_all() {
	for file, _ := range file_map {
		go gofmt(file)
	}
	if cur_file == "" {
		gofmt("")
	}
}
func gofmt(file string) {
	rec, _ := file_map[file]
	var buf []byte
	if file == cur_file {
		buf = []byte(get_source())
	} else {
		buf = rec.buf
	}
	std, error, e := getOutput([]string{os.Getenv("GOBIN") + "/gofmt"}, buf)
	if e != nil {
		println(e.String())
		return
	}
	if file == cur_file {
		error_buf.SetText(string(error))
	} else {
		rec.error = error
	}
	if len(error) == 0 {
		if file == cur_file {
			if string(buf) != string(std) {
				var be, en gtk.GtkTextIter
				source_buf.GetSelectionBounds(&be, &en)
				sel_be := be.GetOffset()
				source_buf.SetText(string(std))
				source_buf.GetIterAtOffset(&be, sel_be)
				move_focus_and_selection(&be, &be)
			}
		} else if string(rec.buf) != string(std) {
			rec.buf = std
			rec.modified = true
			my_iter := tree_view_set_name_iter(file, false)
			if tree_store.IterIsValid(my_iter) {
				var val gtk.GValue
				tree_model.GetValue(my_iter, 0, &val)
				tree_store.Set(my_iter, string('C')+val.GetString()[1:])
			}
			//tree_view_set_name_iter(file, false)
		}
	}
}
func getOutput(args []string, input []byte) (std []byte, error []byte, e os.Error) {
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
	pid, err := os.ForkExec(args[0], args, os.Environ(), "", []*os.File{inpr, stdw, errw})

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
	os.Wait(pid, 0)
	return
}
