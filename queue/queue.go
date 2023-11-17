package queue

import (
	"os"
	"sync"
	"time"

	"github.com/dwilkolek/browse-together-api/clients"
	"github.com/dwilkolek/browse-together-api/dto"
)

type EventQueue interface {
	Initialise()
	GetSnapshot() map[int64]dto.PositionStateDTO
	SessionMemberPositionChange(update dto.PositionStateDTO)
	MemberLeft(memberId int64)
	CloseSession()
	NextMemberId() int64

	OnSessionClosed() <-chan struct{}
	RefreshNeeded() bool
}

func GetEventQueueForSession(sessionId string) EventQueue {
	queue := os.Getenv("QUEUE")
	if queue == "" {
		queue = "IN_MEMORY"
	}

	if queue == "IN_MEMORY" {
		return &InMemoryEventQueue{
			sessionClosedChan: make(chan struct{}),
			memberCount:       0,
			cache:             make(map[int64]dto.PositionStateDTO),
			sessionId:         sessionId,
			closed:            false,
		}
	}

	if queue == "REDIS" {
		return &RedisEventQueue{
			sessionId:         sessionId,
			redisClient:       clients.CreateRedisClient(),
			sessionClosedChan: make(chan struct{}),
			cache:             make(map[int64]dto.PositionStateDTO),
			mu:                sync.Mutex{},
			outdated:          false,
			closed:            false,
		}
	}

	panic("Unknown QUEUE config")
}

func validPositionStates(states map[int64]dto.PositionStateDTO) map[int64]dto.PositionStateDTO {
	var filtered = make(map[int64]dto.PositionStateDTO)
	for memberId, state := range states {
		inactiveForMinute := time.Now().UnixMilli() > state.UpdatedAt+1*time.Minute.Milliseconds()
		validLocation := state.Location != ""
		if !inactiveForMinute && validLocation {
			filtered[memberId] = state
		}
	}
	return filtered
}
