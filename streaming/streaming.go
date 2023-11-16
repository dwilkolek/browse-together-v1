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
	queue.EventQueue
	sessionId string
	clients   map[int64]*websocket.Conn
	lock      sync.Mutex
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
	memberId := state.NextMemberId()
	state.clients[memberId] = conn
	return memberId
}

func JoinSession(sessionId string, conn *websocket.Conn) (int64, *SessionState) {

	log.Printf("Starting position listening %s\n", sessionId)
	mu.Lock()
	defer mu.Unlock()
	if state[sessionId] == nil {
		queue := queue.GetEventQueueForSession(sessionId)
		session := &SessionState{
			EventQueue: queue,
			sessionId:  sessionId,
			clients:    map[int64]*websocket.Conn{},
		}
		session.Initalize()

		go notifyClientsLoop(session, queue)
		go listenForSessionClose(queue, sessionId)
		state[sessionId] = session
	}

	memberId := state[sessionId].addMember(conn)

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
				log.Println("stop " + sessionState.sessionId)
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
	for connClientId, conn := range sessionState.clients {
		toSend := []dto.PositionStateDTO{}
		snapshot := sessionState.GetSnapshot()
		for _, ps := range snapshot {
			if ps.Selector != "" {
				toSend = append(toSend, ps)
			}
		}

		if err = conn.WriteJSON(toSend); err != nil {
			log.Println("write:", err)

			defer func(connClientId int64) {
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
