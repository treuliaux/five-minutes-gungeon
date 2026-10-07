package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type addPlayerFocusedField int

const (
	focusedAddPlayerName addPlayerFocusedField = iota
	focusedAddPlayerHero
)

type addPlayerForm struct {
	nameInput    textinput.Model
	heroInput    selectHeroModel
	focusedField addPlayerFocusedField
	Submitted    bool
	Canceled     bool
}

func newAddPlayerForm(useExtension bool) addPlayerForm {
	ti := textinput.New()
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 26
	ti.SetWidth(26)

	return addPlayerForm{
		nameInput:    ti,
		heroInput:    newSelectHeroModel(useExtension),
		focusedField: focusedAddPlayerName,
	}
}

func (f addPlayerForm) Update(msg tea.Msg) (addPlayerForm, tea.Cmd) {
	var cmd tea.Cmd

	switch f.focusedField {
	case focusedAddPlayerName:
		f.nameInput, cmd = f.nameInput.Update(msg)
	case focusedAddPlayerHero:
		f.heroInput, cmd = f.heroInput.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			f.Canceled = true

			return f, nil

		case "tab", "shift+tab":
			if f.focusedField == focusedAddPlayerName {
				f.focusedField = focusedAddPlayerHero
				f.nameInput.Blur()
			} else {
				f.focusedField = focusedAddPlayerName
				cmd = f.nameInput.Focus()
			}
			return f, cmd

		case "enter":
			name := strings.TrimSpace(f.nameInput.Value())
			if f.focusedField == focusedAddPlayerName && name != "" {
				f.focusedField = focusedAddPlayerHero

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

func (f addPlayerForm) View() string {
	var s strings.Builder
	s.WriteString("=== Add New Player ===\n\n")

	// Name field
	var nameField strings.Builder
	nameField.WriteString("Player Name:\n")
	nameField.WriteString(f.nameInput.View())
	if f.focusedField == focusedAddPlayerName {
		s.WriteString(focusedModelStyle.Render(nameField.String()))
	} else {
		s.WriteString(modelStyle.Render(nameField.String()))
	}

	s.WriteString("\n\n")

	// Class selector
	var classField strings.Builder
	classField.WriteString(f.heroInput.View())
	if f.focusedField == focusedAddPlayerHero {
		s.WriteString(focusedModelStyle.Render(classField.String()))
	} else {
		s.WriteString(modelStyle.Render(classField.String()))
	}

	s.WriteString("\n\n")

	s.WriteString("<Enter> Submit  •  <Esc> Cancel\n")

	return s.String()
}
