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

type TargetKind int

const (
	TargetKindNone TargetKind = iota
	TargetKindDoor
	TargetKindPlayer
	TargetKindTwoPlayers
)

type targetingState struct {
	kind            TargetKind
	sourceCardID    game.CardID
	isAbility       bool
	isArtifact      bool
	artifactID      game.ArtifactID
	actionIndex     game.ArtifactActionIndex
	targetCardID    game.CardID
	targetPlayerIDs map[game.PlayerID]bool
	targetIndex     int
}

type PlayerSwitch int8

const (
	NextPlayer     PlayerSwitch = 1
	PreviousPlayer PlayerSwitch = -1
)

type playModel struct {
	lastPlayerActions []PlayerActionEntry

	currentPlayerIdx int

	selectedCards        map[game.CardID]bool
	currentPlayerHandRow int
	targeting            targetingState
	dismissedPrompt      bool
	promptChoiceIdx      int

	controller client.GameController
	ctx        context.Context

	snapshot game.GameSnapshotDTO
	timer    time.Duration
	config   game.Config
}

type PlayerActionEntry struct {
	CardPlayed  game.CardPlayedEvent
	AbilityUsed game.HeroAbilityUsedEvent
}

func newPlayModel() playModel {
	return playModel{
		selectedCards:     make(map[game.CardID]bool),
		lastPlayerActions: make([]PlayerActionEntry, 7),
	}
}

func (m playModel) Init() playModel {
	m.lastPlayerActions = make([]PlayerActionEntry, len(m.lastPlayerActions))
	m.selectedCards = make(map[game.CardID]bool)
	m.currentPlayerIdx = 0
	m.currentPlayerHandRow = 0
	m.targeting = targetingState{}
	m.dismissedPrompt = false
	m.promptChoiceIdx = 0

	return m
}

func (m playModel) Update(msg tea.Msg) (playModel, tea.Cmd) {
	switch m.snapshot.Status {
	case game.Waiting:
		return m, nil
	default:
	}

	m.currentPlayerHandRow = max(0, min(m.currentPlayerHandRow, m.maxHandRows()-1))

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch m.snapshot.Status {
	case game.Defeat, game.Victory:
		return m.updateGameOver(keyMsg)
	default:
	}

	if m.snapshot.PendingInteraction != nil && slices.Contains(m.snapshot.PendingInteraction.PendingPlayers, m.currentPlayer().Id) {
		return m.updatePendingInteraction(keyMsg)
	}

	switch m.targeting.kind {
	case TargetKindDoor:
		return m.updateDoorTargeting(keyMsg)
	case TargetKindPlayer, TargetKindTwoPlayers:
		return m.updatePlayerTargeting(keyMsg)
	case TargetKindNone:
	}

	return m.updatePlaying(keyMsg)
}

func (m playModel) updateGameOver(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	switch key.String() {
	case "enter":
		return m, backToLobby()
	case "r":
		if m.snapshot.Status != game.Victory {
			break
		}
		if err := m.controller.Dispatch(m.ctx, game.StartCmd{}); err != nil {
			return m.Init(), forwardError(err)
		}
	}

	return m, nil
}

func (m playModel) updatePendingInteraction(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	if m.dismissedPrompt {
		return m.updateDismissedPrompt(key)
	}

	return m.updateActivePrompt(key)
}

func (m playModel) updateDoorTargeting(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	targetList := m.snapshot.OpenedDoors
	if len(targetList) == 0 {
		m.targeting = targetingState{}

		return m, nil
	}

	switch key.String() {
	case "esc":
		m.targeting = targetingState{}

		return m, nil

	case "1", "2":
		idx, err := strconv.Atoi(key.String())
		if err != nil {
			return m, nil
		}

		targetIdx := idx - 1
		switch {
		case targetIdx >= 0 && targetIdx < len(targetList):
			m.targeting.targetIndex = targetIdx
			m.targeting.targetCardID = targetList[targetIdx].Id
		}

		return m, nil

	case "up":
		targetIdx := (m.targeting.targetIndex - 1 + len(targetList)) % len(targetList)
		m.targeting.targetIndex = targetIdx
		m.targeting.targetCardID = targetList[targetIdx].Id

		return m, nil

	case "down":
		targetIdx := (m.targeting.targetIndex + 1 + len(targetList)) % len(targetList)
		m.targeting.targetIndex = targetIdx
		m.targeting.targetCardID = targetList[targetIdx].Id

		return m, nil

	case "space", "enter":
		return m.dispatchTargetedAction()
	}

	return m, nil
}

