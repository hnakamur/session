package session

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// Manager is a session manager.
type Manager struct {
	idManager IDManager
	store     Store
}

// NewManager returns a new session manager.
func NewManager(idManager IDManager, store Store) *Manager {
	return &Manager{
		idManager: idManager,
		store:     store,
	}
}

// LoadOrNew get a session ID from the request and load the session data.
// If a session ID is specified in the request and the session data for that ID
// exists in the session data store, found becomes true.
//
// If a session ID is specified in the request and the session data for that ID
// does not exist, a new session ID will be issued.
// This prevents that a malicious user to use an arbitrary session ID.
//
// After calling LoadOrNew, you need to call Save or Delete once to send the
// session ID to the HTTP client and update the session data in the session data
// store.
// Note you need to call Save even when you do not modify session data,
// since you need to update the expiration time of the session ID and the session data.
func (m *Manager) LoadOrNew(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionDataPtr interface{}) (sessID string, found bool, err error) {
	sessID, err = m.idManager.Get(r)
	if errors.Cause(err) == ErrNotFound {
		sessID, err = m.idManager.Issue()
		return sessID, false, err
	} else if err != nil {
		return "", false, err
	}

	err = m.store.Load(ctx, sessID, sessionDataPtr)
	if errors.Cause(err) == ErrNotFound {
		// NOTE: A malicious user may have set an arbitrary session ID, so issue a new session ID here.
		sessID, err = m.idManager.Issue()
		return sessID, false, err
	} else if err != nil {
		return sessID, false, err
	}
	return sessID, true, nil
}

// Save write the session data in the sessoin data store and renew its expiration time.
// Is also renew the session ID expiration time and send it to the http client.
func (m *Manager) Save(ctx context.Context, w http.ResponseWriter, r *http.Request, sessID string, sessionData interface{}) error {
	err := m.store.Save(ctx, sessID, sessionData)
	if err != nil {
		return err
	}
	return m.idManager.Write(w, sessID)
}

// Delete deletes the session data in the session data store and
// tell the http client to delete the session ID.
func (m *Manager) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, sessID string) error {
	err := m.idManager.Delete(w)
	if err != nil {
		return err
	}
	return m.store.Delete(ctx, sessID)
}
