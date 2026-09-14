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

const cardPlayDebounceDuration = time.Second / 2

type Game struct {
	players             []*Player
	handSize            int
	dungeon             *Dungeon
	playField           Playfield
	isFightingBoss      bool
	lastPlayedCardTimer time.Duration
	inGameTimer         time.Duration
	isTimeFrozen        bool
	Status              Status
}

type GameEngine interface {
	hasActiveDoor(target DungeonCard) bool
	defeatDoor(target DungeonCard) ([]Event, error)
	stopTime(player *Player) []Event
	drawCards(player *Player, count int) ([]Event, error)
	listOtherPlayers(omit *Player) []*Player
	healPlayer(player *Player, amount int) ([]Event, error)
}

func NewGame() *Game {
	return &Game{
		Status: Waiting,
	}
}

func (g *Game) Tick(delta time.Duration) ([]Event, error) {
	var events []Event

	if g.Status == Victory || g.Status == Defeat {
		return events, nil
	}
	if g.Status != Playing {
		return events, fmt.Errorf("game is not in playing state")
	}
	if g.isTimeFrozen {
		delta = 0
	}
	g.inGameTimer += delta
	if g.inGameTimer >= 5*time.Minute {
		g.Status = Defeat

		return append(events, GameLostEvent{}), nil
	}
	if (g.actionDebounce() && !g.isFightingBoss) || !g.playField.isPlayfieldBeaten() {
		return events, nil
	}

	defeatAllDoorsEvents, err := g.playField.defeatAllDoors()
	events = append(events, defeatAllDoorsEvents...)
	if err != nil {
		return events, err
	}
	if g.isFightingBoss {
		g.Status = Victory

		return append(events, GameWonEvent{}), nil
	}
	openDoorEvents, err := g.openDoor()
	if err != nil {
		return append(events, openDoorEvents...), err
	}

	return append(events, openDoorEvents...), nil
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
	var events []Event
	if g.Status != Waiting {
		return events, fmt.Errorf("game is not in waiting state")
	}
	if slices.ContainsFunc(g.players, func(player *Player) bool {
		return player.hero.class == class
	}) {
		return events, fmt.Errorf("player with class %d already exists", class)
	}
	player, err := NewPlayer(name, class)
	if err != nil {
		return events, err
	}
	g.players = append(g.players, player)

	return append(events, PlayerAddedEvent{Player: player}), nil
}

func (g *Game) start() ([]Event, error) {
	var events []Event
	var err error

	if g.Status != Waiting {
		return events, fmt.Errorf("game has already started")
	}
	if len(g.players) < 2 || len(g.players) > 6 {
		return events, fmt.Errorf("game has incorrect number of players: %d", len(g.players))
	}
	g.determineHandSize()
	var cardDrawnEvents []Event
	for _, player := range g.players {
		cardDrawnEvents, err = player.drawCards(g.handSize)
		cardDrawnEvents = append(cardDrawnEvents, cardDrawnEvents...)
		if err != nil {
			return events, err
		}
	}

	g.generateDungeon()
	g.Status = Playing
	g.inGameTimer = 0
	g.lastPlayedCardTimer = 0

	events, err = g.openDoor()
	if err != nil {
		return events, err
	}

	return append([]Event{GameStartedEvent{}}, events...), nil
}

func (g *Game) playCard(p *Player, card PlayerCard) ([]Event, error) {
	var events []Event
	var err error
	if !p.hasCardInHand(card) {
		return events, fmt.Errorf("player does not have card in hand")
	}
	p.removeCardFromHand(card)
	events, err = g.playField.addPlayerCard(p, card)
	if err != nil {
		return nil, err
	}
	g.lastPlayedCardTimer = g.inGameTimer
	if g.isTimeFrozen {
		g.isTimeFrozen = false
		events = append(events, TimeUnfrozenEvent{ByPlayer: p})
	}

	nbToDraw := g.handSize - len(p.hand)
	var cardDrawnEvents []Event
	if nbToDraw > 0 {
		cardDrawnEvents, err = g.drawCards(p, nbToDraw)
	}
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) drawCards(player *Player, count int) ([]Event, error) {
	return player.drawCards(count)
}

