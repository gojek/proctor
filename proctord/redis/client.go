package redis

import (
	"github.com/garyburd/redigo/redis"
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
