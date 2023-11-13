package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/dwilkolek/browse-together/dto"
	"github.com/dwilkolek/browse-together/session"
)

var publicApi string = "ws://127.0.0.1:8080"

func init() {
	apiBase := os.Getenv("API_BASE")
	if apiBase != "" {
		publicApi = apiBase
	}
}

func (s *FiberServer) RegisterFiberRoutes() {
	v1 := s.App.Group("/api/v1/sessions")
	v1.Post("/", s.createSessionHandler)
	v1.Get("/", s.getAllSessionsHandler)
	v1.Get("/:id", s.getSessionHandler)
	v1.Delete("/:id", s.deleteSessionHandler)

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
		return c.JSON(toDto(newSession, publicApi))
	}

	return fiber.NewError(fiber.StatusInternalServerError)
}

func (s *FiberServer) getAllSessionsHandler(c *fiber.Ctx) error {
	sessions := session.GetSessions()
	sessionsDto := make([]dto.SessionDTO, len(sessions))
	for i, session := range sessions {
		sessionsDto[i] = toDto(session, publicApi)
	}
	return c.JSON(sessionsDto)
}

func (s *FiberServer) getSessionHandler(c *fiber.Ctx) error {
	expectedKey := c.Params("id")
	if session, err := session.GetSession(expectedKey); err == nil {
		return c.JSON(toDto(session, publicApi))
	}
	return fiber.NewError(fiber.StatusNotFound)
}

func (s *FiberServer) deleteSessionHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	session.DeleteSession(id)
	return nil
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
func toDto(session session.Session, baseUrl string) dto.SessionDTO {
	return dto.SessionDTO{
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
