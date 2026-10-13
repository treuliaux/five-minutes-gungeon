package tui

import (
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	toastDuration = 2500 * time.Millisecond
	toastWidth    = 56
	toastTop      = 2
)

var toastStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("196")).
	Foreground(lipgloss.Color("231")).
	Padding(0, 2).
	Width(toastWidth)

// toast is a short-lived error message floating over the screen. The id lets an expiry tick ignore itself when a
// newer toast has replaced the one it was started for.
type toast struct {
	id   int
	text string
}

func (t toast) visible() bool {
	return t.text != ""
}

func (t toast) render() string {
	return toastStyle.Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render(icon(iconWarning)+" ERROR") + "\n" + t.text)
}

// localError raises a toast for an invalid action that was caught client-side, through the same path as game errors.
func localError(text string) tea.Cmd {
	return forwardError(errors.New(text))
}

func expireToast(id int) tea.Cmd {
	return tea.Tick(toastDuration, func(time.Time) tea.Msg {
		return toastExpiredMsg{id: id}
	})
}
