package views

import (
	"embed"
	"html/template"
)

//go:embed *.html
var files embed.FS

var (
	LandingTmpl        = parse("landing.html")
	LoginTmpl          = parse("login.html")
	InfoTmpl           = parse("info.html")
	LogoutTmpl         = parse("logout.html")
	ForgotPasswordTmpl = parse("forgot_password.html")
	ResetPasswordTmpl  = parse("reset_password.html")
)

func parse(name string) *template.Template {
	return template.Must(template.New(name).ParseFS(files, name))
}