func (m playModel) updatePlayerTargeting(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	targetList := m.snapshot.Players
	if len(targetList) == 0 {
		m.targeting = targetingState{}

		return m, nil
	}

	switch key.String() {
	case "esc":
		m.targeting = targetingState{}

		return m, nil

	case "1", "2", "3", "4", "5", "6":
		idx, err := strconv.Atoi(key.String())
		if err != nil {
			return m, nil
		}

		targetIdx := idx - 1
		switch {
		case targetIdx >= 0 && targetIdx < len(targetList):
			m.targeting.targetIndex = targetIdx
			if m.targeting.targetPlayerIDs[targetList[targetIdx].Id] {
				delete(m.targeting.targetPlayerIDs, targetList[targetIdx].Id)

				return m, nil
			}
			m.targeting.targetPlayerIDs[targetList[targetIdx].Id] = true
		}

		return m.dispatchTargetedAction()

	case "up":
		targetIdx := (m.targeting.targetIndex - 1 + len(targetList)) % len(targetList)
		m.targeting.targetIndex = targetIdx

		return m, nil

	case "down":
		targetIdx := (m.targeting.targetIndex + 1 + len(targetList)) % len(targetList)
		m.targeting.targetIndex = targetIdx

		return m, nil

	case "space", "enter":
		if m.targeting.targetPlayerIDs[targetList[m.targeting.targetIndex].Id] {
			delete(m.targeting.targetPlayerIDs, targetList[m.targeting.targetIndex].Id)

			return m, nil
		}
		m.targeting.targetPlayerIDs[targetList[m.targeting.targetIndex].Id] = true

		return m.dispatchTargetedAction()
	}

	return m, nil
}

func (m playModel) updatePlaying(key tea.KeyPressMsg) (playModel, tea.Cmd) {
	switch key.String() {
	case "up":
		maxRows := m.maxHandRows()
		m.currentPlayerHandRow = (m.currentPlayerHandRow - 1 + maxRows) % maxRows

		return m, nil

	case "down":
		maxRows := m.maxHandRows()
		m.currentPlayerHandRow = (m.currentPlayerHandRow + 1 + maxRows) % maxRows

		return m, nil

	case "space":
		return m.playSelectedCards()

	case "a":
		return m.useHeroAbility()

	case "d":
		return m.discardCards()

	case "u":
		return m.useArtifact()

	case "tab":
		return m.switchPlayer(NextPlayer), nil

	case "shift+tab":
		return m.switchPlayer(PreviousPlayer), nil

	case "c":
		clear(m.selectedCards)

		return m, nil
	case "1", "2", "3", "4", "5":
		idx, err := strconv.Atoi(key.String())
		if err != nil {
			return m, nil
		}

		return m.toggleSelectCard(idx), nil
	}

	return m, nil
}

func (m playModel) switchPlayer(delta PlayerSwitch) playModel {
	numPlayers := len(m.snapshot.Players)
	if numPlayers == 0 {
		return m
	}

	m.currentPlayerIdx = (m.currentPlayerIdx + int(delta) + numPlayers) % numPlayers
	clear(m.selectedCards)
	m.currentPlayerHandRow = 0
	m.targeting = targetingState{}
	m.dismissedPrompt = false
	m.promptChoiceIdx = 0

	return m
}

func (m playModel) View() string {
	if m.snapshot.Status == game.Waiting {
		return ""
	}

	if m.snapshot.Status == game.Victory {
		return m.renderVictory()
	}

	if m.snapshot.Status == game.Defeat {
		return m.renderDefeat()
	}

	header := m.renderHeaderHUD()

	middleRow1 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderOpenedDoors(),
		lipgloss.JoinVertical(lipgloss.Left, m.renderPlayedHistory(), m.renderPlayfield()),
	)

	middleRow2 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderActivePlayers(),
		m.renderPendingPrompt(),
	)

	middleRow3 := ""
	if m.config.UseExtension {
		middleRow3 = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderArtifacts(),
			m.renderActiveCurses(),
		)
	}

	return m.renderPromptOverlay(lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		middleRow1,
		middleRow2,
		middleRow3,
		m.renderPlayerDashboard(),
		m.renderFooter(),
	))
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
	timer.WriteString("TIME REMAINING: ")
	minutes := int(m.timer.Minutes())
	seconds := int(m.timer.Seconds()) % 60
	milliseconds := int(m.timer.Milliseconds()) % 1000
	timeRemainingString := fmt.Sprintf("%d:%02d:%02d", minutes, seconds, milliseconds/10)
	switch {
	case m.snapshot.IsTimeFrozen:
		timer.WriteString(timerFrozenStyle.Render(fmt.Sprintf("[ %s ] %s FROZEN", timeRemainingString, icon(iconFrozen))))
	case int(m.timer.Seconds()) < 60:
		timer.WriteString(timerLowStyle.Render(fmt.Sprintf("[ %s ]", timeRemainingString)))
	default:
		timer.WriteString(timerStyle.Render(fmt.Sprintf("[ %s ]", timeRemainingString)))
	}
	col2.WriteString(fmt.Sprintf("%s\n", timer.String()))
	col2.WriteString(fmt.Sprintf("STATUS: %s", strings.ToUpper(m.snapshot.Status.String())))

	var col3 strings.Builder
	col3.WriteString(fmt.Sprintf("DOORS LEFT: [ %d ]\n", m.snapshot.RemainingDoorsCount))
	col3.WriteString(fmt.Sprintf("TEAM CARDS IN PLAY: %d", len(m.snapshot.PlayedField)))

	hud.WriteString(playHudStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		playHudColStyle.Width(45).Render(col1.String()),
		playHudColStyle.Width(45).Render(col2.String()),
		playHudLastColStyle.Render(col3.String()),
	)))

	return hud.String()
}

