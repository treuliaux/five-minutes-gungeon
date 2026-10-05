package main

import (
	"charm.land/lipgloss/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

const screenWidth = 132

var (
	modelStyle = lipgloss.NewStyle().
			Width(30).
			Padding(0, 1).
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(30).
				Padding(0, 1).
				Align(lipgloss.Left, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("88"))
	dotStyle     = helpStyle.UnsetMargins()
	appStyle     = lipgloss.NewStyle().Margin(1, 2, 0, 2)
	debugStyle   = lipgloss.NewStyle().Margin(0, 2)
)

var (
	resourceSwordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	resourceArrowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Bold(true)
	resourceShieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	resourceJumpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("128")).Bold(true)
	resourceScrollStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	resourceWildStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("195")).Bold(true)

	timerStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	timerLowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Blink(true)
	timerFrozenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
)

var (
	setupHeaderStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(screenWidth).Margin(0, 0, 1, 0)
	setupHeaderLastColStyle = lipgloss.NewStyle().Padding(0, 2).Width(42)
	setupHeaderColStyle     = setupHeaderLastColStyle.Border(lipgloss.NormalBorder(), false, true, false, false)

	setupRulesStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(1, 3).Width(screenWidth).Margin(0, 0, 1, 0)

	setupSummaryStyle          = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(1, 3).Width(screenWidth)
	setupEasyDifficultyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	setupNormalDifficultyStyle = lipgloss.NewStyle().Foreground(lipgloss.Green)
	setupHardDifficultyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Red)
	setupInsaneDifficultyStyle = setupHardDifficultyStyle.Bold(true).Blink(true)
)

var (
	lobbyPlayersStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width(screenWidth).Margin(0, 0, 1, 0)
	lobbyCommandsStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width((screenWidth/2)-1).Margin(0, 2, 0, 0).Padding(1)
	lobbyBriefingStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Width((screenWidth/2)-1).Margin(0, 2, 0, 0).Padding(0, 1)
)

var (
	playHudStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(screenWidth)
	playHudColStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true, false, false).Padding(0, 2).Height(2)
	playHudLastColStyle = lipgloss.NewStyle().Padding(0, 2)

	playPlayfieldStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(1)

	openedDoorsStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(11)
	doorHeaderStyle  = lipgloss.NewStyle().Bold(true)
	doorStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width((screenWidth / 2) - 4).Height(3)

	playedHistoryStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(8)

	cardStyle         = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(24).Height(7)
	selectedCardStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).BorderForeground(lipgloss.Cyan).Padding(0, 1).Width(24).Height(7)
)

var (
	customBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "├",
		BottomRight: "┤",
	}
	playerStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth)
	playerHudStyle    = lipgloss.NewStyle().Border(customBorder).BorderTop(false).Width(screenWidth)
	playerHeaderStyle = lipgloss.NewStyle().Bold(true)
)

var (
	cursesStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(4)
	cursesHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))

	artifactsStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(4)
	artifactsHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))

	pendingPromptStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(4)
	pendingPromptHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))

	bannerBoxStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).Width(screenWidth).Align(lipgloss.Center)
	victoryTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
	defeatTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
)

var (
	footerStyle = helpStyle
)

func resourceTypeToIcon(rt game.ResourceType) string {
	switch rt {
	case game.Sword:
		return "🗡️"
	case game.Arrow:
		return "🏹"
	case game.Shield:
		return "🛡️"
	case game.Jump:
		return "🦵"
	case game.Scroll:
		return "📜"
	case game.WildCard:
		return "⭐"
	case game.InfiniteSword:
		return "🗡️♾️"
	case game.InfiniteArrow:
		return "🏹♾️"
	case game.InfiniteShield:
		return "🛡️♾️"
	case game.InfiniteJump:
		return "🦵♾️"
	case game.InfiniteScroll:
		return "📜♾️"
	default:
		return ""
	}
}

func styleDifficulty(estimatedDifficulty uint8) string {
	difficultyText := ""
	switch estimatedDifficulty {
	case 0:
		difficultyText = setupEasyDifficultyStyle.Render("EASY")
	case 1, 2:
		difficultyText = setupNormalDifficultyStyle.Render("NORMAL")
	case 3, 4:
		difficultyText = setupHardDifficultyStyle.Render("HARD")
	case 5:
		difficultyText = setupInsaneDifficultyStyle.Render("INSANE")
	}

	return difficultyText
}
