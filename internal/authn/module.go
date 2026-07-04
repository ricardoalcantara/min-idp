package authn

import (
	"github.com/go-minstack/go-minstack/core"
	"github.com/ricardoalcantara/min-idp/internal/users"
)

func newUserAuthenticator(u *users.UserService) UserAuthenticator {
	return u
}

func Register(app *core.App) {
	app.Provide(newUserAuthenticator)
	app.Provide(NewAuthnService)
	app.Provide(NewAuthnController)
	app.Invoke(RegisterRoutes)
}
