package main

import (
	"context"
	"fmt"
	"maps"
	"slices"
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
	lastPlayerActions    []PlayerActionEntry
	currentPlayerIdx     int
	selectedCards        map[game.CardID]bool
	currentPlayerHandRow int

	controller client.GameController
	ctx        context.Context

	snapshot game.GameSnapshotDTO
	timer    time.Duration
}

type PlayerActionEntry struct {
	CardPlayed  game.CardPlayedEvent
	AbilityUsed game.HeroAbilityUsedEvent
}

func newPlayModel() playModel {
	return playModel{
		selectedCards:     make(map[game.CardID]bool),
		lastPlayerActions: make([]PlayerActionEntry, 8),
	}
}

func (m playModel) Update(msg tea.Msg) (playModel, tea.Cmd) {
	if m.snapshot.Status == game.Waiting {
		return m, nil
	}

	m.currentPlayerHandRow = min(m.currentPlayerHandRow, ((len(m.currentPlayer().Hand)+4)/5)-1)

	var cmds []tea.Cmd
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up":
			handSize := len(m.currentPlayer().Hand)
			rowsIdx := (handSize + 4) / 5
			m.currentPlayerHandRow = (m.currentPlayerHandRow - 1 + rowsIdx) % rowsIdx
		case "down":
			handSize := len(m.currentPlayer().Hand)
			rowsIdx := (handSize + 4) / 5
			m.currentPlayerHandRow = (m.currentPlayerHandRow + 1) % rowsIdx
		case "space":
			cmd := m.playSelectedCards()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "a":
			cmd := m.useHeroAbility()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "tab":
			m.currentPlayerIdx = (m.currentPlayerIdx + 1) % len(m.snapshot.Players)
			if m.currentPlayerIdx > len(m.snapshot.Players)-1 {
				m.currentPlayerIdx = 0
			}
			m.selectedCards = make(map[game.CardID]bool)
			m.currentPlayerHandRow = 0
		case "shift+tab":
			m.currentPlayerIdx = (m.currentPlayerIdx - 1) % len(m.snapshot.Players)
			if m.currentPlayerIdx < 0 {
				m.currentPlayerIdx = len(m.snapshot.Players) - 1
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
	if m.snapshot.Status == game.Waiting {
		return ""
	}

	if m.snapshot.Status == game.Victory {
		return "VICTORY!!"
	}

	if m.snapshot.Status == game.Defeat {
		return "DEFEAT (noobs)"
	}

	header := m.renderHeaderHUD()

	doorsAndActionHistory := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderOpenedDoors(),
		lipgloss.JoinVertical(lipgloss.Left, m.renderPlayedHistory(), m.renderPlayfield()),
	)

	//teamRow := lipgloss.JoinHorizontal(
	//	lipgloss.Top,
	//	m.renderArtifacts(),
	//	m.renderPendingPrompt(),
	//)
	//
	playerRow := m.renderPlayerDashboard()
	footer := m.renderFooterKeybindings()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		doorsAndActionHistory,
		//teamRow,
		playerRow,
		footer,
	)
}

