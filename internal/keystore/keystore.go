package keystore

import (
	"context"
	"crypto"

	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
)

type KeyStore interface {
	ActivePrivateKey(ctx context.Context, protocol string) (crypto.PrivateKey, *keystore_entities.SigningKey, error)
	PublicKeys(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error)
	InsertKey(ctx context.Context, key *keystore_entities.SigningKey) error
	RotateKey(ctx context.Context, protocol string, newKey *keystore_entities.SigningKey) error
	Reload(ctx context.Context, protocol string) error
}
