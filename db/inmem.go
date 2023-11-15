package db

import (
	"errors"
	"slices"
	"sync"
)

type InMemoryStore struct {
	sessions []Session
	lock     sync.Mutex
}

func (s *InMemoryStore) StoreSession(session Session) error {
	s.lockMe()
	defer s.releaseMe()
	s.sessions = append(s.sessions, session)
	return nil
}

func (s *InMemoryStore) GetSessions() []Session {
	s.lockMe()
	defer s.releaseMe()
	return s.sessions
}

func (s *InMemoryStore) GetSession(id string) (Session, error) {
	s.lockMe()
	defer s.releaseMe()
	for _, session := range s.sessions {
		if session.Id == id {
			return session, nil
		}
	}

	return Session{}, errors.New("session not found")
}

func (s *InMemoryStore) DeleteSession(id string) error {
	s.lockMe()
	defer s.releaseMe()
	s.sessions = slices.DeleteFunc(s.sessions, func(s Session) bool {
		return s.Id == id
	})

	return nil
}

func (s *InMemoryStore) lockMe() {
	s.lock.Lock()
}

func (s *InMemoryStore) releaseMe() {
	s.lock.Unlock()
}
