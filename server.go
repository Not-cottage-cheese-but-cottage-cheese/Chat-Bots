package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"time"
	types "vezdekod-chat-bots/types"

	vk_api "github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type Message struct {
	Receiver   int
	Message    string
	ImagesDeck *types.Deck
	Keyboard   *types.Keyboard
}

type Server struct {
	api *vk_api.VK
	lp  *longpoll.LongPoll
	// игровые сессии
	sessions map[int]*types.GameSession

	// Все играющие пользователи
	users map[int]*types.User

	baseDeck *types.Deck
}

func newServer(token string, baseDeck *types.Deck) (*Server, error) {
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
	return &Server{
		api:      vk,
		lp:       lp,
		sessions: map[int]*types.GameSession{},
		users:    map[int]*types.User{},
		baseDeck: baseDeck,
	}, nil
}

func (s *Server) Run() error {
	return s.lp.Run()
}

func (s *Server) NewSession() int {
	sessionID := len(s.sessions)

	deck := *s.baseDeck
	rand.Shuffle(len(deck.Images), func(i, j int) { deck.Images[i], deck.Images[j] = deck.Images[j], deck.Images[i] })

	s.sessions[sessionID] = &types.GameSession{
		ID:    sessionID,
		Users: map[int]*types.User{},
		Deck:  *s.baseDeck,
	}
	return sessionID
}

func (s *Server) JoinGame(sessionID int, userID int) {
	session, ok := s.sessions[sessionID]
	if !ok {
		s.NewSession()
	}

	user := &types.User{
		ID:        userID,
		Points:    0,
		SessionID: sessionID,
	}

	session.Users[userID] = user
	s.users[userID] = user
}

func (s *Server) GetUserSessionID(id int) int {
	user, ok := s.users[id]
	if !ok {
		return -1
	}
	return user.SessionID
}

func (s *Server) GetSession(id int) *types.GameSession {
	return s.sessions[id]
}

func (s *Server) StopSession(sessionID int) {
	session := s.sessions[sessionID]
	for id := range session.Users {
		delete(s.users, id)
	}
	delete(s.sessions, sessionID)
}

func (s *Server) GetLP() *longpoll.LongPoll {
	return s.lp
}

func (s *Server) SendMessage(message Message) error {
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