func (m playModel) renderHeaderHUD() string {
	var hud strings.Builder

	var col1 strings.Builder
	col1.WriteString(fmt.Sprintf("DUNGEON: Floor %d (DUNGEON NAME)\n", m.snapshot.Level))
	fighting := "NO"
	if m.snapshot.IsFightingBoss {
		fighting = "YES"
	}
	bossName := "None"
	if m.snapshot.Boss != nil {
		bossName = m.snapshot.Boss.Name
	}
	col1.WriteString(fmt.Sprintf("BOSS: %s (Fighting: %s)", bossName, fighting))

	var col2 strings.Builder
	var timer strings.Builder
	timer.WriteString(fmt.Sprintf("TIME REMAINING: "))
	minutes := int(m.timer.Minutes())
	seconds := int(m.timer.Seconds()) % 60
	milliseconds := int(m.timer.Milliseconds()) % 1000
	timeRemainingString := fmt.Sprintf("%d:%02d:%02d", minutes, seconds, milliseconds/10)
	if m.snapshot.IsTimeFrozen {
		timer.WriteString(timerFrozenStyle.Render(fmt.Sprintf("[ %s ] ❄  FROZEN", timeRemainingString)))
	} else if int(m.timer.Seconds()) < 60 {
		timer.WriteString(timerLowStyle.Render(fmt.Sprintf("[ %s ]", timeRemainingString)))
	} else {
		timer.WriteString(timerStyle.Render(fmt.Sprintf("[ %s ]", timeRemainingString)))
	}
	col2.WriteString(fmt.Sprintf("%s\n", timer.String()))
	col2.WriteString(fmt.Sprintf("STATUS: %s", strings.ToUpper(m.snapshot.Status.String())))

	var col3 strings.Builder
	col3.WriteString(fmt.Sprintf("DOORS LEFT: [ %d ]\n", m.snapshot.RemainingDoorsCount))
	col3.WriteString(fmt.Sprintf("TEAM CARDS IN PLAY: %d", len(m.snapshot.PlayedField)))

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
	topDiscard := "None"
	if len(player.Discard) > 0 {
		topDiscard = player.Discard[len(player.Discard)-1].Name
	}
	hudCol2.WriteString(fmt.Sprintf("Discard: [🗑️ %d cards (Top: %s)]", len(player.Discard), topDiscard))

	var hudCol3 strings.Builder
	hudCol3.WriteString(fmt.Sprintf("Ability: [%s] %s", player.AbilityName, player.AbilityDescription))

	var hud strings.Builder
	hud.WriteString(hudHeader.String())
	hud.WriteString(playerHudStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		hudColStyle.Width(23).Render(hudCol1.String()),
		hudColStyle.Width(51).Render(hudCol2.String()),
		hudLastColStyle.Render(hudCol3.String()),
	)))
	hud.WriteString("\n")

	var playerHand strings.Builder
	rowsCount := max((len(player.Hand)+4)/5, 1)
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
	cardHeader.WriteString(" " + strings.Repeat("─", 16) + "┐")

	var cardBody strings.Builder
	cardBody.WriteString(fmt.Sprintf("%s\n", card.Name))
	switch card.Kind {
	case game.PlayerCardResource:
		cardBody.WriteString("[Resource]\n\n\n")
		for _, r := range card.Resources {
			cardBody.WriteString(fmt.Sprintf("%s ", resourceTypeToIcon(r)))
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
	var doors strings.Builder

	openedDoors := make([]string, len(m.snapshot.OpenedDoors))
	for i, d := range m.snapshot.OpenedDoors {
		openedDoors[i] = m.renderDoor(d, i+1)
	}
	doors.WriteString(lipgloss.JoinVertical(lipgloss.Left, openedDoors...))

	var openedDoorsBoxTitle strings.Builder
	openedDoorsBoxTitle.WriteString("┌── ACTIVE OPENED DOORS ")
	openedDoorsBoxTitle.WriteString(strings.Repeat("─", (screenWidth/2)+5-len(openedDoorsBoxTitle.String())) + "┐\n")

	var openedDoorsBox strings.Builder
	openedDoorsBox.WriteString(doors.String())

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
	doorBody.WriteString("Required: ")
	switch door.Kind {
	case game.CardMonster:
		fallthrough
	case game.CardObstacle:
		fallthrough
	case game.CardPerson:
		fallthrough
	case game.CardMiniBoss:
		fallthrough
	case game.CardBoss:
		for _, r := range door.Resources {
			doorBody.WriteString(resourceTypeToIcon(r))
		}
		doorBody.WriteString("\n")
		doorBody.WriteString("Field Met: ")
		dispatchedResources := m.dispatchResources()[door.Id]
		if len(dispatchedResources) > 0 {
			for _, required := range dispatchedResources {
				doorBody.WriteString(fmt.Sprintf("%s", resourceTypeToIcon(required)))
			}
		}
	case game.CardEvent:
		doorBody.WriteString(door.Description)
	case game.CardCurse:
	}

	var box strings.Builder
	box.WriteString(doorHeader.String() + doorStyle.Render(doorBody.String()))

	return box.String()
}

func (m playModel) currentPlayer() game.PlayerDTO {
	return m.snapshot.Players[m.currentPlayerIdx]
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
		skip := false
		for _, c := range m.currentPlayer().Hand {
			if c.Id != cID {
				continue
			}
			if c.Kind == game.PlayerCardAction && c.TargetType != game.TargetNone && (c.TargetType != game.TargetCard || len(m.snapshot.OpenedDoors) != 1) {
				skip = true
				break
			}
		}
		if !skip {
			err := m.controller.Dispatch(m.ctx, game.PlayCardCmd{
				PlayerID:        m.currentPlayer().Id,
				CardID:          cID,
				TargetCardID:    0,
				TargetPlayerIDs: nil,
			})
			if err != nil {
				return forwardError(err)
			}
		}
		delete(m.selectedCards, cID)
	}

	return nil
}

func (m playModel) useHeroAbility() tea.Cmd {
	if len(m.selectedCards) != 3 {
		return nil
	}
	cardIDs := make([]game.CardID, 3)
	i := 0
	for cID := range maps.Keys(m.selectedCards) {
		cardIDs[i] = cID
		i++
	}
	err := m.controller.Dispatch(m.ctx, game.UseHeroAbilityCmd{
		PlayerID:       m.currentPlayer().Id,
		DiscardCardIDs: cardIDs,
		TargetCardID:   0,
		TargetPlayerID: "",
	})
	if err != nil {
		return forwardError(err)
	}
	for cID := range maps.Keys(m.selectedCards) {
		delete(m.selectedCards, cID)
	}

	return nil
}

