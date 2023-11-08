package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const sessionStreamPrefix = "sessionStream-"
const internal = "internal"

type State struct {
	sessionId    string
	clients      map[int64]*websocket.Conn
	states       map[int64]PositionState
	updateNeeded bool
	lock         sync.Mutex
}

func (state *State) lockMe(reason string) {
	log.Printf("Locking %s: %s\n", reason, state.sessionId)
	state.lock.Lock()
}
func (state *State) unlockMe(reason string) {
	log.Printf("Unlock %s: %s\n", reason, state.sessionId)
	state.lock.Unlock()
}
func (state *State) addClient(clientId int64, conn *websocket.Conn) {
	state.lockMe("addClient")
	defer state.unlockMe("addClient")
	log.Printf("New client. In total %d clients\n", len(state.clients))
	state.clients[clientId] = conn
}

func (state *State) deleteClient(clientId int64) {
	state.lockMe("deleteClient")
	defer state.unlockMe("deleteClient")
	log.Printf("Removing client. In total %d clients\n", len(state.clients))
	state.updateNeeded = true
	delete(state.states, clientId)
	delete(state.clients, clientId)
}
func (state *State) newPosition(newPosition PositionState) {
	state.lockMe("newPosition")
	defer state.unlockMe("newPosition")
	state.states[newPosition.ClientId] = newPosition
	state.updateNeeded = true
}

var state map[string]*State = make(map[string]*State)

func StartListening() {

	pubSub := client.Subscribe(context.Background(), internal)
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
			state[sessionId] = &State{
				sessionId:    sessionId,
				clients:      make(map[int64]*websocket.Conn),
				states:       make(map[int64]PositionState),
				updateNeeded: false,
			}
			go onPositionUpdate(state[sessionId])
			go notifyClientsLoop(state[sessionId])
		}

	}

}

func onPositionUpdate(sessionState *State) {
	client := getRedis()
	// Subscribe to the channel
	pubSub := client.Subscribe(context.Background(), sessionStreamPrefix+sessionState.sessionId)
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
		var positionState PositionState
		err := json.Unmarshal([]byte(msg.Payload), &positionState)
		if err != nil {
			log.Printf("Failed to unmarshal PositionState: %s\n", err)
		}

		fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
		sessionState.newPosition(positionState)
	}
}

func stopPositionStreaming(sessionId string) {
	log.Printf("Stopping position streaming %s\n", sessionId)
	client.Publish(context.Background(), sessionStreamPrefix+sessionId, "STOP").Result()
}

func startPositionStreaming(sessionId string) {
	log.Printf("Starting position streaming %s\n", sessionId)
	client.Publish(context.Background(), internal, "START-"+sessionId).Result()
	client.Publish(context.Background(), sessionStreamPrefix+sessionId, "START").Result()
}

func updatePosition(sessionId string, update PositionState) {
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("Failed to marshal PositionState: %s\n", err)
	}
	client.Publish(context.Background(), sessionStreamPrefix+sessionId, data)
}

func notifyClientsLoop(sessionState *State) {
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

func notifyClients(sessionState *State) {
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
		toSend := []PositionState{}
		for _, ps := range sessionState.states {
			if ps.ElementQuery != "" {
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
