package main

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type selectHeroModel struct {
	cursor  int
	choice  game.HeroClass
	choices []game.HeroClass
}

func newSelectHeroModel(useExtension bool) selectHeroModel {
	choices := game.HeroClassValues()
	if !useExtension {
		choices = slices.DeleteFunc(slices.Clone(choices), func(c game.HeroClass) bool {
			return c == game.Druid || c == game.Shaman
		})
	}

	var firstChoice game.HeroClass
	if len(choices) > 0 {
		firstChoice = choices[0]
	}

	return selectHeroModel{
		choices: choices,
		choice:  firstChoice,
	}
}

func (m selectHeroModel) Update(msg tea.Msg) (selectHeroModel, tea.Cmd) {
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

func (m selectHeroModel) View() string {
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