func (m playModel) renderPlayedHistory() string {
	var playedHistoryBoxTitle strings.Builder
	playedHistoryBoxTitle.WriteString("┌── PLAYERS ACTIONS ")
	playedHistoryBoxTitle.WriteString(strings.Repeat("─", (screenWidth/2)+5-len(playedHistoryBoxTitle.String())) + "┐\n")

	var playedHistoryBox strings.Builder

	for i := range slices.Backward(m.lastPlayerActions) {
		e := m.lastPlayerActions[i]
		if (e.CardPlayed.Card.Id == 0 || e.CardPlayed.ByPlayer.Id == "") && e.AbilityUsed.ByPlayer.Name == "" {
			continue
		}

		playerName := e.CardPlayed.ByPlayer.Name
		if e.AbilityUsed.ByPlayer.Name != "" {
			playerName = e.AbilityUsed.ByPlayer.Name
		}

		var entry strings.Builder
		entry.WriteString(fmt.Sprintf("%s played: [", playerName))
		switch {
		case e.CardPlayed.Card.Id != 0 && e.CardPlayed.ByPlayer.Id != "":
			switch e.CardPlayed.Card.Kind {
			case game.PlayerCardResource:
				for _, r := range e.CardPlayed.Card.Resources {
					entry.WriteString(resourceTypeToIcon(r))
				}
			case game.PlayerCardAction:
				entry.WriteString(e.CardPlayed.Card.Name)
			}
		case e.AbilityUsed.ByPlayer.Name != "":
			entry.WriteString(e.AbilityUsed.ByPlayer.AbilityName)
		}
		entry.WriteString("]")

		if playedHistoryBox.Len() > 0 {
			playedHistoryBox.WriteString("\n")
		}
		playedHistoryBox.WriteString(entry.String())
	}

	var box strings.Builder
	box.WriteString(playedHistoryBoxTitle.String() + playedHistoryStyle.Render(playedHistoryBox.String()))

	return box.String()
}

func (m playModel) renderPlayfield() string {
	var playfieldBoxTitle strings.Builder
	playfieldBoxTitle.WriteString("┌── PLAYFIELD ")
	playfieldBoxTitle.WriteString(strings.Repeat("─", (screenWidth/2)+5-len(playfieldBoxTitle.String())) + "┐\n")

	listResources := make(map[game.ResourceType]int)
	listActions := make(map[string]int)
	for _, c := range m.snapshot.PlayedField {
		for _, r := range c.Resources {
			listResources[r]++
		}
		if c.Kind != game.PlayerCardResource {
			listActions[c.Name]++
		}
	}

	var playfieldBox strings.Builder

	resources := slices.Sorted(maps.Keys(listResources))
	for _, r := range resources {
		playfieldBox.WriteString(fmt.Sprintf("%s x%d ", resourceTypeToIcon(r), listResources[r]))
	}

	var box strings.Builder
	box.WriteString(playfieldBoxTitle.String() + playfieldStyle.Render(playfieldBox.String()))

	return box.String()
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

func (m playModel) dispatchResources() map[game.CardID][]game.ResourceType {
	requirements := m.snapshot.OpenedDoors
	playedCards := m.snapshot.PlayedField

	result := make(map[game.CardID][]game.ResourceType)
	available := make(map[game.ResourceType]int)
	for _, card := range playedCards {
		for _, resource := range card.Resources {
			available[resource]++
		}
	}

	// First pass: exact matches.
	for _, card := range requirements {
		for _, required := range card.Resources {
			if available[required] == 0 {
				continue
			}
			result[card.Id] = append(result[card.Id], required)
			available[required]--
		}
	}

	// Second pass: fill remaining requirements with jokers.
	for _, card := range requirements {
		assigned := make(map[game.ResourceType]int)

		for _, resource := range result[card.Id] {
			assigned[resource]++
		}

		for _, required := range card.Resources {
			if assigned[required] > 0 {
				assigned[required]--
				continue
			}

			if available[game.WildCard] == 0 {
				continue
			}

			result[card.Id] = append(result[card.Id], game.WildCard)
			available[game.WildCard]--
		}
	}

	return result
}

func (m playModel) renderFooterKeybindings() string {
	var footer strings.Builder
	footer.WriteString("[1-5] Select Card   •   [SPACE] Play Cards   •   [A] Hero Ability   •   [U] Use Artifact   •   [Ctrl+C] Quit")

	return footerStyle.Render(footer.String())
}
