package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// IDCookieManager is a session ID manager that uses cookie to send or receive
// a session ID with a HTTP client.
type IDCookieManager struct {
	idByteLen    int
	idEncoder    func(data []byte) string
	sessionIDKey string
	path         string
	domain       string
	maxAge       int
	secure       bool
	httpOnly     bool
}

// NewIDCookieManager creates a new IDCookieManager.
func NewIDCookieManager(options ...IDCookieManagerOption) (*IDCookieManager, error) {
	m := &IDCookieManager{
		idByteLen: 16,
		idEncoder: base64.RawURLEncoding.EncodeToString,
	}
	for _, opt := range options {
		err := opt(m)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return m, nil
}

// IDCookieManagerOption is the type for option of IDCookieManager.
type IDCookieManagerOption func(m *IDCookieManager) error

// SetIDByteLen sets the session ID byte length.
func SetIDByteLen(idByteLen int) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.idByteLen = idByteLen
		return nil
	}
}

// SetIDEncoder sets the session ID encoder.
func SetIDEncoder(idEncoder func(data []byte) string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.idEncoder = idEncoder
		return nil
	}
}

// SetSessionIDKey sets the session ID key in a cookie.
func SetSessionIDKey(sessionIDKey string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.sessionIDKey = sessionIDKey
		return nil
	}
}

// SetPath sets the cookie path.
func SetPath(path string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.path = path
		return nil
	}
}

// SetDomain sets the cookie domain.
func SetDomain(domain string) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.domain = domain
		return nil
	}
}

// SetMaxAge sets the Max-Age in the cookie.
// The precision of the duration is one second and the sub-second part will be ignored.
func SetMaxAge(duration time.Duration) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.maxAge = int(duration / time.Second)
		return nil
	}
}

// SetSecure sets the Secure in the cookie.
func SetSecure(secure bool) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.secure = secure
		return nil
	}
}

// SetHTTPOnly sets the HttpOnly in the cookie.
func SetHTTPOnly(httpOnly bool) IDCookieManagerOption {
	return func(m *IDCookieManager) error {
		m.httpOnly = httpOnly
		return nil
	}
}

// Get gets the session ID from the request.
// To check if the session ID was not specified in the request,
// use errors.Cause(err) == session.ErrNotFound
// where errors is github.com/pkg/errors.
func (m *IDCookieManager) Get(r *http.Request) (string, error) {
	c, err := r.Cookie(m.sessionIDKey)
	if err == http.ErrNoCookie {
		return "", errors.WithStack(ErrNotFound)
	} else if err != nil {
		return "", errors.WithStack(err)
	}

	return c.Value, nil
}

// Issue issues a new session ID.
func (m *IDCookieManager) Issue() (string, error) {
	buf := make([]byte, m.idByteLen)
	_, err := rand.Read(buf)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return m.idEncoder(buf), nil
}

// Write writes a session ID to the request response.
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

// Delete adds a response header which tells the HTTP client to delete the session ID..
func (m *IDCookieManager) Delete(w http.ResponseWriter) error {
	c := &http.Cookie{
		Name:   m.sessionIDKey,
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	return nil
}
