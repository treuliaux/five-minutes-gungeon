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
	Color   DeckColor
	Class   HeroClass
	Name    string
	Ability Ability
}

func NewHeroFromHeroClass(class HeroClass) (*Hero, error) {
	switch class {
	case Sorceress:
		return NewSorceress(), nil
	case Wizard:
		return NewWizard(), nil
	case Huntress:
		return NewHuntress(), nil
	case Ranger:
		return NewRanger(), nil
	case Ninja:
		return NewNinja(), nil
	case Thief:
		return NewThief(), nil
	case Paladin:
		return NewPaladin(), nil
	case Valkyrie:
		return NewValkyrie(), nil
	case Barbarian:
		return NewBarbarian(), nil
	case Gladiator:
		return NewGladiator(), nil
	case Druid:
		return NewDruid(), nil
	case Shaman:
		return NewShaman(), nil
	default:
		return nil, fmt.Errorf("unknown hero class: %d", class)
	}
}

func NewSorceress() *Hero {
	return &Hero{
		Color:   Blue,
		Class:   Sorceress,
		Name:    "Sorceress",
		Ability: TeleportAbility{},
	}
}

func NewWizard() *Hero {
	return &Hero{
		Color:   Blue,
		Class:   Wizard,
		Name:    "Wizard",
		Ability: StopTimeAbility{},
	}
}

func NewHuntress() *Hero {
	return &Hero{
		Color:   Green,
		Class:   Huntress,
		Name:    "Huntress",
		Ability: AnimalCompanionAbility{},
	}
}

func NewRanger() *Hero {
	return &Hero{
		Color:   Green,
		Class:   Ranger,
		Name:    "Ranger",
		Ability: TrickShotAbility{},
	}
}

func NewNinja() *Hero {
	return &Hero{
		Color:   Purple,
		Class:   Ninja,
		Name:    "Ninja",
		Ability: VaultAbility{},
	}
}

func NewThief() *Hero {
	return &Hero{
		Color:   Purple,
		Class:   Thief,
		Name:    "Thief",
		Ability: PickpocketAbility{},
	}
}

func NewPaladin() *Hero {
	return &Hero{
		Color:   Yellow,
		Class:   Paladin,
		Name:    "Paladin",
		Ability: SmiteAbility{},
	}
}

func NewValkyrie() *Hero {
	return &Hero{
		Color:   Yellow,
		Class:   Valkyrie,
		Name:    "Valkyrie",
		Ability: InspireAbility{},
	}
}

func NewBarbarian() *Hero {
	return &Hero{
		Color:   Red,
		Class:   Barbarian,
		Name:    "Barbarian",
		Ability: SlayAbility{},
	}
}

func NewGladiator() *Hero {
	return &Hero{
		Color:   Red,
		Class:   Gladiator,
		Name:    "Gladiator",
		Ability: IntimidateAbility{},
	}
}

func NewDruid() *Hero {
	return &Hero{
		Color:   Black,
		Class:   Druid,
		Name:    "Druid",
		Ability: ForestSpiritsAbility{},
	}
}

func NewShaman() *Hero {
	return &Hero{
		Color:   Black,
		Class:   Shaman,
		Name:    "Shaman",
		Ability: SpiritAnimalAbility{},
	}
}

func (h *Hero) NewExtensionDeck() *Deck {
	switch h.Color {
	case Yellow:
		return NewYellowDeck().IncludeExtension().Shuffle()
	case Red:
		return NewRedDeck().IncludeExtension().Shuffle()
	case Blue:
		return NewBlueDeck().IncludeExtension().Shuffle()
	case Black:
		return NewBlackDeck().IncludeExtension().Shuffle()
	case Green:
		return NewGreenDeck().IncludeExtension().Shuffle()
	case Purple:
		return NewPurpleDeck().IncludeExtension().Shuffle()
	}
	return nil
}

func (h *Hero) NewBaseDeck() *Deck {
	switch h.Color {
	case Yellow:
		return NewYellowDeck().Shuffle()
	case Red:
		return NewRedDeck().Shuffle()
	case Blue:
		return NewBlueDeck().Shuffle()
	case Green:
		return NewGreenDeck().Shuffle()
	case Purple:
		return NewPurpleDeck().Shuffle()
	case Black:
		return NewBlackDeck().Shuffle()
	}

	return nil
}

func (h *Hero) Flip() *Hero {
	switch h.Class {
	case Sorceress:
		return NewWizard()
	case Wizard:
		return NewSorceress()
	case Huntress:
		return NewRanger()
	case Ranger:
		return NewHuntress()
	case Ninja:
		return NewThief()
	case Thief:
		return NewNinja()
	case Paladin:
		return NewValkyrie()
	case Valkyrie:
		return NewPaladin()
	case Barbarian:
		return NewGladiator()
	case Gladiator:
		return NewBarbarian()
	case Druid:
		return NewShaman()
	case Shaman:
		return NewDruid()
	}

	panic("unreachable")
}
