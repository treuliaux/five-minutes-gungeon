package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

// promptContentWidth is the width available to the text inside the modal. It is deliberately a bit smaller than
// promptModalStyle's inner width so that wrapping is always done by us (per line) and never by the modal border.
const promptContentWidth = 60

var (
	promptHintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	promptSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Cyan).Bold(true)
	promptTitleStyle    = lipgloss.NewStyle().Bold(true)

	promptResources = []struct {
		rt   game.ResourceType
		name string
	}{
		{game.Sword, "Sword"},
		{game.Arrow, "Arrow"},
		{game.Shield, "Shield"},
		{game.Jump, "Jump"},
		{game.Scroll, "Scroll"},
	}
)

// promptIsActive reports whether the hero currently controlled has to answer the pending interaction.
func (m playModel) promptIsActive() bool {
	pi := m.snapshot.PendingInteraction
	if pi == nil || pi.Kind == game.InteractionNoInteraction {
		return false
	}

	return slices.Contains(pi.PendingPlayers, m.currentPlayer().Id)
}

func (m playModel) playerName(id game.PlayerID) string {
	for _, p := range m.snapshot.Players {
		if p.Id == id {
			return p.Name
		}
	}

	return string(id)
}

func (m playModel) donationCandidates(player game.PlayerDTO) []game.PlayerDTO {
	candidates := make([]game.PlayerDTO, 0, len(m.snapshot.Players))
	for _, pl := range m.snapshot.Players {
		if pl.Id != player.Id {
			candidates = append(candidates, pl)
		}
	}

	return candidates
}

func discardExpected(pi *game.PendingInteractionDTO, player game.PlayerDTO) int {
	expected := pi.RequiredCounts[player.Id]
	if expected == 0 {
		expected = 1
	}

	return min(expected, len(player.Hand))
}

// selectedInHand counts selected cards that really belong to the given hand (stale selections are ignored).
func (m playModel) selectedInHand(player game.PlayerDTO) int {
	n := 0
	for _, c := range player.Hand {
		if m.selectedCards[c.Id] {
			n++
		}
	}

	return n
}

// promptOptionCount is the number of selectable options of a "choice" prompt.
func (m playModel) promptOptionCount() int {
	pi := m.snapshot.PendingInteraction
	if pi == nil {
		return 0
	}

	switch pi.Kind {
	case game.InteractionTeamChoicePlayer:
		return len(m.snapshot.Players)
	case game.InteractionTeamChoiceResource:
		return len(promptResources)
	case game.InteractionTeamChoiceArtifact:
		return len(m.snapshot.Artifacts)
	case game.InteractionPlayerDonatesHand:
		return len(m.donationCandidates(m.currentPlayer()))
	case game.InteractionPlayerDiscardCards, game.InteractionNoInteraction:
	}

	return 0
}

func (m playModel) updateActivePrompt(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	pi := m.snapshot.PendingInteraction
	isDiscard := pi != nil && pi.Kind == game.InteractionPlayerDiscardCards

	switch key.String() {
	case "esc":
		m.dismissedPrompt = true

		return m, nil

	case "tab":
		return m.switchPlayer(NextPlayer), nil

	case "shift+tab":
		return m.switchPlayer(PreviousPlayer), nil

	case "up", "down":
		delta := 1
		if key.String() == "up" {
			delta = -1
		}

		if isDiscard {
			maxRows := m.maxHandRows()
			m.currentPlayerHandRow = (m.currentPlayerHandRow + delta + maxRows) % maxRows

			return m, nil
		}

		if count := m.promptOptionCount(); count > 0 {
			m.promptChoiceIdx = (m.promptChoiceIdx + delta + count) % count
		}

		return m, nil

	case "1", "2", "3", "4", "5":
		idx, err := strconv.Atoi(key.String())
		if err != nil {
			return m, nil
		}

		if isDiscard {
			return m.toggleDiscardCard(idx), nil
		}

		if idx-1 < m.promptOptionCount() {
			m.promptChoiceIdx = idx - 1
		}

		return m, nil

	case "space", "enter":
		return m.submitPromptChoice()
	}

	return m, nil
}

func (m playModel) updateDismissedPrompt(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	switch key.String() {
	case "space", "enter":
		m.dismissedPrompt = false

		return m, nil

	default:
		return m.updatePlaying(key)
	}
}

