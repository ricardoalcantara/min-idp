package kvstore

import (
	"github.com/go-minstack/go-minstack/core"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("kvstore",
		core.ProvideAs[KVStore](NewDBKVStore),
	)
}
