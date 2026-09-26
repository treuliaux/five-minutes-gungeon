package game

import (
	"math/rand/v2"
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

func (d *Deck) Shuffle() *Deck {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})

	return d
}

func (d *Deck) PutAtop(cards ...PlayerCard) {
	d.Cards = append(d.Cards, cards...)
}

func (d *Deck) PutBelow(cards ...PlayerCard) {
	d.Cards = append(cards, d.Cards...)
}

func (d *Deck) Empty() bool {
	return d.Length() == 0
}

func (d *Deck) Length() int {
	return len(d.Cards)
}

func NewYellowDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 9 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield, Shield}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump}})
	}
	for range 8 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Holy Hand Grenade", Action: HolyHandGrenadeAction{}},
		&ActionCard{Id: nextCardId(), Name: "Heal", Action: HealAction{}},
		&ActionCard{Id: nextCardId(), Name: "Health Potion", Action: HealthPotionAction{}},
		&ActionCard{Id: nextCardId(), Name: "Divine Shield", Action: DivineShieldAction{}},
		&ActionCard{Id: nextCardId(), Name: "Divine Shield", Action: DivineShieldAction{}},
		&ActionCard{Id: nextCardId(), Name: "Smite", Action: SmiteAction{}},
	)
	d := &Deck{
		Color: Yellow,
		Cards: cards,
	}

	return d
}

func NewRedDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword, Sword}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword, Jump}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword, Scroll}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword, Arrow}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword, Shield}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Mighty Leap", Action: MightyLeapAction{}},
		&ActionCard{Id: nextCardId(), Name: "Mighty Leap", Action: MightyLeapAction{}},
		&ActionCard{Id: nextCardId(), Name: "Enrage", Action: EnrageAction{}},
		&ActionCard{Id: nextCardId(), Name: "Enrage", Action: EnrageAction{}},
	)
	d := &Deck{
		Color: Red,
		Cards: cards,
	}

	return d
}

func NewGreenDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow, Arrow}})
	}
	for range 9 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump}})
	}
	for range 4 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}})
	}
	for range 4 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Snipe", Action: SnipeAction{}},
		&ActionCard{Id: nextCardId(), Name: "Healing Herbs", Action: HealingHerbsAction{}},
		&ActionCard{Id: nextCardId(), Name: "Healing Herbs", Action: HealingHerbsAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Action: WildCardAction{}},
	)
	d := &Deck{
		Color: Green,
		Cards: cards,
	}

	return d
}

func NewBlueDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll, Scroll}})
	}
	for range 9 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Cancel", Action: CancelAction{}},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Id: nextCardId(), Name: "Magic Bomb", Action: MagicBombAction{}},
		&ActionCard{Id: nextCardId(), Name: "Magic Bomb", Action: MagicBombAction{}},
		&ActionCard{Id: nextCardId(), Name: "Magic Bomb", Action: MagicBombAction{}},
	)
	d := &Deck{
		Color: Blue,
		Cards: cards,
	}

	return d
}

func NewPurpleDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump, Jump}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Backstab", Action: BackstabAction{}},
		&ActionCard{Id: nextCardId(), Name: "Backstab", Action: BackstabAction{}},
		&ActionCard{Id: nextCardId(), Name: "Backstab", Action: BackstabAction{}},
		&ActionCard{Id: nextCardId(), Name: "Sprint", Action: SprintAction{}},
		&ActionCard{Id: nextCardId(), Name: "Sprint", Action: SprintAction{}},
		&ActionCard{Id: nextCardId(), Name: "Sprint", Action: SprintAction{}},
		&ActionCard{Id: nextCardId(), Name: "Steal", Action: StealAction{}},
		&ActionCard{Id: nextCardId(), Name: "Steal", Action: StealAction{}},
		&ActionCard{Id: nextCardId(), Name: "Donate", Action: DonateAction{}},
	)
	d := &Deck{
		Color: Purple,
		Cards: cards,
	}

	return d
}

func NewBlackDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{InfiniteSword}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{InfiniteArrow}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{InfiniteShield}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{InfiniteScroll}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{InfiniteJump}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Arrow}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Jump}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Tame Creature", Action: TameCreatureAction{}},
		&ActionCard{Id: nextCardId(), Name: "True Sight", Action: TrueSightAction{}},
		&ActionCard{Id: nextCardId(), Name: "Living Vines", Action: LivingVinesAction{}},
		&ActionCard{Id: nextCardId(), Name: "Cleanse", Action: CleanseAction{}},
		&ActionCard{Id: nextCardId(), Name: "Cleanse", Action: CleanseAction{}},
		&ActionCard{Id: nextCardId(), Name: "Ancient Healing", Action: AncientHealingAction{}},
		&ActionCard{Id: nextCardId(), Name: "Ancient Healing", Action: AncientHealingAction{}},
	)
	d := &Deck{
		Color: Black,
		Cards: cards,
	}

	return d
}

func (d *Deck) IncludeExtension() *Deck {
	switch d.Color {
	case Blue:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Time Warp", Action: TimeWarpAction{}},
			&ActionCard{Id: nextCardId(), Name: "Portal", Action: PortalAction{}},
		)
	case Green:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Critical Hit", Action: CriticalHitAction{}},
			&ActionCard{Id: nextCardId(), Name: "Extra Quiver", Action: ExtraQuiverAction{}},
		)
	case Purple:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Throwing Knives", Action: ThrowingKnivesAction{}},
			&ActionCard{Id: nextCardId(), Name: "Throwing Knives", Action: ThrowingKnivesAction{}},
		)
	case Yellow:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Mystic Rune", Action: MysticRuneAction{}},
			&ActionCard{Id: nextCardId(), Name: "Rally", Action: RallyAction{}},
		)
	case Red:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Crush", Action: CrushAction{}},
			&ActionCard{Id: nextCardId(), Name: "Battle Rage", Action: BattleRageAction{}},
		)
	case Black:
	}

	return d
}
