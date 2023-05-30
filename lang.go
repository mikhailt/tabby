package main

import (
	"github.com/mattn/go-gtk/gtksourceview"
)

var (
	lang_map           map[string]*gtksourceview.SourceLanguage
	prev_lang          string = "default"
	default_language   *gtksourceview.SourceLanguage
	source_language_mgr                      = gtksourceview.SourceLanguageManagerGetDefault()
)

func lang_refresh() {
	ext := lang_get_extension(cur_file)
	lang, found := lang_map[ext]
	if !found {
		lang = default_language
		ext = "default"
	}
	if ext != prev_lang {
		source_buf.SetLanguage(lang)
		prev_lang = ext
	}
}

func lang_get_extension(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' || name[i] == '/' {
			return name[i+1:]
		}
	}
	return ""
}

func init_lang() {
	default_language = source_language_mgr.GetLanguage("sh")
	lang_map = map[string]*gtksourceview.SourceLanguage{
		"go":       source_language_mgr.GetLanguage("go"),
		"c":        source_language_mgr.GetLanguage("c"),
		"cpp":      source_language_mgr.GetLanguage("cpp"),
		"h":        source_language_mgr.GetLanguage("cpp"),
		"sh":       source_language_mgr.GetLanguage("sh"),
		"diff":     source_language_mgr.GetLanguage("diff"),
		"patch":    source_language_mgr.GetLanguage("diff"),
		"latex":    source_language_mgr.GetLanguage("latex"),
		"tex":      source_language_mgr.GetLanguage("latex"),
		"Makefile": source_language_mgr.GetLanguage("makefile"),
		"am":       source_language_mgr.GetLanguage("makefile"),
		"in":       source_language_mgr.GetLanguage("makefile"),
		"xml":      source_language_mgr.GetLanguage("xml"),
		"py":       source_language_mgr.GetLanguage("python"),
		"scons":    source_language_mgr.GetLanguage("python"),
		"SConstruct": source_language_mgr.GetLanguage("python"),
		"java":     source_language_mgr.GetLanguage("java"),
		"default":  default_language,
	}
	source_buf.SetLanguage(default_language)
}