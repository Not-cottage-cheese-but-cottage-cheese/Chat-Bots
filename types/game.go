package types

const CARDS_COUNTS = 5

type GameSession struct {
	ID                int
	Users             map[int]*User
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
