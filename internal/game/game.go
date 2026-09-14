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

type Game struct {
	players             []*Player
	handSize            int
	dungeon             *Dungeon
	currentDungeonCard  DungeonCard
	playedCards         []PlayerCard
	lastPlayedCardTimer time.Duration
	inGameTimer         time.Duration
	isTimeFrozen        bool
	Status              Status
}

func NewGame() *Game {
	return &Game{
		Status: Waiting,
	}
}

func (g *Game) Tick(delta time.Duration) ([]Event, error) {
	var events []Event

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
	if (g.inGameTimer-g.lastPlayedCardTimer <= time.Second/2 && !g.isFightingBoss()) || !g.isPlayfieldBeaten() {
		return events, nil
	}
	events = append(events, DoorDefeatedEvent{DungeonCard: g.currentDungeonCard})
	if g.isFightingBoss() {
		g.Status = Victory

		return append(events, GameWonEvent{}), nil
	}
	events = append(events, g.clearField()...)
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
		events, err = g.addPlayer(cmd)
	case StartCmd:
		events, err = g.start()
	case PlayCardCmd:
		events, err = g.playCard(cmd.Player, cmd.Card)
	case DiscardCardCmd:
		events, err = g.discardCard(cmd.Player, cmd.Card)
	case UseHeroAbilityCmd:
		events, err = g.useHeroAbility(cmd)
	}
	if cmd.Reply() != nil {
		cmd.Reply() <- err
	}

	return events, err
}

func (g *Game) addPlayer(cmd AddPlayerCmd) ([]Event, error) {
	var events []Event
	if g.Status != Waiting {
		return events, fmt.Errorf("game is not in waiting state")
	}
	if slices.ContainsFunc(g.players, func(player *Player) bool {
		return player.hero.class == cmd.Class
	}) {
		return events, fmt.Errorf("player with class %d already exists", cmd.Class)
	}
	player, err := NewPlayer(cmd.Name, cmd.Class)
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
	for _, player := range g.players {
		player.DrawCards(g.handSize)
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
	if !p.HasCardInHand(card) {
		return events, fmt.Errorf("player does not have card in hand")
	}
	p.RemoveCardFromHand(card)
	g.playedCards = append(g.playedCards, card)
	g.lastPlayedCardTimer = g.inGameTimer
	g.isTimeFrozen = false

	return append(events, CardPlayedEvent{ByPlayer: p, Card: card}, TimeUnfrozenEvent{ByPlayer: p}), nil
}

func (g *Game) discardCard(p *Player, card PlayerCard) ([]Event, error) {
	var events []Event
	if !p.HasCardInHand(card) {
		return events, fmt.Errorf("player does not have card in hand")
	}
	p.DiscardCard(card)

	return append(events, CardDiscardedEvent{ByPlayer: p, Card: card}), nil
}

func (g *Game) openDoor() ([]Event, error) {
	var events []Event
	if g.currentDungeonCard != nil {
		return events, fmt.Errorf("door is already open")
	}
	card := g.dungeon.OpenDoor()
	g.currentDungeonCard = card

	return append(events, DoorOpenedEvent{DungeonCard: card}), nil
}

func (g *Game) useHeroAbility(cmd UseHeroAbilityCmd) ([]Event, error) {
	var events []Event
	var err error
	player := cmd.Player

	if len(cmd.DiscardCards) != 3 {
		return events, fmt.Errorf("player must discard 3 cards")
	}
	ok := true
	for _, card := range cmd.DiscardCards {
		ok = player.HasCardInHand(card)
		if !ok {
			return events, fmt.Errorf("player does not have card in hand")
		}
	}
	ctx := AbilityContext{
		game:   g,
		player: player,
	}
	switch params := cmd.Params.(type) {
	case TrickShotParams:
		events, err = applyTrickShot(ctx, params)
	case AnimalCompanionParams:
		events, err = applyAnimalCompanion(ctx, params)
	case InspireParams:
		events, err = applyInspire(ctx, params)
	case SmiteParams:
		events, err = applySmite(ctx, params)
	case StopTimeParams:
		events, err = applyStopTime(ctx, params)
	case TeleportParams:
		events, err = applyTeleport(ctx, params)
	case SlayParams:
		events, err = applySlay(ctx, params)
	case IntimidateParams:
		events, err = applyIntimidate(ctx, params)
	case VaultParams:
		events, err = applyVault(ctx, params)
	case PickpocketParams:
		events, err = applyPickpocket(ctx, params)
	case ForestSpiritsParams:
		events, err = applyForestSpirits(ctx, params)
	case SpiritAnimalParams:
		events, err = applySpiritAnimal(ctx, params)
	}
	if err != nil {
		return events, err
	}

	return append([]Event{HeroAbilityUsedEvent{ByPlayer: player}}, events...), nil
}

func (g *Game) clearField() []Event {
	var events []Event
	g.currentDungeonCard = nil
	g.playedCards = make([]PlayerCard, 0)
	g.lastPlayedCardTimer = g.inGameTimer

	return append(events, FieldClearedEvent{})
}

func (g *Game) defeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event
	if target != g.currentDungeonCard {
		return events, fmt.Errorf("target is not the current dungeon card")
	}

	g.currentDungeonCard = nil
	events = append(events, DoorDefeatedEvent{DungeonCard: target})
	if g.currentDungeonCard != nil {
		return events, nil
	}

	events = append(events, g.clearField()...)
	openDoorEvents, err := g.openDoor()
	if err != nil {
		return append(events, openDoorEvents...), err
	}

	return append(events, openDoorEvents...), nil
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
	g.dungeon = NewDungeon()
}

func (g *Game) isPlayfieldBeaten() bool {
	if g.currentDungeonCard == nil {
		return true
	}
	if g.currentDungeonCard.IsBeaten(g.playedCards) {
		return true
	}
	return false
}

func (g *Game) isFightingBoss() bool {
	return g.currentDungeonCard == g.dungeon.boss
}