func (m playModel) renderOpenedDoors() string {
	var openedDoorsBoxTitle strings.Builder
	header := "┌── ACTIVE OPENED DOORS "
	headerStyle := lipgloss.NewStyle()
	boxStyle := openedDoorsStyle

	if m.targeting.kind == TargetKindDoor {
		header = "┌── " + icon(iconTarget) + " SELECT TARGET DOOR "
		headerStyle = headerStyle.Foreground(lipgloss.Cyan)
		boxStyle = openedDoorsStyle.BorderForeground(lipgloss.Cyan)
	}

	var fullTitle strings.Builder
	fullTitle.WriteString(header + strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(header)-1)) + "┐")

	openedDoorsBoxTitle.WriteString(headerStyle.Render(fullTitle.String()) + "\n")

	var doors strings.Builder
	openedDoors := make([]string, len(m.snapshot.OpenedDoors))
	for i, d := range m.snapshot.OpenedDoors {
		openedDoors[i] = m.renderDoor(d, i+1)
	}
	doors.WriteString(lipgloss.JoinVertical(lipgloss.Left, openedDoors...))

	var openedDoorsBox strings.Builder
	openedDoorsBox.WriteString(doors.String())

	var box strings.Builder
	box.WriteString(openedDoorsBoxTitle.String() + boxStyle.Render(openedDoorsBox.String()))

	return box.String()
}

func (m playModel) renderDoor(door game.DungeonCardDTO, idx int) string {
	isTargeted := m.targeting.kind == TargetKindDoor && m.targeting.targetCardID == door.Id

	var doorHeader strings.Builder
	doorBoxTitle := fmt.Sprintf("┌─ [DOOR %d] ", idx)
	headerStyle := lipgloss.NewStyle().Margin(0).Padding(0)
	doorBoxStyle := doorStyle

	if isTargeted {
		doorBoxTitle = fmt.Sprintf("┌─ %s [%d] [DOOR %d - TARGET] ", icon(iconTarget), idx, idx)
		headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
		doorBoxStyle = selectedDoorStyle
	}

	doorHeader.WriteString(doorBoxTitle)
	doorHeader.WriteString(strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(doorBoxTitle))-5) + "┐")

	var doorBody strings.Builder
	doorBody.WriteString(fmt.Sprintf("%s: %s\n", dungeonCardKindToString(door.Kind), door.Name))
	doorBody.WriteString("Required: ")
	switch door.Kind {
	case game.CardMonster, game.CardObstacle, game.CardPerson, game.CardMiniBoss, game.CardBoss:
		for _, r := range door.Resources {
			doorBody.WriteString(resourceTypeToIcon(r))
		}
		doorBody.WriteString("\n")
		doorBody.WriteString("Field Met: ")
		dispatchedResources := m.dispatchResources()[door.Id]
		if len(dispatchedResources) > 0 {
			for _, required := range dispatchedResources {
				doorBody.WriteString(resourceTypeToIcon(required))
			}
		}
	case game.CardEvent:
		doorBody.WriteString(door.Description)
	case game.CardCurse:
	}

	var box strings.Builder
	box.WriteString(lipgloss.JoinVertical(lipgloss.Left, headerStyle.Render(doorHeader.String()), doorBoxStyle.Render(doorBody.String())))

	return box.String()
}

