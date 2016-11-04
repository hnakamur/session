package main

import (
	"context"
	"log"
	"time"

	"bitbucket.org/hnakamur/session"
	"github.com/garyburd/redigo/redis"
)

func main() {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	store, err := session.NewRedisStore(c)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	err = store.Set(ctx, "1234", "foo", "bar2")
	if err != nil {
		log.Fatal(err)
	}

	err = store.Expire(ctx, "1234", time.Second)
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
