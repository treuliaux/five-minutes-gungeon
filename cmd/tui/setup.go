package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type optionName string

const (
	UseExtension       optionName = "UseExtension"
	ResetLevelOnDefeat optionName = "ResetLevelOnDefeat"
)

type gameOption struct {
	Name  optionName
	Value bool
}

type setupModel struct {
	cursorIdx    int
	setupOptions []gameOption
}

func newSetupModel() setupModel {
	setupOptions := make([]gameOption, 2)
	setupOptions[0] = gameOption{
		Name:  UseExtension,
		Value: false,
	}
	setupOptions[1] = gameOption{
		Name:  ResetLevelOnDefeat,
		Value: true,
	}

	return setupModel{
		cursorIdx:    0,
		setupOptions: setupOptions,
	}
}

func (m setupModel) Update(msg tea.Msg) (setupModel, tea.Cmd) {
	var cmds []tea.Cmd

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up":
			if m.cursorIdx > 0 {
				m.cursorIdx--
			} else {
				m.cursorIdx = len(m.setupOptions) - 1
			}
		case "down":
			if m.cursorIdx < len(m.setupOptions)-1 {
				m.cursorIdx++
			} else {
				m.cursorIdx = 0
			}
		case "space":
			m.setupOptions[m.cursorIdx].Value = !m.setupOptions[m.cursorIdx].Value
		case "enter":
			cmds = append(cmds, []tea.Cmd{startSession(m.buildGameConfig()), tea.ClearScreen}...)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m setupModel) View() string {
	var s strings.Builder
	s.WriteString("Game configuration\n\n")

	for i, option := range m.setupOptions {
		cursor := " "
		if m.cursorIdx == i {
			cursor = ">"
		}

		checked := "❌"
		if option.Value {
			checked = "✅"
		}

		s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, option.Name, checked))
	}

	s.WriteString("\nq to quit - <space> to toggle - <enter> to enter lobby\n")

	return s.String()
}

func (m setupModel) buildGameConfig() game.Config {
	cfg := game.Config{}
	for _, opt := range m.setupOptions {
		switch opt.Name {
		case UseExtension:
			cfg.UseExtension = opt.Value
		case ResetLevelOnDefeat:
			cfg.ResetLevelOnDefeat = opt.Value
		}
	}
	cfg.EventReactionTime = 10

	return cfg
}
