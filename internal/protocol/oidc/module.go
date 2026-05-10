package oidc

import "github.com/go-minstack/core"

func Register(app *core.App) {
	app.Provide(NewOIDCController)
	app.Invoke(RegisterRoutes)
}
