package handlers

import (
	"fmt"
	"log"
	s "vezdekod-chat-bots/server"
	"vezdekod-chat-bots/types"

	"github.com/SevereCloud/vksdk/v2/events"
)

type CustomContext struct {
	Server      *s.Server
	Obj         events.MessageNewObject
	UserID      int
	UserIDstr   string
	SessionID   string
	MessageText string
}

func (c *CustomContext) Start() {
	var err error
	if c.SessionID != "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Вы уже начали игру!",
			ImagesDeck: nil,
			Keyboard:   nil,
		})

	} else {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Подключиться к существующей игре или начать новую?",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})
	}
	if err != nil {
		log.Println(err)
	}

}

func (c *CustomContext) Leave() {
	var err error
	if c.SessionID == "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Вы еще не начали игру!",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})

	} else {
		c.Server.LeaveGameForUser(c.UserIDstr)
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Вы покинули игру.\nПодключиться к существующей игре или начать новую?",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})
	}
	if err != nil {
		log.Println(err)
	}

}

func (c *CustomContext) NewGame() {
	// если сессия у игрока уже есть, новую начать не выйдет
	var err error
	if c.SessionID != "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Вы уже начали игру!",
			ImagesDeck: nil,
			Keyboard:   nil,
		})
	} else {
		c.SessionID = c.Server.NewSession(c.UserIDstr)
		c.Server.JoinGame(c.SessionID, c.UserIDstr)

		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    fmt.Sprintf("Удачной игры!\nID вашей сесии: %s", c.SessionID),
			ImagesDeck: nil,
			Keyboard:   types.NewStartNewGameKeyboard().Add(types.NewLeaveKeyboard()),
		})
	}

	if err != nil {
		log.Println(err)
	}
}

func (c *CustomContext) Connect() {
	var err error
	if c.SessionID != "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Вы уже начали игру!",
			ImagesDeck: nil,
			Keyboard:   nil,
		})
	} else {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Выберите сессию или создайте свою",
			ImagesDeck: nil,
			Keyboard:   types.NewSessionsKeyboard(c.Server.GetSessions()).Add(types.NewNewGameKeyboard()),
		})
	}

	if err != nil {
		log.Println(err)
	}
}

func (c *CustomContext) StartGame() {
	var err error
	if c.SessionID == "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Сперва создайте игру!",
			ImagesDeck: nil,
			Keyboard:   nil,
		})
	} else {
		session := c.Server.GetSession(c.SessionID)
		if session.HostID != c.UserIDstr {
			err = c.Server.SendMessage(types.Message{
				Receiver:   c.UserID,
				Message:    "Вы не можете начать игру, т.к. вы не хост!",
				ImagesDeck: nil,
				Keyboard:   nil,
			})
		} else {
			session.StartGame()
		}
	}

	if err != nil {
		log.Println(err)
	}
}

func (c *CustomContext) ConnectToGame() {
	session := c.Server.GetSession(c.MessageText)
	if !session.IsStarted {
		err := c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Ожидайте начала игры",
			ImagesDeck: nil,
			Keyboard:   types.NewLeaveKeyboard(),
		})
		if err != nil {
			log.Println(err)
		}
	} else {
		err := c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Игра уже идет, ожидайте конца раунда",
			ImagesDeck: nil,
			Keyboard:   types.NewLeaveKeyboard(),
		})
		if err != nil {
			log.Println(err)
		}
	}
	c.Server.JoinGame(c.MessageText, c.UserIDstr)
}

func (c *CustomContext) Submit() {
	session := c.Server.GetSession(c.SessionID)
	session.SetImage(c.UserIDstr, c.MessageText)
}

func (c *CustomContext) SendInvalid() {
	var err error

	if c.SessionID == "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Уточните запрос!",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})
	} else {
		session := c.Server.GetSession(c.SessionID)
		var keyboard *types.Keyboard = nil
		if !session.IsStarted {
			if session.HostID == c.UserIDstr {
				keyboard = types.NewStartNewGameKeyboard().Add(types.NewLeaveKeyboard())
			}
		}

		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Уточните запрос!",
			ImagesDeck: nil,
			Keyboard:   keyboard,
		})
	}

	if err != nil {
		log.Println(err)
	}
}
