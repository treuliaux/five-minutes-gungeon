package game

import (
	"fmt"
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

func (p *Player) removeCardFromHand(card PlayerCard) {
	for i, c := range p.hand {
		if c == card {
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			return
		}
	}
}

func (p *Player) discardMultipleCards(cards []PlayerCard) {
	for _, card := range cards {
		p.discardCard(card)
	}
}

func (p *Player) discardCard(card PlayerCard) {
	for i, c := range p.hand {
		if c == card {
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			p.discard = append(p.discard, card)
			return
		}
	}
}

func (p *Player) drawCards(i int) ([]Event, error) {
	var cardDrawnEvents []Event
	if p.deck == nil {
		return cardDrawnEvents, fmt.Errorf("player has no deck")
	}
	for range i {
		draw := p.deck.Draw()
		if draw == nil {
			return cardDrawnEvents, fmt.Errorf("no more cards in deck")
		}
		p.hand = append(p.hand, draw)
		cardDrawnEvents = append(cardDrawnEvents, CardDrawnEvent{ByPlayer: p, Card: draw})
	}

	return cardDrawnEvents, nil
}

func (p *Player) hasCardInHand(card PlayerCard) bool {
	return slices.Contains(p.hand, card)
}

func (p *Player) heal(amount int) int {
	if amount <= 0 || amount > len(p.discard) {
		amount = len(p.discard)
	}
	toHeal := p.discard[len(p.discard)-amount:]
	p.deck.putAtop(toHeal...)
	p.discard = p.discard[:len(p.discard)-amount]

	return amount
}
