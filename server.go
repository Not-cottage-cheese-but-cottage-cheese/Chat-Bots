package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	types "vezdekod-chat-bots/types"

	vk_api "github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	cmap "github.com/orcaman/concurrent-map"
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
	sessions cmap.ConcurrentMap
	// Все играющие пользователи
	users cmap.ConcurrentMap

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
		sessions: cmap.New(),
		users:    cmap.New(),
		baseDeck: baseDeck,
	}, nil
}

func (s *Server) Run() error {
	return s.lp.Run()
}

func (s *Server) NewSession() int {
	sessionID := len(s.sessions.Items())

	deck := *s.baseDeck
	rand.Shuffle(len(deck.Images), func(i, j int) { deck.Images[i], deck.Images[j] = deck.Images[j], deck.Images[i] })

	s.sessions.Set(strconv.Itoa(sessionID), &types.GameSession{
		ID:    sessionID,
		Users: cmap.New(),
		Deck:  *s.baseDeck,
	})
	return sessionID
}

func (s *Server) JoinGame(sessionID int, userID int) {
	sessionI, ok := s.sessions.Get(strconv.Itoa(sessionID))
	if !ok {
		s.NewSession()
	}
	session := sessionI.(*types.GameSession)

	user := &types.User{
		ID:        userID,
		Points:    0,
		SessionID: sessionID,
	}

	session.Users.Set(strconv.Itoa(userID), user)
	s.users.Set(strconv.Itoa(userID), user)
}

func (s *Server) GetUserSessionID(userID int) int {
	userI, ok := s.users.Get(strconv.Itoa(userID))
	if !ok {
		return -1
	}
	user := userI.(*types.User)
	return user.SessionID
}

func (s *Server) GetSession(sessionID int) *types.GameSession {
	sessionI, _ := s.sessions.Get(strconv.Itoa(sessionID))
	return sessionI.(*types.GameSession)
}

func (s *Server) StopSession(sessionID int) {
	sessionI, _ := s.sessions.Get(strconv.Itoa(sessionID))
	session := sessionI.(*types.GameSession)

	for _, id := range session.Users.Keys() {
		s.users.Remove(id)
	}
	s.sessions.Remove(strconv.Itoa(sessionID))
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
