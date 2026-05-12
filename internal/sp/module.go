package sp

import (
	"github.com/go-minstack/core"
	rbac_repositories "github.com/ricardoalcantara/min-idp/internal/rbac/repositories"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
)

func Register(app *core.App) {
	app.Use(core.ProvideAs[SPRepository](sp_repositories.NewSPRepository))
	app.Use(core.ProvideAs[RBACGateRepository](rbac_repositories.NewRBACRepository))
	app.Provide(NewSPService)
	app.Provide(NewSPGateService)
	app.Provide(NewSPController)
}
