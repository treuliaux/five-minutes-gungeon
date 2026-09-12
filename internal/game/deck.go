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

func (d *Deck) PutAtop(cards ...*PlayerCard) {
	for _, card := range cards {
		d.cards = append(d.cards, *card)
	}
}

func FreshYellowDeck() *Deck {
	var cards []PlayerCard
	card := ResourceCard{Resources: []ResourceType{Shield}}
	cards = append(cards, card)
	d := &Deck{
		Color: Yellow,
		cards: cards,
	}

	return d
}

func FreshRedDeck() *Deck {
	var cards []PlayerCard
	card := ResourceCard{Resources: []ResourceType{Sword}}
	cards = append(cards, card)
	d := &Deck{
		Color: Red,
		cards: cards,
	}

	return d
}
