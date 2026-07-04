package users

import (
	"github.com/go-minstack/go-minstack/core"
	user_repositories "github.com/ricardoalcantara/min-idp/internal/users/repositories"
)

func Register(app *core.App) {
	app.Use(core.ProvideAs[UserRepository](user_repositories.NewUserRepository))
	app.Provide(NewUserService)
	app.Provide(NewUserController)
	app.Provide(NewMeController)
}
