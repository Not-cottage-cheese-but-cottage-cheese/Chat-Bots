package types

import (
	"encoding/json"
	"strconv"
)

const (
	START_BUTTON      = "Старт"
	LEAVE_BUTTON      = "Покинуть игру"
	SESSIONS_BUTTON   = "Игровые сессии"
	CONNECT_BUTTON    = "Подключиться"
	NEW_GAME_BUTTON   = "Начать новую игру"
	START_GAME_BUTTON = "Начать игру"
	RESULTS_BUTTON    = "Результаты"
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
	for i := range deck.Images {
		buttons = append(buttons, &Button{
			Action: NewTextAction(strconv.Itoa(i + 1)),
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

func NewResultsKeyboard() *Keyboard {
	button := &Button{
		Action: NewTextAction(RESULTS_BUTTON),
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

func NewNewGameKeyboard() *Keyboard {
	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{
			{
				&Button{
					Action: NewTextAction(NEW_GAME_BUTTON),
					Color:  "primary",
				},
			},
		},
	}
}

func NewGameSelectKeyboard() *Keyboard {
	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{
			{
				&Button{
					Action: NewTextAction(NEW_GAME_BUTTON),
					Color:  "primary",
				},
			},
			{
				&Button{
					Action: NewTextAction(CONNECT_BUTTON),
					Color:  "primary",
				},
			},
		},
	}
}

func NewStartNewGameKeyboard() *Keyboard {
	return &Keyboard{
		OneTime: false,
		Buttons: [][]*Button{
			{
				&Button{
					Action: NewTextAction(START_GAME_BUTTON),
					Color:  "primary",
				},
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

func NewSessionsKeyboard(sessions map[string]*GameSession) *Keyboard {
	buttons := [][]*Button{}

	for id := range sessions {
		buttons = append(buttons, []*Button{
			{
				Action: NewTextAction(id),
				Color:  "primary",
			},
		})
	}

	return &Keyboard{
		Inline:  true,
		Buttons: buttons,
	}
}

func (k Keyboard) String() string {
	bs, _ := json.Marshal(k)
	return string(bs)
}

func (k *Keyboard) Add(kb *Keyboard) *Keyboard {
	k.Buttons = append(k.Buttons, kb.Buttons...)
	return k
}

func (k *Keyboard) AddButtons(bts []*Button) *Keyboard {
	k.Buttons = append(k.Buttons, bts)
	return k
}
