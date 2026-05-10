package keystore

import (
	"context"
	"crypto"
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

func (ks *KeyStoreService) InsertKey(ctx context.Context, key *keystore_entities.SigningKey) error {
	return ks.repo.Insert(ctx, key)
}

func (ks *KeyStoreService) RotateKey(ctx context.Context, protocol string, newKey *keystore_entities.SigningKey) error {
	if existing, err := ks.repo.GetActive(ctx, protocol); err == nil {
		if err := ks.repo.SetStatus(ctx, existing.ID, keystore_entities.StatusPrevious, time.Now().UTC()); err != nil {
			return err
		}
	}
	if err := ks.repo.Insert(ctx, newKey); err != nil {
		return err
	}
	return ks.Reload(ctx, protocol)
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
