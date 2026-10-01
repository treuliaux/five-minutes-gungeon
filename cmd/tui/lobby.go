package main

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type lobbyState int

const (
	lobbyStateMenu lobbyState = iota
	lobbyStateAddPlayer
	lobbyStateChangeHero
)

type LobbyAction int

const (
	StartGameAction LobbyAction = iota
	AddPlayerAction
	ChangeHeroAction
)

func (a LobbyAction) String() string {
	switch a {
	case StartGameAction:
		return "Start Game"
	case AddPlayerAction:
		return "Add Player"
	case ChangeHeroAction:
		return "Change Hero"
	default:
		return "Unknown Action"
	}
}

type lobbyModel struct {
	state        lobbyState
	cursorIdx    int
	lobbyActions []LobbyAction

	addPlayerForm  addPlayerForm
	changeHeroForm changeHeroForm

	controller client.GameController
	ctx        context.Context
	snapshot   game.GameSnapshotDTO
}

func newLobbyModel() lobbyModel {
	lobbyActions := make([]LobbyAction, 3)
	lobbyActions[0] = AddPlayerAction
	lobbyActions[1] = StartGameAction
	lobbyActions[2] = ChangeHeroAction

	return lobbyModel{
		lobbyActions:   lobbyActions,
		addPlayerForm:  newAddPlayerForm(),
		changeHeroForm: newChangeHeroForm(nil),
	}
}

func (m lobbyModel) Update(msg tea.Msg) (lobbyModel, tea.Cmd) {
	switch m.state {
	case lobbyStateMenu:
		return m.updateMenu(msg)

	case lobbyStateAddPlayer:
		return m.updateAddPlayerForm(msg)

	case lobbyStateChangeHero:
		return m.updateChangeHeroForm(msg)
	}

	return m, nil
}

func (m lobbyModel) updateMenu(msg tea.Msg) (lobbyModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up":
			m.cursorIdx = (m.cursorIdx - 1) % len(m.lobbyActions)
			if m.cursorIdx < 0 {
				m.cursorIdx = len(m.lobbyActions) - 1
			}

		case "down":
			m.cursorIdx = (m.cursorIdx + 1) % len(m.lobbyActions)

		case "space", "enter":
			switch m.lobbyActions[m.cursorIdx] {
			case StartGameAction:
				if err := m.controller.Dispatch(m.ctx, game.StartCmd{}); err != nil {
					return m, forwardError(err)
				}
			case AddPlayerAction:
				m.state = lobbyStateAddPlayer
				m.addPlayerForm = newAddPlayerForm()

				return m, m.addPlayerForm.nameInput.Focus()
			case ChangeHeroAction:
				if len(m.snapshot.Players) == 0 {
					break
				}
				playerIDs := make([]game.PlayerID, len(m.snapshot.Players))
				for i, player := range m.snapshot.Players {
					playerIDs[i] = player.Id
				}

				m.state = lobbyStateChangeHero
				m.changeHeroForm = newChangeHeroForm(playerIDs)

				return m, nil
			}
		}
	}

	return m, nil
}

func (m lobbyModel) updateAddPlayerForm(msg tea.Msg) (lobbyModel, tea.Cmd) {
	var cmd tea.Cmd

	m.addPlayerForm, cmd = m.addPlayerForm.Update(msg)
	if m.addPlayerForm.Canceled {
		m.state = lobbyStateMenu
		m.addPlayerForm = newAddPlayerForm()

		return m, nil
	}
	if m.addPlayerForm.Submitted {
		m.state = lobbyStateMenu
		addCmd := game.AddPlayerCmd{Name: m.addPlayerForm.nameInput.Value(), Class: m.addPlayerForm.heroInput.choice}
		if err := m.controller.Dispatch(m.ctx, addCmd); err != nil {
			return m, forwardError(err)
		}
		m.addPlayerForm = newAddPlayerForm()

		return m, nil
	}

	return m, cmd
}

func (m lobbyModel) updateChangeHeroForm(msg tea.Msg) (lobbyModel, tea.Cmd) {
	var cmd tea.Cmd

	m.changeHeroForm, cmd = m.changeHeroForm.Update(msg)
	if m.changeHeroForm.Canceled {
		m.state = lobbyStateMenu

		return m, nil
	}
	if m.changeHeroForm.Submitted {
		m.state = lobbyStateMenu
		addCmd := game.ChangeHeroCmd{PlayerID: m.changeHeroForm.playerInput.choice, Class: m.changeHeroForm.heroInput.choice}
		if err := m.controller.Dispatch(m.ctx, addCmd); err != nil {
			return m, forwardError(err)
		}

		return m, nil
	}

	return m, cmd
}

func (m lobbyModel) View() string {
	var s strings.Builder

	// Render Lobby Header & Current Player Roster from Snapshot
	s.WriteString("=== Game Lobby ===\n\n")
	s.WriteString("Connected Players:\n")
	if len(m.snapshot.Players) == 0 {
		s.WriteString("  (No players added yet)\n")
	} else {
		for _, p := range m.snapshot.Players {
			s.WriteString(fmt.Sprintf("  • %s (%s)\n", p.Name, p.HeroClass))
		}
	}
	s.WriteString("\n------------------------------------\n\n")

	switch m.state {
	case lobbyStateMenu:
		for i, option := range m.lobbyActions {
			cursor := " "
			if m.cursorIdx == i {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, option))
		}
		s.WriteString("\n<Enter> Select Action  •  <q> Quit\n")

	case lobbyStateAddPlayer:
		s.WriteString(m.addPlayerForm.View())

	case lobbyStateChangeHero:
		s.WriteString(m.changeHeroForm.View())
	}

	return s.String()
}
