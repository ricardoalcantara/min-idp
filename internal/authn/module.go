package authn

import "github.com/go-minstack/core"

func Register(app *core.App) {
	app.Provide(NewAuthnService)
	app.Provide(NewAuthnController)
	app.Invoke(RegisterRoutes)
}
