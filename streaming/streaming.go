package streaming

import (
	"log"
	"sync"
	"time"

	"github.com/dwilkolek/browse-together-api/dto"
	"github.com/dwilkolek/browse-together-api/queue"
	"github.com/gofiber/contrib/websocket"
)

type SessionState struct {
	sessionId    string
	clients      map[int64]*websocket.Conn
	states       map[int64]dto.PositionStateDTO
	updateNeeded bool
	lock         sync.Mutex
	queue        queue.EventQueue
	active       bool
}

var mu = sync.Mutex{}
var state map[string]*SessionState = make(map[string]*SessionState)

func (state *SessionState) lockMe(reason string) {
	log.Printf("Locking %s: %s\n", reason, state.sessionId)
	state.lock.Lock()
}
func (state *SessionState) unlockMe(reason string) {
	log.Printf("Unlock %s: %s\n", reason, state.sessionId)
	state.lock.Unlock()
}
func (state *SessionState) addMember(conn *websocket.Conn) int64 {
	state.lockMe("addClient")
	defer state.unlockMe("addClient")
	log.Printf("New client. In total %d clients\n", len(state.clients))
	memberId := state.queue.NextMemberId()
	state.clients[memberId] = conn
	return memberId
}

func (state *SessionState) removeMember(memberId int64) {
	state.lockMe("deleteClient")
	defer state.unlockMe("deleteClient")
	log.Printf("Removing client. In total %d clients\n", len(state.clients))
	delete(state.states, memberId)
	delete(state.clients, memberId)
	state.updateNeeded = true
}
func (state *SessionState) newPosition(newPosition dto.PositionStateDTO) {
	state.lockMe("newPosition")
	defer state.unlockMe("newPosition")
	state.states[newPosition.MemberId] = newPosition
	state.updateNeeded = true
}

func (state *SessionState) SessionMemberPositionChange(newPosition dto.PositionStateDTO) {
	state.queue.SessionMemberPositionChange(newPosition)
}
func (state *SessionState) OnSessionClosed() <-chan struct{} {
	return state.queue.OnSessionClosed()
}
func (state *SessionState) MemberLeft(memberId int64) {
	state.lockMe("MemberLeft")
	defer state.unlockMe("MemberLeft")
	if _, ok := state.states[memberId]; ok {
		state.removeMember(memberId)
		state.queue.MemberLeft(memberId)
	}
}

func JoinSession(sessionId string, conn *websocket.Conn) (int64, *SessionState) {

	log.Printf("Starting position listening %s\n", sessionId)
	mu.Lock()
	defer mu.Unlock()
	if state[sessionId] == nil {
		queue := queue.GetEventQueueForSession(sessionId)
		state[sessionId] = &SessionState{
			sessionId:    sessionId,
			clients:      map[int64]*websocket.Conn{},
			states:       queue.GetSnapshot(),
			updateNeeded: true,
			queue:        queue,
		}

		go onMemberLeave(state[sessionId], queue)
		go onPositionUpdate(state[sessionId], queue)
		go notifyClientsLoop(state[sessionId], queue)
		go listenForSessionClose(queue, sessionId)

	}

	memberId := state[sessionId].addMember(conn)

	return memberId, state[sessionId]

}

func onMemberLeave(sessionState *SessionState, queue queue.EventQueue) {
	memberLeftChan := queue.OnMemberLeave()
	for {
		memberId := <-memberLeftChan
		log.Println("onMemberLeave")
		sessionState.removeMember(memberId)
	}
}

func CloseSession(sessionId string) {

	log.Printf("Closing session %s\n", sessionId)

	mu.Lock()
	defer mu.Unlock()
	if state[sessionId] != nil {
		state[sessionId].queue.CloseSession()
	}

}

func onPositionUpdate(sessionState *SessionState, queue queue.EventQueue) {
	eventChan := queue.OnPositionChanged()
	// Start listening for messages
	for {
		positionState, ok := <-eventChan
		if !ok {
			break
		}

		sessionState.newPosition(positionState)
	}

}

func notifyClientsLoop(sessionState *SessionState, queue queue.EventQueue) {
	ticker := time.NewTicker(1 * time.Millisecond)
	done := queue.OnSessionClosed()
	go func() {
		for {
			select {
			case <-ticker.C:
				notifyClients(sessionState)
			case <-done:
				log.Println("stop " + sessionState.sessionId)
				ticker.Stop()
				return
			}
		}
	}()

}

func notifyClients(sessionState *SessionState) {
	if !sessionState.updateNeeded {
		return
	}
	sessionState.lockMe("notifyClients")
	defer func() {
		sessionState.updateNeeded = false
		sessionState.unlockMe("notifyClients")
	}()
	var err error
	for connClientId, conn := range sessionState.clients {
		toSend := []dto.PositionStateDTO{}
		for _, ps := range sessionState.states {
			if ps.Selector != "" {
				toSend = append(toSend, ps)
			}
		}

		if err = conn.WriteJSON(toSend); err != nil {
			log.Println("write:", err)

			defer func(connClientId int64) {
				delete(sessionState.states, connClientId)
				delete(sessionState.clients, connClientId)
			}(connClientId)
		}
	}
}

func listenForSessionClose(queue queue.EventQueue, sessionId string) {
	<-queue.OnSessionClosed()

	mu.Lock()
	defer mu.Unlock()
	delete(state, sessionId)

}
