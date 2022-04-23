package types

import (
	"encoding/json"
	"fmt"
)

type Keyboard struct {
	OneTime bool        `json:"one_time"`
	Buttons [][]*Button `json:"buttons"`
}

type Button struct {
	Action *Action `json:"action"`
	Color  string  `json:"color"`
}

type Action struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Label   string `json:"label"`
}

func NewTextAction(buttonText string, payload interface{}) *Action {
	return &Action{
		Type:    "text",
		Payload: fmt.Sprintf("{\"button\": \"%v\"}", payload),
		Label:   buttonText,
	}
}

func NewDeckKeyboard(deck *Deck) *Keyboard {
	buttons := []*Button{}
	for i, image := range deck.Images {
		buttons = append(buttons, &Button{
			Action: NewTextAction(image.Name, i+1),
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
		Action: NewTextAction("Старт", "старт"),
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

func NewLeaveKeyboard() *Keyboard {
	button := &Button{
		Action: NewTextAction("Покинуть игру", "Покинуть игру"),
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

func (k Keyboard) String() string {
	bs, _ := json.Marshal(k)
	return string(bs)
}

func (k *Keyboard) AddButtonsFromKeyboard(kb *Keyboard) *Keyboard {
	k.Buttons = append(k.Buttons, kb.Buttons...)

	return k
}
