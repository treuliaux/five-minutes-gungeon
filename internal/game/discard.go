package game

import "slices"

type Discard struct {
	Cards []PlayerCard
}

func NewDiscard() *Discard {
	return &Discard{Cards: make([]PlayerCard, 0)}
}

func (d *Discard) Draw() PlayerCard {
	if len(d.Cards) == 0 {
		return nil
	}
	drawnCard := d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]

	return drawnCard
}

func (d *Discard) DrawResourceCard(resources []ResourceType) PlayerCard {
	for i := len(d.Cards) - 1; i >= 0; i-- {
		c := d.Cards[i]
		switch card := c.(type) {
		case *ResourceCard:
			for _, resource := range resources {
				if slices.Contains(card.Resources, resource) {
					drawnCard := d.Cards[i]
					d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
					return drawnCard
				}
			}
		}
	}

	return nil
}

func (d *Discard) PutAtop(cards ...PlayerCard) {
	d.Cards = append(d.Cards, cards...)
}

func (d *Discard) Empty() bool {
	return d.Length() == 0
}

func (d *Discard) Length() int {
	return len(d.Cards)
}
