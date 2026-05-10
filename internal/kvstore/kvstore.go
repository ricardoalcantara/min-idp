package kvstore

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("kvstore: key not found")

type KVStore interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
}
