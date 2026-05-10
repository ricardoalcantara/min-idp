package kvstore

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("kvstore",
		fx.Provide(
			fx.Annotate(NewDBKVStore, fx.As(new(KVStore))),
		),
	)
}
