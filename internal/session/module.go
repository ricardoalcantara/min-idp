package session

import (
	"github.com/go-minstack/core"
	session_repositories "github.com/ricardoalcantara/min-idp/internal/session/repositories"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(fx.Annotate(
		session_repositories.NewSessionRepository,
		fx.As(new(SessionRepository)),
	))
	app.Provide(NewSessionService)
}
