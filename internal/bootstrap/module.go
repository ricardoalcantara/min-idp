package bootstrap

import (
	"context"

	"github.com/go-minstack/go-minstack/core"
	bootstrap_repositories "github.com/ricardoalcantara/min-idp/internal/bootstrap/repositories"
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
