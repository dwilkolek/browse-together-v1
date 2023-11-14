package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dwilkolek/browse-together/session"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type FiberServer struct {
	*fiber.App
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(),
	}

	server.Use(cors.New())
	server.Use(compress.New())
	server.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	server.Use(func(c *fiber.Ctx) error {

		c.Response().Header.Add("X-XSS-Protection", "1; mode=block")
		c.Response().Header.Add("X-Frame-Options", "DENY")
		c.Response().Header.Add("X-Content-Type-Options", "nosniff")
		c.Response().Header.Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		return c.Next()
	})
	go session.StartStreaming()
	server.RegisterFiberRoutes()
	return server
}

func (server *FiberServer) RedirectToHttpsWww() {
	server.Use(func(c *fiber.Ctx) error {

		host := string(c.Request().Host())
		uri := string(c.Request().RequestURI())
		protocol := c.Protocol()
		redirect := false
		if protocol == "http" {
			redirect = true
			protocol = "https"
		}

		if strings.HasPrefix(host, "www.") {
			redirect = true
			host = host[4:]
		}

		if redirect {
			target := fmt.Sprintf("%s://%s%s", protocol, host, uri)
			log.Println("Redirecting to: " + target)
			return c.Redirect(target, http.StatusMovedPermanently)
		}
		return c.Next()
	})

}
