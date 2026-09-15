package game

import (
	"math/rand"
)

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
	Cards []PlayerCard
}

func (d *Deck) Draw() PlayerCard {
	if len(d.Cards) == 0 {
		return nil
	}
	drawnCard := d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]

	return drawnCard
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) PutAtop(cards ...PlayerCard) {
	d.Cards = append(d.Cards, cards...)
}

func (d *Deck) Empty() bool {
	return d.Length() == 0
}

func (d *Deck) Length() int {
	return len(d.Cards)
}

func NewYellowDeck() *Deck {
	cards := make([]PlayerCard, 0, 12)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield, Scroll}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Shield}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	d := &Deck{
		Color: Yellow,
		Cards: cards,
	}

	return d
}

func NewRedDeck() *Deck {
	cards := make([]PlayerCard, 0, 12)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Scroll}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Shield}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	d := &Deck{
		Color: Red,
		Cards: cards,
	}

	return d
}

func NewGreenDeck() *Deck {
	cards := make([]PlayerCard, 0, 12)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow, Jump}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Arrow}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	d := &Deck{
		Color: Green,
		Cards: cards,
	}

	return d
}

func NewBlueDeck() *Deck {
	cards := make([]PlayerCard, 0, 12)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll, Jump}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield, Scroll}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	d := &Deck{
		Color: Blue,
		Cards: cards,
	}

	return d
}

func NewPurpleDeck() *Deck {
	cards := make([]PlayerCard, 0, 12)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump, Sword}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump, Scroll}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	d := &Deck{
		Color: Purple,
		Cards: cards,
	}

	return d
}

func NewBlackDeck() *Deck {
	cards := make([]PlayerCard, 0, 12)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Shield}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow, Jump}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll, Jump}})
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield, Arrow}})
	}
	d := &Deck{
		Color: Black,
		Cards: cards,
	}

	return d
}
