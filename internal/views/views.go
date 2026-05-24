package views

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/gin"
)

type Template struct {
	*template.Template
}

func (t *Template) Render(ctx *gin.Context, data any) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(ctx.Writer, data)
}

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

func parse(name string) *Template {
	return &Template{template.Must(template.New(name).ParseFS(files, name))}
}
