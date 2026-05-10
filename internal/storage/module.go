package storage

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("storage",
		fx.Provide(NewDB),
	)
}
