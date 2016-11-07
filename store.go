package session

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Load(ctx context.Context, id string, valuePtr interface{}) error
	Save(ctx context.Context, id string, value interface{}) error
	Delete(ctx context.Context, id string) error
	Expire(ctx context.Context, id string, d time.Duration) error

	Close() error
}
