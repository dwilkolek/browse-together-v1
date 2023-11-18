package streaming

import (
	"github.com/dwilkolek/browse-together-api/config"
	"log"
	"sync"
	"time"

	"github.com/dwilkolek/browse-together-api/dto"
	"github.com/dwilkolek/browse-together-api/queue"
	"github.com/gofiber/contrib/websocket"
)

type SessionState struct {
	queue.EventQueue
	sessionId string
	members   map[int64]*websocket.Conn
	lock      sync.Mutex
}

var mu = sync.Mutex{}

var state = make(map[string]*SessionState)

func (state *SessionState) lockMe(reason string) {
	if config.DEBUG {
		log.Printf("Locking %s: %s\n", reason, state.sessionId)
	}
	state.lock.Lock()
}
func (state *SessionState) unlockMe(reason string) {
	if config.DEBUG {
		log.Printf("Unlock %s: %s\n", reason, state.sessionId)
	}
	state.lock.Unlock()
}
func (state *SessionState) addMember(conn *websocket.Conn) int64 {
	state.lockMe("addClient")
	defer state.unlockMe("addClient")
	log.Printf("New client. In total %d members\n", len(state.members))
	memberId := state.NextMemberId()
	state.members[memberId] = conn
	return memberId
}

func JoinSession(sessionId string, conn *websocket.Conn, memberId int64) (int64, *SessionState) {
	log.Printf("Starting position listening %s\n", sessionId)
	mu.Lock()
	defer mu.Unlock()
	if state[sessionId] == nil {
		queueForSession := queue.GetEventQueueForSession(sessionId)
		session := &SessionState{
			EventQueue: queueForSession,
			sessionId:  sessionId,
			members:    map[int64]*websocket.Conn{},
		}
		session.Initialise()

		go notifyClientsLoop(session, queueForSession)
		go listenForSessionClose(queueForSession, sessionId)
		state[sessionId] = session
	}

	if memberId < 1 {
		memberId = state[sessionId].addMember(conn)
	}

	return memberId, state[sessionId]

}

func CloseSession(sessionId string) {
	log.Printf("Closing session %s\n", sessionId)
	mu.Lock()
	defer mu.Unlock()
	if state[sessionId] != nil {
		state[sessionId].CloseSession()
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
				ticker.Stop()
				return
			}
		}
	}()

}

func notifyClients(sessionState *SessionState) {
	if !sessionState.RefreshNeeded() {
		return
	}
	sessionState.lockMe("notifyClients")
	defer func() {
		sessionState.unlockMe("notifyClients")
	}()
	var err error
	var unresponsiveMemberId []int64
	for memberId, conn := range sessionState.members {
		toSend := []dto.PositionStateDTO{}
		snapshot := sessionState.GetSnapshot()
		for _, ps := range snapshot {
			if ps.Selector != "" {
				toSend = append(toSend, ps)
			}
		}

		if err = conn.WriteJSON(toSend); err != nil {
			log.Printf("Member[%d] is not responsive, session %s. %s\n", memberId, sessionState.sessionId, err)
			unresponsiveMemberId = append(unresponsiveMemberId, memberId)
		}
	}
	for _, memberId := range unresponsiveMemberId {
		delete(sessionState.members, memberId)
	}
}

func listenForSessionClose(queue queue.EventQueue, sessionId string) {
	<-queue.OnSessionClosed()

	mu.Lock()
	defer mu.Unlock()
	delete(state, sessionId)

}
