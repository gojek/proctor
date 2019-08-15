package redis

import (
	"github.com/garyburd/redigo/redis"
	"proctor/internal/app/service/infra/config"
	"time"
)

type Client interface {
	GET(string) ([]byte, error)
	SET(string, []byte) error
	KEYS(string) ([]string, error)
	MGET(...interface{}) ([][]byte, error)
}

type redisClient struct {
	connPool *redis.Pool
}

func NewClient() Client {
	connPool, err := newPool()
	if err != nil {
		panic(err.Error())
	}

	return &redisClient{connPool}
}

func newPool() (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     config.Config().RedisMaxActiveConnections / 2,
		MaxActive:   config.Config().RedisMaxActiveConnections,
		IdleTimeout: 5 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", config.Config().RedisAddress) },
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
		Wait: true,
	}

	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	return pool, err
}

func (c *redisClient) GET(key string) ([]byte, error) {
	conn := c.connPool.Get()
	defer conn.Close()

	return redis.Bytes(conn.Do("GET", key))
}

func (c *redisClient) SET(key string, value []byte) error {
	conn := c.connPool.Get()
	defer conn.Close()

	return conn.Send("SET", key, value)
}

func (c *redisClient) KEYS(regex string) ([]string, error) {
	conn := c.connPool.Get()
	defer conn.Close()

	return redis.Strings(conn.Do("KEYS", regex))
}

func (c *redisClient) MGET(keys ...interface{}) ([][]byte, error) {
	conn := c.connPool.Get()
	defer conn.Close()

	return redis.ByteSlices(conn.Do("MGET", keys...))
}
