package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type Model struct {
	controller client.GameController
	ctx        context.Context
	cancel     context.CancelFunc
	step       RunningStep

	setup setupModel
	lobby lobbyModel
	play  playModel

	spinner               spinner.Model
	lastReceivedEvents    []timedEntry[game.Event]
	lastReceivedSnapshots []timedEntry[game.GameSnapshot]
	lastReceivedErrors    []timedEntry[error]
	startTime             time.Time
}

type timedEntry[T any] struct {
	data       T
	receivedAt time.Duration
	valid      bool
}

type RunningStep int

const (
	Setup RunningStep = iota
	Lobby
	Playing
)

func NewModel() Model {
	s := spinner.New()
	s.Style = spinnerStyle

	return Model{
		setup:                 newSetupModel(),
		lobby:                 newLobbyModel(),
		play:                  newPlayModel(),
		spinner:               s,
		lastReceivedEvents:    make([]timedEntry[game.Event], 5),
		lastReceivedSnapshots: make([]timedEntry[game.GameSnapshot], 5),
		lastReceivedErrors:    make([]timedEntry[error], 5),
		startTime:             time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.String()
		if k == "ctrl+c" {
			if m.controller != nil {
				_ = m.controller.Close()
			}
			if m.cancel != nil {
				m.cancel()
			}

			return m, tea.Quit
		}

	case sessionStartedMsg:
		m.controller = msg.controller
		m.ctx = msg.ctx
		m.cancel = msg.cancel
		m.step = Lobby

		m.lobby.controller = msg.controller
		m.lobby.ctx = msg.ctx
		m.lobby.config = msg.config
		m.lobby.estimatedDifficulty = msg.estimatedDifficulty

		m.play.controller = msg.controller
		m.play.ctx = msg.ctx
		m.play.config = msg.config

		// Wait for first event
		return m, tea.Batch(askForSnapshot(m.controller), waitForEvent(m.controller.Events()))

	case backToLobbyMsg:
		m.step = Lobby
		return m, askForSnapshot(m.controller)

	case errMsg:
		// TODO: Display error to user
		entry := timedEntry[error]{data: msg.err, receivedAt: time.Since(m.startTime), valid: true}
		copy(m.lastReceivedErrors, m.lastReceivedErrors[1:])
		m.lastReceivedErrors[len(m.lastReceivedErrors)-1] = entry

		return m, nil

	// Treat incoming game event and wait for next one
	case gameEventMsg:
		m.handleDebugging(msg)

		if _, ok := msg.(game.GameStartedEvent); ok {
			m.step = Playing
			m.play = m.play.Init()
		}
		cmds := []tea.Cmd{waitForEvent(m.controller.Events())}

		var cmd tea.Cmd
		switch m.step {
		case Lobby:
			m.lobby, cmd = m.lobby.Update(msg)
			cmds = append(cmds, askForSnapshot(m.controller))
		case Playing:
			switch ev := msg.(type) {
			case game.TimerTickEvent:
				m.play.timer = ev.TimeLeftDuration
			default:
				m.play, cmd = m.play.Update(msg)
				cmds = append(cmds, askForSnapshot(m.controller))
			}
		default:
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case snapshotMsg:
		entry := timedEntry[game.GameSnapshot]{data: msg, receivedAt: time.Since(m.startTime), valid: true}
		copy(m.lastReceivedSnapshots, m.lastReceivedSnapshots[1:])
		m.lastReceivedSnapshots[len(m.lastReceivedSnapshots)-1] = entry
		switch m.step {
		case Lobby:
			snapshot, ok := msg.(game.GameSnapshotDTO)
			if !ok {
				return m, nil
			}
			m.lobby.snapshot = snapshot
		case Playing:
			snapshot, ok := msg.(game.GameSnapshotDTO)
			if !ok {
				return m, nil
			}
			m.play.snapshot = snapshot
		default:
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd
	}

	var cmd tea.Cmd
	switch m.step {
	case Setup:
		m.setup, cmd = m.setup.Update(msg)
	case Lobby:
		m.lobby, cmd = m.lobby.Update(msg)
	case Playing:
		m.play, cmd = m.play.Update(msg)
	}

	return m, cmd
}

func (m Model) View() tea.View {
	debug := m.debugView()

	var screen strings.Builder
	switch m.step {
	case Setup:
		screen.WriteString(m.setup.View())
	case Lobby:
		screen.WriteString(m.lobby.View())
	case Playing:
		screen.WriteString(m.play.View())
	default:
		screen.WriteString(fmt.Sprintf("Model '%v' not implemented", m.step))
	}

	var total strings.Builder
	total.WriteString(lipgloss.JoinVertical(lipgloss.Left, debug.String(), screen.String()))

	view := tea.NewView(appStyle.Render(total.String()))
	view.AltScreen = true

	return view
}

func (m Model) handleDebugging(msg gameEventMsg) {
	if _, ok := msg.(game.TimerTickEvent); !ok {
		entry := timedEntry[game.Event]{data: msg, receivedAt: time.Since(m.startTime), valid: true}
		copy(m.lastReceivedEvents, m.lastReceivedEvents[1:])
		m.lastReceivedEvents[len(m.lastReceivedEvents)-1] = entry
	}
	if ev, ok := msg.(game.CardPlayedEvent); ok {
		copy(m.play.lastPlayerActions, m.play.lastPlayerActions[1:])
		m.play.lastPlayerActions[len(m.play.lastPlayerActions)-1] = PlayerActionEntry{CardPlayed: ev}
	}
	if ev, ok := msg.(game.HeroAbilityUsedEvent); ok {
		copy(m.play.lastPlayerActions, m.play.lastPlayerActions[1:])
		m.play.lastPlayerActions[len(m.play.lastPlayerActions)-1] = PlayerActionEntry{AbilityUsed: ev}
	}
}
