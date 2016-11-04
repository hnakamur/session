package session

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisStore struct {
	conn redis.Conn
}

func NewRedisStore(address string) (Store, error) {
	conn, err := redis.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &redisStore{
		conn: conn,
	}, nil
}

func (s *redisStore) Get(ctx context.Context, id, key string, valuePtr interface{}) error {
	reply, err := s.conn.Do("HGET", s.formatIDKey(id), key)
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
	_, err = s.conn.Do("HSET", s.formatIDKey(id), key, v)
	log.Printf("Set. id=%s, key=%s, err=%+v", id, key, err)
	return err
}

func (s *redisStore) Remove(ctx context.Context, id, key string) error {
	_, err := s.conn.Do("HDEL", s.formatIDKey(id), key)
	return err
}

func (s *redisStore) RemoveAll(ctx context.Context, id string) error {
	_, err := s.conn.Do("DEL", s.formatIDKey(id))
	return err
}

func (s *redisStore) Expire(ctx context.Context, id string, d time.Duration) error {
	_, err := s.conn.Do("PEXPIRE", s.formatIDKey(id), int64(d/time.Millisecond))
	return err
}

func (s *redisStore) Close() error {
	return s.conn.Close()
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