// toggleDiscardCard toggles the i-th (1-based) card of the displayed hand row, without ever selecting more cards
// than the prompt requires.
func (m playModel) toggleDiscardCard(i int) playModel {
	pi := m.snapshot.PendingInteraction
	player := m.currentPlayer()
	index := m.currentPlayerHandRow*5 + i - 1
	if pi == nil || index < 0 || index >= len(player.Hand) {
		return m
	}

	cardID := player.Hand[index].Id
	if m.selectedCards[cardID] {
		delete(m.selectedCards, cardID)

		return m
	}

	if m.selectedInHand(player) >= discardExpected(pi, player) {
		return m
	}

	m.selectedCards[cardID] = true

	return m
}

func (m playModel) renderPendingPrompt() string {
	var promptBoxTitle strings.Builder
	promptBoxTitle.WriteString("┌── PENDING TEAM PROMPTS / INTERACTIONS ")
	promptBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(promptBoxTitle.String())-1)) + "┐\n")

	pi := m.snapshot.PendingInteraction
	if pi == nil {
		return promptBoxTitle.String() + pendingPromptStyle.Render("(No pending interaction / team prompt)")
	}

	lines := []string{fmt.Sprintf("%s EVENT PROMPT: [%s]", icon(iconWarning), interactionKindToString(pi.Kind))}
	if len(pi.PendingPlayers) > 0 {
		var waiting []string
		for _, pID := range pi.PendingPlayers {
			name := m.playerName(pID)
			if pi.Kind == game.InteractionPlayerDiscardCards {
				waiting = append(waiting, fmt.Sprintf("[%s: WAITING (%d cards)]", name, pi.RequiredCounts[pID]))
			} else {
				waiting = append(waiting, fmt.Sprintf("[%s: WAITING]", name))
			}
		}
		lines = append(lines, "Pending players: "+strings.Join(waiting, " "))
	}

	switch {
	case !m.promptIsActive():
		lines = append(lines, "Action: Not pending for this hero • Press [TAB] to switch hero")
	case m.dismissedPrompt:
		lines = append(lines, "Action: Modal dismissed • Press [ENTER] / [SPACE] to re-open decision modal")
	default:
		lines = append(lines, "Action: Decision modal active • Press [ESC] to dismiss and view playfield")
	}

	return promptBoxTitle.String() + pendingPromptStyle.Render(strings.Join(lines, "\n"))
}

