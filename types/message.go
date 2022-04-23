package types

type Message struct {
	Receiver   int
	Message    string
	ImagesDeck *Deck
	Keyboard   *Keyboard
}
