package main

import (
	"os"
	"strconv"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

var (
	prevDir       string
	lastUnsaved  int = -1
	curFile         string
	fileMap         = make(map[string]FileInfo)
	opt             = &options{
		font:            "Monospace 10",
		spaceNotTab:     false,
		showSearch:      false,
		showError:       false,
		ignoreDotFiles:  true,
		ignoreNames:     make(map[string]bool),
		ignoreExtensions: make(map[string]bool),
	}
)

type FileInfo struct {
	buf      []byte
	modified bool
}

type options struct {
	font             string
	spaceNotTab      bool
	showSearch       bool
	showError        bool
	ignoreDotFiles   bool
	ignoreNames      map[string]bool
	ignoreExtensions map[string]bool
}

func newCb() {
	lastUnsaved++
	file := "unsaved file " + strconv.Itoa(lastUnsaved)
	addFileRecord(file, []byte(""), true)
	fileMap[file].modified = true
	fileTreeStore()
	fileSwitchTo(file)
	treeViewSetCurIter(true)
}

func openCb() {
	fileSaveCurrent()
	dialogOk, dialogFile := fileChooserDialog(OPEN_DIALOG)
	if false == dialogOk {
		return
	}
	readOk, buf := openFileReadToBuf(dialogFile, true)
	if false == readOk {
		return
	}
	if addFileRecord(dialogFile, buf, true) {
		fileTreeStore()
		fileSwitchTo(dialogFile)
	}
}

func openRecCb() {
	dialogOk, dialogDir := fileChooserDialog(OPEN_DIR_DIALOG)
	if false == dialogOk {
		return
	}
	dir, _ := os.OpenFile(dialogDir, os.O_RDONLY, 0)
	if nil == dir {
		bumpMessage("Unable to open directory " + dialogDir)
	}
	openDir(dir, dialogDir, true)
	dir.Close()
	fileTreeStore()
}

func saveCb() {
	if !fileIsSaved(curFile) {
		saveAsCb()
		return
	}
	inotifyRmWatch(curFile)
	defer inotifyAddWatch(curFile)
	file, _ := os.OpenFile(curFile, os.O_CREATE|os.O_WRONLY, 0644)
	if nil == file {
		bumpMessage("Unable to open file for writing: " + curFile)
		return
	}
	fileSaveCurrent()
	rec, _ := fileMap[curFile]
	nbytes, err := file.WriteString(string(rec.buf))
	if nbytes != len(rec.buf) {
		bumpMessage("Error while writing to file: " + curFile)
		println("nbytes = ", nbytes, " errno = ", err)
		return
	}
	file.Truncate(int64(nbytes))
	file.Close()

	sourceBuf.SetModified(false)
	refreshTitle()
}

func saveAsCb() {
	dialogOk, dialogFile := fileChooserDialog(SAVE_DIALOG)
	if false == dialogOk {
		return
	}
	var be, en gtk.TextIter
	sourceBuf.GetStartIter(&be)
	sourceBuf.GetEndIter(&en)
	textToSave := sourceBuf.GetText(&be, &en, true)
	addFileRecord(dialogFile, []byte(textToSave), true)
	fileTreeStore()
	fileToDelete := curFile
	fileSwitchTo(dialogFile)
	deleteFileRecord(fileToDelete)
	fileTreeStore()
	saveCb()
	treeViewSetCurIter(true)
}

func exitCb() {
	// Are-you-sure-you-want-to-exit-because-file-is-unsaved logic will be here.
	sessionSave()
	if nil != listener {
		listener.Close()
	}
	gtk.MainQuit()
}

func closeCb() {
	if "" == curFile {
		return
	}
	closeIt := !sourceBuf.GetModified()
	if !closeIt {
		closeIt = bumpQuestion("This file has been modified. Close it?")
	}
	if closeIt {
		deleteFileRecord(curFile)
		curFile = fileStackPop()
		if 0 == len(fileMap) {
			newCb()
		}
		if "" == curFile {
			// Choose random open file then. Previous if implies that there are some 
			// opened files. At least unsaved.
			for curFile, _ = range fileMap {
				break
			}
		}
		fileSwitchTo(curFile)
		fileTreeStore()
	}
}

func pasteDoneCb() {
	var be, en gtk.TextIter
	sourceBuf.GetStartIter(&be)
	sourceBuf.GetEndIter(&en)
	sourceBuf.RemoveTagByName("instance", &be, &en)
	selectionFlag = false
}

func openFileReadToBuf(name string, verbose bool) (bool, []byte) {
	file, _ := os.OpenFile(name, os.O_RDONLY, 0644)
	if nil == file {
		bumpMessage("Unable to open file for reading: " + name)
		return false, nil
	}
	defer file.Close()
	stat, _ := file.Stat()
	if nil == stat {
		bumpMessage("Unable to stat file: " + name)
		return false, nil
	}
	buf := make([]byte, stat.Size())
	nread, _ := file.Read(buf)
	if nread != int(stat.Size()) {
		bumpMessage("Unable to read whole file: " + name)
		return false, nil
	}
	if nread > 0 {
		if false == glib.Utf8Validate(buf, nread, nil) {
			if verbose {
				bumpMessage("File " + name + " is not correct utf8 text")
			}
			return false, nil
		}
	}
	return true, buf
}

func fileChooserDialog(t int) (bool, string) {
	var action gtk.FileChooserAction
	var okStock string
	if OPEN_DIALOG == t {
		action = gtk.FILE_CHOOSER_ACTION_OPEN
		okStock = gtk.STOCK_OPEN
	} else if SAVE_DIALOG == t {
		action = gtk.FILE_CHOOSER_ACTION_SAVE
		okStock = gtk.STOCK_SAVE
	} else if OPEN_DIR_DIALOG == t {
		action = gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER
		okStock = gtk.STOCK_OPEN
	}
	dialog := gtk.NewFileChooserDialog("", sourceView.GetTopLevelAsWindow(),
		action,
		gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL,
		okStock, gtk.RESPONSE_ACCEPT)
	dialog.SetCurrentFolder(prevDir)
	res := dialog.Run()
	dialogFolder := dialog.GetCurrentFolder()
	dialogFile := dialog.GetFilename()
	dialog.Destroy()
	if gtk.RESPONSE_ACCEPT == res {
		prevDir = dialogFolder
		return true, dialogFile
	}
	return false, ""
}

func errorChkCb(current bool) {
	errorWindow.SetVisible(current)
	opt.showError = current
}

func searchChkCb(current bool) {
	searchView.window.SetVisible(current)
	opt.showSearch = current
}

func noTabChkCb(current bool) {
	opt.spaceNotTab = current
	sourceView.SetInsertSpacesInsteadOfTabs(opt.spaceNotTab)
}

func GofmtCb() {
	gofmt(curFile)
}

func fontCb() {
	dialog := gtk.NewFontSelectionDialog("Pick a font")
	dialog.SetFontName(opt.font)
	if gtk.RESPONSE_OK == dialog.Run() {
		opt.font = dialog.GetFontName()
		sourceView.ModifyFontEasy(opt.font)
	}
	dialog.Destroy()
}