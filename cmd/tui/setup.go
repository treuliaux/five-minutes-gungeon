package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type optionIdx uint8

const (
	UseExtension optionIdx = iota
	ResetLevelOnDefeat
	ReactionWindow
)

type setupModel struct {
	cursorIdx           optionIdx
	useExtension        bool
	resetOnDefeat       bool
	reactionTimeSec     int
	estimatedDifficulty uint8
}

func newSetupModel() setupModel {
	return setupModel{
		cursorIdx:       0,
		useExtension:    false,
		resetOnDefeat:   true,
		reactionTimeSec: 10,
	}
}

func (m setupModel) Update(msg tea.Msg) (setupModel, tea.Cmd) {
	var cmds []tea.Cmd

	m.estimatedDifficulty = m.estimateDifficulty()
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up":
			if m.cursorIdx > 0 {
				m.cursorIdx--
				break
			}
			m.cursorIdx = ReactionWindow

		case "down":
			if m.cursorIdx < ReactionWindow {
				m.cursorIdx++
				break
			}
			m.cursorIdx = 0

		case "left":
			switch m.cursorIdx {
			case UseExtension:
				m.useExtension = !m.useExtension
			case ResetLevelOnDefeat:
				m.resetOnDefeat = !m.resetOnDefeat
			case ReactionWindow:
				if m.reactionTimeSec > 2 {
					m.reactionTimeSec--
				}
			}

		case "right":
			switch m.cursorIdx {
			case UseExtension:
				m.useExtension = !m.useExtension
			case ResetLevelOnDefeat:
				m.resetOnDefeat = !m.resetOnDefeat
			case ReactionWindow:
				if m.reactionTimeSec < 30 {
					m.reactionTimeSec++
				}
			}

		case "space":
			switch m.cursorIdx {
			case UseExtension:
				m.useExtension = !m.useExtension
			case ResetLevelOnDefeat:
				m.resetOnDefeat = !m.resetOnDefeat
			case ReactionWindow:
				m.reactionTimeSec += 5 - (m.reactionTimeSec+5)%5
				if m.reactionTimeSec > 30 {
					m.reactionTimeSec = 2
				}
			}

		case "enter":
			cmds = append(cmds, startSession(m.buildGameConfig(), m.estimatedDifficulty))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m setupModel) View() string {
	var box strings.Builder
	box.WriteString(lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderSetupRules(),
		m.renderSessionSummary(),
		m.renderFooter(),
	))

	return box.String()
}

func (m setupModel) renderHeader() string {

	var setupHeaderCol1 strings.Builder
	setupHeaderCol1.WriteString("GAME CONFIGURATION & MODIFIERS\n")
	setupHeaderCol1.WriteString("Customize your run rules & timers")

	var setupHeaderCol2 strings.Builder
	setupHeaderCol2.WriteString("SETUP WIZARD\n")
	setupHeaderCol2.WriteString("STATUS: Configuration in Progress")

	var setupHeaderCol3 strings.Builder
	setupHeaderCol3.WriteString("GUNGEON ENGINE v1.0\n")
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

func (m setupModel) renderSetupRules() string {
	var setupRulesBoxTitle strings.Builder
	setupRulesBoxTitle.WriteString("┌── ⚙️ DUNGEON SESSION RULES ")
	setupRulesBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth)-lipgloss.Width(setupRulesBoxTitle.String())-1)) + "┐")

	var setupRulesBox strings.Builder
	setupRulesBox.WriteString("[1] Extension Cards, Classes, & Advanced Bosses\n")
	pointer := "   "
	enabledIcon := "✅"
	enabledText := "ENABLED "
	if m.cursorIdx == UseExtension {
		pointer = " ► "
	}
	if m.useExtension == false {
		enabledIcon = "❌"
		enabledText = "DISABLED"
	}
	setupRulesBox.WriteString(fmt.Sprintf("%s [ %s %s ]\n", pointer, enabledIcon, enabledText))
	setupRulesBox.WriteString("    Whether to include game extension \"Curses! Foiled again!\" or not.\n\n")

	setupRulesBox.WriteString("[2] Reset Level Progress on Defeat\n")
	pointer = "   "
	enabledIcon = "✅"
	enabledText = "ENABLED "
	if m.cursorIdx == ResetLevelOnDefeat {
		pointer = " ► "
	}
	if m.resetOnDefeat == false {
		enabledIcon = "❌"
		enabledText = "DISABLED"
	}
	setupRulesBox.WriteString(fmt.Sprintf("%s [ %s %s ]\n", pointer, enabledIcon, enabledText))
	setupRulesBox.WriteString("    Whether or not failing a floor will make you start again at level 1.\n\n")

	setupRulesBox.WriteString("[3] Event Reaction Window\n")
	pointer = "   "
	if m.cursorIdx == ReactionWindow {
		pointer = " ► "
	}
	setupRulesBox.WriteString(fmt.Sprintf("%s [ ◄   %2d SECONDS   ► ]\n", pointer, m.reactionTimeSec))
	setupRulesBox.WriteString("    The time you have to react to an event.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		setupRulesBoxTitle.String(),
		setupRulesStyle.Render(setupRulesBox.String()),
	)
}

