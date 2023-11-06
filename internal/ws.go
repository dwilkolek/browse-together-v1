package internal

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var clients = make(map[int]*websocket.Conn)
var states = make(map[int]PositionState)
var updateNeeded = false

var lock sync.Mutex

func SetupWebsockets(app *fiber.App) {
	ticker := time.NewTicker(16 * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				updateClients()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/cursors", websocket.New(func(c *websocket.Conn) {

		var (
			msg []byte
			err error
		)
		var clientId = len(clients)
		lock.Lock()
		clients[clientId] = c
		lock.Unlock()
		defer deleteClient(clientId)

		log.Printf("New client. In total %d clients\n", len(clients))
		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			var event PositionState
			json.Unmarshal(msg, &event)
			states[clientId] = PositionState{
				ClientId:     clientId,
				X:            event.X,
				Y:            event.Y,
				Height:       event.Height,
				Width:        event.Width,
				ElementQuery: event.ElementQuery,
			}

			updateNeeded = len(clients) > 1
		}

	}))
}

func updateClients() {
	if !updateNeeded {
		return
	}
	lock.Lock()
	log.Printf("Updating %d clients\n", len(clients))
	defer func() {
		updateNeeded = false
		lock.Unlock()
	}()
	var err error
	for connClientId, conn := range clients {
		othersPositions := []PositionState{}

		for stateClientId, state := range states {
			if connClientId != stateClientId {
				othersPositions = append(othersPositions, state)
			}
		}

		if err = conn.WriteJSON(othersPositions); err != nil {
			log.Println("write:", err)
			delete(clients, connClientId)
			delete(states, connClientId)
		}
	}
}

func deleteClient(clientId int) {
	lock.Lock()
	defer lock.Unlock()
	updateNeeded = true
	delete(clients, clientId)
	delete(states, clientId)
}

type PositionState struct {
	ClientId     int     `json:"clientId"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Height       float64 `json:"height"`
	Width        float64 `json:"width"`
	ElementQuery string  `json:"elementQuery"`
}
