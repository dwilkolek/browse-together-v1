package db

import (
	"errors"
	"github.com/google/uuid"
	"slices"
	"sync"
)

type InMemoryStore struct {
	sessions     []Session
	rejoinTokens map[string]int64
	lock         sync.Mutex
}

func (s *InMemoryStore) StoreRejoinToken(memberId int64) string {
	s.lockMe()
	defer s.releaseMe()
	token := uuid.New().String()
	s.rejoinTokens[token] = memberId
	return token
}
func (s *InMemoryStore) GetMemberIdForRejoinToken(token string) (int64, error) {
	s.lockMe()
	defer s.releaseMe()
	memberId, ok := s.rejoinTokens[token]
	if ok {
		return memberId, nil
	} else {
		return 0, errors.New("no such token")
	}
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