func (m setupModel) renderSessionSummary() string {
	var setupSummaryBoxTitle strings.Builder
	setupSummaryBoxTitle.WriteString("┌── 🛠️ DUNGEON SUMMARY ")
	setupSummaryBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth)-lipgloss.Width(setupSummaryBoxTitle.String())-1)) + "┐")

	difficultyText := styleDifficulty(m.estimateDifficulty())

	maxNbPlayers := 5
	if m.useExtension {
		maxNbPlayers = 6
	}
	enabledIcon := "✅"
	enabledText := "ENABLED"
	if m.useExtension == false {
		enabledIcon = "❌"
		enabledText = "DISABLED"
	}
	enabled := fmt.Sprintf("[ %s %s ]", enabledIcon, enabledText)

	var setupSummaryBox strings.Builder
	setupSummaryBox.WriteString(fmt.Sprintf("• Estimated difficulty:     %s\n", difficultyText))
	setupSummaryBox.WriteString(fmt.Sprintf("• Players Limit:            Up to %d players\n", maxNbPlayers))
	setupSummaryBox.WriteString(fmt.Sprintf("• Druid & Shaman:           %s\n", enabled))
	setupSummaryBox.WriteString(fmt.Sprintf("• Curses:                   %s\n", enabled))
	setupSummaryBox.WriteString(fmt.Sprintf("• Bosses special abilities: %s\n", enabled))
	setupSummaryBox.WriteString(fmt.Sprintf("• Extra classes actions:    %s", enabled))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		setupSummaryBoxTitle.String(),
		setupSummaryStyle.Render(setupSummaryBox.String()),
	)
}

func (m setupModel) estimateDifficulty() uint8 {
	var estimatedDifficulty uint8
	if m.resetOnDefeat {
		estimatedDifficulty++
	}
	if m.useExtension {
		estimatedDifficulty += 2
	}
	if m.reactionTimeSec <= 3 {
		estimatedDifficulty += 2
	} else if m.reactionTimeSec < 10 {
		estimatedDifficulty++
	}

	return estimatedDifficulty
}

func (m setupModel) renderFooter() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		helpStyle.Render("  [↑/↓] Select Setting   •   [←/→/SPACE] Adjust Setting   •   [ENTER] Proceed to Lobby   •   [Ctrl+C] Quit"),
	)
}

func (m setupModel) buildGameConfig() game.Config {
	return game.Config{
		UseExtension:       m.useExtension,
		ResetLevelOnDefeat: m.resetOnDefeat,
		EventReactionTime:  m.reactionTimeSec,
	}
}
