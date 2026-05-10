package bootstrap

import (
	"context"

	bootstrap_repositories "github.com/ricardoalcantara/min-idp/internal/bootstrap/repositories"
	"github.com/go-minstack/core"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(bootstrap_repositories.NewBootstrapRepository)
	app.Provide(NewBootstrapService)
	app.Invoke(scheduleRun)
}

func scheduleRun(svc *BootstrapService, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return svc.Run(ctx)
		},
	})
}
