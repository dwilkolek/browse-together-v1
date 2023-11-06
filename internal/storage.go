package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// client
//
//	func init() {
//		client := redis.NewClient(&redis.Options{
//			Addr:     "localhost:6379",
//			Password: "", // no password set
//			DB:       0,  // use default DB
//		})
//	}
var client *redis.Client

var sessionsKey = "sessions"

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping(context.TODO()).Result()
	fmt.Println(pong, err)
}

func StoreSession(session Session) {
	lockSession(session.Id)
	defer releaseSession(session.Id)
	jsonStr, _ := json.Marshal(session)
	_, err := client.SAdd(context.Background(), sessionsKey, jsonStr).Result()

	if err != nil {
		log.Printf("Err! %s\n", err)
	}

	log.Printf("Stored in redis %s\n", session.Id)
}

func GetSessions() []Session {
	var sessions []Session
	value, err := client.SMembers(context.Background(), sessionsKey).Result()
	if err != nil {
		log.Printf("Err! %s\n", err)
		return sessions
	}
	fmt.Println(value)
	for _, sessionStr := range value {
		var session Session
		json.Unmarshal([]byte(sessionStr), &session)
		sessions = append(sessions, session)
	}

	return sessions
}

// func AddClientToSession(clientId string, sessionId string){
// 	client.SAdd(context.TODO(),"clients-"+sessionId, )
// }

func DeleteSession(id string) {
	client.SRem(context.TODO(), "sessions", id)
	client.Del(context.TODO(), "clients-"+id, id)
}

func lockSession(id string) {
	for {
		res, _ := client.SetNX(context.Background(), "lock-"+id, true, time.Minute).Result()
		if res {
			return
		}
	}

}

func releaseSession(id string) {
	client.Del(context.Background(), "lock-"+id)
}
