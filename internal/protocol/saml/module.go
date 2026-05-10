package saml

import "github.com/go-minstack/core"

func Register(app *core.App) {
	app.Provide(NewSAMLController)
	app.Invoke(RegisterRoutes)
}
