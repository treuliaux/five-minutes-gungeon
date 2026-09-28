package main

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("88"))
	dotStyle     = helpStyle.UnsetMargins()
	appStyle     = lipgloss.NewStyle().Margin(1, 2, 0, 2)
	debugStyle   = lipgloss.NewStyle().Margin(0, 2)
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
	lastReceivedEvents    []game.Event
	lastReceivedSnapshots []game.GameSnapshot
	lastReceivedErrors    []error

	error string
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
		lastReceivedEvents:    make([]game.Event, 5),
		lastReceivedSnapshots: make([]game.GameSnapshot, 5),
		lastReceivedErrors:    make([]error, 5),
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.String()
		if k == "q" || k == "ctrl+c" {
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
		m.play.controller = msg.controller
		m.play.ctx = msg.ctx

		// Wait for first event
		return m, tea.Batch(askForSnapshot(m.controller), waitForEvent(m.controller.Events()))

	case errMsg:
		m.lastReceivedErrors = append(m.lastReceivedErrors[1:], msg.err)

		return m, nil

	// Treat incoming event and wait for next one
	case gameEventMsg:
		m.lastReceivedEvents = append(m.lastReceivedEvents[1:], msg)
		if _, ok := msg.(game.GameStartedEvent); ok {
			m.step = Playing
		}

		var cmd tea.Cmd
		switch m.step {
		case Lobby:
			m.lobby, cmd = m.lobby.Update(msg)
		case Playing:
			m.play, cmd = m.play.Update(msg)
		default:
		}

		return m, tea.Batch(cmd, askForSnapshot(m.controller), waitForEvent(m.controller.Events()))

	case snapshotMsg:
		m.lastReceivedSnapshots = append(m.lastReceivedSnapshots[1:], msg)
		m.lobby.snapshot = msg
		m.play.snapshot = msg

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

	return tea.NewView(appStyle.Render(total.String()))
}

func (m Model) debugView() strings.Builder {
	var eventsView strings.Builder
	eventsView.WriteString(m.spinner.View())
	eventsView.WriteString(" Receiving events...")
	eventsView.WriteString("\n\n")
	for _, evt := range m.lastReceivedEvents {
		if evt == nil {
			eventsView.WriteString(dotStyle.Render(strings.Repeat(".", 40)))
			eventsView.WriteString("\n")

			continue
		}
		eventsView.WriteString(dotStyle.Render(formatDebug(fmt.Sprintf("%T", evt), 40)))
		eventsView.WriteString("\n")
	}
	eventsView.WriteString("\n")

	var snapshotsView strings.Builder
	snapshotsView.WriteString(m.spinner.View())
	snapshotsView.WriteString(" Receiving snapshots...")
	snapshotsView.WriteString("\n\n")
	for _, snp := range m.lastReceivedSnapshots {
		if snp == nil {
			snapshotsView.WriteString(dotStyle.Render(strings.Repeat(".", 40)))
			snapshotsView.WriteString("\n")

			continue
		}
		snapshotsView.WriteString(dotStyle.Render(formatDebug(fmt.Sprintf("%T", snp), 40)))
		snapshotsView.WriteString("\n")
	}
	snapshotsView.WriteString("\n")

	var errorsView strings.Builder
	errorsView.WriteString(m.spinner.View())
	errorsView.WriteString(" Receiving errors...")
	errorsView.WriteString("\n\n")
	for _, err := range m.lastReceivedErrors {
		if err == nil {
			errorsView.WriteString(dotStyle.Render(strings.Repeat(".", 40)))
			errorsView.WriteString("\n")

			continue
		}
		errorsView.WriteString(errorStyle.Render(formatDebug(fmt.Sprintf("%v", err.Error()), 40)))
		errorsView.WriteString("\n")
	}
	errorsView.WriteString("\n")

	var debug strings.Builder
	debug.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Top,
		debugStyle.Render(eventsView.String()),
		debugStyle.Render(snapshotsView.String()),
		debugStyle.Render(errorsView.String()),
	))
	return debug
}
