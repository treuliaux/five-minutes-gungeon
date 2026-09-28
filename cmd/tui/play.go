package main

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type PlayAction int

const (
	PlayCardAction PlayAction = iota
	DiscardCardsAction
	UseHeroAbilityAction
	SubmitPromptChoiceAction
	UseArtifactAction
)

func (a PlayAction) String() string {
	switch a {
	case PlayCardAction:
		return "Play Card"
	case DiscardCardsAction:
		return "Discard Cards"
	case UseHeroAbilityAction:
		return "Use Hero Ability"
	case SubmitPromptChoiceAction:
		return "Submit Prompt Choice"
	case UseArtifactAction:
		return "Use Artifact"
	default:
		return "Unknown Action"
	}
}

type playModel struct {
	cursorIdx   int
	playActions []PlayAction

	controller client.GameController
	ctx        context.Context

	snapshot game.GameSnapshot
}

func newPlayModel() playModel {
	playActions := make([]PlayAction, 5)
	playActions[0] = PlayCardAction
	playActions[1] = DiscardCardsAction
	playActions[2] = UseHeroAbilityAction
	playActions[3] = SubmitPromptChoiceAction
	playActions[4] = UseArtifactAction

	return playModel{
		cursorIdx:   0,
		playActions: playActions,
	}
}

func (m playModel) Update(msg tea.Msg) (playModel, tea.Cmd) {
	var cmds []tea.Cmd

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up":
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}
		case "down":
			if m.cursorIdx < len(m.playActions)-1 {
				m.cursorIdx++
			}
		case "space", "enter":
			// execute action
		}
	}

	return m, tea.Batch(cmds...)
}

func (m playModel) View() string {
	var s strings.Builder

	s.WriteString("Dungeon Playfield\n\n")

	for i, option := range m.playActions {
		cursor := " "
		if m.cursorIdx == i {
			cursor = ">"
		}

		s.WriteString(fmt.Sprintf("%s %s\n", cursor, option))
	}

	s.WriteString("\nq to quit - <enter> or <space> to select action\n")

	return s.String()
}
