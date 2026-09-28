package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type heroModel struct {
	cursor  int
	choice  game.HeroClass
	choices []game.HeroClass
}

func newHeroModel() heroModel {
	return heroModel{
		choices: game.HeroClassValues(),
	}
}

func (m heroModel) Update(msg tea.Msg) (heroModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.choice = m.choices[m.cursor]

			return m, nil

		case "down":
			m.cursor = (m.cursor + 1) % len(m.choices)
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up":
			m.cursor = (m.cursor - 1) % len(m.choices)
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m heroModel) View() string {
	s := strings.Builder{}
	s.WriteString("Pick your hero:\n\n")

	for i := range m.choices {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(m.choices[i].String())
		s.WriteString("\n")
	}

	return s.String()
}
