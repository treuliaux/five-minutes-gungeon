package game

import "math/rand"

type DeckColor int

const (
	Blue DeckColor = iota
	Green
	Purple
	Yellow
	Red
	Black
)

type Deck struct {
	Color DeckColor
	cards []PlayerCard
}

func (d *Deck) Draw() PlayerCard {
	if len(d.cards) == 0 {
		return nil
	}
	drawnCard, deck := d.cards[0], d.cards[1:]
	d.cards = deck

	return drawnCard
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

func (d *Deck) putAtop(cards ...PlayerCard) {
	for _, card := range cards {
		d.cards = append(d.cards, card)
	}
}

func NewYellowDeck() *Deck {
	var cards []PlayerCard
	for range 3 {
		cards = append(cards, &ResourceCard{resources: []ResourceType{Shield}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Shield, Scroll}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Sword, Shield}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Jump}})
	}
	d := &Deck{
		Color: Yellow,
		cards: cards,
	}

	return d
}

func NewRedDeck() *Deck {
	var cards []PlayerCard
	for range 3 {
		cards = append(cards, &ResourceCard{resources: []ResourceType{Sword}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Sword, Scroll}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Sword, Shield}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Arrow}})
	}
	d := &Deck{
		Color: Red,
		cards: cards,
	}

	return d
}

func NewGreenDeck() *Deck {
	var cards []PlayerCard
	for range 3 {
		cards = append(cards, &ResourceCard{resources: []ResourceType{Arrow}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Arrow, Jump}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Sword, Arrow}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Scroll}})
	}
	d := &Deck{
		Color: Green,
		cards: cards,
	}

	return d
}

func NewBlueDeck() *Deck {
	var cards []PlayerCard
	for range 3 {
		cards = append(cards, &ResourceCard{resources: []ResourceType{Scroll}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Scroll, Jump}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Shield, Scroll}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Arrow}})
	}
	d := &Deck{
		Color: Blue,
		cards: cards,
	}

	return d
}

func NewPurpleDeck() *Deck {
	var cards []PlayerCard
	for range 3 {
		cards = append(cards, &ResourceCard{resources: []ResourceType{Jump}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Jump, Sword}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Jump, Scroll}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Shield}})
	}
	d := &Deck{
		Color: Purple,
		cards: cards,
	}

	return d
}

func NewBlackDeck() *Deck {
	var cards []PlayerCard
	for range 3 {
		cards = append(cards, &ResourceCard{resources: []ResourceType{Sword, Shield}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Arrow, Jump}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Scroll, Jump}})
		cards = append(cards, &ResourceCard{resources: []ResourceType{Shield, Arrow}})
	}
	d := &Deck{
		Color: Black,
		cards: cards,
	}

	return d
}
