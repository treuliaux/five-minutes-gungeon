package game

import (
	"fmt"
	"slices"
)

type Player struct {
	Name    string
	Hero    *Hero
	Deck    *Deck
	Discard *Discard
	Hand    []PlayerCard
}

func NewPlayer(name string, class HeroClass) (*Player, error) {
	hero, err := NewHeroFromHeroClass(class)
	if err != nil {
		return nil, err
	}

	return &Player{
		Name:    name,
		Hero:    hero,
		Deck:    hero.NewDeck(),
		Discard: NewDiscard(),
		Hand:    make([]PlayerCard, 0),
	}, nil
}

func (p *Player) RemoveCardFromHand(card PlayerCard) {
	for i, c := range p.Hand {
		if c == card {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return
		}
	}
}

func (p *Player) DiscardCards(cards []PlayerCard) ([]Event, error) {
	var events []Event
	for _, card := range cards {
		discardEvents, err := p.DiscardCard(card)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (p *Player) DiscardCard(card PlayerCard) ([]Event, error) {
	if !slices.Contains(p.Hand, card) {
		return nil, fmt.Errorf("player does not have card in hand")
	}

	p.Hand = slices.DeleteFunc(p.Hand, func(c PlayerCard) bool { return c == card })
	p.Discard.PutAtop(card)

	return []Event{CardDiscardedEvent{ByPlayer: p, Card: card}}, nil
}

func (p *Player) DrawCards(count int) ([]Event, error) {
	if p.Deck == nil {
		return nil, fmt.Errorf("player has no deck")
	}

	var events []Event
	for range count {
		drawn := p.Deck.Draw()
		if drawn == nil {
			return events, fmt.Errorf("no more cards in deck")
		}
		p.Hand = append(p.Hand, drawn)
		events = append(events, CardDrawnEvent{ByPlayer: p, Card: drawn})
	}

	return events, nil
}

func (p *Player) DrawCardsFromDiscard(count int) ([]Event, error) {
	if p.Discard.Empty() {
		return nil, fmt.Errorf("player discard is empty")
	}

	var events []Event
	for range count {
		drawn := p.Discard.Draw()
		if drawn == nil {
			return events, fmt.Errorf("no more cards in discard")
		}
		p.Hand = append(p.Hand, drawn)
		events = append(events, CardDrawnFromDiscardEvent{ByPlayer: p, Card: drawn})
	}

	return events, nil
}

func (p *Player) HasCardInHand(card PlayerCard) bool {
	return slices.Contains(p.Hand, card)
}

func (p *Player) Heal(amount int) ([]Event, error) {
	if amount <= 0 || amount > p.Discard.Length() {
		amount = p.Discard.Length()
	}
	var healed []PlayerCard
	for range amount {
		toHeal := p.Discard.Draw()
		if toHeal == nil {
			break
		}
		p.Deck.PutAtop(toHeal)
		healed = append(healed, toHeal)
	}

	return []Event{PlayerHealedEvent{
		Player: p,
		Cards:  healed,
	}}, nil
}
