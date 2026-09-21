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

func (d *Deck) Empty() bool {
	return d.Length() == 0
}

func (d *Deck) Length() int {
	return len(d.Cards)
}

func NewYellowDeck(includeExtension bool) *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 9 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield, Shield}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	for range 8 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	cards = append(cards,
		&ActionCard{Name: "Holy Hand Grenade", Action: HolyHandGrenadeAction{}},
		&ActionCard{Name: "Heal", Action: HealAction{}},
		&ActionCard{Name: "Health Potion", Action: HealthPotionAction{}},
		&ActionCard{Name: "Divine Shield", Action: DivineShieldAction{}},
		&ActionCard{Name: "Divine Shield", Action: DivineShieldAction{}},
		&ActionCard{Name: "Smite", Action: SmiteAction{}},
	)
	if includeExtension {
		cards = append(cards,
			&ActionCard{Name: "Mystic Rune", Action: MysticRuneAction{}},
			&ActionCard{Name: "Rally", Action: RallyAction{}},
		)
	}
	d := &Deck{
		Color: Yellow,
		Cards: cards,
	}

	return d
}

func NewRedDeck(includeExtension bool) *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Sword}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Jump}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Scroll}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Arrow}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword, Shield}})
	}
	cards = append(cards,
		&ActionCard{Name: "Mighty Leap", Action: MightyLeapAction{}},
		&ActionCard{Name: "Mighty Leap", Action: MightyLeapAction{}},
		&ActionCard{Name: "Enrage", Action: EnrageAction{}},
		&ActionCard{Name: "Enrage", Action: EnrageAction{}},
	)
	if includeExtension {
		cards = append(cards,
			&ActionCard{Name: "Crush", Action: CrushAction{}},
			&ActionCard{Name: "Battle Rage", Action: BattleRageAction{}},
		)
	}
	d := &Deck{
		Color: Red,
		Cards: cards,
	}

	return d
}

func NewGreenDeck(includeExtension bool) *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow, Arrow}})
	}
	for range 9 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	for range 4 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	for range 4 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
	}
	cards = append(cards,
		&ActionCard{Name: "Snipe", Action: SnipeAction{}},
		&ActionCard{Name: "Healing Herbs", Action: HealingHerbsAction{}},
		&ActionCard{Name: "Healing Herbs", Action: HealingHerbsAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
		&ActionCard{Name: "Wild Card", Action: WildCardAction{}},
	)
	if includeExtension {
		cards = append(cards,
			&ActionCard{Name: "Critical Hit", Action: CriticalHitAction{}},
			&ActionCard{Name: "Extra Quiver", Action: ExtraQuiverAction{}},
		)
	}
	d := &Deck{
		Color: Green,
		Cards: cards,
	}

	return d
}

func NewBlueDeck(includeExtension bool) *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll, Scroll}})
	}
	for range 9 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	cards = append(cards,
		&ActionCard{Name: "Cancel", Action: CancelAction{}},
		&ActionCard{Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Name: "Fireball", Action: FireballAction{}},
		&ActionCard{Name: "Magic Bomb", Action: MagicBombAction{}},
		&ActionCard{Name: "Magic Bomb", Action: MagicBombAction{}},
		&ActionCard{Name: "Magic Bomb", Action: MagicBombAction{}},
	)
	if includeExtension {
		cards = append(cards,
			&ActionCard{Name: "Time Warp", Action: TimeWarpAction{}},
			&ActionCard{Name: "Portal", Action: PortalAction{}},
		)
	}
	d := &Deck{
		Color: Blue,
		Cards: cards,
	}

	return d
}

func NewPurpleDeck(includeExtension bool) *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump, Jump}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	cards = append(cards,
		&ActionCard{Name: "Backstab", Action: BackstabAction{}},
		&ActionCard{Name: "Backstab", Action: BackstabAction{}},
		&ActionCard{Name: "Backstab", Action: BackstabAction{}},
		&ActionCard{Name: "Sprint", Action: SprintAction{}},
		&ActionCard{Name: "Sprint", Action: SprintAction{}},
		&ActionCard{Name: "Sprint", Action: SprintAction{}},
		&ActionCard{Name: "Steal", Action: StealAction{}},
		&ActionCard{Name: "Steal", Action: StealAction{}},
		&ActionCard{Name: "Donate", Action: DonateAction{}},
	)
	if includeExtension {
		cards = append(cards,
			&ActionCard{Name: "Throwing Knives", Action: ThrowingKnivesAction{}},
			&ActionCard{Name: "Throwing Knives", Action: ThrowingKnivesAction{}},
		)
	}
	d := &Deck{
		Color: Purple,
		Cards: cards,
	}

	return d
}

func NewBlackDeck() *Deck {
	cards := make([]PlayerCard, 0, 42)
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{InfiniteSword}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{InfiniteArrow}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{InfiniteShield}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{InfiniteScroll}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{InfiniteJump}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Arrow}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Shield}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Scroll}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Resources: []ResourceType{Jump}})
	}
	cards = append(cards,
		&ActionCard{Name: "Tame Creature", Action: TameCreatureAction{}},
		&ActionCard{Name: "True Sight", Action: TrueSightAction{}},
		&ActionCard{Name: "Living Vines", Action: LivingVinesAction{}},
		&ActionCard{Name: "Cleanse", Action: CleanseAction{}},
		&ActionCard{Name: "Cleanse", Action: CleanseAction{}},
		&ActionCard{Name: "Ancient Healing", Action: AncientHealingAction{}},
		&ActionCard{Name: "Ancient Healing", Action: AncientHealingAction{}},
	)
	d := &Deck{
		Color: Black,
		Cards: cards,
	}

	return d
}
