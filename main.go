package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No such config file")
		} else {
			log.Println("Read config error")
		}

		log.Println("Get from OS env")
		viper.Set("GROUP_TOKEN", os.Getenv("GROUP_TOKEN"))
		viper.Set("SECRET", os.Getenv("SECRET"))
		viper.Set("PORT", os.Getenv("PORT"))
	}
}

func main() {
	groupToken := viper.GetString("GROUP_TOKEN")
	secretToken := viper.GetString("SECRET")

	baseDeck, err := types.NewDeckFromFiles("./images", "keywords.txt")
	if err != nil {
		log.Fatal(err)
	}

	server, err := s.NewServer(groupToken, secretToken, baseDeck)
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
				// Начать игру
				ctx.StartGame()
			} else if _, err := uuid.Parse(ctx.MessageText); err == nil && ctx.SessionID == "" {
				// Пришел уид И игровой сессии нет
				ctx.ConnectToGame()
			} else if _, err := strconv.Atoi(ctx.MessageText); err == nil && ctx.SessionID != "" {
				// пришло число
				ctx.Submit()
			} else if strings.EqualFold(ctx.MessageText, types.RESULTS_BUTTON) {
				// Результаты
				ctx.Results()
			} else if isAlbumURL(ctx.MessageText) && ctx.SessionID != "" {
				// сменить колоду
				ctx.SetDeck()
			} else {
				// пришло нечто невалидное
				ctx.SendInvalid()
			}

		}()
	})

	log.Println(viper.GetString("PORT"))

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("pong")
			fmt.Fprintf(w, "pong")
		})
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", viper.GetString("PORT")), nil))
	}()

	log.Println("Start server")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
