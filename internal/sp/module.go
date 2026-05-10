package sp

import (
	"github.com/go-minstack/core"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
)

func Register(app *core.App) {
	app.Provide(sp_repositories.NewSPRepository)
	app.Provide(NewSPService)
	app.Provide(NewSPController)
}
