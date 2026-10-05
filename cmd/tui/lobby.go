package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type lobbyState int

const (
	lobbyStateMenu lobbyState = iota
	lobbyStateAddPlayer
	lobbyStateChangeHero
)

type LobbyAction uint8

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
	state               lobbyState
	cursorIdx           LobbyAction
	estimatedDifficulty uint8
	config              game.Config

	addPlayerForm  addPlayerForm
	changeHeroForm changeHeroForm

	controller client.GameController
	ctx        context.Context
	snapshot   game.GameSnapshotDTO
}

func newLobbyModel() lobbyModel {
	return lobbyModel{
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
			if m.cursorIdx > 0 {
				m.cursorIdx--
				break
			}
			m.cursorIdx = ChangeHeroAction

		case "down":
			if m.cursorIdx < ChangeHeroAction {
				m.cursorIdx++
				break
			}
			m.cursorIdx = StartGameAction

		case "1", "2", "3":
			idx, err := strconv.Atoi(msg.String())
			if err != nil {
				return m, nil
			}
			if idx > 3 {
				return m, nil
			}
			m.cursorIdx = LobbyAction(idx - 1)

			return m.triggerCommand()

		case "space", "enter":
			return m.triggerCommand()
		}
	}

	return m, nil
}

func (m lobbyModel) triggerCommand() (lobbyModel, tea.Cmd) {
	switch m.cursorIdx {
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
	var box strings.Builder
	switch m.state {
	case lobbyStateMenu:
		box.WriteString(lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderHeader(),
			m.renderPlayersTable(),
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				m.renderCommands(),
				m.renderBriefing(),
			),
			m.renderFooter(),
		))

	case lobbyStateAddPlayer:
		box.WriteString(m.addPlayerForm.View())

	case lobbyStateChangeHero:
		box.WriteString(m.changeHeroForm.View())
	}

	return box.String()
}

func (m lobbyModel) renderHeader() string {

	var setupHeaderCol1 strings.Builder
	setupHeaderCol1.WriteString("GAME LOBBY: Expedition Prep Room\n")
	setupHeaderCol1.WriteString("MODE: Local Hot-seat / Co-op")

	var setupHeaderCol2 strings.Builder
	setupHeaderCol2.WriteString(fmt.Sprintf("DUNGEON NEXT STAGE: Floor %d\n", m.snapshot.Level))
	setupHeaderCol2.WriteString(fmt.Sprintf("DIFFICULTY: %s", styleDifficulty(m.estimatedDifficulty)))

	var setupHeaderCol3 strings.Builder
	maxNbPlayers := 5
	if m.config.UseExtension {
		maxNbPlayers = 6
	}
	setupHeaderCol3.WriteString(fmt.Sprintf("PARTY STATUS: %d / %d\n", len(m.snapshot.Players), maxNbPlayers))
	setupHeaderCol3.WriteString("MODE: Local Session hot-seat")

	var setupHeaderBox strings.Builder
	setupHeaderBox.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Left,
		setupHeaderColStyle.Render(setupHeaderCol1.String()),
		setupHeaderColStyle.Render(setupHeaderCol2.String()),
		setupHeaderLastColStyle.Render(setupHeaderCol3.String()),
	))

	return setupHeaderStyle.Render(setupHeaderBox.String())
}

