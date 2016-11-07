package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/hnakamur/session"
)

var sessionManager *session.Manager

type mySession struct {
	Counter int `json:"counter"`
}

const sessionIDKey = "SessionID"

func viewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sess := mySession{}
	sessID, _, err := sessionManager.LoadOrNew(ctx, w, r, &sess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess.Counter++

	err = sessionManager.Save(ctx, w, r, sessID, sess)
	if err != nil {
		log.Printf("after sessionManager.Save. err=%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hello, counter=%d\n", sess.Counter)
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	sess := mySession{}
	sessID, _, err := sessionManager.LoadOrNew(ctx, w, r, &sess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = sessionManager.Delete(ctx, w, r, sessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Deleted sesion")
}

func main() {
	sessionMaxAge := time.Minute
	sessionIDManager, err := session.NewIDCookieManager(
		session.SetSessionIDKey(sessionIDKey),
		session.SetMaxAge(sessionMaxAge))
	if err != nil {
		log.Fatal(err)
	}
	sessionStore, err := session.NewRedisStore(":6379",
		session.SetRedisPoolMaxIdle(2),
		session.SetExpiration(sessionMaxAge))
	if err != nil {
		log.Fatal(err)
	}
	defer sessionStore.Close()
	sessionManager = session.NewManager(sessionIDManager, sessionStore)

	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/sign-out/", signOutHandler)
	http.ListenAndServe(":8080", nil)
}
