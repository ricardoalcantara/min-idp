package keystore

import (
	"context"
	"crypto"

	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
)

type KeyStore interface {
	ActivePrivateKey(ctx context.Context, protocol string) (crypto.PrivateKey, *keystore_entities.SigningKey, error)
	PublicKeys(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error)
	ListAllKeys(ctx context.Context) ([]*keystore_entities.SigningKey, error)
	ListKeysByProtocol(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error)
	GenerateKey(ctx context.Context, protocol string) (*keystore_entities.SigningKey, error)
	InsertKey(ctx context.Context, key *keystore_entities.SigningKey) error
	RotateKey(ctx context.Context, protocol string, newKey *keystore_entities.SigningKey) error
	GenerateAndRotate(ctx context.Context, protocol string) error
	Reload(ctx context.Context, protocol string) error
}
