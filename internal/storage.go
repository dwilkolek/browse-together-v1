package internal

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

var sessionsKey = "sessions"
var sessionPrefix = "session-"
var lockPrefix = "lock-"
var sessionClientIdPrefix = "sessionClientIdCounter-"

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

	startPositionStreaming(session.Id)
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

	stopPositionStreaming(id)
	return nil
}

func GetNewClientId(sessionId string) int64 {
	newValue, err := client.Incr(context.Background(), sessionClientIdPrefix+sessionId).Result()
	if err != nil {
		panic(err)
	}
	return newValue
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
