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

func NewPlayer(name string, class HeroClass, includeExtension bool) (*Player, error) {
	if !includeExtension && (class == Druid || class == Shaman) {
		return nil, fmt.Errorf("druid/shaman is not available without the extension enabled")
	}
	hero, err := NewHeroFromHeroClass(class)
	if err != nil {
		return nil, err
	}

	deck := hero.NewBaseDeck()
	if includeExtension {
		deck.IncludeExtension()
	}

	return &Player{
		Name:    name,
		Hero:    hero,
		Deck:    deck,
		Discard: NewDiscard(),
		Hand:    make([]PlayerCard, 0),
	}, nil
}

func (p *Player) RemoveCardFromHand(card PlayerCard) {
	idx := slices.Index(p.Hand, card)
	if idx != -1 {
		p.Hand = slices.Delete(p.Hand, idx, idx+1)
	}
}

func (p *Player) DiscardCards(cards []PlayerCard) ([]Event, error) {
	var events []Event
	for _, card := range slices.Clone(cards) {
		discardEvents, err := p.DiscardCard(card)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (p *Player) DiscardCard(card PlayerCard) ([]Event, error) {
	idx := slices.Index(p.Hand, card)
	if idx == -1 {
		return nil, fmt.Errorf("player does not have card in hand")
	}

	p.Hand = slices.Delete(p.Hand, idx, idx+1)
	if p.Discard != nil {
		p.Discard.PutAtop(card)
	}

	return []Event{CardDiscardedEvent{ByPlayer: p, Card: card}}, nil
}

func (p *Player) DrawCardsFromDeck(count int) ([]Event, error) {
	if count <= 0 {
		return nil, nil
	}
	if p.Deck == nil {
		return nil, fmt.Errorf("player has no deck")
	}
	if p.Deck.Empty() {
		return nil, nil
	}

	var events []Event
	count = min(count, p.Deck.Length())
	for range count {
		drawn := p.Deck.Draw()
		p.Hand = append(p.Hand, drawn)
		events = append(events, CardDrawnFromDeckEvent{ByPlayer: p, Card: drawn})
	}

	return events, nil
}

func (p *Player) DrawCardsFromDiscard(count int) ([]Event, error) {
	if count <= 0 {
		return nil, nil
	}
	if p.Discard == nil {
		return nil, fmt.Errorf("player has no deck")
	}
	if p.Discard.Empty() {
		return nil, nil
	}

	var events []Event
	count = min(count, p.Discard.Length())
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

func (p *Player) DrawResourceCardsFromDiscard(resourceTypes []ResourceType) ([]Event, error) {
	if p.Discard == nil {
		return nil, fmt.Errorf("player has no discard")
	}

	var events []Event
	for {
		drawn := p.Discard.DrawResourceCard(resourceTypes)
		if drawn == nil {
			break
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

func (p *Player) ArtifactHeal() ([]Event, error) {
	var healed []PlayerCard
	for range p.Discard.Length() {
		toHeal := p.Discard.Draw()
		if toHeal == nil {
			break
		}
		p.Deck.PutBelow(toHeal)
		healed = append(healed, toHeal)
	}

	return []Event{PlayerHealedEvent{
		Player: p,
		Cards:  healed,
	}}, nil
}

func (p *Player) FlipHeroMat() []Event {
	oldHero := p.Hero
	p.Hero = p.Hero.Flip()

	return []Event{HeroMatFlippedEvent{
		Player: p,
		From:   oldHero,
		To:     p.Hero,
	}}
}

func (p *Player) VoidHand(effect GameCurseEffect, game *Game) ([]Event, error) {
	voidedCards := slices.Clone(p.Hand)
	p.Hand = make([]PlayerCard, 0)
	events := []Event{PlayerHandVoidedEvent{Player: p, VoidedCards: voidedCards}}

	refillEvents, err := game.RefillPlayerHand(p)
	events = append(events, refillEvents...)
	if err != nil {
		return events, err
	}

	return events, &CurseRuleViolatedError{Curse: effect, ByPlayer: p}
}
