package saml

import "github.com/go-minstack/core"

func Register(app *core.App) {
	app.Provide(NewSAMLService)
	app.Provide(NewSAMLController)
	app.Invoke(RegisterRoutes)
}
