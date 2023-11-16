package clients

import (
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisLock sync.Mutex = sync.Mutex{}

func CreateRedisClient() *redis.Client {
	if redisClient == nil {
		redisLock.Lock()
		defer redisLock.Unlock()
		if redisClient == nil {
			addr := os.Getenv("REDIS_URL")
			if addr == "" {
				addr = "redis://default:@localhost:6379/0"

			}
			opt, err := redis.ParseURL(addr)
			fmt.Printf("%v\n", opt)
			if err != nil {
				panic(err)
			}

			redisClient = redis.NewClient(opt)
		}
	}
	return redisClient
}
