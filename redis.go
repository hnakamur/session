package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

type RedisStore struct {
	pool        *redis.Pool
	expiration  time.Duration
	formatIDKey func(id string) string
	encodeValue func(value interface{}) ([]byte, error)
	decodeValue func(data []byte, valuePtr interface{}) error
}

func NewRedisStore(address string, options ...RedisStoreOption) (*RedisStore, error) {
	c := defaultRedisStoreConfig()
	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}

	return &RedisStore{
		pool:        newRedisPool(address, c),
		expiration:  c.expiration,
		formatIDKey: c.formatIDKey,
		encodeValue: c.encodeValue,
		decodeValue: c.decodeValue,
	}, nil
}

type redisStoreConfig struct {
	password               string
	poolMaxIdle            int
	poolMaxActive          int
	poolIdleTimeout        time.Duration
	poolBorrowTestDuration time.Duration
	expiration             time.Duration
	formatIDKey            func(id string) string
	encodeValue            func(value interface{}) ([]byte, error)
	decodeValue            func(data []byte, valuePtr interface{}) error
}

func defaultRedisStoreConfig() *redisStoreConfig {
	return &redisStoreConfig{
		poolMaxIdle:            3,
		poolIdleTimeout:        240 * time.Second,
		poolBorrowTestDuration: time.Minute,
		expiration:             5 * time.Minute,
		formatIDKey: func(id string) string {
			return "sess:" + id
		},
		encodeValue: json.Marshal,
		decodeValue: json.Unmarshal,
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

func SetExpiration(expiration time.Duration) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.expiration = expiration
		return nil
	}
}

func SetFormatIDKey(formatIDKey func(id string) string) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.formatIDKey = formatIDKey
		return nil
	}
}

func SetEncodeValue(encodeValue func(value interface{}) ([]byte, error)) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.encodeValue = encodeValue
		return nil
	}
}

func SetDecodeValue(decodeValue func(data []byte, valuePtr interface{}) error) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.decodeValue = decodeValue
		return nil
	}
}

func newRedisPool(address string, c *redisStoreConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.poolMaxIdle,
		MaxActive:   c.poolMaxActive,
		IdleTimeout: c.poolIdleTimeout,
		Wait:        true,
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

func (s *RedisStore) Load(ctx context.Context, id string, valuePtr interface{}) error {
	conn := s.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("GET", s.formatIDKey(id))
	if err != nil {
		return err
	}
	if reply == nil {
		return ErrNotFound
	}
	return s.decodeValue(reply.([]byte), valuePtr)
}

func (s *RedisStore) Save(ctx context.Context, id string, value interface{}) error {
	conn := s.pool.Get()
	defer conn.Close()

	v, err := s.encodeValue(value)
	if err != nil {
		return err
	}
	seconds := int64(s.expiration / time.Second)
	_, err = conn.Do("SETEX", s.formatIDKey(id), seconds, v)
	if err != nil {
		return errors.WithStack(err)
	}
	return err
}

func (s *RedisStore) Delete(ctx context.Context, id string) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", s.formatIDKey(id))
	return err
}

func (s *RedisStore) Close() error {
	return s.pool.Close()
}
