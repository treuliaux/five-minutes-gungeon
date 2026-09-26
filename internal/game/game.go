package game

import (
	"fmt"
	"slices"
	"time"
)

type Status int

const (
	Waiting Status = iota
	Victory
	Defeat
	Playing
)

const (
	cardPlayDebounceDuration = time.Second / 2
	gameDuration             = 5 * time.Minute
)

type Game struct {
	Players                []*Player
	HandSize               int
	Dungeon                *Dungeon
	PlayField              *Playfield
	IsFightingBoss         bool
	LastPlayedCardTimer    time.Duration
	InGameTimer            time.Duration
	RealElapsedTime        time.Duration
	IsTimeFrozen           bool
	Status                 Status
	UseExtension           bool
	PendingInteraction     PendingInteraction
	CurseExpectingDiscards map[*Player]int
	PlayerCardsMap         map[CardID]PlayerCard
	DungeonCardsMap        map[CardID]DungeonCard
}

type GameEngine interface {
	DefeatDoor(target DungeonCard) ([]Event, error)
	OpenDoor() ([]Event, error)
	SendDungeonCardBottomDungeon(card DungeonCard) ([]Event, error)
	StopTime(player *Player) ([]Event, error)
	ClearStopTimeCurse()

	ListPlayers() []*Player
	ActiveDoors(f *DoorsFilter) []DungeonCard
	ListArtifacts() []*ArtifactCard
	DungeonCardByID(id CardID) (DungeonCard, error)
	PlayerCardsByIDs(id []CardID) ([]PlayerCard, error)
	PlayerByID(id PlayerID) (*Player, error)
	ArtifactByID(id ArtifactID) (*ArtifactCard, error)

	PlayArbitraryCard(player *Player, card PlayerCard) ([]Event, error)
	RemoveDungeonCard(card DungeonCard) ([]Event, error)
	RemovePlayerCard(card PlayerCard) ([]Event, error)
	DiscardTopCardFromDungeon() ([]Event, error)
	RefillPlayerHand(player *Player) ([]Event, error)
}

func NewGame() *Game {
	return &Game{
		PlayField:              NewPlayfield(),
		Status:                 Waiting,
		UseExtension:           true,
		CurseExpectingDiscards: make(map[*Player]int),
		DungeonCardsMap:        make(map[CardID]DungeonCard, 62),
		PlayerCardsMap:         make(map[CardID]PlayerCard, 252),
	}
}

func (g *Game) Tick(delta time.Duration) ([]Event, error) {
	g.RealElapsedTime += delta
	if g.Status == Victory || g.Status == Defeat {
		return nil, nil
	}
	if g.Status != Playing {
		return nil, fmt.Errorf("game is not in playing state")
	}
	if g.IsTimeFrozen {
		delta = 0
	}
	g.InGameTimer += delta
	if g.InGameTimer >= gameDuration {
		g.Status = Defeat

		return []Event{GameLostEvent{}}, nil
	}
	if g.PendingInteraction != nil {
		return nil, nil
	}
	events, err := g.resolveActiveEvents()
	if err != nil {
		return events, err
	}
	if (g.actionDebounce() && !g.IsFightingBoss) || !g.PlayField.IsPlayfieldBeaten() {
		return events, nil
	}

	defeatAllDoorsEvents, err := g.PlayField.DefeatAllDoors()
	events = append(events, defeatAllDoorsEvents...)
	if err != nil {
		return events, err
	}
	if g.IsFightingBoss {
		g.Status = Victory
		events = append(events, GameWonEvent{})

		return events, nil
	}
	openDoorEvents, err := g.OpenDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) Apply(cmd Command) ([]Event, error) {
	var err error
	var events []Event

	switch cmd := cmd.(type) {
	case AddPlayerCmd:
		events, err = g.addPlayer(cmd)
	case StartCmd:
		events, err = g.start()
	case PlayCardCmd:
		events, err = g.playCard(cmd)
	case DiscardCardsCmd:
		events, err = g.discardCard(cmd)
	case UseHeroAbilityCmd:
		events, err = g.useHeroAbility(cmd)
	case SubmitPromptChoiceCmd:
		events, err = g.submitPromptChoice(cmd)
	case UseArtifactCmd:
		events, err = g.useArtifact(cmd)
	}
	if cmd.Reply() != nil {
		cmd.Reply() <- err
	}

	return events, err
}

