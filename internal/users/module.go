package users

import (
	"github.com/go-minstack/core"
	user_repositories "github.com/ricardoalcantara/min-idp/internal/users/repositories"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(fx.Annotate(
		user_repositories.NewUserRepository,
		fx.As(new(UserRepository)),
	))
	app.Provide(NewUserService)
	app.Provide(NewUserController)
	app.Provide(NewMeController)
}
