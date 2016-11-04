package session

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Get(ctx context.Context, id, key string, valuePtr interface{}) error
	Set(ctx context.Context, id, key string, value interface{}) error
	Remove(ctx context.Context, id, key string) error
	RemoveAll(ctx context.Context, id string) error
	Expire(ctx context.Context, id string, d time.Duration) error

	Close() error
}
