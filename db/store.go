package db

import (
	"os"
	"sync"
)

type Db interface {
	StoreSession(session Session) error
	GetSessions() []Session
	GetSession(id string) (Session, error)
	DeleteSession(id string) error
}

type Session struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Creator      string `json:"creator"`
	BaseLocation string `json:"baseLocation"`
}

var lock sync.Mutex
var db Db

func GetDb() Db {
	if db == nil {
		lock.Lock()
		defer lock.Unlock()
		if db == nil {
			mode := os.Getenv("STORAGE")
			if mode == "" {
				mode = "IN_MEMORY"
			}

			if mode == "IN_MEMORY" {
				db = &InMemoryStore{
					sessions: []Session{},
					lock:     sync.Mutex{},
				}
			}

			if mode == "REDIS" {
				store := CreateRedisStore()
				db = &store
			}
			if db == nil {
				panic("No storage configured")
			}

		}
	}
	return db
}
