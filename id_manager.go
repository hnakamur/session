package session

import "net/http"

// IDManager manages the session ID
type IDManager interface {
	// Get the session ID from the request
	Get(r *http.Request) (string, error)

	// Issue a new session ID
	Issue() (string, error)

	// Write the session ID to the response
	Write(w http.ResponseWriter, sessID string) error

	// Delete tells the HTTP client to remove the session ID
	Delete(w http.ResponseWriter) error
}
