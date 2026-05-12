package keystore

import (
	"github.com/go-minstack/core"
	keystore_repositories "github.com/ricardoalcantara/min-idp/internal/keystore/repositories"
)

func Register(app *core.App) {
	app.Provide(keystore_repositories.NewKeyRepository)
	app.Use(core.ProvideAs[KeyStore](NewKeyStoreService))
	app.Provide(NewKeyStoreController)
}
