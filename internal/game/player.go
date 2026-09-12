package game

import (
	"slices"

	"github.com/gookit/goutil/dump"
)

type Player struct {
	Name    string
	Hero    *Hero
	Deck    *Deck
	Discard []PlayerCard
	Hand    []PlayerCard
}

func NewPlayer(name string, hero *Hero) *Player {
	return &Player{
		Name:    name,
		Hero:    hero,
		Deck:    hero.GetNewDeck(),
		Discard: []PlayerCard{},
		Hand:    []PlayerCard{},
	}
}

func (p *Player) RemoveCard(card PlayerCard) {
	for i, c := range p.Hand {
		if c == card {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return
		}
	}
}

func (p *Player) DrawCards(i int) {
	for range i {
		draw := p.Deck.Draw()
		if draw == nil {
			return
		}
		p.Hand = append(p.Hand, draw)
	}
}

func (p *Player) HasCard(card PlayerCard) bool {
	for _, c := range p.Hand {
		dump.V(card)
		dump.V(c)
	}
	return slices.Contains(p.Hand, card)
}

func (p *Player) Heal() {
	for _, card := range p.Discard {
		p.Deck.PutAtop(&card)
	}
	p.Discard = make([]PlayerCard, 0)
}
