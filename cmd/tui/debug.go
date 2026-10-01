package main

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

func formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	sec := int(d / time.Second)
	ms := int((d % time.Second) / time.Millisecond)
	return fmt.Sprintf("%ds%dms", sec, ms)
}

func formatDebugRow(left, right string, totalWidth int) string {
	leftRunes := []rune(left)
	rightRunes := []rune(right)

	availableForLeft := max(totalWidth-len(rightRunes)-1, 0)

	if len(leftRunes) > availableForLeft {
		leftRunes = leftRunes[:availableForLeft]
	}

	dotCount := max(totalWidth-len(leftRunes)-len(rightRunes), 0)

	return string(leftRunes) + strings.Repeat(".", dotCount) + string(rightRunes)
}

func (m Model) debugView() strings.Builder {
	const panelWidth = 40

	// 1. Events Panel
	var eventsView strings.Builder
	eventsView.WriteString(m.spinner.View())
	eventsView.WriteString(" Receiving events...\n\n")
	for _, entry := range m.lastReceivedEvents {
		if !entry.valid {
			eventsView.WriteString(dotStyle.Render(strings.Repeat(".", panelWidth)) + "\n")
			continue
		}
		left := fmt.Sprintf("%T", entry.data)
		right := formatDuration(entry.receivedAt)
		eventsView.WriteString(dotStyle.Render(formatDebugRow(left, right, panelWidth)) + "\n")
	}
	eventsView.WriteString("\n")

	// 2. Snapshots Panel
	var snapshotsView strings.Builder
	snapshotsView.WriteString(m.spinner.View())
	snapshotsView.WriteString(" Receiving snapshots...\n\n")
	for _, entry := range m.lastReceivedSnapshots {
		if !entry.valid {
			snapshotsView.WriteString(dotStyle.Render(strings.Repeat(".", panelWidth)) + "\n")
			continue
		}
		left := fmt.Sprintf("%T", entry.data)
		right := formatDuration(entry.receivedAt)
		snapshotsView.WriteString(dotStyle.Render(formatDebugRow(left, right, panelWidth)) + "\n")
	}
	snapshotsView.WriteString("\n")

	// 3. Errors Panel
	var errorsView strings.Builder
	errorsView.WriteString(m.spinner.View())
	errorsView.WriteString(" Receiving errors...\n\n")
	for _, entry := range m.lastReceivedErrors {
		if !entry.valid {
			errorsView.WriteString(errorStyle.Render(strings.Repeat(".", panelWidth)) + "\n")
			continue
		}
		left := entry.data.Error()
		right := formatDuration(entry.receivedAt)
		errorsView.WriteString(errorStyle.Render(formatDebugRow(left, right, panelWidth)) + "\n")
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
