package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var sessions = make(map[string]Session)

func SetupSessionApi(app *fiber.App) {
	v1 := app.Group("/api/v1/sessions")
	v1.Post("/", func(c *fiber.Ctx) error {
		var cmd CreateSessionV1Cmd
		if err := c.BodyParser(&cmd); err != nil {
			return err
		}
		newSession := Session{
			Id:      uuid.New().String(),
			Name:    cmd.Name,
			Active:  true,
			Creator: cmd.Creator,
			Users:   1,
		}
		sessions[newSession.Id] = newSession
		StoreSession(newSession)
		return c.JSON(newSession)
	})
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(GetSessions())
	})
	v1.Get("/:id", func(c *fiber.Ctx) error {
		expectedKey := c.Params("id")
		for key, session := range sessions {
			if key == expectedKey {
				return c.JSON(session)
			}
		}
		return fiber.NewError(fiber.StatusNotFound)
	})
	v1.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		delete(sessions, id)
		return nil
	})
}

type CreateSessionV1Cmd struct {
	Name         string `json:"name"`
	BaseLocation string `json:"baseLocation"`
	Creator      string `json:"creator"`
}

type CloseSessionV1Cmd struct {
	Id string `json:"id"`
}

type Session struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Creator string `json:"creator"`
	Users   int    `json:"users"`
}
