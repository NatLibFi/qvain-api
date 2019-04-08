package main

import (
	"fmt"

	"github.com/CSCfi/qvain-api/internal/redis"
	redigo "github.com/gomodule/redigo/redis" // real redis package, for helper functions
)

const REDIS_NETWORK = "unix"
const REDIS_ADDRESS = "/home/wouter/.redis.sock"

func main() {
	fmt.Println("redis tester")

	pool := redis.NewRedisPool(REDIS_NETWORK, REDIS_ADDRESS)
	conn := pool.Get()
	defer conn.Close()

	// old redis: SETEX key seconds value
	//conn.Do("SETEX", "testkey", 10, "hello world")
	// new redis: SET key value [EX seconds] [PX milliseconds] [NX|XX]
	conn.Do("SET", "testkey", "hello world", "EX", 10)
	/*
		s, err := redis.String(conn.Do("GET", "testkey"))
		if err != nil {
			panic(err)
		}
	*/
	res, err := conn.Do("GET", "testkey")
	if err != nil {
		panic(err)
	}
	fmt.Println("key:", string(res.([]byte)))
	res, err = redigo.String(res, err)
	fmt.Println("helper:", res)

	res, err = conn.Do("TTL", "testkey")
	if err != nil {
		panic(err)
	}
	fmt.Printf("ttl: %d (%T)\n", res, res)
}
