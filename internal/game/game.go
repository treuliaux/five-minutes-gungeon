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
	Players             []*Player
	HandSize            int
	Dungeon             *Dungeon
	PlayField           *Playfield
	IsFightingBoss      bool
	LastPlayedCardTimer time.Duration
	InGameTimer         time.Duration
	IsTimeFrozen        bool
	Status              Status
	UseExtension        bool
	PendingInteraction  PendingInteraction
}

type GameEngine interface {
	DefeatDoor(target DungeonCard) ([]Event, error)
	OpenDoor() ([]Event, error)
	SendDungeonCardBottomDungeon(card DungeonCard) ([]Event, error)
	StopTime(player *Player) ([]Event, error)

	ListPlayers() []*Player
	ActiveDoors(f *DoorsFilter) []DungeonCard
	ListArtifacts() []*ArtifactCard

	PlayArbitraryCard(player *Player, card PlayerCard) ([]Event, error)
	RemoveDungeonCard(card DungeonCard) ([]Event, error)
	RemovePlayerCard(card PlayerCard) ([]Event, error)
	DiscardTopCardFromDungeon() ([]Event, error)
}

func NewGame() *Game {
	return &Game{
		PlayField:          NewPlayfield(),
		Status:             Waiting,
		UseExtension:       true,
		PendingInteraction: nil,
	}
}

func (g *Game) Tick(delta time.Duration) ([]Event, error) {
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
		events, err = g.addPlayer(cmd.Name, cmd.Class)
	case StartCmd:
		events, err = g.start()
	case PlayCardCmd:
		events, err = g.playCard(cmd.Player, cmd.Card)
	case DiscardCardCmd:
		events, err = g.discardCard(cmd.Player, cmd.Card)
	case UseHeroAbilityCmd:
		events, err = g.useHeroAbility(cmd.Player, cmd.DiscardCards, cmd.Ability)
	case SubmitEventChoiceCmd:
		events, err = g.submitEventChoice(cmd.Player, cmd.TargetPlayer, cmd.Cards, cmd.Resource)
	case UseArtifactCmd:
		events, err = g.useArtifact(cmd.Player, cmd.Artifact, cmd.ActionIndex, cmd.Target)
	}
	if cmd.Reply() != nil {
		cmd.Reply() <- err
	}

	return events, err
}

