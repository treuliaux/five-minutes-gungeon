package game

import (
	"math/rand/v2"
)

//go:generate go run github.com/fairjungle/enumer -type=DeckColor -json
type DeckColor uint8

const (
	NoColor DeckColor = iota
	Blue
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
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Sacred Shield", Resources: []ResourceType{Shield}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Ardent Defender", Resources: []ResourceType{Shield, Shield}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Leap", Resources: []ResourceType{Jump}})
	}
	for range 8 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Judgment", Resources: []ResourceType{Scroll}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Holy Strike", Resources: []ResourceType{Sword}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Javelin Throw", Resources: []ResourceType{Arrow}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Holy Hand Grenade", Description: "Defeat any card", Action: HolyHandGrenadeAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Heal", Description: "Target player put their [Discard Pile] atop their [Deck]", Action: HealAction{}, TargetType: TargetPlayer},
		&ActionCard{Id: nextCardId(), Name: "Health Potion", Description: "All players draw 3 cards from their [Discard Pile]", Action: HealthPotionAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Divine Shield", Description: "[Freeze] timer - All players draw a card", Action: DivineShieldAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Divine Shield", Description: "[Freeze] timer - All players draw a card", Action: DivineShieldAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Smite", Description: "Defeat a [Monster]", Action: SmiteAction{}, TargetType: TargetCard},
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
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Cleave", Resources: []ResourceType{Sword}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Whirlwind", Resources: []ResourceType{Sword, Sword}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Block", Resources: []ResourceType{Shield}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Battle Cry", Resources: []ResourceType{Scroll}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Thrown Axe", Resources: []ResourceType{Arrow}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Rush", Resources: []ResourceType{Jump}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Leaping Strike", Resources: []ResourceType{Sword, Jump}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Flaming Sword", Resources: []ResourceType{Sword, Scroll}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Martial Combo", Resources: []ResourceType{Sword, Arrow}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Defensive Stance", Resources: []ResourceType{Sword, Shield}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Mighty Leap", Description: "Defeat an [Obstacle]", Action: MightyLeapAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Mighty Leap", Description: "Defeat an [Obstacle]", Action: MightyLeapAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Enrage", Description: "Choose 2 players - They draw 3 cards", Action: EnrageAction{}, TargetType: TargetTwoPlayers},
		&ActionCard{Id: nextCardId(), Name: "Enrage", Description: "Choose 2 players - They draw 3 cards", Action: EnrageAction{}, TargetType: TargetTwoPlayers},
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
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Multi-Shot", Resources: []ResourceType{Arrow, Arrow}})
	}
	for range 9 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Steady Shot", Resources: []ResourceType{Arrow}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Athletics", Resources: []ResourceType{Jump}})
	}
	for range 4 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Arcane Shot", Resources: []ResourceType{Scroll}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Feign Death", Resources: []ResourceType{Shield}})
	}
	for range 4 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Lunge", Resources: []ResourceType{Sword}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Snipe", Description: "Defeat a [Person]", Action: SnipeAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Healing Herbs", Description: "Choose a player - They draw 4 cards from their [Discard Pile]", Action: HealingHerbsAction{}, TargetType: TargetPlayer},
		&ActionCard{Id: nextCardId(), Name: "Healing Herbs", Description: "Choose a player - They draw 4 cards from their [Discard Pile]", Action: HealingHerbsAction{}, TargetType: TargetPlayer},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Wild Card", Description: "Counts as any [Resource]", Action: WildCardAction{}, TargetType: TargetNone},
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
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Pyroblast", Resources: []ResourceType{Scroll, Scroll}})
	}
	for range 9 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Arcane Missiles", Resources: []ResourceType{Scroll}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Wand Shot", Resources: []ResourceType{Arrow}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Staff Strike", Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Ice Barrier", Resources: []ResourceType{Shield}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Blink", Resources: []ResourceType{Jump}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Cancel", Description: "Cancel an Event", Action: CancelAction{}},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Description: "Defeat a [Monster]", Action: FireballAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Description: "Defeat a [Monster]", Action: FireballAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Description: "Defeat a [Monster]", Action: FireballAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Fireball", Description: "Defeat a [Monster]", Action: FireballAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Magic Bomb", Description: "Counts as one of each [Resource]", Action: MagicBombAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Magic Bomb", Description: "Counts as one of each [Resource]", Action: MagicBombAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Magic Bomb", Description: "Counts as one of each [Resource]", Action: MagicBombAction{}, TargetType: TargetNone},
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
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Grappling hook", Resources: []ResourceType{Jump, Jump}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Acrobatics", Resources: []ResourceType{Jump}})
	}
	for range 7 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Sneak Attack", Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Evasion", Resources: []ResourceType{Shield}})
	}
	for range 6 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Stolen Scroll", Resources: []ResourceType{Scroll}})
	}
	for range 3 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Throw Dagger", Resources: []ResourceType{Arrow}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Backstab", Description: "Defeat a [Person]", Action: BackstabAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Backstab", Description: "Defeat a [Person]", Action: BackstabAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Backstab", Description: "Defeat a [Person]", Action: BackstabAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Sprint", Description: "Defeat an [Obstacle]", Action: SprintAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Sprint", Description: "Defeat an [Obstacle]", Action: SprintAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Sprint", Description: "Defeat an [Obstacle]", Action: SprintAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Steal", Description: "Steal another's player whole [Hand]", Action: StealAction{}, TargetType: TargetPlayer},
		&ActionCard{Id: nextCardId(), Name: "Steal", Description: "Steal another's player whole [Hand]", Action: StealAction{}, TargetType: TargetPlayer},
		&ActionCard{Id: nextCardId(), Name: "Donate", Description: "Give your [Hand] to target player", Action: DonateAction{}, TargetType: TargetPlayer},
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
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Tiger's Fury", Resources: []ResourceType{InfiniteSword}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Thorns Storm", Resources: []ResourceType{InfiniteArrow}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Earth Elemental", Resources: []ResourceType{InfiniteShield}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Elements Fury", Resources: []ResourceType{InfiniteScroll}})
	}
	for range 2 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Ancient Winds", Resources: []ResourceType{InfiniteJump}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Maul", Resources: []ResourceType{Sword}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Thorn Volley", Resources: []ResourceType{Arrow}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Barkskin", Resources: []ResourceType{Shield}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Wrath", Resources: []ResourceType{Scroll}})
	}
	for range 5 {
		cards = append(cards, &ResourceCard{Id: nextCardId(), Name: "Swiftness", Resources: []ResourceType{Jump}})
	}
	cards = append(cards,
		&ActionCard{Id: nextCardId(), Name: "Tame Creature", Description: "Defeat a [Monster]", Action: TameCreatureAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "True Sight", Description: "Defeat an [Obstacle]", Action: TrueSightAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Living Vines", Description: "Defeat an [Person]", Action: LivingVinesAction{}, TargetType: TargetCard},
		&ActionCard{Id: nextCardId(), Name: "Cleanse", Description: "Cure a [Curse]", Action: CleanseAction{}, TargetType: TargetCurse},
		&ActionCard{Id: nextCardId(), Name: "Cleanse", Description: "Cure a [Curse]", Action: CleanseAction{}, TargetType: TargetCurse},
		&ActionCard{Id: nextCardId(), Name: "Ancient Healing", Description: "Every player draw 2 cards from their [Discard Pile]", Action: AncientHealingAction{}, TargetType: TargetNone},
		&ActionCard{Id: nextCardId(), Name: "Ancient Healing", Description: "Every player draw 2 cards from their [Discard Pile]", Action: AncientHealingAction{}, TargetType: TargetNone},
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
			&ActionCard{Id: nextCardId(), Name: "Time Warp", Description: "[Freeze] timer", Action: TimeWarpAction{}, TargetType: TargetNone},
			&ActionCard{Id: nextCardId(), Name: "Portal", Description: "Put target [Door] at the bottom of the [Dungeon Deck]", Action: PortalAction{}, TargetType: TargetCard},
		)
	case Green:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Critical Hit", Description: "Defeat a [Monster]", Action: CriticalHitAction{}, TargetType: TargetCard},
			&ActionCard{Id: nextCardId(), Name: "Extra Quiver", Description: "All players draw 2 cards", Action: ExtraQuiverAction{}, TargetType: TargetNone},
		)
	case Purple:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Throwing Knives", Description: "Counts as 3 [Resource] of any type", Action: ThrowingKnivesAction{}, TargetType: TargetNone},
			&ActionCard{Id: nextCardId(), Name: "Throwing Knives", Description: "Counts as 3 [Resource] of any type", Action: ThrowingKnivesAction{}, TargetType: TargetNone},
		)
	case Yellow:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Mystic Rune", Description: "Every player draw 1 card for each active [Curse]", Action: MysticRuneAction{}, TargetType: TargetNone},
			&ActionCard{Id: nextCardId(), Name: "Rally", Description: "Draw every cards from your [Discard Pile] that has a 🗡️ or a 🛡️", Action: RallyAction{}, TargetType: TargetNone},
		)
	case Red:
		d.Cards = append(d.Cards,
			&ActionCard{Id: nextCardId(), Name: "Crush", Description: "Defeat a [Mini-Boss]", Action: CrushAction{}, TargetType: TargetCard},
			&ActionCard{Id: nextCardId(), Name: "Battle Rage", Description: "Put a [Curse] at the bottom of [Dungeon Deck]", Action: BattleRageAction{}, TargetType: TargetCurse},
		)
	case Black:
	case NoColor:
	}

	return d.Shuffle()
}