func (m playModel) renderPlayedHistory() string {
	var playedHistoryBoxTitle strings.Builder
	playedHistoryBoxTitle.WriteString("┌── PLAYERS ACTIONS ")
	playedHistoryBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(playedHistoryBoxTitle.String())-1)) + "┐\n")

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
	playfieldBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(playfieldBoxTitle.String())-1)) + "┐\n")

	listResources := make(map[game.ResourceType]int)
	for _, c := range m.snapshot.PlayedField {
		for _, r := range c.Resources {
			listResources[r]++
		}
	}

	var playfieldBox strings.Builder

	resources := slices.Sorted(maps.Keys(listResources))
	for _, r := range resources {
		playfieldBox.WriteString(fmt.Sprintf("%s x%d ", resourceTypeToIcon(r), listResources[r]))
	}

	var box strings.Builder
	box.WriteString(playfieldBoxTitle.String() + playPlayfieldStyle.Render(playfieldBox.String()))

	return box.String()
}

func (m playModel) renderActivePlayers() string {
	isPlayerTargeting := m.targeting.kind == TargetKindPlayer || m.targeting.kind == TargetKindTwoPlayers

	var activePlayersBoxTitle strings.Builder
	header := "┌── PLAYERS "
	headerStyle := lipgloss.NewStyle()
	boxStyle := pendingPromptStyle

	if isPlayerTargeting {
		header = "┌── " + icon(iconTarget) + " SELECT TARGET PLAYERS "
		headerStyle = headerStyle.Foreground(lipgloss.Cyan)
		boxStyle = pendingPromptStyle.BorderForeground(lipgloss.Cyan)
	}

	var fullTitle strings.Builder
	fullTitle.WriteString(header + strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(header)-1)) + "┐")

	activePlayersBoxTitle.WriteString(headerStyle.Render(fullTitle.String()) + "\n")

	var columns [2][]string
	for i, p := range m.snapshot.Players {
		playerRowStyle := lipgloss.NewStyle()
		pointer := "  "
		targeted := "  "
		if isPlayerTargeting && i == m.targeting.targetIndex {
			playerRowStyle = playerRowStyle.Foreground(lipgloss.Cyan)
			pointer = "> "
		}
		if m.targeting.targetPlayerIDs[p.Id] {
			targeted = icon(iconTarget)
		}
		columns[0] = append(columns[0], playerRowStyle.Render(fmt.Sprintf("%s%s%s (%s)", pointer, targeted, p.Name, p.HeroName)))
		columns[1] = append(columns[1], fmt.Sprintf("    %s %2d - %s %2d - %s %2d", icon(iconHand), len(p.Hand), icon(iconDeck), p.DeckCount, icon(iconDiscard), len(p.Discard)))
	}

	players := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, columns[0]...),
		lipgloss.JoinVertical(lipgloss.Left, columns[1]...),
	)

	var box strings.Builder
	box.WriteString(activePlayersBoxTitle.String() + boxStyle.Render(players))

	return box.String()
}

func (m playModel) renderArtifacts() string {
	var artifactsBoxTitle strings.Builder
	artifactsBoxTitle.WriteString("┌── TEAM ARTIFACTS ")
	artifactsBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(artifactsBoxTitle.String())-1)) + "┐\n")

	var artifacts strings.Builder
	if len(m.snapshot.Artifacts) == 0 {
		artifacts.WriteString("(No team artifacts)\n")

		return artifactsBoxTitle.String() + artifactsStyle.Render(artifacts.String())
	}

	for _, a := range m.snapshot.Artifacts {
		status := "[ READY ]"
		switch a.Used {
		case true:
			status = "[ USED  ]"
		}
		artifacts.WriteString(fmt.Sprintf("• [%s] %s  %s\n  \"%s\"\n", a.Color, a.Name, status, a.Description))
	}

	return artifactsBoxTitle.String() + artifactsStyle.Render(artifacts.String())
}

func (m playModel) renderActiveCurses() string {
	var cursesBoxTitle strings.Builder
	cursesBoxTitle.WriteString("┌── ACTIVE CURSES ")
	cursesBoxTitle.WriteString(strings.Repeat("─", max(0, (screenWidth/2)-lipgloss.Width(cursesBoxTitle.String())-1)) + "┐\n")

	var curses strings.Builder
	if len(m.snapshot.ActiveCurses) == 0 {
		curses.WriteString("(No active curses)\n")

		return cursesBoxTitle.String() + cursesStyle.Render(curses.String())
	}
	for _, c := range m.snapshot.ActiveCurses {
		curses.WriteString(fmt.Sprintf("%s [%s]\n  %s\n", icon(iconCurse), c.Name, c.Description))
	}

	return cursesBoxTitle.String() + cursesStyle.Render(curses.String())
}

