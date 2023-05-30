package main

import (
	"bytes"
	"github.com/mattn/go-gtk/gtk"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func gofmtAll() {
	for file := range fileMap {
		gofmt(file)
	}
	fileTreeStore()
}

func gofmt(file string) {
	rec := fileMap[file]
	var buf []byte
	if file == curFile {
		buf = []byte(getSource())
	} else {
		buf = rec.buf
	}
	bin, err := getGofmtBin()
	if err != nil {
		tabbyLog(err.Error())
		return
	}
	std, err, e := getOutput([]string{bin}, buf)
	if e != nil {
		tabbyLog(e.Error())
		return
	}
	if file == curFile {
		errorBuf.SetText(string(err))
	} else {
		rec.error = err
	}
	if len(err) != 0 {
		return
	}
	if file == curFile {
		if string(buf) != string(std) {
			be, en := gtk.TextIter{}, gtk.TextIter{}
			sourceBuf.GetSelectionBounds(&be, &en)
			selBeOffset := be.GetOffset()
			sourceBuf.SetText(string(std))
			sourceBuf.GetIterAtOffset(&be, selBeOffset)
			moveFocusAndSelection(&be, &be)
		}
	} else if string(rec.buf) != string(std) {
		rec.buf = std
		rec.modified = true
	}
}

func getOutput(args []string, input []byte) ([]byte, []byte, error) {
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

	var stdBuf, errBuf bytes.Buffer
	io.Copy(&stdBuf, stdr)
	std := stdBuf.Bytes()
	stdBuf.Reset()
	io.Copy(&errBuf, errr)
	errBytes := errBuf.Bytes()

	inpr.Close()
	stdr.Close()
	errr.Close()
	pid.Wait()
	return std, errBytes, nil
}

func getGofmtBin() (string, error) {
	bin := os.Getenv("GOBIN")
	if bin != "" {
		bin = filepath.Join(bin, "gofmt")
	} else {
		var err error
		bin, err = exec.LookPath("gofmt")
		if err != nil {
			return "", err
		}
	}
	return bin, nil
}