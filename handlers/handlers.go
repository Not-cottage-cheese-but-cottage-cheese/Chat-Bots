package handlers

import (
	"fmt"
	"log"
	"regexp"
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
			Receiver: c.UserID,
			Message: fmt.Sprintf(
				`Удачной игры!
				ID вашей сесии: %s
				Сообщите данный ID своим друзьям, чтобы вы смогли сыграть вместе!
				Вы можете начать игру со стандартной колодой или прислать ссылку на альбом с новой колодой`,
				c.SessionID),
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
			Keyboard: types.NewSessionsKeyboard(c.Server.GetSessions()).
				Add(types.NewNewGameKeyboard()),
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
			Keyboard:   types.NewGameSelectKeyboard(),
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
	if session == nil {
		err := c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Сессия не найдена",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})
		if err != nil {
			log.Println(err)
		}
		return
	}

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
	session.SetUserPick(c.UserIDstr, c.MessageText)
}

func (c *CustomContext) Results() {
	var err error

	if c.SessionID == "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Вы еще не начали игру!",
			ImagesDeck: nil,
			Keyboard:   types.NewGameSelectKeyboard(),
		})
	} else {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    c.Server.GetSession(c.SessionID).String(),
			ImagesDeck: nil,
			Keyboard:   nil,
		})
	}
	if err != nil {
		log.Println(err)
	}
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
				keyboard = types.NewStartNewGameKeyboard().
					Add(types.NewLeaveKeyboard())
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

func (c *CustomContext) SetDeck() {
	var err error

	text := ""
	if c.Server.GetSession(c.SessionID).IsStarted {
		text = "Нельзя сменить колоду во время игры"
	} else if c.UserIDstr != c.Server.GetSession(c.SessionID).HostID {
		text = "Только хост может установить колоду"
	}
	if text != "" {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    text,
			ImagesDeck: nil,
			Keyboard:   nil,
		})
		if err != nil {
			log.Println(err)
		}
		return
	}

	re := regexp.MustCompile(`(https?:\/\/)?vk.com\/album(-?\d+)_(\d+)`)
	match := re.FindStringSubmatch(c.MessageText)

	ownerID := match[2]
	albumID := match[3]

	deck, err := c.Server.GetAlbumDeck(ownerID, albumID)
	if err != nil {
		log.Println(err)
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Что-то пошло не так",
			ImagesDeck: nil,
			Keyboard:   nil,
		})
	} else if len(deck.Images) == 0 {
		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    "Выбранный альбом не подходит для колоды",
			ImagesDeck: nil,
			Keyboard:   nil,
		})
	} else {
		c.Server.GetSession(c.SessionID).Deck = *deck

		err = c.Server.SendMessage(types.Message{
			Receiver:   c.UserID,
			Message:    fmt.Sprintf("Колода успешно установлена!\nЕё размер составит %d карт", len(deck.Images)),
			ImagesDeck: nil,
			Keyboard:   types.NewStartNewGameKeyboard().Add(types.NewLeaveKeyboard()),
		})
	}
	if err != nil {
		log.Println(err)
	}

}