func (m playModel) renderPlayerDashboard() string {
	player := m.currentPlayer()
	if player.Id == "" {
		return ""
	}

	var hudHeader strings.Builder
	playerBoxTitle := fmt.Sprintf("┌── YOUR HERO: %s (%s) ", strings.ToUpper(player.HeroClass.String()), player.Name)
	hudHeader.WriteString(playerHeaderStyle.Render(playerBoxTitle))
	hudHeader.WriteString(strings.Repeat("─", max(0, screenWidth-lipgloss.Width(playerBoxTitle)-1)) + "┐\n")

	var hudCol1 strings.Builder
	hudCol1.WriteString(fmt.Sprintf("[Deck]\n %s %d cards", icon(iconDeck), player.DeckCount))

	var hudCol2 strings.Builder
	topDiscard := "None"
	if len(player.Discard) > 0 {
		topDiscard = player.Discard[len(player.Discard)-1].Name
	}
	hudCol2.WriteString(fmt.Sprintf("[Discard]\n %s %d cards (Top: %s)", icon(iconDiscard), len(player.Discard), topDiscard))

	var hudCol3 strings.Builder
	hudCol3.WriteString(fmt.Sprintf("[Ability] %s\n %s", player.AbilityName, player.AbilityDescription))

	var hud strings.Builder
	hud.WriteString(hudHeader.String())
	hud.WriteString(playerHudStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		playHudColStyle.Width(23).Render(hudCol1.String()),
		playHudColStyle.Width(51).Render(hudCol2.String()),
		playHudLastColStyle.Render(hudCol3.String()),
	)))
	hud.WriteString("\n")

	var playerHand strings.Builder
	playerHand.WriteString(fmt.Sprintf(
		"YOUR HAND (Row: %d/%d - Selected: %d - Total: %d):\n",
		m.currentPlayerHandRow+1,
		m.maxHandRows(),
		len(m.selectedCards),
		len(player.Hand),
	))

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
		cardBody.WriteString(card.Description)
	}

	style := cardStyle
	headerStyle := lipgloss.NewStyle().Margin(0).Padding(0)
	if m.selectedCards[card.Id] {
		style = selectedCardStyle
		headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	}

	var box strings.Builder
	box.WriteString(lipgloss.JoinVertical(lipgloss.Left, headerStyle.Render(cardHeader.String()), style.Render(cardBody.String())))

	return box.String()
}

func (m playModel) renderVictory() string {
	var s strings.Builder
	s.WriteString(bannerBoxStyle.Render(victoryTitleStyle.Render(icon(iconVictory) + " VICTORY ACHIEVED! The Gungeon has been conquered! " + icon(iconVictory) + "\n\nAll dungeon doors cleared and boss defeated!")))
	s.WriteString("\n\n" + footerStyle.Render("[ENTER] Return to Lobby  •  [R] Start Next Floor  •  [Ctrl+C] Quit"))

	return s.String()
}

func (m playModel) renderDefeat() string {
	var s strings.Builder
	s.WriteString(bannerBoxStyle.Render(defeatTitleStyle.Render(icon(iconDefeat) + " DEFEAT - The Gungeon claimed your party! " + icon(iconDefeat) + "\n\nTime expired or your team was overwhelmed.")))
	s.WriteString("\n\n" + footerStyle.Render("[ENTER] Return to Lobby  •  [Ctrl+C] Quit"))

	return s.String()
}

func (m playModel) renderFooter() string {
	var footer strings.Builder
	switch m.targeting.kind {
	case TargetKindDoor:
		footer.WriteString("[↑/↓] / [1-2] Select Target Door • [ENTER] / [SPACE] Confirm Target • [ESC] Cancel Targeting • [Ctrl+C] Quit")

		return footerStyle.Render(footer.String())
	case TargetKindPlayer, TargetKindTwoPlayers:
		footer.WriteString("[↑/↓] / [1-6] Select Target Player • [ENTER] / [SPACE] Confirm Target • [ESC] Cancel Targeting • [Ctrl+C] Quit")

		return footerStyle.Render(footer.String())
	case TargetKindNone:
	}
	if m.promptIsActive() && !m.dismissedPrompt {
		footer.WriteString("[1-5] Choose Option • [TAB] Switch Hero • [ENTER] / [SPACE] Submit Choice • [ESC] Dismiss Modal • [Ctrl+C] Quit")

		return footerStyle.Render(footer.String())
	}

	footer.WriteString("[1-5] Select Card • [SPACE] Play Cards • [C] Clear Selection • [A] Hero Ability • [D] Discard Cards (Curse)")
	if m.config.UseExtension {
		footer.WriteString(" • [U] Use Artifact")
	}
	footer.WriteString(" • [Ctrl+C] Quit")

	return footerStyle.Render(footer.String())
}

