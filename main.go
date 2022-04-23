package main

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"vezdekod-chat-bots/handlers"
	s "vezdekod-chat-bots/server"
	types "vezdekod-chat-bots/types"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("no such config file")
		} else {
			log.Println("read config error")
		}
		log.Fatal(err)
	}
}

func main() {
	token := viper.GetString("group_token")

	baseDeck, err := types.NewDeckFromFiles("./images", "keywords.txt")
	if err != nil {
		log.Fatal(err)
	}

	baseDeck.Images = baseDeck.Images[:9]

	server, err := s.NewServer(token, baseDeck)
	if err != nil {
		log.Fatal(err)
	}

	server.GetLP().MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		go func() {

			ctx := &handlers.CustomContext{
				Server:      server,
				Obj:         obj,
				UserID:      obj.Message.PeerID,
				UserIDstr:   strconv.Itoa(obj.Message.PeerID),
				SessionID:   server.GetUserSessionID(strconv.Itoa(obj.Message.PeerID)),
				MessageText: obj.Message.Text,
			}

			log.Printf("got message %s from %d (session %s)", ctx.MessageText, ctx.UserID, ctx.SessionID)

			if strings.EqualFold(ctx.MessageText, types.START_BUTTON) {
				// Старт
				ctx.Start()
			} else if strings.EqualFold(ctx.MessageText, types.LEAVE_BUTTON) {
				// Покинуть игру
				ctx.Leave()
			} else if strings.EqualFold(ctx.MessageText, types.NEW_GAME_BUTTON) {
				// Создать новую игру
				ctx.NewGame()
			} else if strings.EqualFold(ctx.MessageText, types.CONNECT_BUTTON) {
				// Подключиться
				ctx.Connect()
			} else if strings.EqualFold(ctx.MessageText, types.START_GAME_BUTTON) {
				ctx.StartGame()
			} else if _, err := uuid.Parse(ctx.MessageText); err == nil && ctx.SessionID == "" {
				// Пришел уид И игровой сессии нет
				ctx.ConnectToGame()
			} else if _, err := strconv.Atoi(ctx.MessageText); err == nil && ctx.SessionID != "" {
				ctx.Submit()
			} else {
				ctx.SendInvalid()
			}

		}()
	})

	log.Println("Start server")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
