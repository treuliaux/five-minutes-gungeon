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
	Status              Status
}

func NewGame() *Game {
	return &Game{
		Status: Waiting,
	}
}

func (g *Game) Tick(delta time.Duration) error {
	if g.Status != Playing {
		return fmt.Errorf("game is not in playing state")
	}
	g.inGameTimer += delta
	if g.inGameTimer >= 5*time.Minute {
		g.Status = Defeat

		return nil
	}
	if (g.inGameTimer-g.lastPlayedCardTimer <= time.Second/2 && !g.isFightingBoss()) || !g.isPlayfieldBeaten() {
		return nil
	}
	if g.isFightingBoss() {
		g.Status = Victory

		return nil
	}
	g.clearField()
	if err := g.openDoor(); err != nil {
		return err
	}

	return nil
}

func (g *Game) isFightingBoss() bool {
	return g.currentDungeonCard == g.dungeon.boss
}

func (g *Game) Apply(cmd Command) {
	var err error
	switch cmd := cmd.(type) {
	case AddPlayerCmd:
		err = g.addPlayer(cmd)
	case StartCmd:
		err = g.start()
	case PlayCardCmd:
		err = g.playCard(cmd.Player, cmd.Card)
	case DiscardCardCmd:
		err = g.discardCard(cmd.Player, cmd.Card)
	case FreezeCmd:
		//err = g.freeze(cmd.Player)
	}
	if cmd.Reply() != nil {
		cmd.Reply() <- err
	}
}

func (g *Game) addPlayer(cmd AddPlayerCmd) error {
	if g.Status != Waiting {
		return fmt.Errorf("game is not in waiting state")
	}
	if slices.ContainsFunc(g.players, func(player *Player) bool {
		return player.hero.class == cmd.Class
	}) {
		return fmt.Errorf("player with class %d already exists", cmd.Class)
	}
	player, err := NewPlayer(cmd.Name, cmd.Class)
	if err != nil {
		return err
	}
	g.players = append(g.players, player)

	return nil
}

func (g *Game) start() error {
	if g.Status != Waiting {
		return fmt.Errorf("game has already started")
	}
	if len(g.players) < 2 || len(g.players) > 6 {
		return fmt.Errorf("game has incorrect number of players: %d", len(g.players))
	}
	g.determineHandSize()
	for _, player := range g.players {
		player.DrawCards(g.handSize)
	}

	g.generateDungeon()
	g.Status = Playing
	g.inGameTimer = 0
	g.lastPlayedCardTimer = 0

	if err := g.openDoor(); err != nil {
		return err
	}

	return nil
}

func (g *Game) playCard(p *Player, card PlayerCard) error {
	if !p.HasCardInHand(card) {
		return fmt.Errorf("player does not have card in hand")
	}
	p.RemoveCardFromHand(card)
	g.playedCards = append(g.playedCards, card)
	g.lastPlayedCardTimer = g.inGameTimer

	return nil
}

func (g *Game) discardCard(p *Player, card PlayerCard) error {
	if !p.HasCardInHand(card) {
		return fmt.Errorf("player does not have card in hand")
	}
	p.DiscardCard(card)

	return nil
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

func (g *Game) openDoor() error {
	if g.currentDungeonCard != nil {
		return fmt.Errorf("door is already open")
	}
	card := g.dungeon.OpenDoor()
	g.currentDungeonCard = card

	return nil
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

func (g *Game) clearField() {
	g.currentDungeonCard = nil
	g.playedCards = make([]PlayerCard, 0)
	g.lastPlayedCardTimer = g.inGameTimer
}
