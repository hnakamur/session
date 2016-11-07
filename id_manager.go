package session

import "net/http"

type IDManager interface {
	Get(r *http.Request) (string, error)
	Issue() (string, error)
	Write(w http.ResponseWriter, sessID string) error
	Delete(w http.ResponseWriter) error
}
