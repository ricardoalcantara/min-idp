package keystore

import (
	"context"
	"crypto"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ricardoalcantara/min-idp/internal/config"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	keystore_repositories "github.com/ricardoalcantara/min-idp/internal/keystore/repositories"
	"go.uber.org/fx"
)

type loadedKey struct {
	meta       *keystore_entities.SigningKey
	privateKey crypto.PrivateKey
}

type KeyStoreService struct {
	repo      *keystore_repositories.KeyRepository
	masterKey []byte
	log       *slog.Logger

	mu       sync.RWMutex
	active   map[string]*loadedKey
	previous map[string]*loadedKey
}

func NewKeyStoreService(repo *keystore_repositories.KeyRepository, cfg *config.Config, lc fx.Lifecycle, log *slog.Logger) (KeyStore, error) {
	masterKey, err := localcrypto.DecodeMasterKey(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("keystore: %w", err)
	}

	ks := &KeyStoreService{
		repo:      repo,
		masterKey: masterKey,
		log:       log,
		active:    make(map[string]*loadedKey),
		previous:  make(map[string]*loadedKey),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, proto := range []string{keystore_entities.ProtocolOIDC, keystore_entities.ProtocolSAML} {
				if err := ks.Reload(ctx, proto); err != nil {
					return fmt.Errorf("keystore: reload %s: %w", proto, err)
				}
			}
			return nil
		},
	})

	return ks, nil
}

func (ks *KeyStoreService) decryptKey(key *keystore_entities.SigningKey) (crypto.PrivateKey, error) {
	plainPEM, err := localcrypto.Decrypt(ks.masterKey, key.PrivateKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("keystore: decrypt key %d: %w", key.ID, err)
	}
	return localcrypto.ParsePrivateKeyPEM(plainPEM)
}

func (ks *KeyStoreService) ActivePrivateKey(_ context.Context, protocol string) (crypto.PrivateKey, *keystore_entities.SigningKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	lk, ok := ks.active[protocol]
	if !ok {
		return nil, nil, fmt.Errorf("keystore: no active key for %s", protocol)
	}
	return lk.privateKey, lk.meta, nil
}

func (ks *KeyStoreService) PublicKeys(_ context.Context, protocol string) ([]*keystore_entities.SigningKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	var keys []*keystore_entities.SigningKey
	if lk, ok := ks.active[protocol]; ok {
		keys = append(keys, lk.meta)
	}
	if lk, ok := ks.previous[protocol]; ok {
		keys = append(keys, lk.meta)
	}
	return keys, nil
}

func (ks *KeyStoreService) ListAllKeys(ctx context.Context) ([]*keystore_entities.SigningKey, error) {
	return ks.repo.ListAll(ctx)
}

func (ks *KeyStoreService) ListKeysByProtocol(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error) {
	return ks.repo.ListByProtocol(ctx, protocol)
}

func (ks *KeyStoreService) InsertKey(ctx context.Context, key *keystore_entities.SigningKey) error {
	return ks.repo.Create(key)
}

func (ks *KeyStoreService) RotateKey(ctx context.Context, protocol string, newKey *keystore_entities.SigningKey) error {
	if existing, err := ks.repo.GetActive(ctx, protocol); err == nil {
		if err := ks.repo.SetStatus(ctx, existing.ID, keystore_entities.StatusPrevious, time.Now().UTC()); err != nil {
			return err
		}
	}
	if err := ks.repo.Create(newKey); err != nil {
		return err
	}
	return ks.Reload(ctx, protocol)
}

func (ks *KeyStoreService) GenerateKey(_ context.Context, protocol string) (*keystore_entities.SigningKey, error) {
	var privPEM, pubPEM, certPEM []byte
	var alg string

	switch protocol {
	case keystore_entities.ProtocolOIDC:
		key, err := localcrypto.GenerateECKey(elliptic.P256())
		if err != nil {
			return nil, fmt.Errorf("keystore: generate oidc key: %w", err)
		}
		privPEM, _ = localcrypto.MarshalPrivateKeyPEM(key)
		pubPEM, _ = localcrypto.MarshalPublicKeyPEM(key.Public())
		alg = "ES256"
	case keystore_entities.ProtocolSAML:
		key, err := localcrypto.GenerateRSAKey(2048)
		if err != nil {
			return nil, fmt.Errorf("keystore: generate saml key: %w", err)
		}
		privPEM, _ = localcrypto.MarshalPrivateKeyPEM(key)
		pubPEM, _ = localcrypto.MarshalPublicKeyPEM(key.Public())
		_, certPEM, _ = localcrypto.GenerateSelfSignedCert(key, "min-idp", 10*365*24*time.Hour)
		alg = "RS256"
	default:
		return nil, fmt.Errorf("keystore: unsupported protocol %q", protocol)
	}

	encrypted, err := localcrypto.Encrypt(ks.masterKey, privPEM)
	if err != nil {
		return nil, fmt.Errorf("keystore: encrypt key: %w", err)
	}

	now := time.Now().UTC()
	return &keystore_entities.SigningKey{
		Protocol:            protocol,
		KID:                 generateKID(),
		Algorithm:           alg,
		PrivateKeyEncrypted: encrypted,
		PublicKey:           string(pubPEM),
		Certificate:         string(certPEM),
		Status:              keystore_entities.StatusActive,
		ActivatedAt:         &now,
	}, nil
}

func (ks *KeyStoreService) GenerateAndRotate(ctx context.Context, protocol string) error {
	newKey, err := ks.GenerateKey(ctx, protocol)
	if err != nil {
		return err
	}
	return ks.RotateKey(ctx, protocol, newKey)
}

func generateKID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (ks *KeyStoreService) Reload(ctx context.Context, protocol string) error {
	keys, err := ks.repo.ListPublished(ctx, protocol)
	if err != nil {
		return err
	}
	ks.mu.Lock()
	defer ks.mu.Unlock()
	delete(ks.active, protocol)
	delete(ks.previous, protocol)
	for _, k := range keys {
		priv, err := ks.decryptKey(k)
		if err != nil {
			ks.log.Warn("keystore: skipping unloadable key", "id", k.ID, "err", err)
			continue
		}
		lk := &loadedKey{meta: k, privateKey: priv}
		switch k.Status {
		case keystore_entities.StatusActive:
			ks.active[protocol] = lk
		case keystore_entities.StatusPrevious:
			ks.previous[protocol] = lk
		}
	}
	return nil
}
