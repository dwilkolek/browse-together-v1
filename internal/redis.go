package internal

import (
	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func init() {
	client = getRedis()
}

func getRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
