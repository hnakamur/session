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

func (m *Manager) LoadOrNew(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionDataPtr interface{}) (sessID string, isNew bool, err error) {
	sessID, err = m.idManager.GetOrIssue(w, r)
	if err != nil {
		return "", false, err
	}

	err = m.store.Load(ctx, sessID, sessionDataPtr)
	if err != nil && err != ErrNotFound {
		return "", false, err
	}

	return sessID, err == ErrNotFound, nil
}

func (m *Manager) Save(ctx context.Context, w http.ResponseWriter, r *http.Request, sessID string, sessionData interface{}) error {
	return m.store.Save(ctx, sessID, sessionData)
}
