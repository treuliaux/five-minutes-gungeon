package game

import (
	"slices"
)

type Player struct {
	name    string
	hero    *Hero
	deck    *Deck
	discard []PlayerCard
	hand    []PlayerCard
}

func NewPlayer(name string, class HeroClass) (*Player, error) {
	hero, err := NewHeroFromHeroClass(class)
	if err != nil {
		return nil, err
	}

	return &Player{
		name:    name,
		hero:    hero,
		deck:    hero.NewDeck(),
		discard: []PlayerCard{},
		hand:    []PlayerCard{},
	}, nil
}

func (p *Player) RemoveCardFromHand(card PlayerCard) {
	for i, c := range p.hand {
		if c == card {
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			return
		}
	}
}

func (p *Player) DiscardMultipleCards(cards []PlayerCard) {
	for _, card := range cards {
		p.DiscardCard(card)
	}
}

func (p *Player) DiscardCard(card PlayerCard) {
	for i, c := range p.hand {
		if c == card {
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			p.discard = append(p.discard, card)
			return
		}
	}
}

func (p *Player) DrawCards(i int) {
	for range i {
		draw := p.deck.Draw()
		if draw == nil {
			return
		}
		p.hand = append(p.hand, draw)
	}
}

func (p *Player) HasCardInHand(card PlayerCard) bool {
	return slices.Contains(p.hand, card)
}

func (p *Player) Heal() {
	for _, card := range p.discard {
		p.deck.PutAtop(card)
	}
	p.discard = make([]PlayerCard, 0)
}
