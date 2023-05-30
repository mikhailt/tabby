package main

import (
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
	"strings"
	"strconv"
)

var global bool
var globalMap map[string]int
var insertSet bool
var prevGlobal bool
var scopeEn gtk.TextIter
var fnrBeMark *gtk.TextMark
var fnrEnMark *gtk.TextMark

func fnrCb() {
	fnrDialog()
}

func fnrDialog() {
	var fnrCnt int
	var scopeBe gtk.TextIter
	if MAX_SEL_LEN < len(sourceSelection()) {
		sourceBuf.GetSelectionBounds(&scopeBe, &scopeEn)
	} else {
		sourceBuf.GetStartIter(&scopeBe)
		sourceBuf.GetEndIter(&scopeEn)
	}
	fnrBeMark = sourceBuf.CreateMark("fnr_be", &scopeBe, true)
	fnrEnMark = sourceBuf.CreateMark("fnr_en", &scopeEn, false)

	dialog := gtk.NewDialog()
	dialog.SetTitle("Find and Replace")
	dialog.AddButton("_Find Next", gtk.RESPONSE_OK)
	dialog.AddButton("_Replace", gtk.RESPONSE_YES)
	dialog.AddButton("Replace _All", gtk.RESPONSE_APPLY)
	dialog.AddButton("_Close", gtk.RESPONSE_CLOSE)

	entry := findEntryWithHistory()
	replacement := findEntryWithHistory()

	globalButton := gtk.NewCheckButtonWithLabel("Global")
	globalButton.SetVisible(true)
	globalButton.SetActive(prevGlobal)

	vbox := dialog.GetVBox()
	vbox.Add(entry)
	vbox.Add(replacement)
	vbox.Add(globalButton)

	findNextButton := dialog.GetWidgetForResponse(int(gtk.RESPONSE_OK))
	replaceButton := dialog.GetWidgetForResponse(int(gtk.RESPONSE_YES))
	replaceAllButton := dialog.GetWidgetForResponse(int(gtk.RESPONSE_APPLY))
	closeButton := dialog.GetWidgetForResponse(int(gtk.RESPONSE_CLOSE))

	findNextButton.Connect("clicked", func() {
		fnrPreCb(globalButton, &insertSet)
		if !fnrFindNext(entry.GetActiveText(), prevGlobal, &mapFilled, &globalMap) {
			fnrCloseAndReport(dialog, fnrCnt)
		}
	}, nil)
	findNextButton.AddAccelerator("clicked", accelGroup, gdk.KEY_Return, 0, gtk.ACCEL_VISIBLE)

	replaceButton.Connect("clicked", func() {
		fnrPreCb(globalButton, &insertSet)
		done, nextFound := fnrReplace(entry.GetActiveText(), replacement.GetActiveText(), prevGlobal, &mapFilled, &globalMap)
		fnrCnt += done
		if !nextFound {
			fnrCloseAndReport(dialog, fnrCnt)
		}
	}, nil)

	replaceAllButton.Connect("clicked", func() {
		insertSet = false
		fnrPreCb(globalButton, &insertSet)
		fnrCnt += fnrReplaceAllLocal(entry.GetActiveText(), replacement.GetActiveText())
		if prevGlobal {
			fnrCnt += fnrReplaceAllGlobal(entry.GetActiveText(), replacement.GetActiveText())
			fileTreeStore()
		}
		fnrCloseAndReport(dialog, fnrCnt)
	}, nil)

	closeButton.Connect("clicked", func() { dialog.Destroy() }, nil)

	dialog.Run()
}

func fnrReplaceAllLocal(entry string, replacement string) int {
	cnt := 0
	var t bool = true
	if !fnrFindNext(entry, false, &t, nil) {
		return 0
	}
	for {
		done, nextFound := fnrReplace(entry, replacement, false, &t, nil)
		cnt += done
		if !nextFound {
			break
		}
	}
	return cnt
}