func (m playModel) currentPlayer() game.PlayerDTO {
	if len(m.snapshot.Players) == 0 || m.currentPlayerIdx >= len(m.snapshot.Players) {
		return game.PlayerDTO{}
	}

	return m.snapshot.Players[m.currentPlayerIdx]
}

func (m playModel) toggleSelectCard(i int) playModel {
	hand := m.currentPlayer().Hand
	index := m.currentPlayerHandRow*5 + i - 1

	if index < 0 || index >= len(hand) {
		return m
	}

	cardID := hand[index].Id
	if m.selectedCards[cardID] {
		delete(m.selectedCards, cardID)

		return m
	}

	m.selectedCards[cardID] = true

	return m
}

func (m playModel) playSelectedCards() (playModel, tea.Cmd) {
	for cID := range m.selectedCards {
		targetType := game.TargetNone
		for _, c := range m.currentPlayer().Hand {
			if c.Id != cID {
				continue
			}
			if c.Kind == game.PlayerCardAction {
				targetType = c.TargetType
			}
			break
		}

		switch targetType {
		case game.TargetCard:
			if len(m.snapshot.OpenedDoors) == 2 {
				m.targeting = targetingState{
					kind:         TargetKindDoor,
					sourceCardID: cID,
					targetCardID: m.snapshot.OpenedDoors[0].Id,
					targetIndex:  0,
				}

				return m, nil
			}
			if len(m.snapshot.OpenedDoors) == 0 {
				return m, nil
			}
			err := m.controller.Dispatch(m.ctx, game.PlayCardCmd{
				PlayerID:     m.currentPlayer().Id,
				CardID:       cID,
				TargetCardID: m.snapshot.OpenedDoors[0].Id,
			})
			delete(m.selectedCards, cID)
			if err != nil {
				return m, forwardError(err)
			}

			continue

		case game.TargetPlayer:
			m.targeting = targetingState{
				kind:            TargetKindPlayer,
				sourceCardID:    cID,
				targetPlayerIDs: make(map[game.PlayerID]bool),
				targetIndex:     0,
			}

			return m, nil

		case game.TargetTwoPlayers:
			m.targeting = targetingState{
				kind:            TargetKindTwoPlayers,
				sourceCardID:    cID,
				targetPlayerIDs: make(map[game.PlayerID]bool),
				targetIndex:     0,
			}

			return m, nil

		case game.TargetCurse:

		case game.TargetNone:
			err := m.controller.Dispatch(m.ctx, game.PlayCardCmd{
				PlayerID:     m.currentPlayer().Id,
				CardID:       cID,
				TargetCardID: 0,
			})
			delete(m.selectedCards, cID)
			if err != nil {
				return m, forwardError(err)
			}
		}
	}

	return m, nil
}

func (m playModel) useHeroAbility() (playModel, tea.Cmd) {
	if len(m.selectedCards) != 3 {
		return m, nil
	}

	cardIDs := make([]game.CardID, 3)
	i := 0
	for cID := range m.selectedCards {
		cardIDs[i] = cID
		i++
	}

	player := m.currentPlayer()
	switch heroAbilityTargetKind(player.HeroClass) {
	case TargetKindDoor:
		if len(m.snapshot.OpenedDoors) == 0 {
			return m, nil
		}
		if len(m.snapshot.OpenedDoors) == 2 {
			m.targeting = targetingState{
				kind:         TargetKindDoor,
				isAbility:    true,
				targetCardID: m.snapshot.OpenedDoors[0].Id,
				targetIndex:  0,
			}

			return m, nil
		}
		err := m.controller.Dispatch(m.ctx, game.UseHeroAbilityCmd{
			PlayerID:       player.Id,
			DiscardCardIDs: cardIDs,
			TargetCardID:   m.snapshot.OpenedDoors[0].Id,
		})
		clear(m.selectedCards)
		if err != nil {
			return m, forwardError(err)
		}

		return m, nil

	case TargetKindPlayer:
		m.targeting = targetingState{
			kind:            TargetKindPlayer,
			isAbility:       true,
			targetPlayerIDs: make(map[game.PlayerID]bool),
			targetIndex:     0,
		}

		return m, nil

	case TargetKindTwoPlayers:
	case TargetKindNone:
	}

	err := m.controller.Dispatch(m.ctx, game.UseHeroAbilityCmd{
		PlayerID:       player.Id,
		DiscardCardIDs: cardIDs,
		TargetCardID:   0,
		TargetPlayerID: "",
	})
	clear(m.selectedCards)
	if err != nil {
		return m, forwardError(err)
	}

	return m, nil
}

