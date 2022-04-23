package types

import (
	"strconv"

	cmap "github.com/orcaman/concurrent-map"
)

const (
	CARDS_COUNTS = 5
	POINTS       = 3
)

type GameSession struct {
	ID                int
	Users             cmap.ConcurrentMap
	Deck              Deck
	SelectedImageName string
	SelectedImage     int
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

func (gs *GameSession) AddPointToPlayer(userID int, points int) {
	strUserID := strconv.Itoa(userID)
	userI, _ := gs.Users.Get(strUserID)
	user := userI.(*User)

	user.Points += points

	gs.Users.Set(strUserID, user)
}

func (gs *GameSession) GetPlayerPoints(userID int) int {
	strUserID := strconv.Itoa(userID)
	userI, _ := gs.Users.Get(strUserID)
	user := userI.(*User)

	return user.Points
}
