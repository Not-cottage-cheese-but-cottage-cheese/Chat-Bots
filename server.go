package main

import (
	"bytes"
	"fmt"
	"strings"
	types "vezdekod-chat-bots/types"

	vk_api "github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type Message struct {
	Receiver int
	Message  string
	Images   []*types.Image
	Keyboard *types.Keyboard
}

type Server struct {
	api   *vk_api.VK
	lp    *longpoll.LongPoll
	users map[int]*types.User
}

func newServer(token string) (*Server, error) {
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

	return &Server{
		api:   vk,
		lp:    lp,
		users: map[int]*types.User{},
	}, nil
}

func (s *Server) Run() error {
	return s.lp.Run()
}

func (s *Server) UserExist(id int) bool {
	_, ok := s.users[id]
	return ok
}

func (s *Server) NewUser(user *types.User) {
	s.users[user.ID] = user
}

func (s *Server) TryGetUser(id int) (*types.User, bool) {
	user, ok := s.users[id]
	return user, ok
}

func (s *Server) GetLP() *longpoll.LongPoll {
	return s.lp
}

func (s *Server) SendMessage(message Message) error {
	builder := params.NewMessagesSendBuilder()

	attachs := []string{}
	for _, image := range message.Images {
		resp, err := s.api.UploadMessagesPhoto(message.Receiver, bytes.NewReader(image.ImgBytes))
		if err != nil {
			return nil
		}
		photo := resp[0]
		attachs = append(attachs, fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID))
	}
	builder.Keyboard(message.Keyboard.String())
	builder.Attachment(strings.Join(attachs, ","))
	builder.Message(message.Message)
	builder.RandomID(0)
	builder.PeerID(message.Receiver)

	s.api.MessagesSend(vk_api.Params(builder.Params))

	return nil
}
