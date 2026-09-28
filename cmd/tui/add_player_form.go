package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type addPlayerFocusedField int

const (
	focusedName addPlayerFocusedField = iota
	focusedHero
)

var (
	modelStyle = lipgloss.NewStyle().
			Width(30).
			Height(3).
			Padding(0, 1).
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(30).
				Height(3).
				Padding(0, 1).
				Align(lipgloss.Left, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
)

type addPlayerForm struct {
	nameInput    textinput.Model
	heroInput    heroModel
	focusedField addPlayerFocusedField
	Submitted    bool
	Canceled     bool
}

func newAddPlayerForm() addPlayerForm {
	ti := textinput.New()
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 26
	ti.SetWidth(26)

	return addPlayerForm{
		nameInput:    ti,
		heroInput:    newHeroModel(),
		focusedField: focusedName,
	}
}

func (f addPlayerForm) Update(msg tea.Msg) (addPlayerForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			f.Canceled = true

			return f, nil

		case "tab", "shift+tab":
			if f.focusedField == focusedName {
				f.focusedField = focusedHero
				f.nameInput.Blur()
			} else {
				f.focusedField = focusedName
				cmd = f.nameInput.Focus()
			}
			return f, cmd

		case "enter":
			name := strings.TrimSpace(f.nameInput.Value())
			if f.focusedField == focusedName && name != "" {
				f.focusedField = focusedHero

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

	switch f.focusedField {
	case focusedName:
		f.nameInput, cmd = f.nameInput.Update(msg)
	case focusedHero:
		f.heroInput, cmd = f.heroInput.Update(msg)
	}

	return f, cmd
}

func (f addPlayerForm) View() string {
	var s strings.Builder
	s.WriteString("=== Add New Player ===\n\n")
	//s.WriteString(fmt.Sprintf("STEP: %v\n", f.focusedField))

	// Name field
	var nameField strings.Builder
	nameField.WriteString("Player Name:\n")
	nameField.WriteString(f.nameInput.View())
	if f.focusedField == focusedName {
		s.WriteString(focusedModelStyle.Render(nameField.String()))
	} else {
		s.WriteString(modelStyle.Render(nameField.String()))
	}

	s.WriteString("\n\n")

	// Class selector
	var classField strings.Builder
	classField.WriteString(f.heroInput.View())
	if f.focusedField == focusedHero {
		s.WriteString(focusedModelStyle.Render(classField.String()))
	} else {
		s.WriteString(modelStyle.Render(classField.String()))
	}

	s.WriteString("\n\n")

	s.WriteString("<Enter> Submit  •  <Esc> Cancel\n")

	return s.String()
}
