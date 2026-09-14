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
	color DeckColor
	class HeroClass
	name  string
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
		color: Blue,
		class: Sorceress,
		name:  "Sorceress",
	}
}

func newWizard() *Hero {
	return &Hero{
		color: Blue,
		class: Wizard,
		name:  "Wizard",
	}
}

func newHuntress() *Hero {
	return &Hero{
		color: Green,
		class: Huntress,
		name:  "Huntress",
	}
}

func newRanger() *Hero {
	return &Hero{
		color: Green,
		class: Ranger,
		name:  "Ranger",
	}
}

func newNinja() *Hero {
	return &Hero{
		color: Purple,
		class: Ninja,
		name:  "Ninja",
	}
}

func newThief() *Hero {
	return &Hero{
		color: Purple,
		class: Thief,
		name:  "Thief",
	}
}

func newPaladin() *Hero {
	return &Hero{
		color: Yellow,
		class: Paladin,
		name:  "Paladin",
	}
}

func newValkyrie() *Hero {
	return &Hero{
		color: Yellow,
		class: Valkyrie,
		name:  "Valkyrie",
	}
}

func newBarbarian() *Hero {
	return &Hero{
		color: Red,
		class: Barbarian,
		name:  "Barbarian",
	}
}

func newGladiator() *Hero {
	return &Hero{
		color: Red,
		class: Gladiator,
		name:  "Gladiator",
	}
}

func newDruid() *Hero {
	return &Hero{
		color: Black,
		class: Druid,
		name:  "Druid",
	}
}

func newShaman() *Hero {
	return &Hero{
		color: Black,
		class: Shaman,
		name:  "Shaman",
	}
}

func (p *Hero) NewDeck() *Deck {
	switch p.color {
	case Yellow:
		return NewYellowDeck()
	case Red:
		return NewRedDeck()
	case Blue:
		return NewBlueDeck()
	case Black:
		return NewBlackDeck()
	case Green:
		return NewGreenDeck()
	case Purple:
		return NewPurpleDeck()
	}
	return nil
}
