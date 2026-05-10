package keystore

import (
	"github.com/go-minstack/core"
	keystore_repositories "github.com/ricardoalcantara/min-idp/internal/keystore/repositories"
	"go.uber.org/fx"
)

func Register(app *core.App) {
	app.Provide(keystore_repositories.NewKeyRepository)
	app.Provide(fx.Annotate(NewKeyStoreService, fx.As(new(KeyStore))))
	app.Provide(NewKeyStoreController)
}
