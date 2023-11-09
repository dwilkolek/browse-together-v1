package internal

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func init() {
	client = getRedis()
}

func getRedis() *redis.Client {
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
