package server

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
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
	api     *vk_api.VK
	userApi *vk_api.VK
	lp      *longpoll.LongPoll
	// игровые сессии
	sessions cmap.ConcurrentMap
	// Все играющие пользователи
	users cmap.ConcurrentMap

	baseDeck *types.Deck
}

func NewServer(groupToken string, secretToken string, baseDeck *types.Deck) (*Server, error) {
	api := vk_api.NewVK(groupToken)
	userApi := vk_api.NewVK(secretToken)

	groups, err := api.GroupsGetByID(nil)
	if err != nil {
		return nil, err
	}

	if len(groups) != 1 {
		return nil, err
	}

	group := groups[0]

	lp, err := longpoll.NewLongPoll(api, group.ID)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	server := &Server{
		api:      api,
		userApi:  userApi,
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
		ID:                  sessionID,
		Users:               cmap.New(),
		Deck:                *s.baseDeck,
		SelectedImageNumber: -1,
		HostID:              hostID,
		PlayerQueue:         cmap.New(),
		Result:              make(chan cmap.ConcurrentMap),
		Messages:            make(chan types.Message),
		IsStarted:           false,
	}

	gs.NameGetter = func(cm cmap.ConcurrentMap) []*types.User {
		resp, err := s.api.UsersGet(vk_api.Params{
			"user_ids": strings.Join(cm.Keys(), ","),
		})
		if err != nil {
			log.Println(err)
			return []*types.User{}
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

		return users
	}

	s.sessions.Set(sessionID, gs)

	go func() {
		for {
			select {
			case result := <-gs.Result:
				users := gs.NameGetter(result)

				for _, user := range users {
					userID, _ := strconv.Atoi(user.ID)
					err := s.SendMessage(types.Message{
						Receiver:   userID,
						Message:    fmt.Sprintf("Игра окончена!\n%s", gs.String()),
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
			PickedImage: -1,
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

	if session.PlayerQueue.Count() == 0 && !session.IsStarted {
		s.StopSession(session.ID)
	} else if !session.IsStarted && userID == session.HostID {
		for _, id := range session.PlayerQueue.Keys() {
			numID, _ := strconv.Atoi(id)
			err := s.SendMessage(types.Message{
				Receiver:   numID,
				Message:    "Хост покинул игру :(",
				ImagesDeck: nil,
				Keyboard:   types.NewGameSelectKeyboard(),
			})

			if err != nil {
				log.Println(err)
			}

			session.RemovePlayer(id)
			s.users.Remove(id)
		}

		s.StopSession(session.ID)
	}
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
	if sessionI == nil {
		return nil
	}
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
			resp, err := s.api.UploadMessagesPhoto(message.Receiver, image.GetReader())
			if err != nil {
				log.Println(err)
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

	_, err := s.api.MessagesSend(vk_api.Params(builder.Params))

	log.Printf("send message to %d with error %v", message.Receiver, err)

	return err
}

func (s *Server) GetAlbumDeck(ownerID string, albumID string) (*types.Deck, error) {
	owner, _ := strconv.Atoi(ownerID)
	album, _ := strconv.Atoi(albumID)

	resp, err := s.userApi.PhotosGet(vk_api.Params{
		"owner_id": owner,
		"album_id": album,
		"count":    types.MAX_DECK_SIZE,
	})
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`\s+`)

	deck := &types.Deck{}
	for _, imageInfo := range resp.Items {
		keywords := re.Split(imageInfo.Text, -1)

		if len(keywords) == 1 && keywords[0] == "" {
			continue
		}

		if len(keywords) > 0 {
			fmt.Println(len(keywords), keywords)
			deck.Images = append(deck.Images, &types.Image{
				ID:       uuid.NewString(),
				Name:     imageInfo.Title,
				ImgBytes: nil,
				URL:      imageInfo.MaxSize().URL,
				Keywords: keywords,
			})
		}

	}

	if len(deck.Images) == 0 {
		return &types.Deck{}, nil
	}
	// перемешаем колоду
	rand.Shuffle(len(deck.Images), func(i, j int) { deck.Images[i], deck.Images[j] = deck.Images[j], deck.Images[i] })

	return deck, nil
}
