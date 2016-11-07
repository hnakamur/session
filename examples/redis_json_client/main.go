package main

import (
	"context"
	"log"
	"time"

	"bitbucket.org/hnakamur/session"
)

type mySession struct {
	Foo string `json:"foo"`
}

func main() {
	store, err := session.NewRedisStore(":6379",
		session.SetRedisPoolMaxIdle(2),
		session.SetExpiration(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	session := mySession{Foo: "bar"}

	ctx := context.Background()
	err = store.Save(ctx, "1234", session)
	if err != nil {
		log.Fatal(err)
	}

	session2 := mySession{}
	err = store.Load(ctx, "1234", &session2)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("session2=%+v", session2)

	time.Sleep(2 * time.Second)
	session3 := mySession{}
	err = store.Load(ctx, "1234", &session3)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("session3=%+v", session3)
}
