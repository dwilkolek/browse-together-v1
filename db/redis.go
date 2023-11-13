package db

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client = getRedisClient()

func getRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = "redis://default:@localhost:6379/0"

	}
	opt, err := redis.ParseURL(addr)
	fmt.Printf("%v\n", opt)
	if err != nil {
		panic(err)
	}

	return redis.NewClient(opt)
}
