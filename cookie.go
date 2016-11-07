package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

type IDCookieManager struct {
	sessionIDKey string
	path         string
	domain       string
	maxAge       int
	secure       bool
	httpOnly     bool
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

func SetPath(path string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.path = path
		return nil
	}
}

func SetDomain(domain string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.domain = domain
		return nil
	}
}

func SetMaxAge(duration time.Duration) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.maxAge = int(duration / time.Second)
		return nil
	}
}

func SetSecure(secure bool) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.secure = secure
		return nil
	}
}

func SetHTTPOnly(httpOnly bool) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.httpOnly = httpOnly
		return nil
	}
}

func (m *IDCookieManager) Get(r *http.Request) (string, error) {
	c, err := r.Cookie(m.sessionIDKey)
	if err == http.ErrNoCookie {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}

	return c.Value, nil
}

func (m *IDCookieManager) Issue() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (m *IDCookieManager) Write(w http.ResponseWriter, sessID string) error {
	c := &http.Cookie{
		Name:     m.sessionIDKey,
		Value:    sessID,
		Path:     m.path,
		Domain:   m.domain,
		MaxAge:   m.maxAge,
		Secure:   m.secure,
		HttpOnly: m.httpOnly,
	}
	http.SetCookie(w, c)
	return nil
}

func (m *IDCookieManager) Delete(w http.ResponseWriter) error {
	c := &http.Cookie{
		Name:   m.sessionIDKey,
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	return nil
}