func (g *Game) DefeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event

	if _, ok := target.(*CurseCard); ok {
		return nil, fmt.Errorf("curses cannot be defeated")
	}

	defeatDoorEvents, err := g.PlayField.DefeatDoor(target)
	events = append(events, defeatDoorEvents...)
	if err != nil {
		return events, err
	}

	if _, ok := target.(*BossMat); ok || g.IsFightingBoss {
		g.Status = Victory
		events = append(events, GameWonEvent{})

		return events, nil
	}

	if g.PlayField.HasDoorsOpened() {
		return events, nil
	}

	events = append(events, g.clearField()...)

	openDoorEvents, err := g.OpenDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) StopTime(player *Player) ([]Event, error) {
	if g.IsTimeFrozen {
		return nil, fmt.Errorf("time is already frozen")
	}
	if g.HasActiveCurseEffect(TimeCannotBeStopped) {
		return player.VoidHand(TimeCannotBeStopped, g)
	}
	if g.HasActiveCurseEffect(ThreeDiscardsWhenTimeStops) {
		for _, p := range g.Players {
			g.CurseExpectingDiscards[p] = 3
		}
	}
	g.IsTimeFrozen = true

	return []Event{TimeFrozenEvent{ByPlayerID: player.Id}}, nil
}

func (g *Game) ClearStopTimeCurse() {
	g.CurseExpectingDiscards = make(map[*Player]int, len(g.Players))
}

func (g *Game) ListPlayers() []*Player {
	return g.Players
}

func (g *Game) ListArtifacts() []*ArtifactCard {
	return g.PlayField.Artifacts
}

func (g *Game) ActiveDoors(f *DoorsFilter) []DungeonCard {
	if f == nil {
		f = NewDoorsFilter().AddEverything()
	}
	var picked []DungeonCard
	var playfieldHaystack []DungeonCard
	for _, door := range g.PlayField.OpenedDoors {
		playfieldHaystack = append(playfieldHaystack, door)
	}
	for _, curse := range g.PlayField.ActiveCurses {
		playfieldHaystack = append(playfieldHaystack, curse)
	}
	for _, door := range playfieldHaystack {
		switch d := door.(type) {
		case *BossMat:
			if f.IncludeBossMat {
				picked = append(picked, d)
			}
		case *CurseCard:
			if slices.Contains(f.ChallengeKinds, ChallengeCurse) {
				picked = append(picked, d)
			}
		case *EventCard:
			if slices.Contains(f.ChallengeKinds, ChallengeEvent) {
				picked = append(picked, d)
			}
		case *MiniBossCard:
			if slices.Contains(f.ChallengeKinds, ChallengeMiniBoss) {
				picked = append(picked, d)
			}
		case *DoorCard:
			if slices.Contains(f.DoorKinds, d.Type) {
				picked = append(picked, d)
			}
		}
	}

	return picked
}

func (g *Game) SendDungeonCardBottomDungeon(card DungeonCard) ([]Event, error) {
	switch card.(type) {
	case *BossMat, *EventCard:
		return nil, fmt.Errorf("card cannot be sent back to dungeon")
	}
	events, err := g.PlayField.RemoveDungeonCard(g, card)
	if err != nil {
		return events, err
	}
	g.Dungeon.PutDoorBelowDeck(card)

	if len(g.PlayField.OpenedDoors) == 0 {
		openDoorEvents, err := g.OpenDoor()
		events = append(events, openDoorEvents...)
		if err != nil {
			return events, err
		}
	}

	return append(events, DungeonCardSentToBottomEvent{CardID: card.ID()}), nil
}

func (g *Game) DiscardTopCardFromDungeon() ([]Event, error) {
	if len(g.Dungeon.Doors) == 0 {
		return nil, nil
	}

	return []Event{DungeonCardDiscardedEvent{CardID: g.Dungeon.OpenDoor().ID()}}, nil
}

func (g *Game) PlayArbitraryCard(p *Player, card PlayerCard) ([]Event, error) {
	return g.PlayField.AddPlayerCard(p, card)
}

