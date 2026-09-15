package game

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

func (d *Discard) PutAtop(cards ...PlayerCard) {
	d.Cards = append(d.Cards, cards...)
}

func (d *Discard) Empty() bool {
	return d.Length() == 0
}

func (d *Discard) Length() int {
	return len(d.Cards)
}
