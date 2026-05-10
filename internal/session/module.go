package session

import (
	"github.com/go-minstack/core"
	session_repositories "github.com/ricardoalcantara/min-idp/internal/session/repositories"
)

func Register(app *core.App) {
	app.Provide(session_repositories.NewSessionRepository)
	app.Provide(NewSessionService)
}
