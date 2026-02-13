package handlers

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*
var embeddedFS embed.FS

func LoadTemplates() *template.Template {
	sub, err := fs.Sub(embeddedFS, "templates")
	if err != nil {
		panic(err)
	}
	tmpl := template.Must(template.ParseFS(sub, "*.html"))
	return tmpl

}
