package types

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map"
)

const (
	CARDS_COUNTS = 5
	POINTS       = 3
	TURN_TIME    = 10
)

type GameSession struct {
	ID                string              // ID сессии
	Users             cmap.ConcurrentMap  // Список игроков (текущий)  [string]*User
	Deck              Deck                // колода сессии
	SelectedImageName string              // выбранная на ход картинка
	HostID            string              // ID хоста
	PlayerQueue       cmap.ConcurrentMap  // ожидающие подключения
	Sender            func(Message) error // для рассылки сообщений
	Done              chan string         // для отправки, что сессия кончилась
	IsStarted         bool                // началась ли игра
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
	if selected == nil {
		return "", result
	}

	gs.SelectedImageName = selected.Name
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

func (gs *GameSession) SetImage(userID, number string) {
	userI, _ := gs.Users.Get(userID)
	user := userI.(*User)

	user.SessionInfo.PickedImage = number

	gs.Users.Set(userID, user)
}

func (gs *GameSession) String() string {
	users := []*User{}
	for _, user := range gs.Users.Items() {
		users = append(users, user.(*User))
	}

	sort.SliceStable(users, func(i, j int) bool {
		return users[i].SessionInfo.Points > users[j].SessionInfo.Points
	})

	usrsStings := []string{}
	for i, user := range users {
		usrsStings = append(usrsStings, fmt.Sprintf("%d) %s: %d", i+1, user.ID, user.SessionInfo.Points))
	}

	return strings.Join(usrsStings, "\n")
}

func (gs *GameSession) StartGame() {
	go func() {
		gs.IsStarted = true
		for {
			var err error

			for id, user := range gs.PlayerQueue.Items() {
				gs.Users.Set(id, user)
			}
			gs.PlayerQueue.Clear()

			keyword, deck := gs.NextTurn()

			// карты кончились
			if len(deck.Images) == 0 {
				for _, users := range []cmap.ConcurrentMap{gs.Users /*, gs.PlayerQueue*/} {
					for _, id := range users.Keys() {
						userID, _ := strconv.Atoi(id)
						err = gs.Sender(Message{
							Receiver:   userID,
							Message:    fmt.Sprintf("Игра окончена!\n%s", gs.String()),
							ImagesDeck: nil,
							Keyboard:   NewStartKeyboard(),
						})
					}
				}

				if err != nil {
					log.Println(err)
				}

				gs.Done <- gs.ID

				return
			} else {
				for _, id := range gs.Users.Keys() {
					userID, _ := strconv.Atoi(id)
					err = gs.Sender(Message{
						Receiver:   userID,
						Message:    fmt.Sprintf("Время на обдумывание %d секунд.\nКлючевое слово: %s", TURN_TIME, keyword),
						ImagesDeck: deck,
						Keyboard: NewDeckKeyboard(deck).
							Add(NewLeaveKeyboard()),
					})
					if err != nil {
						log.Println(err)
					}
				}
			}

			time.Sleep(time.Second * TURN_TIME)

			for _, userI := range gs.Users.Items() {
				user := userI.(*User)
				id, _ := strconv.Atoi(user.ID)

				text := ""
				if user.SessionInfo.PickedImage == gs.SelectedImageName {
					gs.AddPointToPlayer(user.ID, POINTS)
					text = "Верно!"
				} else {
					text = fmt.Sprintf("Было близко! Правильный ответ: %s, ваш ответ: %s", gs.SelectedImageName, user.SessionInfo.PickedImage)
				}

				err = gs.Sender(Message{
					Receiver:   id,
					Message:    fmt.Sprintf("%s\nУ вас %d баллов!", text, gs.GetPlayerPoints(user.ID)),
					ImagesDeck: nil,
					Keyboard:   NewEmptyKeyboard(),
				})
				if err != nil {
					log.Println(err)
				}

				user.SessionInfo.PickedImage = ""
				gs.Users.Set(user.ID, user)
			}

		}
	}()

}
