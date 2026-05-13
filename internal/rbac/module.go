package rbac

import (
	"github.com/go-minstack/core"
	rbac_repositories "github.com/ricardoalcantara/min-idp/internal/rbac/repositories"
)

func Register(app *core.App) {
	app.Use(core.ProvideAs[RBACRepository](rbac_repositories.NewRBACRepository))
	app.Provide(NewRBACService)
	app.Provide(NewRBACController)
}