func (m lobbyModel) renderPlayersTable() string {
	maxNbPlayers := 5
	if m.config.UseExtension {
		maxNbPlayers = 6
	}

	var rows [][]string
	for i := 0; i < maxNbPlayers; i++ {
		if i >= len(m.snapshot.Players) {
			rows = append(rows, []string{strconv.Itoa(i + 1), "[Open Slot]", "---", "---", "---", "WAITING"})
			continue
		}
		p := m.snapshot.Players[i]
		rows = append(rows, []string{strconv.Itoa(i + 1), p.Name, p.HeroName, p.AbilityDescription, heroToDeckType(p.HeroClass), "READY ✅"})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.White)).
		BorderRight(false).BorderTop(false).BorderLeft(false).BorderBottom(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				switch col {
				case 0:
					return lipgloss.NewStyle().Width(4).Padding(0, 1)
				case 1:
					return lipgloss.NewStyle().Width(20).Padding(0, 1)
				case 2:
					return lipgloss.NewStyle().Width(13).Padding(0, 1)
				case 3:
					return lipgloss.NewStyle().Width(60).Padding(0, 1)
				case 4:
					return lipgloss.NewStyle().Width(17).Padding(0, 1)
				case 5:
					return lipgloss.NewStyle().Width(11).Padding(0, 1)
				}
			}

			return lipgloss.NewStyle().Padding(0, 1)
		}).
		Headers("#", "Player Name", "Hero Class", "Special Ability", "Deck Type", "Status").
		Rows(rows...).
		Width(screenWidth - 2)

	var lobbyPlayersBoxTitle strings.Builder
	lobbyPlayersBoxTitle.WriteString("┌── 🛡️ ADVENTURING PARTY ROSTER ")
	lobbyPlayersBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth)-len([]rune(lobbyPlayersBoxTitle.String()))-1)) + "┐")

	return lipgloss.JoinVertical(lipgloss.Left,
		lobbyPlayersBoxTitle.String(),
		lobbyPlayersStyle.Render(t.Render()),
	)
}

func (m lobbyModel) renderCommands() string {
	var lobbyCommandsBoxTitle strings.Builder
	lobbyCommandsBoxTitle.WriteString("┌── LOBBY COMMAND CENTER ")
	lobbyCommandsBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2-1)-len([]rune(lobbyCommandsBoxTitle.String()))-1)) + "┐")

	var lobbyCommandsBox strings.Builder
	pointer := "  "
	if m.cursorIdx == StartGameAction {
		pointer = "► "
	}
	lobbyCommandsBox.WriteString(fmt.Sprintf("%s[1] START EXPEDITION\n", pointer))

	pointer = "  "
	if m.cursorIdx == AddPlayerAction {
		pointer = "► "
	}
	lobbyCommandsBox.WriteString(fmt.Sprintf("%s[2] Add New Player (Hotseat)\n", pointer))

	pointer = "  "
	if m.cursorIdx == ChangeHeroAction {
		pointer = "► "
	}
	lobbyCommandsBox.WriteString(fmt.Sprintf("%s[3] Change Hero Class", pointer))

	return lipgloss.JoinVertical(lipgloss.Left,
		lobbyCommandsBoxTitle.String(),
		lobbyCommandsStyle.Render(lobbyCommandsBox.String()),
	)
}

func (m lobbyModel) renderBriefing() string {
	var lobbyBriefingBoxTitle strings.Builder
	lobbyBriefingBoxTitle.WriteString("┌── EXPEDITION BRIEFING & HINTS ")
	lobbyBriefingBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2-1)-len([]rune(lobbyBriefingBoxTitle.String()))-1)) + "┐")

	var lobbyBriefingBox strings.Builder
	lobbyBriefingBox.WriteString("📌 Quick Rules:\n")
	lobbyBriefingBox.WriteString("• Work in real-time to match door requirements\n")
	lobbyBriefingBox.WriteString("• Hero abilities cost 3 discarded cards to trigger\n")
	lobbyBriefingBox.WriteString(fmt.Sprintf("• You have %ds to react to an event\n", m.config.EventReactionTime))
	lobbyBriefingBox.WriteString("• Communicate to beat the boss before the time runs out!\n\n")
	lobbyBriefingBox.WriteString("💡 Tip: Played card are permanently lost!")

	return lipgloss.JoinVertical(lipgloss.Left,
		lobbyBriefingBoxTitle.String(),
		lobbyBriefingStyle.Render(lobbyBriefingBox.String()),
	)
}

func (m lobbyModel) renderFooter() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		helpStyle.Render("  [↑/↓] Navigate Commands   •   [ENTER/SPACE] Select Command   •   [1-3] Quick Select   •   [Ctrl+C] Quit"),
	)
}

func heroToDeckType(class game.HeroClass) string {
	switch class {
	case game.Sorceress, game.Wizard:
		return "Magic/Control"
	case game.Huntress, game.Ranger:
		return "Range/Versatile"
	case game.Ninja, game.Thief:
		return "Melee/Control"
	case game.Paladin, game.Valkyrie:
		return "Tank/Support"
	case game.Barbarian, game.Gladiator:
		return "Melee/Power"
	case game.Druid, game.Shaman:
		return "Versatile/Heal"
	}

	return "Standard"
}
