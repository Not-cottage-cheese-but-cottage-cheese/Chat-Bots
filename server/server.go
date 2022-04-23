package server

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	types "vezdekod-chat-bots/types"

	vk_api "github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
)

type Server struct {
	api *vk_api.VK
	lp  *longpoll.LongPoll
	// игровые сессии
	sessions cmap.ConcurrentMap
	// Все играющие пользователи
	users cmap.ConcurrentMap

	baseDeck *types.Deck

	done chan string
}

func NewServer(token string, baseDeck *types.Deck) (*Server, error) {
	vk := vk_api.NewVK(token)

	groups, err := vk.GroupsGetByID(nil)
	if err != nil {
		return nil, err
	}

	if len(groups) != 1 {
		return nil, err
	}

	group := groups[0]

	lp, err := longpoll.NewLongPoll(vk, group.ID)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	server := &Server{
		api:      vk,
		lp:       lp,
		sessions: cmap.New(),
		users:    cmap.New(),
		baseDeck: baseDeck,
		done:     make(chan string),
	}

	go func() {
		for {
			select {
			case id := <-server.done:
				server.StopSession(id)
				log.Printf("session %s stopped", id)
			}
		}
	}()

	return server, nil
}

func (s *Server) Run() error {
	return s.lp.Run()
}

func (s *Server) NewSession(hostID string) string {
	sessionID := uuid.New().String()

	deck := *s.baseDeck
	rand.Shuffle(len(deck.Images), func(i, j int) { deck.Images[i], deck.Images[j] = deck.Images[j], deck.Images[i] })

	s.sessions.Set(sessionID, &types.GameSession{
		ID:                sessionID,
		Users:             cmap.New(),
		Deck:              *s.baseDeck,
		SelectedImageName: "",
		HostID:            hostID,
		PlayerQueue:       cmap.New(),
		Sender: func(m types.Message) error {
			return s.SendMessage(m)
		},
		Done: s.done,
	})
	return sessionID
}

func (s *Server) JoinGame(sessionID string, userID string) {
	sessionI, _ := s.sessions.Get(sessionID)
	session := sessionI.(*types.GameSession)

	user := &types.User{
		ID: userID,
		SessionInfo: types.SessionInfo{
			SessionID:   sessionID,
			Points:      0,
			PickedImage: "",
		},
	}

	session.PlayerQueue.Set(userID, user)
	s.users.Set(userID, user)
}

func (s *Server) LeaveGameForUser(userID string) {
	userI, _ := s.users.Get(userID)
	user := userI.(*types.User)

	sessionI, _ := s.sessions.Get(user.SessionInfo.SessionID)
	session := sessionI.(*types.GameSession)

	session.RemovePlayer(userID)
	s.users.Remove(userID)
}

func (s *Server) GetUserSessionID(userID string) string {
	userI, ok := s.users.Get(userID)
	if !ok {
		return ""
	}
	user := userI.(*types.User)
	return user.SessionInfo.SessionID
}

func (s *Server) GetSession(sessionID string) *types.GameSession {
	sessionI, _ := s.sessions.Get(sessionID)
	return sessionI.(*types.GameSession)
}

func (s *Server) StopSession(sessionID string) {
	sessionI, _ := s.sessions.Get(sessionID)
	session := sessionI.(*types.GameSession)

	for _, id := range session.Users.Keys() {
		intID, _ := strconv.Atoi(id)
		_ = s.SendMessage(types.Message{
			Receiver:   intID,
			Message:    "Подключиться к существующей игре или начать новую",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})
		s.users.Remove(id)
	}
	s.sessions.Remove(sessionID)
}

func (s *Server) GetSessions() map[string]*types.GameSession {
	result := map[string]*types.GameSession{}

	for id, sessionI := range s.sessions.Items() {
		result[id] = sessionI.(*types.GameSession)
	}

	return result
}

func (s *Server) GetLP() *longpoll.LongPoll {
	return s.lp
}

func (s *Server) SendMessage(message types.Message) error {
	builder := params.NewMessagesSendBuilder()

	attachs := []string{}

	if message.ImagesDeck != nil {
		for _, image := range message.ImagesDeck.Images {
			resp, err := s.api.UploadMessagesPhoto(message.Receiver, bytes.NewReader(image.ImgBytes))
			if err != nil {
				return nil
			}
			photo := resp[0]
			attachs = append(attachs, fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID))
		}

		builder.Attachment(strings.Join(attachs, ","))
	}
	if message.Keyboard != nil {
		builder.Keyboard(message.Keyboard.String())
	}
	builder.Message(message.Message)
	builder.RandomID(0)
	builder.PeerID(message.Receiver)

	s.api.MessagesSend(vk_api.Params(builder.Params))

	return nil
}
