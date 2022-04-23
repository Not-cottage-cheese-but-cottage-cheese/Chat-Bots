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
		go func() {
			userID := obj.Message.PeerID
			userIDstr := strconv.Itoa(userID)
			sessionID := server.GetUserSessionID(userIDstr)

			log.Printf("got message %s from %d (session %s)", obj.Message.Text, userID, sessionID)

			messageText := obj.Message.Text

			if strings.EqualFold(messageText, types.START_BUTTON) {
				if sessionID == "" {
					sessionID = server.NewSession()
					server.JoinGame(sessionID, userIDstr)

					err := server.SendMessage(Message{
						Receiver:   userID,
						Message:    "Удачной игры!",
						ImagesDeck: nil,
						Keyboard:   types.NewEmptyKeyboard(),
					})
					if err != nil {
						log.Println(err)
					}
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
				if sessionID == "" {
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
					session.AddPointToPlayer(userID, types.POINTS)
					text = "Верно!"
				} else {
					text = fmt.Sprintf("Было близко! Правильный ответ: %s", session.SelectedImageName)
				}
				err := server.SendMessage(Message{
					Receiver:   userID,
					Message:    fmt.Sprintf("%s\nУ вас %d баллов!", text, session.GetPlayerPoints(userID)),
					ImagesDeck: nil,
					Keyboard:   types.NewEmptyKeyboard(),
				})
				if err != nil {
					log.Println(err)
				}
			} else if strings.EqualFold(messageText, types.LEAVE_BUTTON) {
				if sessionID == "" {
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
				} else {
					err := server.SendMessage(Message{
						Receiver:   userID,
						Message:    fmt.Sprintf("Спасибо за игру! Ваш результат: %d баллов!", server.GetSession(sessionID).GetPlayerPoints(userID)),
						ImagesDeck: nil,
						Keyboard:   types.NewStartKeyboard(),
					})
					if err != nil {
						log.Println(err)
					}

					server.StopSession(sessionID)
					return
				}
			} else {
				err := server.SendMessage(Message{
					Receiver:   userID,
					Message:    "Уточните запрос!",
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
					Message:    fmt.Sprintf("Игра окончена! Ваш результат: %d", session.GetPlayerPoints(userID)),
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
					Keyboard: types.NewDeckKeyboard(deck).
						AddButtonsFromKeyboard(types.NewLeaveKeyboard()),
				})
				if err != nil {
					log.Println(err)
				}
			}
		}()
	})

	log.Println("Start server")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