func (g *Game) RemovePlayerCard(card PlayerCard) ([]Event, error) {
	return g.PlayField.RemovePlayerCard(card)
}

func (g *Game) RemoveDungeonCard(card DungeonCard) ([]Event, error) {
	return g.PlayField.RemoveDungeonCard(g, card)
}

func (g *Game) OpenDoor() ([]Event, error) {
	if g.Dungeon == nil {
		return nil, nil
	}
	var events []Event
	var err error
	loop := true
	for loop || (g.HasActiveCurseEffect(DoorsOpenInPairs) && len(g.PlayField.OpenedDoors) != 2) {
		loop = len(g.PlayField.OpenedDoors) == 0
		card := g.Dungeon.OpenDoor()
		if card == nil {
			break
		}
		var addDungeonCardEvents []Event
		addDungeonCardEvents, err = g.PlayField.AddDungeonCard(card, g)
		events = append(events, addDungeonCardEvents...)
		if err != nil {
			return events, err
		}
		switch card.(type) {
		case *DoorCard, *MiniBossCard, *EventCard:
			loop = false
		case *BossMat:
			g.IsFightingBoss = true
			loop = false
		case *CurseCard:
		}
	}

	return events, nil
}

func (g *Game) HasActiveCurseEffect(effect GameCurseEffect) bool {
	if g.PlayField == nil {
		return false
	}
	for _, curse := range g.PlayField.ActiveCurses {
		if curse.Effect == effect {
			return true
		}
	}

	return false
}

func (g *Game) TargetHandSize() int {
	if g.HasActiveCurseEffect(HandSizeLimitedToThree) {
		return min(g.HandSize, 3)
	}

	return g.HandSize
}

func (g *Game) addPlayer(cmd AddPlayerCmd) ([]Event, error) {
	if g.Status != Waiting {
		return nil, fmt.Errorf("game is not in waiting state")
	}
	if slices.ContainsFunc(g.Players, func(player *Player) bool {
		return player.Hero.Class == cmd.Class
	}) {
		return nil, fmt.Errorf("player with class %d already exists", cmd.Class)
	}
	player, err := NewPlayer(cmd.Name, cmd.Class, g.UseExtension)
	if err != nil {
		return nil, err
	}
	g.Players = append(g.Players, player)

	return []Event{PlayerAddedEvent{PlayerID: player.Id}}, nil
}

