package sp

import (
	"github.com/go-minstack/core"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(fx.Annotate(
		sp_repositories.NewSPRepository,
		fx.As(new(SPRepository)),
	))
	app.Provide(NewSPService)
	app.Provide(NewSPController)
}
