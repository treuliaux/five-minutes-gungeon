package game

import "fmt"

type HeroClass int

const (
	Sorceress HeroClass = iota
	Wizard
	Huntress
	Ranger
	Ninja
	Thief
	Paladin
	Valkyrie
	Barbarian
	Gladiator
	Druid
	Shaman
)

type Hero struct {
	Color DeckColor
	Class HeroClass
	Name  string
}

func NewHeroFromHeroClass(class HeroClass) (*Hero, error) {
	switch class {
	case Sorceress:
		return newSorceress(), nil
	case Wizard:
		return newWizard(), nil
	case Huntress:
		return newHuntress(), nil
	case Ranger:
		return newRanger(), nil
	case Ninja:
		return newNinja(), nil
	case Thief:
		return newThief(), nil
	case Paladin:
		return newPaladin(), nil
	case Valkyrie:
		return newValkyrie(), nil
	case Barbarian:
		return newBarbarian(), nil
	case Gladiator:
		return newGladiator(), nil
	case Druid:
		return newDruid(), nil
	case Shaman:
		return newShaman(), nil
	default:
		return nil, fmt.Errorf("unknown hero class: %d", class)
	}
}

func newSorceress() *Hero {
	return &Hero{
		Color: Blue,
		Class: Sorceress,
		Name:  "Sorceress",
	}
}

func newWizard() *Hero {
	return &Hero{
		Color: Blue,
		Class: Wizard,
		Name:  "Wizard",
	}
}

func newHuntress() *Hero {
	return &Hero{
		Color: Green,
		Class: Huntress,
		Name:  "Huntress",
	}
}

func newRanger() *Hero {
	return &Hero{
		Color: Green,
		Class: Ranger,
		Name:  "Ranger",
	}
}

func newNinja() *Hero {
	return &Hero{
		Color: Purple,
		Class: Ninja,
		Name:  "Ninja",
	}
}

func newThief() *Hero {
	return &Hero{
		Color: Purple,
		Class: Thief,
		Name:  "Thief",
	}
}

func newPaladin() *Hero {
	return &Hero{
		Color: Yellow,
		Class: Paladin,
		Name:  "Paladin",
	}
}

func newValkyrie() *Hero {
	return &Hero{
		Color: Yellow,
		Class: Valkyrie,
		Name:  "Valkyrie",
	}
}

func newBarbarian() *Hero {
	return &Hero{
		Color: Red,
		Class: Barbarian,
		Name:  "Barbarian",
	}
}

func newGladiator() *Hero {
	return &Hero{
		Color: Red,
		Class: Gladiator,
		Name:  "Gladiator",
	}
}

func newDruid() *Hero {
	return &Hero{
		Color: Black,
		Class: Druid,
		Name:  "Druid",
	}
}

func newShaman() *Hero {
	return &Hero{
		Color: Black,
		Class: Shaman,
		Name:  "Shaman",
	}
}

func (h *Hero) NewDeck(includeExtension bool) *Deck {
	switch h.Color {
	case Yellow:
		return NewYellowDeck(includeExtension).Shuffle()
	case Red:
		return NewRedDeck(includeExtension).Shuffle()
	case Blue:
		return NewBlueDeck(includeExtension).Shuffle()
	case Black:
		return NewBlackDeck().Shuffle()
	case Green:
		return NewGreenDeck(includeExtension).Shuffle()
	case Purple:
		return NewPurpleDeck(includeExtension).Shuffle()
	}
	return nil
}
