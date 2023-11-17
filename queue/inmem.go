package queue

import (
	"sync"

	"github.com/dwilkolek/browse-together-api/dto"
)

type InMemoryEventQueue struct {
	sessionId         string
	sessionClosedChan chan struct{}
	memberCount       int64
	cache             map[int64]dto.PositionStateDTO
	mu                sync.Mutex
	outdated          bool
	closed            bool
}

func (q *InMemoryEventQueue) Initialise() {
	//noop
}
func (q *InMemoryEventQueue) RefreshNeeded() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.outdated && !q.closed
}
func (q *InMemoryEventQueue) GetSnapshot() map[int64]dto.PositionStateDTO {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.outdated = false
	q.cache = validPositionStates(q.cache)
	return q.cache
}
func (q *InMemoryEventQueue) SessionMemberPositionChange(update dto.PositionStateDTO) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.cache, update.MemberId)
	if q.closed {
		return
	}

	q.outdated = true
	q.cache[update.MemberId] = update
}
func (q *InMemoryEventQueue) MemberLeft(memberId int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	delete(q.cache, memberId)
	q.outdated = true
}
func (q *InMemoryEventQueue) CloseSession() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	close(q.sessionClosedChan)
}
func (q *InMemoryEventQueue) NextMemberId() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.memberCount += 1
	return q.memberCount
}

func (q *InMemoryEventQueue) OnSessionClosed() <-chan struct{} {
	return q.sessionClosedChan
}
