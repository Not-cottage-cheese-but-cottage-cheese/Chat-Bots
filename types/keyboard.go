package types

import (
	"encoding/json"
	"fmt"
	"strconv"
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

func NewTextAction(buttonID int) *Action {
	return &Action{
		Type:    "text",
		Payload: fmt.Sprintf("{\"button\": \"%d\"}", buttonID),
		Label:   strconv.Itoa(buttonID),
	}
}

func NewKeyboard(buttonsCount int) *Keyboard {
	buttons := []*Button{}
	for i := 0; i < buttonsCount; i++ {
		buttons = append(buttons, &Button{
			Action: NewTextAction(i + 1),
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

func (k Keyboard) String() string {
	bs, _ := json.Marshal(k)
	return string(bs)
}
