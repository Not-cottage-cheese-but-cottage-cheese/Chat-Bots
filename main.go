package main

import (
	"context"
	"log"
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

	server, err := newServer(token)
	if err != nil {
		log.Fatal(err)
	}

	baseDeck, err := types.NewDeckFromFiles("./images", "")
	if err != nil {
		log.Fatal(err)
	}

	server.GetLP().MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		log.Printf("got message %s from %d", obj.Message.Text, obj.Message.PeerID)
		if strings.ToLower(obj.Message.Text) == "старт" {
			// новый пользователь
			user, ok := server.TryGetUser(obj.Message.PeerID)

			if !ok {
				user = &types.User{
					ID:     obj.Message.PeerID,
					Points: 0,
					Deck:   *baseDeck,
				}
				server.NewUser(user)
			}

			text := ""
			// обновим карты
			if len(user.Deck.Images) == 0 {
				text = "Ваши карты кончились. Перемешиваю колоду"
				user.Deck = *baseDeck
			}

			err := server.SendMessage(Message{
				Receiver: user.ID,
				Message:  text,
				Images:   user.Deck.GetCards(5).Images,
				Keyboard: types.NewKeyboard(len(user.Deck.GetCards(5).Images)),
			})
			if err != nil {
				log.Println(err)
			}
		}
	})

	log.Println("Start Long Poll")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
