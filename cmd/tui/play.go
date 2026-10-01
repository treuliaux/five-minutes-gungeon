package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type PlayAction int

const (
	PlayCardAction PlayAction = iota
	DiscardCardsAction
	UseHeroAbilityAction
	SubmitPromptChoiceAction
	UseArtifactAction
)

func (a PlayAction) String() string {
	switch a {
	case PlayCardAction:
		return "Play Card"
	case DiscardCardsAction:
		return "Discard Cards"
	case UseHeroAbilityAction:
		return "Use Hero Ability"
	case SubmitPromptChoiceAction:
		return "Submit Prompt Choice"
	case UseArtifactAction:
		return "Use Artifact"
	default:
		return "Unknown Action"
	}
}

type playModel struct {
	currentPlayerIdx     int
	selectedCards        map[game.CardID]bool
	currentPlayerHandRow int

	controller client.GameController
	ctx        context.Context

	snapshot game.GameSnapshot
	timer    time.Duration
}

func newPlayModel() playModel {
	return playModel{
		selectedCards: make(map[game.CardID]bool),
	}
}

func (m playModel) Update(msg tea.Msg) (playModel, tea.Cmd) {
	snapshot, ok := m.snapshot.(game.GameSnapshotDTO)
	if !ok {
		return m, nil
	}

	var cmds []tea.Cmd

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up":
			handSize := len(m.currentPlayer().Hand)
			rows := (handSize + 4) / 5
			m.currentPlayerHandRow = (m.currentPlayerHandRow - 1 + rows) % rows
		case "down":
			handSize := len(m.currentPlayer().Hand)
			rows := (handSize + 4) / 5
			m.currentPlayerHandRow = (m.currentPlayerHandRow + 1) % rows
		case "space":
			cmd := m.playSelectedCards()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "tab":
			m.currentPlayerIdx = (m.currentPlayerIdx + 1) % len(snapshot.Players)
			if m.currentPlayerIdx > len(snapshot.Players)-1 {
				m.currentPlayerIdx = 0
			}
			m.selectedCards = make(map[game.CardID]bool)
			m.currentPlayerHandRow = 0
		case "shift+tab":
			m.currentPlayerIdx = (m.currentPlayerIdx - 1) % len(snapshot.Players)
			if m.currentPlayerIdx < 0 {
				m.currentPlayerIdx = len(snapshot.Players) - 1
			}
			m.selectedCards = make(map[game.CardID]bool)
			m.currentPlayerHandRow = 0
		case "c":
			m.selectedCards = make(map[game.CardID]bool)
		case "1", "2", "3", "4", "5":
			if idx, err := strconv.Atoi(msg.String()); err == nil {
				m.toggleSelectCard(idx)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m playModel) View() string {
	snapshot, ok := m.snapshot.(game.GameSnapshotDTO)
	if !ok {
		return ""
	}

	if snapshot.Status == game.Victory {
		return "VICTORY!!"
	}

	if snapshot.Status == game.Defeat {
		return "DEFEAT (noobs)"
	}

	header := m.renderHeaderHUD()

	doorsAndCurses := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderOpenedDoors(),
		//lipgloss.JoinVertical(lipgloss.Left, m.renderActiveCurses(), m.renderPlayedField()),
	)

	//teamRow := lipgloss.JoinHorizontal(
	//	lipgloss.Top,
	//	m.renderArtifacts(),
	//	m.renderPendingPrompt(),
	//)
	//
	playerRow := m.renderPlayerDashboard()
	//footer := m.renderFooterKeybindings()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		doorsAndCurses,
		//teamRow,
		playerRow,
		//footer,
	)
}

func (m playModel) renderHeaderHUD() string {
	snapshot, ok := m.snapshot.(game.GameSnapshotDTO)
	if !ok {
		return ""
	}
	var hud strings.Builder

	var col1 strings.Builder
	col1.WriteString(fmt.Sprintf("DUNGEON: Floor %d (DUNGEON NAME)\n", snapshot.Level))
	fighting := "NO"
	if snapshot.IsFightingBoss {
		fighting = "YES"
	}
	bossName := "None"
	if snapshot.Boss != nil {
		bossName = snapshot.Boss.Name
	}
	col1.WriteString(fmt.Sprintf("BOSS: %s (Fighting: %s)", bossName, fighting))

	var col2 strings.Builder
	var timer strings.Builder
	timer.WriteString(fmt.Sprintf("TIME REMAINING: "))
	timeRemainingSeconds := int(m.timer.Seconds())
	timeRemainingString := fmt.Sprintf("%d:%02d", timeRemainingSeconds/60, timeRemainingSeconds%60)
	if snapshot.IsTimeFrozen {
		timer.WriteString(timerFrozenStyle.Render(fmt.Sprintf("[ %s ] ❄ FROZEN", timeRemainingString)))
	} else if timeRemainingSeconds < 60 {
		timer.WriteString(timerLowStyle.Render(fmt.Sprintf("[ %s ]", timeRemainingString)))
	} else {
		timer.WriteString(timerStyle.Render(fmt.Sprintf("[ %s ]", timeRemainingString)))
	}
	col2.WriteString(fmt.Sprintf("%s\n", timer.String()))
	col2.WriteString(fmt.Sprintf("STATUS: %s", strings.ToUpper(snapshot.Status.String())))

	var col3 strings.Builder
	col3.WriteString(fmt.Sprintf("DOORS LEFT: [ %d ]\n", snapshot.RemainingDoorsCount))
	col3.WriteString(fmt.Sprintf("TEAM CARDS IN PLAY: %d", len(snapshot.PlayedField)))

	hud.WriteString(hudStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		hudColStyle.Width(45).Render(col1.String()),
		hudColStyle.Width(45).Render(col2.String()),
		hudLastColStyle.Render(col3.String()),
	)))

	return hud.String()
}

