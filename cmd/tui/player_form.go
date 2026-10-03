package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type selectPlayerModel struct {
	cursor  int
	choice  game.PlayerID
	choices []game.PlayerID
}

func newSelectPlayerModel(players []game.PlayerID) selectPlayerModel {
	choice := game.PlayerID("")
	if len(players) > 0 {
		choice = players[0]
	}

	return selectPlayerModel{
		choices: players,
		choice:  choice,
	}
}

func (m selectPlayerModel) Update(msg tea.Msg) (selectPlayerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if len(m.choices) == 0 {
				return m, nil
			}
			m.choice = m.choices[m.cursor]

			return m, nil

		case "down":
			if len(m.choices) == 0 {
				return m, nil
			}
			m.cursor = (m.cursor + 1) % len(m.choices)
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up":
			if len(m.choices) == 0 {
				return m, nil
			}
			m.cursor = (m.cursor - 1) % len(m.choices)
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m selectPlayerModel) View() string {
	s := strings.Builder{}
	s.WriteString("Pick a player:\n\n")

	for i := range m.choices {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(string(m.choices[i]))
		s.WriteString("\n")
	}

	return s.String()
}
