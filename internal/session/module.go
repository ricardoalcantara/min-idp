package session

import (
	"github.com/go-minstack/go-minstack/core"
	session_repositories "github.com/ricardoalcantara/min-idp/internal/session/repositories"
)

func Register(app *core.App) {
	app.Use(core.ProvideAs[SessionRepository](session_repositories.NewSessionRepository))
	app.Provide(NewSessionService)
	app.Provide(NewCookieTokenService)
}