func (m playModel) renderPlayerDashboard() string {
	player := m.currentPlayer()
	if player.Id == "" {
		return ""
	}

	var hudHeader strings.Builder
	playerBoxTitle := fmt.Sprintf("┌── YOUR HERO: %s (%s) ", strings.ToUpper(player.HeroClass.String()), player.Name)
	hudHeader.WriteString(playerHeaderStyle.Render(playerBoxTitle))
	hudHeader.WriteString(strings.Repeat("─", 137-len(playerBoxTitle)) + "┐\n")

	var hudCol1 strings.Builder
	hudCol1.WriteString(fmt.Sprintf("Deck: [🂠 %d cards]", player.DeckCount))

	var hudCol2 strings.Builder
	hudCol2.WriteString(fmt.Sprintf("Discard: [🗑️ %d cards (Top: XXX)]", len(player.Discard)))

	var hudCol3 strings.Builder
	hudCol3.WriteString(fmt.Sprintf("Ability: [ %s ]", player.AbilityName))

	var hud strings.Builder
	hud.WriteString(hudHeader.String())
	hud.WriteString(playerHudStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		hudColStyle.Width(35).Render(hudCol1.String()),
		hudColStyle.Width(50).Render(hudCol2.String()),
		hudLastColStyle.Render(hudCol3.String()),
	)))
	hud.WriteString("\n")

	var playerHand strings.Builder
	rowsCount := min((len(player.Hand)+4)/5, 1)
	playerHand.WriteString(fmt.Sprintf("YOUR HAND (Row: %d/%d - Selected: %d - Total: %d):\n", m.currentPlayerHandRow+1, rowsCount, len(m.selectedCards), len(player.Hand)))

	start := m.currentPlayerHandRow * 5
	end := min(start+5, len(player.Hand))
	playerCardsToDraw := player.Hand[start:end]
	playerCards := make([]string, len(playerCardsToDraw))
	for i, s := range playerCardsToDraw {
		playerCards[i] = m.renderPlayerCard(s, i+1)
	}
	playerHand.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, playerCards...))

	var box strings.Builder
	box.WriteString(hud.String() + playerStyle.Render(playerHand.String()))

	return box.String()
}

func (m playModel) renderPlayerCard(card game.PlayerCardDTO, idx int) string {
	var cardHeader strings.Builder
	cardHeader.WriteString("┌─ ")
	cardHeader.WriteString(fmt.Sprintf("[%d]", idx))
	//cardHeader.WriteString(cardHeaderStyle.Render(cardBoxTitle))
	cardHeader.WriteString(" " + strings.Repeat("─", 16) + "┐")

	var cardBody strings.Builder
	cardBody.WriteString(fmt.Sprintf("%s\n", card.Name))
	switch card.Kind {
	case game.PlayerCardResource:
		cardBody.WriteString("[Resource]\n\n\n")
		for _, r := range card.Resources {
			switch r {
			case game.NoResource:
			case game.Sword:
				cardBody.WriteString("🗡️ ")
			case game.Arrow:
				cardBody.WriteString("🏹 ")
			case game.Shield:
				cardBody.WriteString("🛡️ ")
			case game.Jump:
				cardBody.WriteString("🦵 ")
			case game.Scroll:
				cardBody.WriteString("📜 ")
			case game.WildCard:
			case game.InfiniteSword:
				cardBody.WriteString("🗡️♾️ ")
			case game.InfiniteArrow:
				cardBody.WriteString("🏹♾️ ")
			case game.InfiniteShield:
				cardBody.WriteString("🛡️♾️ ")
			case game.InfiniteJump:
				cardBody.WriteString("🦵♾️ ")
			case game.InfiniteScroll:
				cardBody.WriteString("📜♾️ ")
			}
		}
	case game.PlayerCardAction:
		cardBody.WriteString("[Action]\n\n")
		cardBody.WriteString(fmt.Sprintf("%s", card.Description))
	}

	var style lipgloss.Style
	var headerStyle lipgloss.Style
	style = cardStyle
	headerStyle = lipgloss.NewStyle().Margin(0).Padding(0)
	if m.selectedCards[card.Id] {
		style = selectedCardStyle
		headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	}
	var box strings.Builder
	box.WriteString(lipgloss.JoinVertical(lipgloss.Left, headerStyle.Render(cardHeader.String()), style.Render(cardBody.String())))

	return box.String()
}

