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

func NewRedisStore(address string, options ...RedisStoreOption) (Store, error) {
	c := defaultRedisStoreConfig()
	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}

	return &redisStore{
		pool: newRedisPool(address, c),
	}, nil
}

type redisStoreConfig struct {
	password               string
	poolMaxIdle            int
	poolMaxActive          int
	poolIdleTimeout        time.Duration
	poolBorrowTestDuration time.Duration
}

func defaultRedisStoreConfig() *redisStoreConfig {
	return &redisStoreConfig{
		poolMaxIdle:            3,
		poolIdleTimeout:        240 * time.Second,
		poolBorrowTestDuration: time.Minute,
	}
}

type RedisStoreOption func(c *redisStoreConfig) error

func SetRedisPassword(password string) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.password = password
		return nil
	}
}

func SetRedisPoolMaxIdle(maxIdle int) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolMaxIdle = maxIdle
		return nil
	}
}

func SetRedisPoolMaxActive(maxActive int) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolMaxActive = maxActive
		return nil
	}
}

func SetRedisPoolIdleTimeout(idleTimeout time.Duration) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolIdleTimeout = idleTimeout
		return nil
	}
}

func SetRedisBorrowPoolTestDuration(duration time.Duration) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolBorrowTestDuration = duration
		return nil
	}
}

func newRedisPool(address string, c *redisStoreConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.poolMaxIdle,
		MaxActive:   c.poolMaxActive,
		IdleTimeout: c.poolIdleTimeout,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			if c.password != "" {
				if _, err := conn.Do("AUTH", c.password); err != nil {
					conn.Close()
					return nil, err
				}
			}
			return conn, err
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < c.poolBorrowTestDuration {
				return nil
			}
			_, err := conn.Do("PING")
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
