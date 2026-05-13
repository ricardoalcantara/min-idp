package views

import (
	"embed"
	"html/template"
)

//go:embed *.html
var files embed.FS

var (
	LoginTmpl  = parse("login.html")
	InfoTmpl   = parse("info.html")
	LogoutTmpl = parse("logout.html")
)

func parse(name string) *template.Template {
	return template.Must(template.New(name).ParseFS(files, name))
}
