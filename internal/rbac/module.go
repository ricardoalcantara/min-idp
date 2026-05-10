package rbac

import (
	"github.com/go-minstack/core"
	rbac_repositories "github.com/ricardoalcantara/min-idp/internal/rbac/repositories"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(fx.Annotate(
		rbac_repositories.NewRBACRepository,
		fx.As(new(RBACRepository)),
	))
	app.Provide(rbac_repositories.NewGroupRepository)
	app.Provide(NewRBACService)
	app.Provide(NewRBACController)
	app.Provide(NewGroupController)
	app.Provide(NewGroupService)
}
