package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

type IDCookieManager struct {
	sessionIDKey string
}

func NewIDCookieManager(options ...IDCookieManagerOption) (*IDCookieManager, error) {
	m := new(IDCookieManager)
	for _, opt := range options {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

type IDCookieManagerOption func(m *IDCookieManager) error

func SetSessionIDKey(sessionIDKey string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.sessionIDKey = sessionIDKey
		return nil
	}
}

func (m *IDCookieManager) GetOrIssue(w http.ResponseWriter, r *http.Request) (string, error) {
	c, err := r.Cookie(m.sessionIDKey)
	if err != nil && err != http.ErrNoCookie {
		return "", err
	}

	if err == http.ErrNoCookie {
		sid, err := m.issueSessionID()
		if err != nil {
			return "", err
		}
		c = &http.Cookie{
			Name:  m.sessionIDKey,
			Value: sid,
		}
		http.SetCookie(w, c)
	}
	return c.Value, nil
}

func (m *IDCookieManager) issueSessionID() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
