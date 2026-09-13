package game

import "fmt"

type HeroAbility func()

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
	color   DeckColor
	class   HeroClass
	name    string
	ability HeroAbility
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
		color:   Blue,
		class:   Sorceress,
		name:    "Sorceress",
		ability: func() {},
	}
}

func newWizard() *Hero {
	return &Hero{
		color:   Blue,
		class:   Wizard,
		name:    "Wizard",
		ability: func() {},
	}
}

func newHuntress() *Hero {
	return &Hero{
		color:   Green,
		class:   Huntress,
		name:    "Huntress",
		ability: func() {},
	}
}

func newRanger() *Hero {
	return &Hero{
		color:   Green,
		class:   Ranger,
		name:    "Ranger",
		ability: func() {},
	}
}

func newNinja() *Hero {
	return &Hero{
		color:   Purple,
		class:   Ninja,
		name:    "Ninja",
		ability: func() {},
	}
}

func newThief() *Hero {
	return &Hero{
		color:   Purple,
		class:   Thief,
		name:    "Thief",
		ability: func() {},
	}
}

func newPaladin() *Hero {
	return &Hero{
		color:   Yellow,
		class:   Paladin,
		name:    "Paladin",
		ability: func() {},
	}
}

func newValkyrie() *Hero {
	return &Hero{
		color:   Yellow,
		class:   Valkyrie,
		name:    "Valkyrie",
		ability: func() {},
	}
}

func newBarbarian() *Hero {
	return &Hero{
		color:   Red,
		class:   Barbarian,
		name:    "Barbarian",
		ability: func() {},
	}
}

func newGladiator() *Hero {
	return &Hero{
		color:   Red,
		class:   Gladiator,
		name:    "Gladiator",
		ability: func() {},
	}
}

func newDruid() *Hero {
	return &Hero{
		color:   Black,
		class:   Druid,
		name:    "Druid",
		ability: func() {},
	}
}

func newShaman() *Hero {
	return &Hero{
		color:   Black,
		class:   Shaman,
		name:    "Shaman",
		ability: func() {},
	}
}

func (p *Hero) NewDeck() *Deck {
	switch p.color {
	case Yellow:
		return NewYellowDeck()
	case Red:
		return NewRedDeck()
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