func (g *Game) addPlayer(name string, class HeroClass) ([]Event, error) {
	if g.Status != Waiting {
		return nil, fmt.Errorf("game is not in waiting state")
	}
	if slices.ContainsFunc(g.Players, func(player *Player) bool {
		return player.Hero.Class == class
	}) {
		return nil, fmt.Errorf("player with class %d already exists", class)
	}
	player, err := NewPlayer(name, class, g.UseExtension)
	if err != nil {
		return nil, err
	}
	g.Players = append(g.Players, player)

	return []Event{PlayerAddedEvent{Player: player}}, nil
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
	g.Status = Playing
	g.InGameTimer = 0
	g.LastPlayedCardTimer = 0

	var events []Event
	events = append(events, GameStartedEvent{})

	for _, player := range g.Players {
		drawEvents, err := g.refillPlayerHand(player)
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

func (g *Game) playCard(p *Player, card PlayerCard) ([]Event, error) {
	if g.PendingInteraction != nil {
		return nil, fmt.Errorf("cannot play cards: waiting for event resolution (%s)", g.PendingInteraction)
	}
	if !p.HasCardInHand(card) {
		return nil, fmt.Errorf("player does not have card in hand")
	}

	p.RemoveCardFromHand(card)
	fieldEvents, err := g.PlayField.AddPlayerCard(p, card)
	if err != nil {
		p.Hand = append(p.Hand, card)
		return nil, err
	}

	var events []Event
	g.LastPlayedCardTimer = g.InGameTimer
	events = append(events, fieldEvents...)
	wasTimeFrozen := g.IsTimeFrozen
	if g.IsTimeFrozen {
		g.IsTimeFrozen = false
		events = append(events, TimeUnfrozenEvent{ByPlayer: p})
	}

	var actionEvents []Event
	if ac, ok := card.(*ActionCard); ok {
		ctx := CardActionContext{
			Engine: g,
			Player: p,
			Card:   card,
		}
		actionEvents, err = ac.Action.Execute(ctx)
		if err != nil {
			p.Hand = append(p.Hand, card)
			g.PlayField.Field = slices.DeleteFunc(g.PlayField.Field, func(c PlayerCard) bool { return c == card })
			g.IsTimeFrozen = wasTimeFrozen
			return nil, err
		}
	}
	events = append(events, actionEvents...)

	cardDrawnEvents, err := g.refillPlayerHand(p)
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) DefeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event

	defeatDoorEvents, err := g.PlayField.DefeatDoor(target)
	events = append(events, defeatDoorEvents...)
	if err != nil {
		return events, err
	}
	if t, ok := target.(*CurseCard); ok {
		if t.Cure == nil {
			return events, nil
		}
		ctx := CardCurseContext{
			Engine: g,
			Card:   t,
		}
		cureEvents, err := t.Cure.Execute(ctx)
		events = append(events, cureEvents...)
		if err != nil {
			return events, err
		}

		return events, nil
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
	g.IsTimeFrozen = true

	return []Event{TimeFrozenEvent{ByPlayer: player}}, nil
}

func (g *Game) ListPlayers() []*Player {
	return g.Players
}

func (g *Game) ListArtifacts() []*ArtifactCard {
	return g.PlayField.Artifacts
}

type DoorsFilter struct {
	DoorKinds      []DoorKind
	ChallengeKinds []ChallengeKind
	IncludeBossMat bool
}

func NewDoorsFilter() *DoorsFilter {
	return &DoorsFilter{}
}
func (d *DoorsFilter) AddEverything() *DoorsFilter {
	d.IncludeBossMat = true
	d.ChallengeKinds = []ChallengeKind{ChallengeEvent, ChallengeCurse, ChallengeMiniBoss}
	d.DoorKinds = []DoorKind{DoorObstacle, DoorPerson, DoorMonster}

	return d
}
func (d *DoorsFilter) AddAllDoors() *DoorsFilter {
	d.DoorKinds = []DoorKind{DoorObstacle, DoorPerson, DoorMonster}

	return d
}
func (d *DoorsFilter) AddDoors(doorKinds ...DoorKind) *DoorsFilter {
	for _, dk := range doorKinds {
		if !slices.Contains(d.DoorKinds, dk) {
			d.DoorKinds = append(d.DoorKinds, dk)
		}
	}

	return d
}
func (d *DoorsFilter) AddCurses() *DoorsFilter {
	if !slices.Contains(d.ChallengeKinds, ChallengeCurse) {
		d.ChallengeKinds = append(d.ChallengeKinds, ChallengeCurse)
	}

	return d
}
func (d *DoorsFilter) AddEvents() *DoorsFilter {
	if !slices.Contains(d.ChallengeKinds, ChallengeEvent) {
		d.ChallengeKinds = append(d.ChallengeKinds, ChallengeEvent)
	}

	return d
}
func (d *DoorsFilter) AddMiniBoss() *DoorsFilter {
	if !slices.Contains(d.ChallengeKinds, ChallengeMiniBoss) {
		d.ChallengeKinds = append(d.ChallengeKinds, ChallengeMiniBoss)
	}

	return d
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
	events, err := g.PlayField.RemoveDungeonCard(card)
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

	return append(events, DungeonCardSentToBottomEvent{Card: card}), nil
}

func (g *Game) DiscardTopCardFromDungeon() ([]Event, error) {
	if len(g.Dungeon.Doors) == 0 {
		return nil, nil
	}

	return []Event{DungeonCardDiscardedEvent{Card: g.Dungeon.OpenDoor()}}, nil
}

func (g *Game) PlayArbitraryCard(p *Player, card PlayerCard) ([]Event, error) {
	return g.PlayField.AddPlayerCard(p, card)
}

func (g *Game) RemovePlayerCard(card PlayerCard) ([]Event, error) {
	return g.PlayField.RemovePlayerCard(card)
}

func (g *Game) RemoveDungeonCard(card DungeonCard) ([]Event, error) {
	return g.PlayField.RemoveDungeonCard(card)
}

func (g *Game) discardCard(p *Player, card PlayerCard) ([]Event, error) {
	return p.DiscardCard(card)
}

func (g *Game) OpenDoor() ([]Event, error) {
	var events []Event
	var err error
	loop := true
	for loop {
		loop = len(g.PlayField.OpenedDoors) == 0
		var addDungeonCardEvents []Event
		card := g.Dungeon.OpenDoor()
		addDungeonCardEvents, err = g.PlayField.AddDungeonCard(card, g.InGameTimer)
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

func (g *Game) useHeroAbility(player *Player, discardCards []PlayerCard, ability Ability) ([]Event, error) {
	if g.PendingInteraction != nil {
		return nil, fmt.Errorf("cannot play cards: waiting for event resolution (%s)", g.PendingInteraction)
	}
	discardCards = uniqueCards(discardCards)
	if len(discardCards) != 3 {
		return nil, fmt.Errorf("player must discard 3 different cards")
	}
	for _, card := range discardCards {
		if !player.HasCardInHand(card) {
			return nil, fmt.Errorf("player does not have card in hand")
		}
	}

	var events []Event
	for _, card := range discardCards {
		discardEvents, err := player.DiscardCard(card)
		events = append(events, discardEvents...)
		if err != nil {
			return nil, err
		}
	}

	ctx := AbilityContext{
		Engine: g,
		Player: player,
	}
	abilityEvents, err := ability.Execute(ctx)
	if err != nil {
		for _, c := range discardCards {
			if !slices.Contains(player.Hand, c) {
				player.Hand = append(player.Hand, c)
			}
		}
		return nil, err
	}

	events = append(events, HeroAbilityUsedEvent{ByPlayer: player})
	events = append(events, abilityEvents...)

	cardDrawnEvents, err := g.refillPlayerHand(player)
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) useArtifact(player *Player, artifact *ArtifactCard, index ArtifactActionIndex, target DungeonCard) ([]Event, error) {
	if !g.UseExtension {
		return nil, fmt.Errorf("extension must be enabled to use artifacts")
	}
	if g.PendingInteraction != nil {
		return nil, fmt.Errorf("cannot play cards: waiting for event resolution (%s)", g.PendingInteraction)
	}
	if !g.PlayField.ArtifactCanBePlayed(artifact) {
		return nil, fmt.Errorf("artifact cannot be played")
	}

	ctx := ArtifactActionContext{
		Engine:       g,
		Player:       player,
		Artifact:     artifact,
		ChosenAction: index,
		Target:       target,
	}
	artifactEvents, err := artifact.Action.Execute(ctx)
	if err != nil {
		return nil, err
	}
	artifact.Used = true

	events := []Event{ArtifactUsedEvent{ByPlayer: player}}

	return append(events, artifactEvents...), nil
}

func (g *Game) refillPlayerHand(player *Player) ([]Event, error) {
	if g.HandSize == 0 {
		g.determineHandSize()
	}

	var events []Event
	var err error
	nbToDraw := g.HandSize - len(player.Hand)
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
	g.LastPlayedCardTimer = g.InGameTimer

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
	return g.InGameTimer-g.LastPlayedCardTimer <= cardPlayDebounceDuration
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

		ctx := CardEventContext{
			Engine: g,
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
				Kind:          interaction.Kind(),
				RequiredCount: interaction.RequiredCount(),
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

func (g *Game) submitEventChoice(player *Player, targetPlayer *Player, cards []PlayerCard, resource *ResourceType) ([]Event, error) {
	if g.PendingInteraction == nil {
		return nil, fmt.Errorf("no pending player interaction")
	}
	ctx := CardEventContext{
		Engine: g,
		Card:   g.PendingInteraction.EventCard(),
		Input:  g.PendingInteraction,
	}

	var events []Event
	var resolveEventFunc func(CardEventContext) ([]Event, error)
	switch i := g.PendingInteraction.(type) {
	case *TeamChoicePlayerInteraction:
		if targetPlayer == nil {
			return nil, fmt.Errorf("expected target player")
		}
		i.CollectedChoices[player] = targetPlayer
		delete(i.PendingPlayers, player)
		events = append(events, PlayerEventChoiceSubmittedEvent{Player: player})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	case *PlayerDonatesHandInteraction:
		if targetPlayer == nil {
			return nil, fmt.Errorf("expected target player")
		}
		i.CollectedChoices[player] = targetPlayer
		delete(i.PendingPlayers, player)
		events = append(events, PlayerEventChoiceSubmittedEvent{Player: player})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	case *PlayerDiscardCardsInteraction:
		cards := uniqueCards(cards)
		expectedCount := min(i.RequiredCount(), len(player.Hand))
		if len(cards) != expectedCount {
			return nil, fmt.Errorf("expected %d cards to discard, got %d", expectedCount, len(cards))
		}
		for _, c := range cards {
			if !player.HasCardInHand(c) {
				return nil, fmt.Errorf("player does not have card in hand")
			}
		}
		i.CollectedChoices[player] = cards
		delete(i.PendingPlayers, player)
		events = append(events, PlayerEventChoiceSubmittedEvent{Player: player})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	case *TeamChoiceResourceInteraction:
		if resource == nil {
			return nil, fmt.Errorf("expected resource")
		}
		i.CollectedChoices[player] = *resource
		delete(i.PendingPlayers, player)
		events = append(events, PlayerEventChoiceSubmittedEvent{Player: player})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		resolveEventFunc = i.OnComplete

	default:
		return nil, fmt.Errorf("unexpected interaction type: %v", i)
	}

	return g.finalizeEventInteraction(ctx, resolveEventFunc)
}

func (g *Game) finalizeEventInteraction(ctx CardEventContext, onComplete func(CardEventContext) ([]Event, error)) ([]Event, error) {
	eventCard := g.PendingInteraction.EventCard()
	g.PendingInteraction = nil

	actionEvents, err := onComplete(ctx)
	if err != nil {
		return actionEvents, err
	}

	for _, p := range g.Players {
		drawEvents, err := g.refillPlayerHand(p)
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

func uniqueCards[T comparable](inputSlice []T) []T {
	uniqueSlice := make([]T, 0, len(inputSlice))
	seen := make(map[T]bool, len(inputSlice))
	for _, element := range inputSlice {
		if !seen[element] {
			uniqueSlice = append(uniqueSlice, element)
			seen[element] = true
		}
	}

	return uniqueSlice
}