// renderPromptOverlay draws the decision modal centered on top of baseView.
//
// Every line of the modal is rendered on its own (never a styled string containing "\n"), wrapped to
// promptContentWidth, so lipgloss can't pad/shift the lines of a multi-line block.
func (m playModel) renderPromptOverlay(baseView string) string {
	pi := m.snapshot.PendingInteraction
	if pi == nil || m.dismissedPrompt || !m.promptIsActive() {
		return baseView
	}

	lineStyle := lipgloss.NewStyle().Width(promptContentWidth)
	var lines []string
	add := func(style lipgloss.Style, text string) {
		lines = append(lines, style.Width(promptContentWidth).Render(text))
	}
	blank := func() { lines = append(lines, "") }

	add(pendingPromptHeaderStyle, "TEAM DECISION: "+strings.ToUpper(interactionKindToString(pi.Kind)))
	blank()

	if len(pi.PendingPlayers) > 0 {
		add(promptTitleStyle, "Pending party members:")
		for _, pID := range pi.PendingPlayers {
			if pi.Kind == game.InteractionPlayerDiscardCards {
				add(lineStyle, fmt.Sprintf("  - [%s] (must discard %d card(s))", m.playerName(pID), pi.RequiredCounts[pID]))
			} else {
				add(lineStyle, fmt.Sprintf("  - [%s] (waiting)", m.playerName(pID)))
			}
		}
		blank()
	}

	player := m.currentPlayer()
	add(lineStyle, fmt.Sprintf("Active Acting Hero: %s (%s)", strings.ToUpper(player.HeroClass.String()), player.Name))
	add(promptHintStyle, "(Press [TAB] / [Shift+TAB] to switch active player)")
	blank()

	option := func(i int, text, selectedSuffix string, selected bool) {
		if selected {
			add(promptSelectedStyle, fmt.Sprintf("> [%d] %s%s", i+1, text, selectedSuffix))

			return
		}
		add(lineStyle, fmt.Sprintf("  [%d] %s", i+1, text))
	}

	switch pi.Kind {
	case game.InteractionTeamChoicePlayer:
		add(lineStyle, "Choose target party member:")
		for i, p := range m.snapshot.Players {
			option(i, fmt.Sprintf("%s (%s)", p.Name, p.HeroClass.String()), "  (SELECTED)", i == m.promptChoiceIdx)
		}

	case game.InteractionTeamChoiceResource:
		add(lineStyle, "Choose resource type:")
		for i, r := range promptResources {
			option(i, resourceTypeToIcon(r.rt)+" "+r.name, "  (SELECTED)", i == m.promptChoiceIdx)
		}

	case game.InteractionTeamChoiceArtifact:
		add(lineStyle, "Choose team artifact:")
		if len(m.snapshot.Artifacts) == 0 {
			add(promptHintStyle, "  (No artifact available)")
		}
		for i, a := range m.snapshot.Artifacts {
			option(i, fmt.Sprintf("[%s] %s: \"%s\"", a.Color, a.Name, a.Description), "  (SELECTED)", i == m.promptChoiceIdx)
		}

	case game.InteractionPlayerDiscardCards:
		expected := discardExpected(pi, player)
		add(lineStyle, fmt.Sprintf("Select %d card(s) from hand to discard:", expected))

		start := min(m.currentPlayerHandRow*5, len(player.Hand))
		end := min(start+5, len(player.Hand))
		for i, c := range player.Hand[start:end] {
			option(i, c.Name, "  [SELECTED FOR DISCARD]", m.selectedCards[c.Id])
		}
		if m.maxHandRows() > 1 {
			add(promptHintStyle, fmt.Sprintf("  Hand row %d/%d • [↑/↓] to change row", m.currentPlayerHandRow+1, m.maxHandRows()))
		}
		blank()
		add(lineStyle, fmt.Sprintf("Selected: %d / %d cards", m.selectedInHand(player), expected))

	case game.InteractionPlayerDonatesHand:
		add(lineStyle, "Choose teammate to receive your hand:")
		candidates := m.donationCandidates(player)
		if len(candidates) == 0 {
			add(promptHintStyle, "  (No teammate available)")
		}
		for i, p := range candidates {
			option(i, fmt.Sprintf("%s (%s)", p.Name, p.HeroClass.String()), "  (SELECTED)", i == m.promptChoiceIdx)
		}

	case game.InteractionNoInteraction:
		return baseView
	}

	blank()
	hint := "[1-5] Choose • [↑/↓] Move • [ENTER]/[SPACE] Submit • [ESC] Dismiss"
	if pi.Kind == game.InteractionPlayerDiscardCards {
		hint = "[1-5] Toggle card • [ENTER]/[SPACE] Submit • [ESC] Dismiss"
	}
	add(promptHintStyle, hint)

	return overlayCenter(baseView, promptModalStyle.Render(strings.Join(lines, "\n")))
}

// overlayCenter draws modal on top of base, centered, keeping the rest of base visible around it.
func overlayCenter(base, modal string) string {
	top := max(0, (strings.Count(base, "\n")+1-(strings.Count(modal, "\n")+1))/2)

	return overlayAt(base, modal, top)
}

// overlayAt draws modal on top of base, horizontally centered, starting at line top.
func overlayAt(base, modal string, top int) string {
	baseLines := strings.Split(base, "\n")
	modalLines := strings.Split(modal, "\n")

	baseWidth := lipgloss.Width(base)
	modalWidth := lipgloss.Width(modal)
	left := max(0, (baseWidth-modalWidth)/2)

	for len(baseLines) < top+len(modalLines) {
		baseLines = append(baseLines, "")
	}

	for i, ml := range modalLines {
		line := baseLines[top+i]
		if w := lipgloss.Width(line); w < left+modalWidth {
			line += strings.Repeat(" ", left+modalWidth-w)
		}

		baseLines[top+i] = ansi.Truncate(line, left, "") + ansi.ResetStyle + ml + ansi.ResetStyle + ansi.TruncateLeft(line, left+modalWidth, "")
	}

	return strings.Join(baseLines, "\n")
}
