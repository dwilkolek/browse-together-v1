package db

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"strconv"
	"time"

	"github.com/dwilkolek/browse-together-api/clients"
	"github.com/redis/go-redis/v9"
)

const sessionPrefix = "session-"
const lockPrefix = "lock-"
const rejoinPrefix = "rejoin-"

type RedisStore struct {
	*redis.Client
}

func CreateRedisStore() RedisStore {
	return RedisStore{
		clients.CreateRedisClient(),
	}
}

func (s *RedisStore) StoreRejoinToken(memberId int64) string {
	token := uuid.New().String()
	s.Client.Set(context.Background(), rejoinPrefix+token, memberId, time.Hour).Result()
	return token
}
func (s *RedisStore) GetMemberIdForRejoinToken(token string) (int64, error) {
	result, err := s.Client.Get(context.Background(), rejoinPrefix+token).Result()
	if err != nil {
		return 0, err
	}
	parseInt, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, err
	}
	return parseInt, nil
}

func (s *RedisStore) StoreSession(session Session) error {
	s.lockSession(session.Id)
	defer s.releaseSession(session.Id)
	jsonStr, _ := json.Marshal(session)

	if _, err := s.Set(context.Background(), sessionPrefix+session.Id, jsonStr, 8*time.Hour).Result(); err != nil {
		log.Printf("Failed storing session: %s\n", err)
		return err
	}

	log.Printf("Stored in redis %s\n", session.Id)
	return nil
}

func (s *RedisStore) GetSessions() []Session {
	var keys []string
	var cursor uint64
	for {

		var err error
		keys, cursor, err = s.Client.Scan(context.Background(), cursor, sessionPrefix+"*", 0).Result()
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			keys = append(keys, key[len(sessionPrefix):])
		}

		if cursor == 0 { // no more keys
			break
		}
	}

	var sessions = make([]Session, 0)

	for _, sessionId := range keys {
		if session, err := s.GetSession(sessionId); err == nil {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

func (s *RedisStore) GetSession(id string) (Session, error) {
	var session Session
	value, err := s.Get(context.Background(), sessionPrefix+id).Result()
	if err != nil {
		log.Printf("No such session %s: %s\n", id, err)
		return session, err
	}
	err = json.Unmarshal([]byte(value), &session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func (s *RedisStore) DeleteSession(id string) error {
	s.lockSession(id)
	defer s.releaseSession(id)
	if _, err := s.Del(context.TODO(), sessionPrefix+id).Result(); err != nil {
		log.Printf("Failed to remove session %s\n", id)
		return err
	}

	return nil
}

func (s *RedisStore) lockSession(id string) {
	log.Printf("Locking %s\n", id)
	for {
		res, _ := s.SetNX(context.Background(), lockPrefix+id, true, time.Minute).Result()
		if res {
			return
		}
	}

}

func (s *RedisStore) releaseSession(id string) {
	log.Printf("Releasing lock: %s\n", id)
	s.Del(context.Background(), lockPrefix+id)
}