func (g *Game) start() ([]Event, error) {
	if g.Status != Waiting {
		return nil, fmt.Errorf("game has already started")
	}
	if len(g.Players) < 2 || len(g.Players) > 6 {
		return nil, fmt.Errorf("game has incorrect number of players: %d", len(g.Players))
	}

	g.generateDungeon()
	if g.UseExtension {
		var deckColors []DeckColor
		for _, p := range g.Players {
			deckColors = append(deckColors, p.Hero.Color)
		}
		g.PlayField.SetupArtifacts(deckColors)
	}
	g.registerCards()

	g.Status = Playing
	g.CurseExpectingDiscards = make(map[*Player]int, len(g.Players))

	var events []Event
	events = append(events, GameStartedEvent{})

	for _, player := range g.Players {
		drawEvents, err := g.RefillPlayerHand(player)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	openDoorEvents, err := g.OpenDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) playCard(cmd PlayCardCmd) ([]Event, error) {
	if g.PendingInteraction != nil {
		return nil, fmt.Errorf("cannot play cards: waiting for event resolution (%s)", g.PendingInteraction)
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	c, err := g.PlayerCardByID(cmd.CardID)
	if err != nil {
		return nil, err
	}
	if !p.HasCardInHand(c) {
		return nil, fmt.Errorf("player does not have card in hand")
	}
	if g.HasActiveCurseEffect(HandSizeLimitedToThree) && len(p.Hand) > 3 {
		return p.VoidHand(HandSizeLimitedToThree, g)
	}
	if g.HasActiveCurseEffect(ThreeDiscardsWhenTimeStops) {
		if _, ok := g.CurseExpectingDiscards[p]; ok {
			delete(g.CurseExpectingDiscards, p)
			return p.VoidHand(ThreeDiscardsWhenTimeStops, g)
		}
	}
	if _, ok := c.(*ActionCard); ok && g.HasActiveCurseEffect(ActionsCannotBePlayed) {
		return p.VoidHand(ActionsCannotBePlayed, g)
	}

	p.RemoveCardFromHand(c)
	fieldEvents, err := g.PlayField.AddPlayerCard(p, c)
	if err != nil {
		p.Hand = append(p.Hand, c)
		return nil, err
	}

	var events []Event
	g.LastPlayedCardTimer = g.RealElapsedTime
	events = append(events, fieldEvents...)
	wasTimeFrozen := g.IsTimeFrozen
	if g.IsTimeFrozen {
		g.IsTimeFrozen = false
		events = append(events, TimeUnfrozenEvent{ByPlayerID: p.Id})
	}

	var actionEvents []Event
	if ac, ok := c.(*ActionCard); ok {
		var targetCard DungeonCard
		if cmd.TargetCardID != 0 {
			targetCard, err = g.DungeonCardByID(cmd.TargetCardID)
			if err != nil {
				p.Hand = append(p.Hand, c)

				return events, err
			}
		}
		var targetPlayers []*Player
		if len(cmd.TargetPlayerIDs) > 0 {
			targetPlayers, err = g.PlayersByIDs(cmd.TargetPlayerIDs)
			if err != nil {
				p.Hand = append(p.Hand, c)

				return events, err
			}
		}
		ctx := &CardActionContext{
			engine:        g,
			Player:        p,
			Card:          c,
			TargetCard:    targetCard,
			TargetPlayers: targetPlayers,
		}
		actionEvents, err = ac.Action.Execute(ctx)
		if err != nil {
			p.Hand = append(p.Hand, c)
			g.PlayField.Field = slices.DeleteFunc(g.PlayField.Field, func(item PlayerCard) bool { return item == c })
			g.IsTimeFrozen = wasTimeFrozen

			return nil, err
		}
	}
	events = append(events, actionEvents...)

	cardDrawnEvents, err := g.RefillPlayerHand(p)
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) useHeroAbility(cmd UseHeroAbilityCmd) ([]Event, error) {
	if g.PendingInteraction != nil {
		return nil, fmt.Errorf("cannot play cards: waiting for event resolution (%s)", g.PendingInteraction)
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	discardCards, err := g.PlayerCardsByIDs(cmd.DiscardCardIDs)
	if err != nil {
		return nil, err
	}
	var targetCard DungeonCard
	if cmd.TargetCardID != 0 {
		targetCard, err = g.DungeonCardByID(cmd.TargetCardID)
		if err != nil {
			return nil, err
		}
	}
	var targetPlayer *Player
	if cmd.TargetPlayerID != "" {
		targetPlayer, err = g.PlayerByID(cmd.TargetPlayerID)
		if err != nil {
			return nil, err
		}
	}
	if g.HasActiveCurseEffect(AbilitiesCannotBePlayed) {
		return p.VoidHand(AbilitiesCannotBePlayed, g)
	}
	if g.HasActiveCurseEffect(HandSizeLimitedToThree) && len(p.Hand) > 3 {
		return p.VoidHand(HandSizeLimitedToThree, g)
	}
	if g.HasActiveCurseEffect(ThreeDiscardsWhenTimeStops) {
		if _, ok := g.CurseExpectingDiscards[p]; ok {
			delete(g.CurseExpectingDiscards, p)
			return p.VoidHand(ThreeDiscardsWhenTimeStops, g)
		}
	}

	discardCards = uniqueCards(discardCards)
	if len(discardCards) != 3 {
		return nil, fmt.Errorf("player must discard 3 different cards")
	}
	for _, card := range discardCards {
		if !p.HasCardInHand(card) {
			return nil, fmt.Errorf("player does not have card in hand")
		}
	}

	var events []Event
	for _, card := range discardCards {
		discardEvents, err := p.DiscardCard(card)
		events = append(events, discardEvents...)
		if err != nil {
			return nil, err
		}
	}

	ctx := &AbilityContext{
		engine:       g,
		Player:       p,
		TargetCard:   targetCard,
		TargetPlayer: targetPlayer,
	}
	abilityEvents, err := p.Hero.Ability.Execute(ctx)
	if err != nil {
		for _, c := range discardCards {
			if !slices.Contains(p.Hand, c) {
				p.Hand = append(p.Hand, c)
			}
		}
		return nil, err
	}

	events = append(events, HeroAbilityUsedEvent{ByPlayerID: p.Id})
	events = append(events, abilityEvents...)

	cardDrawnEvents, err := g.RefillPlayerHand(p)
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) useArtifact(cmd UseArtifactCmd) ([]Event, error) {
	if !g.UseExtension {
		return nil, fmt.Errorf("extension must be enabled to use artifacts")
	}
	if g.PendingInteraction != nil {
		return nil, fmt.Errorf("cannot play cards: waiting for event resolution (%s)", g.PendingInteraction)
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	var target DungeonCard
	if cmd.TargetID != 0 {
		target, err = g.DungeonCardByID(cmd.TargetID)
		if err != nil {
			return nil, err
		}
	}
	artifact, err := g.ArtifactByID(cmd.ArtifactID)
	if err != nil {
		return nil, err
	}
	if !g.PlayField.ArtifactCanBePlayed(artifact) {
		return nil, fmt.Errorf("artifact cannot be played")
	}

	ctx := ArtifactActionContext{
		engine:       g,
		Player:       p,
		Artifact:     artifact,
		ChosenAction: cmd.ActionIndex,
		Target:       target,
	}
	artifactEvents, err := artifact.Action.Execute(ctx)
	if err != nil {
		return nil, err
	}
	artifact.Used = true

	events := []Event{ArtifactUsedEvent{ByPlayerID: p.Id}}

	return append(events, artifactEvents...), nil
}

func (g *Game) RefillPlayerHand(player *Player) ([]Event, error) {
	if g.HandSize == 0 {
		g.determineHandSize()
	}

	var events []Event
	var err error
	nbToDraw := g.TargetHandSize() - len(player.Hand)
	if nbToDraw > 0 {
		events, err = player.DrawCardsFromDeck(nbToDraw)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (g *Game) clearField() []Event {
	fieldEvents, err := g.PlayField.ClearField()
	if err != nil {
		return nil
	}
	g.LastPlayedCardTimer = g.RealElapsedTime

	return fieldEvents
}

func (g *Game) determineHandSize() {
	switch len(g.Players) {
	case 2:
		g.HandSize = 5
	case 3:
		g.HandSize = 4
	case 4, 5, 6:
		g.HandSize = 3
	}
}

func (g *Game) generateDungeon() {
	if g.Dungeon == nil {
		if g.UseExtension {
			g.Dungeon = NewExtensionDungeon(1, len(g.Players))

			return
		}
		g.Dungeon = NewBaseDungeon(1, len(g.Players))
	}
}

func (g *Game) actionDebounce() bool {
	return g.RealElapsedTime-g.LastPlayedCardTimer <= cardPlayDebounceDuration
}

func (g *Game) resolveActiveEvents() ([]Event, error) {
	var events []Event
	for _, card := range slices.Clone(g.PlayField.OpenedDoors) {
		eventCard, ok := card.(*EventCard)
		if !ok || eventCard.Action == nil {
			continue
		}
		if g.InGameTimer-eventCard.OpenedTime < 2*time.Second {
			return nil, nil
		}

		ctx := &CardEventContext{
			engine: g,
			Card:   eventCard,
		}
		if interaction := eventCard.Action.Interaction(ctx); interaction != nil {
			g.PendingInteraction = interaction

			initEvents, err := interaction.Init(ctx)
			events = append(events, initEvents...)
			if err != nil {
				return events, err
			}

			promptEvent := EventPromptOpenedEvent{
				Kind:           interaction.Kind(),
				RequiredCounts: interaction.RequiredCounts(),
			}
			return append(events, promptEvent), nil
		}
		resolveEvents, err := g.PlayField.ResolveEvent(ctx)
		events = append(events, resolveEvents...)
		if err != nil {
			return events, err
		}

		defeatDoorEvents, err := g.DefeatDoor(eventCard)
		events = append(events, defeatDoorEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (g *Game) submitPromptChoice(cmd SubmitPromptChoiceCmd) ([]Event, error) {
	if g.PendingInteraction == nil {
		return nil, fmt.Errorf("no pending player interaction")
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	var targetPlayer *Player
	if cmd.TargetPlayerID != "" {
		targetPlayer, err = g.PlayerByID(cmd.TargetPlayerID)
		if err != nil {
			return nil, err
		}
	}
	cards, err := g.PlayerCardsByIDs(cmd.CardIDs)
	if err != nil {
		return nil, err
	}

	var events []Event
	var resolveEventFunc func(Context) ([]Event, error)
	switch i := g.PendingInteraction.(type) {
	case *TeamChoicePlayerInteraction:
		if targetPlayer == nil {
			return nil, fmt.Errorf("expected target player")
		}
		i.CollectedChoices[p] = targetPlayer
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	case *PlayerDonatesHandInteraction:
		if targetPlayer == nil {
			return nil, fmt.Errorf("expected target player")
		}
		i.CollectedChoices[p] = targetPlayer
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	case *PlayerDiscardCardsInteraction:
		cards := uniqueCards(cards)
		expectedCount := min(i.RequiredCounts()[p.Id], len(p.Hand))
		if len(cards) != expectedCount {
			return nil, fmt.Errorf("expected %d cards to discard, got %d", expectedCount, len(cards))
		}
		for _, c := range cards {
			if !p.HasCardInHand(c) {
				return nil, fmt.Errorf("p does not have card in hand")
			}
		}
		i.CollectedChoices[p] = cards
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	case *TeamChoiceResourceInteraction:
		if cmd.Resource == nil {
			return nil, fmt.Errorf("expected resource")
		}
		i.CollectedChoices[p] = *cmd.Resource
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	default:
		return nil, fmt.Errorf("unexpected interaction type: %v", i)
	}

	ctx := &CardEventContext{
		engine: g,
		Card:   g.PendingInteraction.Card(),
		Input:  g.PendingInteraction,
	}

	return g.finalizeEventInteraction(ctx, resolveEventFunc)
}

func (g *Game) finalizeEventInteraction(ctx *CardEventContext, onComplete func(Context) ([]Event, error)) ([]Event, error) {
	eventCard := g.PendingInteraction.Card()
	g.PendingInteraction = nil

	actionEvents, err := onComplete(ctx)
	if err != nil {
		return actionEvents, err
	}

	for _, p := range g.Players {
		drawEvents, err := g.RefillPlayerHand(p)
		actionEvents = append(actionEvents, drawEvents...)
		if err != nil {
			return actionEvents, err
		}
	}

	defeatEvents, err := g.DefeatDoor(eventCard)
	if err != nil {
		return append(actionEvents, defeatEvents...), err
	}

	return append(actionEvents, defeatEvents...), nil
}

func (g *Game) discardCard(cmd DiscardCardsCmd) ([]Event, error) {
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	cards, err := g.PlayerCardsByIDs(cmd.CardIDs)
	if err != nil {
		return nil, err
	}

	if _, ok := g.CurseExpectingDiscards[p]; ok {
		g.CurseExpectingDiscards[p] -= len(cards)
		if g.CurseExpectingDiscards[p] <= 0 {
			delete(g.CurseExpectingDiscards, p)
		}
	}

	events, err := p.DiscardCards(cards)
	if err != nil {
		return events, err
	}
	refillEvents, err := g.RefillPlayerHand(p)
	events = append(events, refillEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

func (g *Game) registerCards() {
	if g.DungeonCardsMap == nil {
		g.DungeonCardsMap = make(map[CardID]DungeonCard, 64)
	}
	if g.PlayerCardsMap == nil {
		g.PlayerCardsMap = make(map[CardID]PlayerCard, 256)
	}
	if g.Status != Waiting {
		return
	}

	for _, c := range append(g.Dungeon.Doors, g.Dungeon.Boss) {
		g.DungeonCardsMap[c.ID()] = c
	}

	for _, p := range g.Players {
		for _, c := range p.Deck.Cards {
			g.PlayerCardsMap[c.ID()] = c
		}
	}
}

func uniqueCards[T IdentifiableCard](inputSlice []T) []T {
	uniqueSlice := make([]T, 0, len(inputSlice))
	seen := make(map[CardID]bool, len(inputSlice))
	for _, element := range inputSlice {
		if !seen[element.ID()] {
			uniqueSlice = append(uniqueSlice, element)
			seen[element.ID()] = true
		}
	}

	return uniqueSlice
}
