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
}

type GameEngine interface {
	HasActiveDoor(target DungeonCard) bool
	DefeatDoor(target DungeonCard) ([]Event, error)
	StopTime(player *Player) ([]Event, error)
	DrawCardsFromDeck(player *Player, count int) ([]Event, error)
	DrawCardsFromDiscard(player *Player, count int) ([]Event, error)
	DrawResourceCardsFromDiscard(player *Player, resources []ResourceType) ([]Event, error)
	ListOtherPlayers(omit *Player) []*Player
	ListPlayers() []*Player
	HealPlayer(player *Player, amount int) ([]Event, error)
	PlayArbitraryCard(player *Player, card PlayerCard) ([]Event, error)
	RemovePlayerCardFromPlayfield(card PlayerCard) ([]Event, error)
	SendDungeonCardBottomDungeon(card DungeonCard) ([]Event, error)
	GetActiveCurses() []DungeonCard
	RemoveCurseFromPlayfield(card *CurseCard) ([]Event, error)
	GetActiveDoorsOfType(doorKinds []DoorKind, challengeKinds []ChallengeKind, includeBossMat bool) []DungeonCard
}

func NewGame() *Game {
	return &Game{
		PlayField:    NewPlayfield(),
		Status:       Waiting,
		UseExtension: true,
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
	openDoorEvents, err := g.openDoor()
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
	g.determineHandSize()
	for _, player := range g.Players {
		if _, err := player.DrawCardsFromDeck(g.HandSize); err != nil {
			return nil, err
		}
	}

	g.generateDungeon(g.UseExtension)
	g.Status = Playing
	g.InGameTimer = 0
	g.LastPlayedCardTimer = 0

	var events []Event
	events = append(events, GameStartedEvent{})

	openDoorEvents, err := g.openDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) playCard(p *Player, card PlayerCard) ([]Event, error) {
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

	cardDrawnEvents, err2 := g.refillPlayerHand(p)
	events = append(events, cardDrawnEvents...)
	if err2 != nil {
		return events, err2
	}

	return events, nil
}

func (g *Game) DrawCardsFromDeck(player *Player, count int) ([]Event, error) {
	return player.DrawCardsFromDeck(count)
}

func (g *Game) DrawCardsFromDiscard(player *Player, count int) ([]Event, error) {
	return player.DrawCardsFromDiscard(count)
}

func (g *Game) DrawResourceCardsFromDiscard(player *Player, resourceTypes []ResourceType) ([]Event, error) {
	return player.DrawResourceCardsFromDiscard(resourceTypes)
}

func (g *Game) DefeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event

	defeatDoorEvents, err := g.PlayField.DefeatDoor(target)
	events = append(events, defeatDoorEvents...)
	if err != nil {
		return events, err
	}

	if g.IsFightingBoss {
		g.Status = Victory
		events = append(events, GameWonEvent{})

		return events, nil
	}

	if g.PlayField.HasDoorsOpened() {
		return events, nil
	}

	events = append(events, g.clearField()...)

	openDoorEvents, err2 := g.openDoor()
	events = append(events, openDoorEvents...)
	if err2 != nil {
		return events, err2
	}

	return events, nil
}

func (g *Game) HasActiveDoor(target DungeonCard) bool {
	return g.PlayField.HasActiveDoor(target)
}

func (g *Game) StopTime(player *Player) ([]Event, error) {
	if g.IsTimeFrozen {
		return nil, fmt.Errorf("time is already frozen")
	}
	g.IsTimeFrozen = true

	return []Event{TimeFrozenEvent{ByPlayer: player}}, nil
}

func (g *Game) ListOtherPlayers(omit *Player) []*Player {
	var players []*Player
	for _, player := range g.Players {
		if player == omit {
			continue
		}
		players = append(players, player)
	}

	return players
}

func (g *Game) ListPlayers() []*Player {
	return g.Players
}

func (g *Game) HealPlayer(p *Player, amount int) ([]Event, error) {
	return p.Heal(amount)
}

func (g *Game) GetActiveDoorsOfType(doorKinds []DoorKind, challengeKinds []ChallengeKind, includeBossMat bool) []DungeonCard {
	var picked []DungeonCard
	for _, door := range append(g.PlayField.OpenedDoors, g.GetActiveCurses()...) {
		switch d := door.(type) {
		case *BossMat:
			if includeBossMat {
				picked = append(picked, d)
			}
		case *CurseCard:
			if slices.Contains(challengeKinds, ChallengeCurse) {
				picked = append(picked, d)
			}
		case *EventCard:
			if slices.Contains(challengeKinds, ChallengeEvent) {
				picked = append(picked, d)
			}
		case *MiniBossCard:
			if slices.Contains(challengeKinds, ChallengeMiniBoss) {
				picked = append(picked, d)
			}
		case *DoorCard:
			if slices.Contains(doorKinds, d.Type) {
				picked = append(picked, d)
			}
		}
	}

	return picked
}

func (g *Game) GetActiveCurses() []DungeonCard {
	b := make([]DungeonCard, len(g.PlayField.ActiveCurses))
	for i := range g.PlayField.ActiveCurses {
		b[i] = g.PlayField.ActiveCurses[i]
	}

	return b
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
		openDoorEvents, err2 := g.openDoor()
		events = append(events, openDoorEvents...)
		if err2 != nil {
			return events, err2
		}
	}

	return append(events, DungeonCardSentToBottomEvent{Card: card}), nil
}

func (g *Game) PlayArbitraryCard(p *Player, card PlayerCard) ([]Event, error) {
	return g.PlayField.AddPlayerCard(p, card)
}

func (g *Game) RemovePlayerCardFromPlayfield(card PlayerCard) ([]Event, error) {
	return g.PlayField.RemovePlayerCard(card)
}

func (g *Game) RemoveCurseFromPlayfield(card *CurseCard) ([]Event, error) {
	return g.PlayField.RemoveDungeonCard(card)
}

func (g *Game) discardCard(p *Player, card PlayerCard) ([]Event, error) {
	return p.DiscardCard(card)
}

func (g *Game) openDoor() ([]Event, error) {
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

	cardDrawnEvents, err2 := g.refillPlayerHand(player)
	events = append(events, cardDrawnEvents...)
	if err2 != nil {
		return events, err2
	}

	return events, nil
}

func (g *Game) refillPlayerHand(player *Player) ([]Event, error) {
	var events []Event
	var err error
	nbToDraw := g.HandSize - len(player.Hand)
	if nbToDraw > 0 {
		events, err = g.DrawCardsFromDeck(player, nbToDraw)
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

func (g *Game) generateDungeon(extensionEnabled bool) {
	if g.Dungeon == nil {
		g.Dungeon = NewDungeon(1, len(g.Players), extensionEnabled)
	}
}

func (g *Game) actionDebounce() bool {
	return g.InGameTimer-g.LastPlayedCardTimer <= cardPlayDebounceDuration
}

func (g *Game) resolveActiveEvents() ([]Event, error) {
	var events []Event
	for _, card := range g.PlayField.OpenedDoors {
		eventCard, ok := card.(*EventCard)
		if !ok {
			continue
		}
		resolveEvents, err := g.PlayField.ResolveEvent(eventCard, g)
		events = append(events, resolveEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
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
