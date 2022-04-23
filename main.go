package main

import (
	"context"
	"log"
	"strings"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
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

	vk := api.NewVK(token)

	groups, err := vk.GroupsGetByID(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(groups) != 1 {
		log.Fatal("something goes wrong")
	}

	group := groups[0]

	lp, err := longpoll.NewLongPoll(vk, group.ID)
	if err != nil {
		log.Fatal(err)
	}

	users := map[int]*types.User{}
	baseDeck, err := types.NewDeckFromFiles("./images", "")
	if err != nil {
		log.Fatal(err)
	}
	baseDeck.Images = baseDeck.Images[:10]

	lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		log.Printf("got message %s from %d", obj.Message.Text, obj.Message.PeerID)
		if strings.ToLower(obj.Message.Text) == "старт" {
			user, ok := users[obj.Message.PeerID]

			// новый пользователь
			if !ok {
				user = &types.User{
					ID:     obj.Message.PeerID,
					Points: 0,
					Deck:   *baseDeck,
				}
				users[obj.Message.PeerID] = user
			}

			text := ""
			// обновим карты
			if len(user.Deck.Images) == 0 {
				text = "Ваши карты кончились. Перемешиваю колоду"
				user.Deck = *baseDeck
			}

			message, err := newMessageWithImages(vk, user, text, user.Deck.GetCards(5).Images)
			if err != nil {
				log.Fatal(err)
			}

			_, err = vk.MessagesSend(api.Params(message.Params))
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	log.Println("Start Long Poll")
	if err := lp.Run(); err != nil {
		log.Fatal(err)
	}
}
