package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

// RedisStore is a session data store which uses Redis as a backend.
// It uses a connection pool and you can use one RedisStore concurrently
// from multiple goroutines at the same time.
type RedisStore struct {
	pool        *redis.Pool
	expiration  time.Duration
	formatIDKey func(id string) string
	encodeValue func(value interface{}) ([]byte, error)
	decodeValue func(data []byte, valuePtr interface{}) error
}

// NewRedisStore creates a RedisStore.
func NewRedisStore(address string, options ...RedisStoreOption) (*RedisStore, error) {
	c := defaultRedisStoreConfig()
	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, errors.WithStack(err)
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

// RedisStoreOption is the option type for NewRedisStore.
type RedisStoreOption func(c *redisStoreConfig) error

// SetRedisPassword sets the password for the redis server.
func SetRedisPassword(password string) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.password = password
		return nil
	}
}

// SetRedisPoolMaxIdle sets the max idle worker count in the redis connection pool.
func SetRedisPoolMaxIdle(maxIdle int) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolMaxIdle = maxIdle
		return nil
	}
}

// SetRedisPoolMaxActive sets the max active worker count in the redis connection pool.
func SetRedisPoolMaxActive(maxActive int) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolMaxActive = maxActive
		return nil
	}
}

// SetRedisPoolIdleTimeout sets the max idle timeout for the redis connection pool.
func SetRedisPoolIdleTimeout(idleTimeout time.Duration) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolIdleTimeout = idleTimeout
		return nil
	}
}

// SetRedisBorrowPoolTestDuration sets the test duration for the redis connection pool.
func SetRedisBorrowPoolTestDuration(duration time.Duration) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.poolBorrowTestDuration = duration
		return nil
	}
}

// SetExpiration sets the expiration duration for session data.
// The precision is one second and the sub-second part is ignored.
func SetExpiration(expiration time.Duration) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.expiration = expiration
		return nil
	}
}

// SetFormatIDKey sets a function to format the session key for a session ID.
func SetFormatIDKey(formatIDKey func(id string) string) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.formatIDKey = formatIDKey
		return nil
	}
}

// SetEncodeValue sets a function to encode a session data.
func SetEncodeValue(encodeValue func(value interface{}) ([]byte, error)) RedisStoreOption {
	return func(c *redisStoreConfig) error {
		c.encodeValue = encodeValue
		return nil
	}
}

// SetDecodeValue sets a function to decode a session data.
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
				return nil, errors.WithStack(err)
			}
			if c.password != "" {
				if _, err := conn.Do("AUTH", c.password); err != nil {
					conn.Close()
					return nil, errors.WithStack(err)
				}
			}
			return conn, nil
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < c.poolBorrowTestDuration {
				return nil
			}
			_, err := conn.Do("PING")
			if err != nil {
				return errors.WithStack(err)
			}
		},
	}
}

// Load loads a session data from the store.
// The first argument ctx is not used and ignored in the current implementation.
func (s *RedisStore) Load(ctx context.Context, id string, valuePtr interface{}) error {
	conn := s.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("GET", s.formatIDKey(id))
	if err != nil {
		return errors.WithStack(err)
	}
	if reply == nil {
		return errors.WithStack(ErrNotFound)
	}
	err = s.decodeValue(reply.([]byte), valuePtr)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Save saves a session data to the store.
// The first argument ctx is not used and ignored in the current implementation.
func (s *RedisStore) Save(ctx context.Context, id string, value interface{}) error {
	conn := s.pool.Get()
	defer conn.Close()

	v, err := s.encodeValue(value)
	if err != nil {
		return errors.WithStack(err)
	}
	seconds := int64(s.expiration / time.Second)
	_, err = conn.Do("SETEX", s.formatIDKey(id), seconds, v)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Delete deletes a session data from the store.
// The first argument ctx is not used and ignored in the current implementation.
func (s *RedisStore) Delete(ctx context.Context, id string) error {
	conn := s.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", s.formatIDKey(id))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Close shutdowns the connection pool to the redis server.
func (s *RedisStore) Close() error {
	return s.pool.Close()
}
