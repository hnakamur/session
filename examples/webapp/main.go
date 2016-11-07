package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/hnakamur/session"
)

var sessionIDManager session.IDManager
var sessionStore *session.RedisStore

type mySession struct {
	Counter int `json:"counter"`
}

const sessionIDKey = "SessionID"

func viewHandler(w http.ResponseWriter, r *http.Request) {
	sessID, err := sessionIDManager.GetOrIssue(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess := mySession{}
	err = sessionStore.Load(context.Background(), sessID, &sess)
	log.Printf("after Load. sess=%+v, err=%+v", sess, err)
	if err != nil && err != session.ErrNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess.Counter++

	err = sessionStore.Save(context.Background(), sessID, sess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hello, counter=%d\n", sess.Counter)
}

func main() {
	var err error
	sessionIDManager, err = session.NewIDCookieManager(
		session.SetSessionIDKey(sessionIDKey))
	sessionStore, err = session.NewRedisStore(":6379",
		session.SetRedisPoolMaxIdle(2),
		session.SetAutoExpire(time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	defer sessionStore.Close()

	http.HandleFunc("/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