func (m playModel) discardCards() (playModel, tea.Cmd) {
	cardIDs := make([]game.CardID, len(m.selectedCards))
	i := 0
	for cID := range m.selectedCards {
		cardIDs[i] = cID
		i++
	}

	err := m.controller.Dispatch(m.ctx, game.DiscardCardsCmd{
		PlayerID: m.currentPlayer().Id,
		CardIDs:  cardIDs,
	})
	clear(m.selectedCards)
	if err != nil {
		return m, forwardError(err)
	}

	return m, nil
}

func (m playModel) useArtifact() (playModel, tea.Cmd) {
	if !m.config.UseExtension {
		return m, nil
	}

	var readyArtifacts []game.ArtifactDTO
	for _, a := range m.snapshot.Artifacts {
		if !a.Used {
			readyArtifacts = append(readyArtifacts, a)
		}
	}
	if len(readyArtifacts) == 0 {
		return m, nil
	}

	art := readyArtifacts[0]
	switch art.Id {
	case game.BattleAxeID, game.TheInfinityScrollID:
		if len(m.snapshot.OpenedDoors) == 0 {
			return m, nil
		}
		if len(m.snapshot.OpenedDoors) == 2 {
			m.targeting = targetingState{
				kind:         TargetKindDoor,
				isArtifact:   true,
				artifactID:   art.Id,
				actionIndex:  game.FirstArtifactAction,
				targetCardID: m.snapshot.OpenedDoors[0].Id,
				targetIndex:  0,
			}

			return m, nil
		}

		err := m.controller.Dispatch(m.ctx, game.UseArtifactCmd{
			PlayerID:    m.currentPlayer().Id,
			ArtifactID:  art.Id,
			ActionIndex: game.FirstArtifactAction,
			TargetID:    m.snapshot.OpenedDoors[0].Id,
		})
		if err != nil {
			return m, forwardError(err)
		}

		return m, nil
	}

	err := m.controller.Dispatch(m.ctx, game.UseArtifactCmd{
		PlayerID:   m.currentPlayer().Id,
		ArtifactID: art.Id,
	})
	if err != nil {
		return m, forwardError(err)
	}

	return m, nil
}

func (m playModel) dispatchTargetedAction() (playModel, tea.Cmd) {
	if m.targeting.targetCardID == 0 && len(m.targeting.targetPlayerIDs) == 0 {
		m.targeting = targetingState{}

		return m, nil
	}

	targetCardID := m.targeting.targetCardID
	targetPlayerIDs := make([]game.PlayerID, len(m.targeting.targetPlayerIDs))
	i := 0
	for pID := range maps.Keys(m.targeting.targetPlayerIDs) {
		targetPlayerIDs[i] = pID
		i++
	}

	var cmd tea.Cmd

	switch {
	case m.targeting.isAbility:
		if len(m.selectedCards) == 3 {
			var targetPlayerID game.PlayerID
			if len(targetPlayerIDs) > 0 {
				targetPlayerID = targetPlayerIDs[0]
			}
			cardIDs := make([]game.CardID, 3)
			i := 0
			for cID := range m.selectedCards {
				cardIDs[i] = cID
				i++
			}
			err := m.controller.Dispatch(m.ctx, game.UseHeroAbilityCmd{
				PlayerID:       m.currentPlayer().Id,
				DiscardCardIDs: cardIDs,
				TargetCardID:   targetCardID,
				TargetPlayerID: targetPlayerID,
			})
			if err != nil {
				cmd = forwardError(err)
			}
			clear(m.selectedCards)
		}

	case m.targeting.isArtifact:
		err := m.controller.Dispatch(m.ctx, game.UseArtifactCmd{
			PlayerID:    m.currentPlayer().Id,
			ArtifactID:  m.targeting.artifactID,
			ActionIndex: m.targeting.actionIndex,
			TargetID:    targetCardID,
		})
		if err != nil {
			cmd = forwardError(err)
		}

	case m.targeting.sourceCardID != 0:
		if m.targeting.kind == TargetKindTwoPlayers && len(targetPlayerIDs) < 2 {
			return m, nil
		}

		cID := m.targeting.sourceCardID
		err := m.controller.Dispatch(m.ctx, game.PlayCardCmd{
			PlayerID:        m.currentPlayer().Id,
			CardID:          cID,
			TargetCardID:    targetCardID,
			TargetPlayerIDs: targetPlayerIDs,
		})
		if err != nil {
			cmd = forwardError(err)
		}
		delete(m.selectedCards, cID)
	}

	m.targeting = targetingState{}

	return m, cmd
}

