package views

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/gin"
)

type Template[T any] struct {
	*template.Template
}

func (t *Template[T]) Render(ctx *gin.Context, data T) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(ctx.Writer, data)
}

//go:embed *.html
var files embed.FS

var (
	LandingTmpl        = parse[struct{}]("landing.html")
	LoginTmpl          = parse[LoginViewModel]("login.html")
	InfoTmpl           = parse[InfoViewModel]("info.html")
	LogoutTmpl         = parse[LogoutViewModel]("logout.html")
	ForgotPasswordTmpl = parse[ForgotPasswordViewModel]("forgot_password.html")
	ResetPasswordTmpl  = parse[ResetPasswordViewModel]("reset_password.html")
)

func parse[T any](name string) *Template[T] {
	return &Template[T]{template.Must(template.New(name).ParseFS(files, name))}
}
