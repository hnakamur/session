package session

import "net/http"

type IDManager interface {
	GetOrIssue(w http.ResponseWriter, r *http.Request) (string, error)
}