func (m playModel) renderOpenedDoors() string {
	snapshot, ok := m.snapshot.(game.GameSnapshotDTO)
	if !ok {
		return ""
	}

	var doors strings.Builder

	openedDoors := make([]string, len(snapshot.OpenedDoors))
	for i, d := range snapshot.OpenedDoors {
		openedDoors[i] = m.renderDoor(d, i+1)
	}
	doors.WriteString(lipgloss.JoinVertical(lipgloss.Left, openedDoors...))

	var openedDoorsBoxTitle strings.Builder
	openedDoorsBoxTitle.WriteString("┌── ACTIVE OPENED DOORS ")
	openedDoorsBoxTitle.WriteString(strings.Repeat("─", (screenWidth/2)+5-len(openedDoorsBoxTitle.String())) + "┐\n")

	var openedDoorsBox strings.Builder
	openedDoorsBox.WriteString(openedDoorsHeaderStyle.Render(doors.String()))

	var box strings.Builder
	box.WriteString(openedDoorsBoxTitle.String() + openedDoorsStyle.Render(openedDoorsBox.String()))

	return box.String()
}

func (m playModel) renderDoor(door game.DungeonCardDTO, idx int) string {
	var doorHeader strings.Builder
	doorBoxTitle := fmt.Sprintf("┌─ [DOOR %d] ", idx)
	doorHeader.WriteString(doorHeaderStyle.Render(doorBoxTitle))
	doorHeader.WriteString(strings.Repeat("─", (screenWidth/2)-1-len(doorBoxTitle)) + "┐\n")

	var doorBody strings.Builder
	doorBody.WriteString(fmt.Sprintf("%s: %s\n", dungeonCardKindToString(door.Kind), door.Name))
	doorBody.WriteString("Required:")
	for _, r := range door.Resources {
		switch r {
		case game.NoResource:
		case game.Sword:
			doorBody.WriteString("🗡️ ")
		case game.Arrow:
			doorBody.WriteString("🏹 ")
		case game.Shield:
			doorBody.WriteString("🛡️ ")
		case game.Jump:
			doorBody.WriteString("🦵 ")
		case game.Scroll:
			doorBody.WriteString("📜 ")
		default:
		}
	}
	doorBody.WriteString("\n")
	doorBody.WriteString("Field Met: ")

	var box strings.Builder
	box.WriteString(doorHeader.String() + doorStyle.Render(doorBody.String()))

	return box.String()
}

func (m playModel) currentPlayer() game.PlayerDTO {
	snapshot, ok := m.snapshot.(game.GameSnapshotDTO)
	if !ok {
		return game.PlayerDTO{}
	}

	return snapshot.Players[m.currentPlayerIdx]
}

func (m playModel) toggleSelectCard(i int) {
	hand := m.currentPlayer().Hand
	index := m.currentPlayerHandRow*5 + i - 1

	if index < 0 || index >= len(hand) {
		return
	}

	cardID := hand[index].Id
	if m.selectedCards[cardID] {
		delete(m.selectedCards, cardID)
	} else {
		m.selectedCards[cardID] = true
	}
}

func (m playModel) playSelectedCards() tea.Cmd {
	for cID := range m.selectedCards {
		isAction := false
		for _, c := range m.currentPlayer().Hand {
			if c.Id != cID {
				continue
			}
			if c.Kind == game.PlayerCardAction {
				isAction = true
				break
			}
		}
		if isAction {
			continue
		}
		err := m.controller.Dispatch(m.ctx, game.PlayCardCmd{
			PlayerID:        m.currentPlayer().Id,
			CardID:          cID,
			TargetCardID:    0,
			TargetPlayerIDs: nil,
		})
		if err != nil {
			return forwardError(err)
		}
		delete(m.selectedCards, cID)
	}

	return nil
}

func dungeonCardKindToString(kind game.DungeonCardKind) string {
	switch kind {
	case game.CardMonster:
		return "Monster"
	case game.CardObstacle:
		return "Obstacle"
	case game.CardPerson:
		return "Person"
	case game.CardMiniBoss:
		return "Mini-Boss"
	case game.CardEvent:
		return "Event"
	case game.CardCurse:
		return "Curse"
	case game.CardBoss:
		return "Boss"
	}

	return ""
}
