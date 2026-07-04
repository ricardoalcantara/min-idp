package oidc

import (
	"github.com/go-minstack/go-minstack/core"
	oidc_repositories "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/repositories"
)

func Register(app *core.App) {
	app.Use(core.ProvideAs[OAuthTokenRepository](oidc_repositories.NewOAuthTokenRepository))
	app.Provide(NewOIDCService)
	app.Provide(NewOIDCController)
	app.Invoke(RegisterRoutes)
}
