package main

import (
	"github.com/mattn/go-gtk/gtksourceview"
)

var lang_map map[string]*gtksourceview.SourceLanguage

var prev_lang string = "default"

func lang_refresh() {
	ext := lang_get_extension(cur_file)
	lang, found := lang_map[ext]
	if !found {
		lang = lang_map["default"]
		ext = "default"
	}
	if ext == prev_lang {
		return
	}
	prev_lang = ext
	source_buf.SetLanguage(lang)
}

func lang_get_extension(name string) string {
	for y := len(name) - 1; y >= 0; y-- {
		if ('.' == name[y]) || ('/' == name[y]) {
			return name[y+1:]
		}
	}
	return ""
}

func init_lang() {
	lang_man := gtksourceview.SourceLanguageManagerGetDefault()
	lang_map = make(map[string]*gtksourceview.SourceLanguage)
	lang_map["go"] = lang_man.GetLanguage("go")
	lang_map["c"] = lang_man.GetLanguage("c")
	lang_map["h"] = lang_man.GetLanguage("cpp")
	lang_map["sh"] = lang_man.GetLanguage("sh")
	lang_map["diff"] = lang_man.GetLanguage("diff")
	lang_map["patch"] = lang_man.GetLanguage("diff")
	lang_map["cpp"] = lang_man.GetLanguage("cpp")
	lang_map["hpp"] = lang_man.GetLanguage("cpp")
	lang_map["cc"] = lang_man.GetLanguage("cpp")
	lang_map["latex"] = lang_man.GetLanguage("latex")
	lang_map["tex"] = lang_man.GetLanguage("latex")
	lang_map["Makefile"] = lang_man.GetLanguage("makefile")
	lang_map["am"] = lang_man.GetLanguage("makefile")
	lang_map["in"] = lang_man.GetLanguage("makefile")
	lang_map["xml"] = lang_man.GetLanguage("xml")
	lang_map["py"] = lang_man.GetLanguage("python")
	lang_map["scons"] = lang_man.GetLanguage("python")
	lang_map["SConstruct"] = lang_man.GetLanguage("python")
	lang_map["java"] = lang_man.GetLanguage("java")
	
	lang_map["default"] = lang_man.GetLanguage("sh")
	
	source_buf.SetLanguage(lang_map["default"])
}
