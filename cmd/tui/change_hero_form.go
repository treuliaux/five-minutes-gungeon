package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type changeHeroFocusedField int

const (
	focusedChangeHeroPlayer changeHeroFocusedField = iota
	focusedChangeHeroHero
)

type changeHeroForm struct {
	playerInput  selectPlayerModel
	heroInput    selectHeroModel
	focusedField changeHeroFocusedField
	Submitted    bool
	Canceled     bool
}

func newChangeHeroForm(players []game.PlayerID, useExtension bool) changeHeroForm {
	return changeHeroForm{
		playerInput:  newSelectPlayerModel(players),
		heroInput:    newSelectHeroModel(useExtension),
		focusedField: focusedChangeHeroPlayer,
	}
}

func (f changeHeroForm) Update(msg tea.Msg) (changeHeroForm, tea.Cmd) {
	var cmd tea.Cmd

	switch f.focusedField {
	case focusedChangeHeroPlayer:
		f.playerInput, cmd = f.playerInput.Update(msg)
	case focusedChangeHeroHero:
		f.heroInput, cmd = f.heroInput.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			f.Canceled = true

			return f, nil

		case "tab", "shift+tab":
			switch f.focusedField {
			case focusedChangeHeroPlayer:
				f.focusedField = focusedChangeHeroHero
			case focusedChangeHeroHero:
				f.focusedField = focusedChangeHeroPlayer
			}

			return f, cmd

		case "enter":
			name := strings.TrimSpace(string(f.playerInput.choice))
			if f.focusedField == focusedChangeHeroPlayer && name != "" {
				f.focusedField = focusedChangeHeroHero

				return f, nil
			}
			if name != "" {
				f.heroInput, cmd = f.heroInput.Update(msg)
				f.Submitted = true

				return f, nil
			}

			return f, nil
		}
	}

	return f, cmd
}

func (f changeHeroForm) View() string {
	var s strings.Builder
	s.WriteString("=== Choose hero for player ===\n\n")

	// Name field
	var nameField strings.Builder
	nameField.WriteString(f.playerInput.View())
	if f.focusedField == focusedChangeHeroPlayer {
		s.WriteString(focusedModelStyle.Render(nameField.String()))
	} else {
		s.WriteString(modelStyle.Render(nameField.String()))
	}

	s.WriteString("\n\n")

	// Class selector
	var classField strings.Builder
	classField.WriteString(f.heroInput.View())
	if f.focusedField == focusedChangeHeroHero {
		s.WriteString(focusedModelStyle.Render(classField.String()))
	} else {
		s.WriteString(modelStyle.Render(classField.String()))
	}

	s.WriteString("\n\n")

	s.WriteString("<Enter> Submit  •  <Esc> Cancel\n")

	return s.String()
}
