package game

type Event interface {
	isEvent()
}

type GameStartedEvent struct {
}

func (GameStartedEvent) isEvent() {}

type GameLostEvent struct {
}

func (GameLostEvent) isEvent() {}

type GameWonEvent struct {
}

func (GameWonEvent) isEvent() {}

type DoorDefeatedEvent struct {
	DungeonCard DungeonCard
}

func (DoorDefeatedEvent) isEvent() {}

type FieldClearedEvent struct {
}

func (FieldClearedEvent) isEvent() {}

type PlayerAddedEvent struct {
	Player *Player
}

func (PlayerAddedEvent) isEvent() {}

type CardPlayedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (CardPlayedEvent) isEvent() {}

type CardRemovedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (CardRemovedEvent) isEvent() {}

type DoorCardRemovedEvent struct {
	ByPlayer *Player
	Card     DungeonCard
}

func (DoorCardRemovedEvent) isEvent() {}

type DungeonCardSentToBottomEvent struct {
	Card DungeonCard
}

func (DungeonCardSentToBottomEvent) isEvent() {}

type DungeonCardDiscardedEvent struct {
	Card DungeonCard
}

func (DungeonCardDiscardedEvent) isEvent() {}

type CurseActivatedEvent struct {
	Card DungeonCard
}

func (CurseActivatedEvent) isEvent() {}

type CurseRemovedEvent struct {
	Card DungeonCard
}

func (CurseRemovedEvent) isEvent() {}

type ActionCardPlayedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (ActionCardPlayedEvent) isEvent() {}

type EventCounteredEvent struct {
	ByPlayer  *Player
	EventCard DungeonCard
	WithCard  any
}

func (EventCounteredEvent) isEvent() {}

type CardDrawnFromDeckEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (CardDrawnFromDeckEvent) isEvent() {}

type CardDrawnFromDiscardEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (CardDrawnFromDiscardEvent) isEvent() {}

type CardDiscardedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (CardDiscardedEvent) isEvent() {}

type TimeFrozenEvent struct {
	ByPlayer *Player
}

func (TimeFrozenEvent) isEvent() {}

type TimeUnfrozenEvent struct {
	ByPlayer *Player
}

func (TimeUnfrozenEvent) isEvent() {}

type DoorOpenedEvent struct {
	DungeonCard DungeonCard
}

func (DoorOpenedEvent) isEvent() {}

type HeroAbilityUsedEvent struct {
	ByPlayer *Player
}

func (HeroAbilityUsedEvent) isEvent() {}

type ArtifactUsedEvent struct {
	ByPlayer *Player
}

func (ArtifactUsedEvent) isEvent() {}

type ArtifactReEnabledEvent struct {
	Artifact *ArtifactCard
}

func (ArtifactReEnabledEvent) isEvent() {}

type PlayerHealedEvent struct {
	Player *Player
	Cards  []PlayerCard
}

func (PlayerHealedEvent) isEvent() {}

type ExtensionToggledEvent struct {
	Enabled bool
}

func (ExtensionToggledEvent) isEvent() {}

type HandDonatedEvent struct {
	From  *Player
	To    *Player
	Cards []PlayerCard
}

func (HandDonatedEvent) isEvent() {}

type HandStolenEvent struct {
	From  *Player
	To    *Player
	Cards []PlayerCard
}

func (HandStolenEvent) isEvent() {}

type EventPromptOpenedEvent struct {
	Kind           InteractionKind
	RequiredCounts map[*Player]int
}

func (EventPromptOpenedEvent) isEvent() {}

type PlayerEventChoiceSubmittedEvent struct {
	Player *Player
}

func (PlayerEventChoiceSubmittedEvent) isEvent() {}

type HeroMatFlippedEvent struct {
	Player *Player
	From   *Hero
	To     *Hero
}

func (HeroMatFlippedEvent) isEvent() {}

type PlayerHandVoidedEvent struct {
	Player      *Player
	VoidedCards []PlayerCard
}

func (PlayerHandVoidedEvent) isEvent() {}
