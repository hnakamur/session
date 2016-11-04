package session

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisStore struct {
	pool *redis.Pool
}

func NewRedisStore(address string) (Store, error) {
	return &redisStore{
		pool: newRedisPool(address, ""),
	}, nil
}

func newRedisPool(address, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (s *redisStore) Get(ctx context.Context, id, key string, valuePtr interface{}) error {
	conn := s.pool.Get()
	reply, err := conn.Do("HGET", s.formatIDKey(id), key)
	if err != nil {
		return err
	}
	if reply == nil {
		return ErrNotFound
	}
	if valuePtr == nil {
		return nil
	}
	return s.decodeValue(reply.([]byte), valuePtr)
}

func (s *redisStore) Set(ctx context.Context, id, key string, value interface{}) error {
	v, err := s.encodeValue(value)
	if err != nil {
		return err
	}
	conn := s.pool.Get()
	_, err = conn.Do("HSET", s.formatIDKey(id), key, v)
	log.Printf("Set. id=%s, key=%s, err=%+v", id, key, err)
	return err
}

func (s *redisStore) Remove(ctx context.Context, id, key string) error {
	conn := s.pool.Get()
	_, err := conn.Do("HDEL", s.formatIDKey(id), key)
	return err
}

func (s *redisStore) RemoveAll(ctx context.Context, id string) error {
	conn := s.pool.Get()
	_, err := conn.Do("DEL", s.formatIDKey(id))
	return err
}

func (s *redisStore) Expire(ctx context.Context, id string, d time.Duration) error {
	conn := s.pool.Get()
	_, err := conn.Do("PEXPIRE", s.formatIDKey(id), int64(d/time.Millisecond))
	return err
}

func (s *redisStore) Close() error {
	return s.pool.Close()
}

func (s *redisStore) formatIDKey(id string) string {
	return "sess:" + id
}

func (s *redisStore) encodeValue(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (s *redisStore) decodeValue(data []byte, valuePtr interface{}) error {
	return json.Unmarshal(data, valuePtr)
}
