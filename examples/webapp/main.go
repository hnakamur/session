package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

const sessionIDKey = "SessionID"

func viewHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionIDKey)
	if err == nil {
		log.Printf("cookie=%+v", c)
	} else if err == http.ErrNoCookie {
		sid, err := issueSessionID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c := &http.Cookie{
			Name:  sessionIDKey,
			Value: sid,
		}
		http.SetCookie(w, c)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Hello")
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}

func issueSessionID() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
