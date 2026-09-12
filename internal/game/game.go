package game

import (
	"fmt"
	"time"
)

type GameStatus int

const (
	Waiting GameStatus = iota
	Victory
	Defeat
	Playing
)

type Game struct {
	Players            []*Player
	Dungeon            *Dungeon
	CurrentDungeonCard *DungeonCard
	PlayedCards        []PlayerCard
	Timer              time.Duration
	Status             GameStatus
}

func NewGame() *Game {
	return &Game{
		Status: Waiting,
	}
}

func (g *Game) AddPlayer(player *Player) {
	g.Players = append(g.Players, player)
}

func (g *Game) GenerateDungeon() {
	g.Dungeon = NewDungeon()
}

func (g *Game) Start() error {
	if g.Status != Waiting {
		return fmt.Errorf("game has already started")
	}
	if len(g.Players) < 2 || len(g.Players) > 6 {
		return fmt.Errorf("game has incorrect number of players: %d", len(g.Players))
	}
	handSize := 0
	switch len(g.Players) {
	case 2:
		handSize = 5
	case 3:
		handSize = 4
	case 4, 5, 6:
		handSize = 3
	}
	for _, player := range g.Players {
		player.DrawCards(handSize)
	}

	g.Status = Playing
	g.Timer = 0

	return nil
}

func (g *Game) OpenDoor() {
	g.PlayedCards = make([]PlayerCard, 0)
	card := g.Dungeon.OpenDoor()
	g.CurrentDungeonCard = card
}

func (g *Game) PlayCard(p *Player, card PlayerCard) {
	if !p.HasCard(card) {
		return
	}
	p.RemoveCard(card)
	g.PlayedCards = append(g.PlayedCards, card)
}