func (g *Game) discardCard(p *Player, card PlayerCard) ([]Event, error) {
	var events []Event
	if !p.hasCardInHand(card) {
		return events, fmt.Errorf("player does not have card in hand")
	}
	p.discardCard(card)

	return append(events, CardDiscardedEvent{ByPlayer: p, Card: card}), nil
}

func (g *Game) openDoor() ([]Event, error) {
	card := g.dungeon.OpenDoor()
	events, err := g.playField.addDungeonCard(card)
	if err != nil {
		return []Event{}, err
	}
	if _, ok := card.(*BossMat); ok {
		g.isFightingBoss = true
	}

	return events, nil
}

func (g *Game) useHeroAbility(player *Player, discardCards []PlayerCard, ability Ability) ([]Event, error) {
	var events []Event
	var err error

	if len(discardCards) != 3 {
		return events, fmt.Errorf("player must discard 3 cards")
	}
	ok := true
	for _, card := range discardCards {
		ok = player.hasCardInHand(card)
		if !ok {
			return events, fmt.Errorf("player does not have card in hand")
		}
	}
	ctx := AbilityContext{
		engine: g,
		player: player,
	}
	events, err = ability.execute(ctx)
	if err != nil {
		return events, err
	}

	var discardedCardEvents []Event
	for _, card := range discardCards {
		discardCardEvents, err := g.discardCard(player, card)
		discardedCardEvents = append(discardedCardEvents, discardCardEvents...)
		if err != nil {
			return append(events, discardCardEvents...), err
		}
	}
	events = append(discardedCardEvents, events...)

	return append([]Event{HeroAbilityUsedEvent{ByPlayer: player}}, events...), nil
}

func (g *Game) clearField() []Event {
	var events []Event
	fieldEvents, err := g.playField.clearField()
	if err != nil {
		return nil
	}
	g.lastPlayedCardTimer = g.inGameTimer

	return append(events, fieldEvents...)
}

func (g *Game) defeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event

	defeatDoorEvents, err := g.playField.defeatDoor(target)
	events = append(events, defeatDoorEvents...)
	if err != nil {
		return append(events, defeatDoorEvents...), err
	}

	if g.isFightingBoss {
		g.Status = Victory

		return append(events, GameWonEvent{}), nil
	}

	if g.playField.hasDoorsOpened() {
		return events, nil
	}

	events = append(events, g.clearField()...)

	openDoorEvents, err := g.openDoor()
	if err != nil {
		return append(events, openDoorEvents...), err
	}

	return append(events, openDoorEvents...), nil
}

func (g *Game) hasActiveDoor(target DungeonCard) bool {
	return g.playField.hasActiveDoor(target)
}

func (g *Game) stopTime(player *Player) []Event {
	g.isTimeFrozen = true

	return []Event{TimeFrozenEvent{ByPlayer: player}}
}

func (g *Game) determineHandSize() {
	switch len(g.players) {
	case 2:
		g.handSize = 5
	case 3:
		g.handSize = 4
	case 4, 5, 6:
		g.handSize = 3
	}
}

func (g *Game) generateDungeon() {
	if g.dungeon == nil {
		g.dungeon = NewDungeon()
	}
}

func (g *Game) actionDebounce() bool {
	return g.inGameTimer-g.lastPlayedCardTimer <= cardPlayDebounceDuration
}

func (g *Game) listOtherPlayers(omit *Player) []*Player {
	var players []*Player
	for _, player := range g.players {
		if player != omit {
			players = append(players, player)
		}
	}

	return players
}

func (g *Game) healPlayer(p *Player, amount int) ([]Event, error) {
	amount = p.heal(amount)
	return []Event{PlayerHealedEvent{
		Player: p,
		amount: amount,
	}}, nil
}
