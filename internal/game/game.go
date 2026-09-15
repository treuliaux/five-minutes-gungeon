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
	Players             []*Player
	HandSize            int
	Dungeon             *Dungeon
	PlayField           Playfield
	IsFightingBoss      bool
	LastPlayedCardTimer time.Duration
	InGameTimer         time.Duration
	IsTimeFrozen        bool
	Status              Status
}

type GameEngine interface {
	HasActiveDoor(target DungeonCard) bool
	DefeatDoor(target DungeonCard) ([]Event, error)
	StopTime(player *Player) []Event
	DrawCards(player *Player, count int) ([]Event, error)
	ListOtherPlayers(omit *Player) []*Player
	HealPlayer(player *Player, amount int) ([]Event, error)
}

func NewGame() *Game {
	return &Game{
		Status: Waiting,
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
	if g.InGameTimer >= 5*time.Minute {
		g.Status = Defeat

		return []Event{GameLostEvent{}}, nil
	}
	if (g.actionDebounce() && !g.IsFightingBoss) || !g.PlayField.IsPlayfieldBeaten() {
		return nil, nil
	}

	var events []Event
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
	player, err := NewPlayer(name, class)
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
		if _, err := player.DrawCards(g.HandSize); err != nil {
			return nil, err
		}
	}

	g.generateDungeon()
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
	events, err := g.PlayField.AddPlayerCard(p, card)
	if err != nil {
		return nil, err
	}
	g.LastPlayedCardTimer = g.InGameTimer
	if g.IsTimeFrozen {
		g.IsTimeFrozen = false
		events = append(events, TimeUnfrozenEvent{ByPlayer: p})
	}

	nbToDraw := g.HandSize - len(p.Hand)
	if nbToDraw > 0 {
		cardDrawnEvents, err := g.DrawCards(p, nbToDraw)
		events = append(events, cardDrawnEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (g *Game) DrawCards(player *Player, count int) ([]Event, error) {
	return player.DrawCards(count)
}

func (g *Game) discardCard(p *Player, card PlayerCard) ([]Event, error) {
	return p.DiscardCard(card)
}

func (g *Game) openDoor() ([]Event, error) {
	card := g.Dungeon.OpenDoor()
	events, err := g.PlayField.AddDungeonCard(card)
	if err != nil {
		return nil, err
	}
	if _, ok := card.(*BossMat); ok {
		g.IsFightingBoss = true
	}

	return events, nil
}

func (g *Game) useHeroAbility(player *Player, discardCards []PlayerCard, ability Ability) ([]Event, error) {
	if len(discardCards) != 3 {
		return nil, fmt.Errorf("player must discard 3 cards")
	}
	for _, card := range discardCards {
		if !player.HasCardInHand(card) {
			return nil, fmt.Errorf("player does not have card in hand")
		}
	}
	ctx := AbilityContext{
		Engine: g,
		Player: player,
	}
	abilityEvents, err := ability.Execute(ctx)
	if err != nil {
		return nil, err
	}

	var events []Event
	events = append(events, HeroAbilityUsedEvent{ByPlayer: player})

	for _, card := range discardCards {
		discardEvents, err := player.DiscardCard(card)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	events = append(events, abilityEvents...)

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

	openDoorEvents, err := g.openDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) HasActiveDoor(target DungeonCard) bool {
	return g.PlayField.HasActiveDoor(target)
}

func (g *Game) StopTime(player *Player) []Event {
	g.IsTimeFrozen = true

	return []Event{TimeFrozenEvent{ByPlayer: player}}
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
		g.Dungeon = NewDungeon()
	}
}

func (g *Game) actionDebounce() bool {
	return g.InGameTimer-g.LastPlayedCardTimer <= cardPlayDebounceDuration
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

func (g *Game) HealPlayer(p *Player, amount int) ([]Event, error) {
	return p.Heal(amount)
}
