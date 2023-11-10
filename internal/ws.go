package internal

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func SetupWebsockets(app *fiber.App) {

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:sessionId/cursors", websocket.New(func(c *websocket.Conn) {
		sessionId := c.Params("sessionId")
		log.Printf("Trying to connect to session %s\n", sessionId)
		var (
			msg []byte
			err error
		)
		if state[sessionId] == nil {
			log.Printf("No session %s. Closing\n", sessionId)
			defer c.Close()
			return
		}
		var clientId = GetNewClientId(sessionId)
		state[sessionId].addClient(clientId, c)
		defer state[sessionId].deleteClient(clientId)

		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			var event UpdatePositionCmdDTO
			json.Unmarshal(msg, &event)
			updatePosition(sessionId, PositionStateDTO{
				MemberId: clientId,
				X:        event.X,
				Y:        event.Y,
				Selector: event.Selector,
				Location: event.Location,
			})
		}

	}))
}
