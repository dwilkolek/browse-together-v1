package internal

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func SetupSessionApi(app *fiber.App) {
	var publicApi string = "ws://127.0.0.1:8080"
	apiBase := os.Getenv("API_BASE")
	if apiBase != "" {
		publicApi = apiBase
	}

	v1 := app.Group("/api/v1/sessions")
	v1.Post("/", func(c *fiber.Ctx) error {
		var cmd CreateSessionV1Cmd
		if err := c.BodyParser(&cmd); err != nil {
			return err
		}
		newSession := Session{
			Id:           uuid.New().String(),
			Name:         cmd.Name,
			Creator:      cmd.Creator,
			BaseLocation: cmd.BaseLocation,
		}
		if err := StoreSession(newSession); err == nil {
			return c.JSON(newSession.toDto(publicApi))
		}

		return fiber.NewError(fiber.StatusInternalServerError)
	})
	v1.Get("/", func(c *fiber.Ctx) error {
		sessions := GetSessions()
		sessionsDto := make([]SessionDTO, len(sessions))
		for i, session := range sessions {
			sessionsDto[i] = session.toDto(publicApi)
		}
		return c.JSON(sessionsDto)
	})
	v1.Get("/:id", func(c *fiber.Ctx) error {
		expectedKey := c.Params("id")
		if session, err := GetSession(expectedKey); err == nil {
			return c.JSON(session.toDto(publicApi))
		}
		return fiber.NewError(fiber.StatusNotFound)
	})
	v1.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		DeleteSession(id)
		return nil
	})
}

func (session Session) toDto(baseUrl string) SessionDTO {
	return SessionDTO{
		Id:                session.Id,
		JoinUrl:           fmt.Sprintf("%s/ws/%s/cursors", baseUrl, session.Id),
		Name:              session.Name,
		BaseUrl:           session.BaseLocation,
		CreatorIdentifier: session.Creator,
	}
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
	Id           string `json:"id"`
	Name         string `json:"name"`
	Creator      string `json:"creator"`
	BaseLocation string `json:"baseLocation"`
}
