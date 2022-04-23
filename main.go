package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/spf13/viper"

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

	server, err := newServer(token, baseDeck)
	if err != nil {
		log.Fatal(err)
	}

	server.GetLP().MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		userID := obj.Message.PeerID
		sessionID := server.GetUserSessionID(userID)

		log.Printf("got message %s from %d (session %d)", obj.Message.Text, userID, sessionID)

		messageText := obj.Message.Text

		if strings.ToLower(messageText) == "старт" {
			if sessionID == -1 {
				sessionID = server.NewSession()
				server.JoinGame(sessionID, userID)
			} else {
				err := server.SendMessage(Message{
					Receiver:   userID,
					Message:    "Вы уже начали игру!",
					ImagesDeck: nil,
					Keyboard:   nil,
				})
				if err != nil {
					log.Println(err)
				}
				return
			}
		} else if _, err := strconv.Atoi(messageText); err == nil {
			if sessionID == -1 {
				err := server.SendMessage(Message{
					Receiver:   userID,
					Message:    "Сперва начните игру!",
					ImagesDeck: nil,
					Keyboard:   types.NewStartKeyboard(),
				})
				if err != nil {
					log.Println(err)
				}
				return
			}
			session := server.GetSession(sessionID)
			text := ""
			if messageText == session.SelectedImageName {
				session.Users[userID].Points += 3
				text = "Верно!"
			} else {
				text = fmt.Sprintf("Было близко! Правильный ответ: %s", session.SelectedImageName)
			}
			err := server.SendMessage(Message{
				Receiver:   userID,
				Message:    fmt.Sprintf("%s\nУ вас %d баллов!", text, session.Users[userID].Points),
				ImagesDeck: nil,
				Keyboard:   nil,
			})
			if err != nil {
				log.Println(err)
			}
		} else {
			err := server.SendMessage(Message{
				Receiver:   userID,
				Message:    "Уточните запрос/ответ",
				ImagesDeck: nil,
				Keyboard:   nil,
			})
			if err != nil {
				log.Println(err)
			}
			return
		}

		session := server.GetSession(sessionID)
		keyword, deck := session.NextTurn()

		if len(deck.Images) == 0 {
			err := server.SendMessage(Message{
				Receiver:   userID,
				Message:    fmt.Sprintf("Игра окончена! Ваш результат: %d", session.Users[userID].Points),
				ImagesDeck: nil,
				Keyboard:   types.NewStartKeyboard(),
			})
			if err != nil {
				log.Println(err)
			}
			server.StopSession(sessionID)
		} else {
			err := server.SendMessage(Message{
				Receiver:   userID,
				Message:    keyword,
				ImagesDeck: deck,
				Keyboard:   types.NewDeckKeyboard(deck),
			})
			if err != nil {
				log.Println(err)
			}
		}

	})

	log.Println("Start server")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
