package main

import "charm.land/lipgloss/v2"

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
	hudStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(screenWidth)
	hudColStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true, false, false).Padding(0, 2)
	hudLastColStyle = lipgloss.NewStyle().Padding(0, 2)
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
	cardStyle         = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(24).Height(7)
	selectedCardStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).BorderForeground(lipgloss.Cyan).Padding(0, 1).Width(24).Height(7)
)

var (
	openedDoorsStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(11)
)

var (
	playedHistoryStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(8)
)

var (
	playfieldStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width(screenWidth / 2).Height(1)
)

var (
	doorStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(false).Padding(0, 1).Width((screenWidth / 2) - 4).Height(3)
	doorHeaderStyle = lipgloss.NewStyle().Bold(true)
)

var (
	footerStyle = helpStyle
)
