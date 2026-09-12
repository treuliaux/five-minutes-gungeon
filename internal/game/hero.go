package game

type HeroAbility func()

type Hero struct {
	Color   DeckColor
	Name    string
	Ability HeroAbility
}

func Paladin() *Hero {
	return &Hero{
		Color:   Yellow,
		Name:    "Paladin",
		Ability: func() {},
	}
}

func Barbarian() *Hero {
	return &Hero{
		Color:   Red,
		Name:    "Barbarian",
		Ability: func() {},
	}
}

func (p *Hero) GetNewDeck() *Deck {
	switch p.Color {
	case Yellow:
		return FreshYellowDeck()
	case Red:
		return FreshRedDeck()
	case Blue:
		return nil
	case Black:
		return nil
	case Green:
		return nil
	case Purple:
		return nil
	}
	return nil
}
