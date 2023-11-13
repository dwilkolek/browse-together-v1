package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dwilkolek/browse-together/db"
	"github.com/dwilkolek/browse-together/dto"
	"github.com/gofiber/contrib/websocket"
)

const sessionStreamPrefix = "sessionStream-"
const internal = "internal"

type SessionState struct {
	sessionId    string
	clients      map[int64]*websocket.Conn
	states       map[int64]dto.PositionStateDTO
	updateNeeded bool
	lock         sync.Mutex
}

var state map[string]*SessionState = make(map[string]*SessionState)

func (state *SessionState) lockMe(reason string) {
	log.Printf("Locking %s: %s\n", reason, state.sessionId)
	state.lock.Lock()
}
func (state *SessionState) unlockMe(reason string) {
	log.Printf("Unlock %s: %s\n", reason, state.sessionId)
	state.lock.Unlock()
}
func (state *SessionState) addMember(clientId int64, conn *websocket.Conn) {
	state.lockMe("addClient")
	defer state.unlockMe("addClient")
	log.Printf("New client. In total %d clients\n", len(state.clients))
	state.clients[clientId] = conn
}

func (state *SessionState) deleteClient(clientId int64) {
	state.lockMe("deleteClient")
	defer state.unlockMe("deleteClient")
	log.Printf("Removing client. In total %d clients\n", len(state.clients))
	state.updateNeeded = true
	delete(state.states, clientId)
	delete(state.clients, clientId)
}
func (state *SessionState) newPosition(newPosition dto.PositionStateDTO) {
	state.lockMe("newPosition")
	defer state.unlockMe("newPosition")
	state.states[newPosition.MemberId] = newPosition
	state.updateNeeded = true
}

func stopPositionStreaming(sessionId string) {
	log.Printf("Stopping position streaming %s\n", sessionId)
	db.Client.Publish(context.Background(), sessionStreamPrefix+sessionId, "STOP").Result()
}

func startPositionStreaming(sessionId string) {
	log.Printf("Starting position streaming %s\n", sessionId)
	db.Client.Publish(context.Background(), internal, "START-"+sessionId).Result()
	db.Client.Publish(context.Background(), sessionStreamPrefix+sessionId, "START").Result()
}

func StartStreaming() {

	pubSub := db.Client.Subscribe(context.Background(), internal)
	defer pubSub.Close()
	// Channel to receive subscription messages
	subscriptionChannel := pubSub.Channel()

	// Start listening for messages
	for {
		msg, ok := <-subscriptionChannel
		if !ok {
			log.Panicf("Stopped listetning to internal channel")
			break
		}
		// fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)

		if strings.HasPrefix(msg.Payload, "START-") {
			sessionId := strings.TrimPrefix(msg.Payload, "START-")
			log.Printf("Starting position listening %s\n", sessionId)
			state[sessionId] = &SessionState{
				sessionId:    sessionId,
				clients:      make(map[int64]*websocket.Conn),
				states:       make(map[int64]dto.PositionStateDTO),
				updateNeeded: false,
			}
			go onPositionUpdate(state[sessionId])
			go notifyClientsLoop(state[sessionId])
		}

	}

}

func onPositionUpdate(sessionState *SessionState) {
	// Subscribe to the channel
	pubSub := db.Client.Subscribe(context.Background(), sessionStreamPrefix+sessionState.sessionId)
	defer pubSub.Close()

	// Channel to receive subscription messages
	subscriptionChannel := pubSub.Channel()

	// Start listening for messages
	for {
		msg, ok := <-subscriptionChannel
		if !ok {
			break
		}
		if msg.Payload == "START" {
			continue
		}
		if msg.Payload == "STOP" {
			delete(state, sessionState.sessionId)
			break
		}
		var positionState dto.PositionStateDTO
		err := json.Unmarshal([]byte(msg.Payload), &positionState)
		if err != nil {
			log.Printf("Failed to unmarshal PositionStateDTO: %s\n", err)
		}

		fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
		sessionState.newPosition(positionState)
	}
}

func (state *SessionState) UpdatePosition(update dto.PositionStateDTO) {
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("Failed to marshal PositionStateDTO: %s\n", err)
	}
	db.Client.Publish(context.Background(), sessionStreamPrefix+state.sessionId, data)
}

func notifyClientsLoop(sessionState *SessionState) {
	ticker := time.NewTicker(16 * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				notifyClients(sessionState)
			case <-quit:
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
			delete(sessionState.states, connClientId)
			delete(sessionState.clients, connClientId)
		}
	}
}

func (state *SessionState) JoinSession(conn *websocket.Conn) int64 {
	memberUd, err := db.Client.Incr(context.Background(), sessionClientIdPrefix+state.sessionId).Result()
	if err != nil {
		panic(err)
	}
	state.addMember(memberUd, conn)
	return memberUd
}

func (state *SessionState) LeaveSession(clientId int64) {
	state.deleteClient(clientId)
}

func GetSessionState(sessionId string) (*SessionState, error) {
	if state[sessionId] == nil {
		log.Printf("No session %s. Closing\n", sessionId)
		return nil, errors.New("no session %s")
	}
	return state[sessionId], nil
}
