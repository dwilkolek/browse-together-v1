package queue

import "github.com/dwilkolek/browse-together-api/dto"

type EventQueue interface {
	GetSnapshot() map[int64]dto.PositionStateDTO
	SessionMemberPositionChange(update dto.PositionStateDTO)
	MemberLeft(memberId int64)
	CloseSession()
	NextMemberId() int64

	OnPositionChanged() <-chan dto.PositionStateDTO
	OnSessionClosed() <-chan struct{}
	OnMemberLeave() <-chan int64
}

func GetEventQueueForSession(sessionId string) EventQueue {
	return &InMemoryEventQueue{
		onPositionChangedChan: make(chan dto.PositionStateDTO),
		sessionClosedChan:     make(chan struct{}),
		memberCount:           0,
		cache:                 make(map[int64]dto.PositionStateDTO),
		sessionId:             sessionId,
		memberLeaveChan:       make(chan int64),
	}
}
