package redis

import (
	"time"

	"github.com/gojektech/proctor/proctord/config"

	"github.com/garyburd/redigo/redis"
)

func newPool() (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     config.RedisMaxActiveConnections() / 2,
		MaxActive:   config.RedisMaxActiveConnections(),
		IdleTimeout: 5 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", config.RedisAddress()) },
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
