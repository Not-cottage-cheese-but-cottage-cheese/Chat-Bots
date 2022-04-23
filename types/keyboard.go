package types

import (
	"encoding/json"
)

const (
	START_BUTTON = "Старт"
	LEAVE_BUTTON = "Покинуть игру"
)

type Keyboard struct {
	OneTime bool        `json:"one_time"`
	Inline  bool        `json:"inline"`
	Buttons [][]*Button `json:"buttons"`
}

type Button struct {
	Action *Action `json:"action"`
	Color  string  `json:"color"`
}

type Action struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

func NewTextAction(label string) *Action {
	return &Action{
		Type:  "text",
		Label: label,
	}
}

func NewDeckKeyboard(deck *Deck) *Keyboard {
	buttons := []*Button{}
	for _, image := range deck.Images {
		buttons = append(buttons, &Button{
			Action: NewTextAction(image.Name),
			Color:  "primary",
		})
	}
	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{
			buttons,
		},
	}
}

func NewStartKeyboard() *Keyboard {
	button := &Button{
		Action: NewTextAction(START_BUTTON),
		Color:  "primary",
	}

	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{
			{
				button,
			},
		},
	}
}

func NewEmptyKeyboard() *Keyboard {
	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{},
	}
}

func NewLeaveKeyboard() *Keyboard {
	button := &Button{
		Action: NewTextAction(LEAVE_BUTTON),
		Color:  "primary",
	}

	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{
			{
				button,
			},
		},
	}
}

// func NewSessionsKeyboard(sessions ) *Keyboard {

// }

func (k Keyboard) String() string {
	bs, _ := json.Marshal(k)
	return string(bs)
}

func (k *Keyboard) AddButtonsFromKeyboard(kb *Keyboard) *Keyboard {
	k.Buttons = append(k.Buttons, kb.Buttons...)

	return k
}