func fnrReplaceAllGlobal(entry string, replacement string) int {
	totalCnt := 0
	lent := len(entry)
	lrep := len(replacement)
	inds := make(map[int]int)
	for file, rec := range fileMap {
		if file == curFile {
			continue
		}
		cnt := 0
		scope := rec.buf[:]
		for {
			pos := strings.Index(string(scope), entry)
			if -1 == pos {
				break
			}
			inds[cnt] = pos
			cnt++
			scope = scope[pos+lent:]
		}
		if 0 == cnt {
			continue
		}
		buf := make([]byte, len(rec.buf)+cnt*(lrep-lent))
		scope = rec.buf[:]
		destScope := buf[:]
		for y := 0; y < cnt; y++ {
			shift := inds[y]
			copy(destScope, scope[:shift])
			destScope = destScope[shift:]
			copy(destScope, replacement)
			destScope = destScope[lrep:]
			scope = scope[shift+lent:]
		}
		copy(destScope, scope)
		rec.buf = buf
		rec.modified = true
		totalCnt += cnt
	}
	return totalCnt
}

func fnrPreCb(globalButton *gtk.CheckButton, insertSet *bool) {
	prevGlobal = globalButton.GetActive()
	fnrRefreshScope(prevGlobal)
	fnrSetInsert(insertSet)
}

func fnrCloseAndReport(dialog *gtk.Dialog, fnrCnt int) {
	dialog.Destroy()
	bumpMessage(strconv.Itoa(fnrCnt) + " replacements were done.")
}

func fnrSetInsert(insertSet *bool) {
	if false == *insertSet {
		*insertSet = true
		var scopeBe gtk.TextIter
		getIterAtMarkByName("fnr_be", &scopeBe)
		sourceBuf.MoveMarkByName("insert", &scopeBe)
		sourceBuf.MoveMarkByName("selection_bound", &scopeBe)
	}
}

func fnrRefreshScope(global bool) {
	var be gtk.TextIter
	if global {
		sourceBuf.GetStartIter(&be)
		sourceBuf.GetEndIter(&scopeEn)
		sourceBuf.CreateMark("fnr_be", &be, true)
		sourceBuf.CreateMark("fnr_en", &scopeEn, false)
	}
}

func fnrFindNext(pattern string, global bool, mapFilled *bool, m *map[string]int) bool {
	var be, en gtk.TextIter
	getIterAtMarkByName("fnr_en", &scopeEn)
	getIterAtMarkByName("selection_bound", &en)
	if en.ForwardSearch(pattern, 0, &be, &en, &scopeEn) {
		moveFocusAndSelection(&be, &en)
		return true
	}
	// Have to switch to next file or to beginning of current depending on <global>.
	if global {
		// Switch to next file.
		fnrFindNextFillGlobalMap(pattern, m, mapFilled)
		nextFile := popStringFromMap(m)
		if "" == nextFile {
			return false
		}
		fileSaveCurrent()
		fileSwitchTo(nextFile)
		fnrRefreshScope(true)
		sourceBuf.GetStartIter(&be)
		sourceBuf.MoveMarkByName("insert", &be)
		sourceBuf.MoveMarkByName("selection_bound", &be)
		return fnrFindNext(pattern, global, mapFilled, m)
	} else {
		// Temporary fix. Is there necessity to search the document all over again?
		return false
		// Start search from beginning of scope.
		// getIterAtMarkByName("fnr_be", &be)
		// if be.ForwardSearch(pattern, 0, &be, &en, &scopeEn) {
		//	moveFocusAndSelection(&be, &en)
		//	return true
		//} else {
		//	return false
		//}
	}
	return false
}

func fnrFindNextFillGlobalMap(pattern string, m *map[string]int, mapFilled *bool) {
	if *mapFilled {
		return
	}
	*mapFilled = true
	*m = make(map[string]int)
	for file, rec := range fileMap {
		if curFile == file {
			continue
		}
		if -1 != strings.Index(string(rec.buf), pattern) {
			(*m)[file] = 1
		}
	}
}

// Returns (done, nextFound)
func fnrReplace(entry string, replacement string, global bool, mapFilled *bool, globalMap *map[string]int) (int, bool) {
	if entry != sourceSelection() {
		return 0, true
	}
	sourceBuf.DeleteSelection(false, true)
	sourceBuf.InsertAtCursor(replacement)
	var be, en gtk.TextIter
	sourceBuf.GetSelectionBounds(&be, &en)
	sourceBuf.MoveMarkByName("insert", &en)
	return 1, fnrFindNext(entry, global, mapFilled, globalMap)
}

func popStringFromMap(m *map[string]int) string {
	if 0 == len(*m) {
		return ""
	}
	for s, _ := range *m {
		delete(*m, s)
		return s
	}
	return ""
}

func getIterAtMarkByName(markName string, iter *gtk.TextIter) {
	mark := sourceBuf.GetMark(markName)
	sourceBuf.GetIterAtMark(iter, mark)
}