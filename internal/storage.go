package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

var sessionsKey = "sessions"
var sessionPrefix = "session-"
var clientPrefix = "client-"
var lockPrefix = "lock-"

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping(context.TODO()).Result()
	fmt.Println(pong, err)
}

func StoreSession(session Session) error {
	lockSession(session.Id)
	defer releaseSession(session.Id)
	jsonStr, _ := json.Marshal(session)

	if _, err := client.SAdd(context.Background(), sessionsKey, session.Id).Result(); err != nil {
		log.Printf("Err! %s\n", err)
		return err
	}

	if _, err := client.Set(context.Background(), sessionPrefix+session.Id, jsonStr, 0).Result(); err != nil {
		log.Printf("Err! %s\n", err)
		return err
	}
	log.Printf("Stored in redis %s\n", session.Id)
	return nil
}

func GetSessions() []Session {
	var sessions []Session = make([]Session, 0)
	value, err := client.SMembers(context.Background(), sessionsKey).Result()
	if err != nil {
		log.Printf("Err! %s\n", err)
		return sessions
	}
	for _, sessionId := range value {
		if session, err := GetSession(sessionId); err == nil {
			sessions = append(sessions, session)
		}

	}

	return sessions
}

func GetSession(id string) (Session, error) {
	var session Session
	value, err := client.Get(context.Background(), sessionPrefix+id).Result()
	if err != nil {
		log.Printf("Err! %s\n", err)
		return session, err
	}
	json.Unmarshal([]byte(value), &session)

	return session, nil
}

func DeleteSession(id string) error {
	lockSession(id)
	defer releaseSession(id)
	if _, err := client.SRem(context.TODO(), sessionsKey, id).Result(); err != nil {
		log.Printf("Failed to remove from sessions %s\n", id)
		return err
	}
	if _, err := client.Del(context.TODO(), sessionPrefix+id).Result(); err != nil {
		log.Printf("Failed to remove session %s\n", id)
		return err
	}
	return nil
}

func lockSession(id string) {
	log.Printf("Locking %s\n", id)
	for {
		res, _ := client.SetNX(context.Background(), lockPrefix+id, true, time.Minute).Result()
		if res {
			return
		}
	}

}

func releaseSession(id string) {
	log.Printf("Releasing lock: %s\n", id)
	client.Del(context.Background(), lockPrefix+id)
}
