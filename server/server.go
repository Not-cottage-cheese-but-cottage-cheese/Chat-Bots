package server

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"sort"
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
	}

	return server, nil
}

func (s *Server) Run() error {
	return s.lp.Run()
}

func (s *Server) NewSession(hostID string) string {
	sessionID := uuid.New().String()

	deck := *s.baseDeck
	rand.Shuffle(len(deck.Images), func(i, j int) { deck.Images[i], deck.Images[j] = deck.Images[j], deck.Images[i] })

	gs := &types.GameSession{
		ID:                sessionID,
		Users:             cmap.New(),
		Deck:              *s.baseDeck,
		SelectedImageName: "",
		HostID:            hostID,
		PlayerQueue:       cmap.New(),
		Result:            make(chan cmap.ConcurrentMap),
		Messages:          make(chan types.Message),
		IsStarted:         false,
	}

	s.sessions.Set(sessionID, gs)

	go func() {
		for {
			select {
			case result := <-gs.Result:
				resp, err := s.api.UsersGet(vk_api.Params{
					"user_ids": strings.Join(result.Keys(), ","),
				})
				if err != nil {
					log.Println(err)
					return
				}
				users := []*types.User{}
				for _, userResp := range resp {
					userI, _ := gs.Users.Get(strconv.Itoa(userResp.ID))
					user := userI.(*types.User)
					user.FullName = userResp.FirstName + " " + userResp.LastName
					users = append(users, user)
				}

				sort.SliceStable(users, func(i, j int) bool {
					return users[i].SessionInfo.Points > users[j].SessionInfo.Points
				})

				resTable := []string{}
				for i, user := range users {
					resTable = append(resTable, fmt.Sprintf("%d)[id%s|%s] -> %d", i+1, user.ID, user.FullName, user.SessionInfo.Points))
				}

				for _, user := range users {
					userID, _ := strconv.Atoi(user.ID)
					err = s.SendMessage(types.Message{
						Receiver:   userID,
						Message:    fmt.Sprintf("Игра окончена!\n%s", strings.Join(resTable, "\n")),
						ImagesDeck: nil,
						Keyboard:   types.NewStartKeyboard(),
					})

					if err != nil {
						log.Println(err)
					}
				}

				log.Printf("session %s ended", gs.ID)
				s.StopSession(gs.ID)
				return
			case message := <-gs.Messages:
				s.SendMessage(message)
			}
		}
	}()

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
