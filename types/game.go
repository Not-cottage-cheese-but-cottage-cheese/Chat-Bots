package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map"
)

const (
	CARDS_COUNTS  = 5
	POINTS        = 3
	TURN_TIME     = 15
	MAX_DECK_SIZE = 100
)

type GameSession struct {
	ID                  string                           // ID сессии
	Users               cmap.ConcurrentMap               // Список игроков (текущий)  [string]*User
	Deck                Deck                             // колода сессии
	SelectedImageNumber int                              // выбранная на ход картинка
	HostID              string                           // ID хоста
	PlayerQueue         cmap.ConcurrentMap               // ожидающие подключения
	Result              chan cmap.ConcurrentMap          // результаты
	Messages            chan Message                     // для отправки сообщений
	IsStarted           bool                             // началась ли игра
	NameGetter          func(cmap.ConcurrentMap) []*User // получение имен пользователей с формированием рейтинговой таблицы
}

func (gs *GameSession) NextTurn() (string, *Deck) {
	count := CARDS_COUNTS

	if len(gs.Deck.Images) < count {
		count = len(gs.Deck.Images)
	}

	result := &Deck{
		Images: gs.Deck.Images[:count],
	}
	keyword, selected := result.GetUniqKeywordWithImage()
	if selected == -1 {
		return "", result
	}

	gs.SelectedImageNumber = selected
	gs.Deck.Images = gs.Deck.Images[count:]
	return keyword, result
}

func (gs *GameSession) AddPointToPlayer(userID string, points int) {
	userI, _ := gs.Users.Get(userID)
	user := userI.(*User)

	user.SessionInfo.Points += points

	gs.Users.Set(userID, user)
}

func (gs *GameSession) GetPlayerPoints(userID string) int {
	userI, _ := gs.Users.Get(userID)
	user := userI.(*User)

	return user.SessionInfo.Points
}

func (gs *GameSession) RemovePlayer(userID string) {
	gs.PlayerQueue.Remove(userID)
	gs.Users.Remove(userID)
}

func (gs *GameSession) IsEmpty() bool {
	return gs.PlayerQueue.IsEmpty() && gs.Users.IsEmpty()
}

func (gs *GameSession) SetUserPick(userID, number string) {
	userI, _ := gs.Users.Get(userID)
	user := userI.(*User)

	user.SessionInfo.PickedImage, _ = strconv.Atoi(number)

	gs.Users.Set(userID, user)
}

func (gs *GameSession) String() string {
	users := gs.NameGetter(gs.Users)

	resTable := []string{}
	for i, user := range users {
		resTable = append(resTable, fmt.Sprintf("%d)[id%s|%s] -> %d", i+1, user.ID, user.FullName, user.SessionInfo.Points))
	}

	return strings.Join(resTable, "\n")
}

func (gs *GameSession) StartGame() {
	go func() {
		gs.IsStarted = true
		for {
			for id, user := range gs.PlayerQueue.Items() {
				gs.Users.Set(id, user)
			}
			gs.PlayerQueue.Clear()

			// карты кончились
			if len(gs.Deck.Images) == 0 {
				gs.Result <- gs.Users
				return
			}

			keyword, deck := gs.NextTurn()

			for _, id := range gs.Users.Keys() {
				userID, _ := strconv.Atoi(id)
				gs.Messages <- Message{
					Receiver:   userID,
					Message:    fmt.Sprintf("Время на обдумывание %d секунд.\nКлючевое слово: %s", TURN_TIME, keyword),
					ImagesDeck: deck,
					Keyboard: NewDeckKeyboard(deck).
						Add(NewResultsKeyboard()).
						Add(NewLeaveKeyboard()),
				}
			}

			time.Sleep(time.Second * TURN_TIME)

			for _, userI := range gs.Users.Items() {
				user := userI.(*User)
				id, _ := strconv.Atoi(user.ID)

				text := ""
				if user.SessionInfo.PickedImage == gs.SelectedImageNumber {
					gs.AddPointToPlayer(user.ID, POINTS)
					text = "Верно!"
				} else {
					text = fmt.Sprintf("Было близко! Правильный ответ: %d", gs.SelectedImageNumber)
				}

				gs.Messages <- Message{
					Receiver:   id,
					Message:    fmt.Sprintf("%s\nВаше количество баллов: %d", text, gs.GetPlayerPoints(user.ID)),
					ImagesDeck: nil,
					Keyboard:   NewEmptyKeyboard(),
				}

				user.SessionInfo.PickedImage = -1
				gs.Users.Set(user.ID, user)
			}

		}
	}()

}
