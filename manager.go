package session

import (
	"context"
	"net/http"
)

type Manager struct {
	idManager IDManager
	store     Store
}

func NewManager(idManager IDManager, store Store) *Manager {
	return &Manager{
		idManager: idManager,
		store:     store,
	}
}

func (m *Manager) LoadOrNew(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionDataPtr interface{}) (sessID string, found bool, err error) {
	sessID, err = m.idManager.Get(r)
	if err == ErrNotFound {
		sessID, err = m.idManager.Issue()
		return sessID, false, err
	} else if err != nil {
		return "", false, err
	}

	err = m.store.Load(ctx, sessID, sessionDataPtr)
	if err == ErrNotFound {
		// NOTE: A malicious user may have set arbitrary session ID, so issue a new session ID here.
		sessID, err = m.idManager.Issue()
		return sessID, false, err
	} else if err != nil {
		return sessID, false, err
	}
	return sessID, true, nil
}

func (m *Manager) Save(ctx context.Context, w http.ResponseWriter, r *http.Request, sessID string, sessionData interface{}) error {
	err := m.store.Save(ctx, sessID, sessionData)
	if err != nil {
		return err
	}
	return m.idManager.Write(w, sessID)
}

func (m *Manager) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, sessID string) error {
	err := m.idManager.Delete(w)
	if err != nil {
		return err
	}
	return m.store.Delete(ctx, sessID)
}
