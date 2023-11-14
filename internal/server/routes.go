package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/dwilkolek/browse-together/dto"
	"github.com/dwilkolek/browse-together/session"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Get("/health", func(c *fiber.Ctx) error {
		return c.Send(nil)
	})

	v1 := s.App.Group("/api/v1/sessions")
	v1.Post("/", s.createSessionHandler)
	v1.Get("/", s.getAllSessionsHandler)
	v1.Get("/:id", s.getSessionHandler)
	v1.Delete("/:id", s.deleteSessionHandler)

	v1.Get("/:id/joinUrl", s.getJoinSessionHandler)

	s.App.Get("/ws/:sessionId/cursors", websocket.New(s.sessionHandler))
}

func (s *FiberServer) createSessionHandler(c *fiber.Ctx) error {
	var cmd CreateSessionV1Cmd
	if err := c.BodyParser(&cmd); err != nil {
		return err
	}
	newSession := session.Session{
		Id:           uuid.New().String(),
		Name:         cmd.Name,
		Creator:      cmd.Creator,
		BaseLocation: cmd.BaseLocation,
	}

	if err := session.StoreSession(newSession); err == nil {
		return c.JSON(toDto(newSession))
	}

	return fiber.NewError(fiber.StatusInternalServerError)
}

func (s *FiberServer) getAllSessionsHandler(c *fiber.Ctx) error {
	sessions := session.GetSessions()
	sessionsDto := make([]dto.SessionDTO, len(sessions))
	for i, session := range sessions {
		sessionsDto[i] = toDto(session)
	}
	return c.JSON(sessionsDto)
}

func (s *FiberServer) getSessionHandler(c *fiber.Ctx) error {
	expectedKey := c.Params("id")
	if session, err := session.GetSession(expectedKey); err == nil {
		return c.JSON(toDto(session))
	}
	return fiber.NewError(fiber.StatusNotFound)
}

func (s *FiberServer) deleteSessionHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	session.DeleteSession(id)
	return nil
}
func (s *FiberServer) getJoinSessionHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(map[string]string{
		"joinUrl": fmt.Sprintf("/ws/%s/cursors", id),
	})
}

func (s *FiberServer) sessionHandler(c *websocket.Conn) {
	sessionId := c.Params("sessionId")
	log.Printf("Trying to connect to session %s\n", sessionId)
	var (
		msg []byte
		err error
	)
	defer c.Close()

	sessionState, err := session.GetSessionState(sessionId)
	if err != nil {
		log.Printf("No session state: %s\n", err)
		return
	}
	memberId := sessionState.JoinSession(c)

	for {
		if _, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		var event dto.UpdatePositionCmdDTO
		json.Unmarshal(msg, &event)

		sessionState.UpdatePosition(dto.PositionStateDTO{
			MemberId: memberId,
			X:        event.X,
			Y:        event.Y,
			Selector: event.Selector,
			Location: event.Location,
		})
	}

}
func toDto(session session.Session) dto.SessionDTO {
	return dto.SessionDTO{
		Id:                session.Id,
		JoinUrl:           fmt.Sprintf("/api/v1/sessions/%s/join", session.Id),
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
