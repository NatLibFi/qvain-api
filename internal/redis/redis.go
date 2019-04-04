// Package redis wraps the redigo redis package.
package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisPool struct {
	*redis.Pool
}

func NewRedisPool(net, address string) *RedisPool {
	/*
		func newPool(addr string) *redis.Pool {
			return &redis.Pool{
				MaxIdle: 3,
				IdleTimeout: 10 * time.Second,
				//Dial: func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
				Dial: func() (redis.Conn, error) { return redis.Dial(net, address) },
			}
		}
	*/

	return &RedisPool{
		&redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 10 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.Dial(net, address) },
		},
	}
}

//func DialConnectTimeout(d time.Duration) DialOption
/*
func Dial() (redis.Conn, error) {
	c, err := redis.Dial("tcp", ":6379", redis.DialConnectTimeout(5*time.Second))
	if err != nil {
		// handle error
	}
	defer c.Close()
}
*/
