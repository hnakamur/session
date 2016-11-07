package session

import (
	"context"
	"errors"
)

// ErrNotFound is the error returned when a session ID or a session data is not found.
var ErrNotFound = errors.New("not found")

// Store is an interface for loading or saving a session data to the store.
type Store interface {
	// Load loads a session data from the store.
	Load(ctx context.Context, id string, valuePtr interface{}) error

	// Save saves a session data to the store.
	Save(ctx context.Context, id string, value interface{}) error

	// Delete deletes a session data from the store.
	Delete(ctx context.Context, id string) error
}
