package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/hnakamur/session"
)

var sessionStore *session.RedisStore

type mySession struct {
	Counter int `json:"counter"`
}

const sessionIDKey = "SessionID"

func viewHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionIDKey)
	log.Printf("c=%+v, err=%+v", c, err)
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err == http.ErrNoCookie {
		sid, err := issueSessionID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c = &http.Cookie{
			Name:  sessionIDKey,
			Value: sid,
		}
		http.SetCookie(w, c)
	}

	sessID := c.Value
	sess := mySession{}
	err = sessionStore.Load(context.Background(), sessID, &sess)
	log.Printf("after Load. sess=%+v, err=%+v", sess, err)
	if err != nil && err != session.ErrNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err == nil {
		sess.Counter++
	}

	err = sessionStore.Save(context.Background(), sessID, sess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hello, counter=%d\n", sess.Counter)
}

func main() {
	var err error
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

func issueSessionID() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