func (m playModel) submitPromptChoice() (playModel, tea.Cmd) {
	pi := m.snapshot.PendingInteraction
	if pi == nil {
		return m, nil
	}

	p := m.currentPlayer()

	switch pi.Kind {
	case game.InteractionTeamChoicePlayer:
		if m.promptChoiceIdx < 0 || m.promptChoiceIdx >= len(m.snapshot.Players) {
			return m, nil
		}

		targetPlayer := m.snapshot.Players[m.promptChoiceIdx]
		err := m.controller.Dispatch(m.ctx, game.SubmitPromptChoiceCmd{
			PlayerID:       p.Id,
			TargetPlayerID: targetPlayer.Id,
		})
		if err != nil {
			return m, forwardError(err)
		}

	case game.InteractionTeamChoiceResource:
		resources := []game.ResourceType{game.Sword, game.Arrow, game.Shield, game.Jump, game.Scroll}
		if m.promptChoiceIdx < 0 || m.promptChoiceIdx >= len(resources) {
			return m, nil
		}

		res := resources[m.promptChoiceIdx]
		err := m.controller.Dispatch(m.ctx, game.SubmitPromptChoiceCmd{
			PlayerID: p.Id,
			Resource: res,
		})
		if err != nil {
			return m, forwardError(err)
		}

	case game.InteractionTeamChoiceArtifact:
		if m.promptChoiceIdx < 0 || m.promptChoiceIdx >= len(m.snapshot.Artifacts) {
			return m, nil
		}

		art := m.snapshot.Artifacts[m.promptChoiceIdx]
		err := m.controller.Dispatch(m.ctx, game.SubmitPromptChoiceCmd{
			PlayerID:         p.Id,
			TargetArtifactID: art.Id,
		})
		if err != nil {
			return m, forwardError(err)
		}

	case game.InteractionPlayerDiscardCards:
		expected := discardExpected(pi, p)
		cardIDs := make([]game.CardID, 0, expected)
		for _, c := range p.Hand {
			if m.selectedCards[c.Id] {
				cardIDs = append(cardIDs, c.Id)
			}
		}
		if len(cardIDs) != expected {
			return m, nil
		}

		err := m.controller.Dispatch(m.ctx, game.SubmitPromptChoiceCmd{
			PlayerID: p.Id,
			CardIDs:  cardIDs,
		})
		if err != nil {
			return m, forwardError(err)
		}

		clear(m.selectedCards)

	case game.InteractionPlayerDonatesHand:
		candidates := m.donationCandidates(p)
		if len(candidates) == 0 || m.promptChoiceIdx < 0 || m.promptChoiceIdx >= len(candidates) {
			return m, nil
		}

		targetPlayer := candidates[m.promptChoiceIdx]
		err := m.controller.Dispatch(m.ctx, game.SubmitPromptChoiceCmd{
			PlayerID:       p.Id,
			TargetPlayerID: targetPlayer.Id,
		})
		if err != nil {
			return m, forwardError(err)
		}
	case game.InteractionNoInteraction:
	}

	// Only reached when the choice was dispatched (invalid choices return early), so the prompt stays open
	// until the player really answered it.
	m.promptChoiceIdx = 0

	return m, nil
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

func (m playModel) maxHandRows() int {
	handLen := len(m.currentPlayer().Hand)
	if handLen == 0 {
		return 1
	}

	return (handLen + 4) / 5
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

func interactionKindToString(kind game.InteractionKind) string {
	switch kind {
	case game.InteractionTeamChoicePlayer:
		return "Team Choice: Select Player"
	case game.InteractionTeamChoiceResource:
		return "Team Choice: Select Resource"
	case game.InteractionTeamChoiceArtifact:
		return "Team Choice: Select Artifact"
	case game.InteractionPlayerDiscardCards:
		return "Player Action: Discard Cards"
	case game.InteractionPlayerDonatesHand:
		return "Player Action: Donate Hand"
	default:
		return "Team Prompt"
	}
}

func heroAbilityTargetKind(class game.HeroClass) TargetKind {
	switch class {
	case game.Sorceress, game.Paladin, game.Barbarian, game.Gladiator, game.Ranger, game.Ninja:
		return TargetKindDoor
	case game.Huntress, game.Shaman:
		return TargetKindPlayer
	case game.Wizard, game.Thief, game.Valkyrie, game.Druid:
		return TargetKindNone
	}

	return TargetKindNone
}
