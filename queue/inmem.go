package queue

import (
	"fmt"
	"log"
	"sync"

	"github.com/dwilkolek/browse-together-api/dto"
)

type InMemoryEventQueue struct {
	sessionId             string
	onPositionChangedChan chan dto.PositionStateDTO
	sessionClosedChan     chan struct{}
	memberLeaveChan       chan int64
	memberCount           int64
	cache                 map[int64]dto.PositionStateDTO
	mu                    sync.Mutex
}

func (q *InMemoryEventQueue) GetSnapshot() map[int64]dto.PositionStateDTO {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.cache
}
func (q *InMemoryEventQueue) SessionMemberPositionChange(update dto.PositionStateDTO) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.cache, update.MemberId)
	q.cache[update.MemberId] = update
	q.onPositionChangedChan <- update
}
func (q *InMemoryEventQueue) MemberLeft(memberId int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	count := len(q.cache)
	delete(q.cache, memberId)
	if count-len(q.cache) == 0 {
		return
	}
	fmt.Printf("<-- %d leaft %d\n", count-len(q.cache), memberId)
	q.memberLeaveChan <- memberId
}
func (q *InMemoryEventQueue) CloseSession() {
	q.mu.Lock()
	defer q.mu.Unlock()
	log.Printf("Closing session INMEM %s\n", q.sessionId)
	log.Printf("Cache size: %d", len(q.cache))
	close(q.memberLeaveChan)
	log.Println("memberLeaveChan")
	close(q.onPositionChangedChan)
	log.Println("onPositionChangedChan")
	close(q.sessionClosedChan)
	log.Println("sessionClosedChan")
}
func (q *InMemoryEventQueue) NextMemberId() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.memberCount += 1
	log.Printf("New member. In total %d\n", q.memberCount)
	return q.memberCount
}
func (q *InMemoryEventQueue) OnPositionChanged() <-chan dto.PositionStateDTO {
	return q.onPositionChangedChan
}
func (q *InMemoryEventQueue) OnSessionClosed() <-chan struct{} {
	return q.sessionClosedChan
}
func (q *InMemoryEventQueue) OnMemberLeave() <-chan int64 {
	return q.memberLeaveChan
}
