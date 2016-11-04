package main

import (
	"context"
	"log"
	"time"

	"bitbucket.org/hnakamur/session"
)

func main() {
	store, err := session.NewRedisStore(":6379",
		session.SetRedisPoolMaxIdle(2),
		session.SetAutoExpire(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	err = store.Set(ctx, "1234", "foo", "bar2")
	if err != nil {
		log.Fatal(err)
	}

	var v interface{}
	err = store.Get(ctx, "1234", "foo", &v)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("v=%+v", v)

	time.Sleep(2 * time.Second)
	err = store.Get(ctx, "1234", "foo", &v)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("v=%+v", v)
}
