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
	CardID CardID
}

func (DoorDefeatedEvent) isEvent() {}

type FieldClearedEvent struct {
}

func (FieldClearedEvent) isEvent() {}

type PlayerAddedEvent struct {
	PlayerID PlayerID
}

func (PlayerAddedEvent) isEvent() {}

type CardPlayedEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (CardPlayedEvent) isEvent() {}

type CardRemovedEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (CardRemovedEvent) isEvent() {}

type DoorCardRemovedEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (DoorCardRemovedEvent) isEvent() {}

type DungeonCardSentToBottomEvent struct {
	CardID CardID
}

func (DungeonCardSentToBottomEvent) isEvent() {}

type DungeonCardDiscardedEvent struct {
	CardID CardID
}

func (DungeonCardDiscardedEvent) isEvent() {}

type CurseActivatedEvent struct {
	CardID CardID
}

func (CurseActivatedEvent) isEvent() {}

type CurseRemovedEvent struct {
	CardID CardID
}

func (CurseRemovedEvent) isEvent() {}

type ActionCardPlayedEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (ActionCardPlayedEvent) isEvent() {}

type EventCounteredEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
	WithCardID any
}

func (EventCounteredEvent) isEvent() {}

type CardDrawnFromDeckEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (CardDrawnFromDeckEvent) isEvent() {}

type CardDrawnFromDiscardEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (CardDrawnFromDiscardEvent) isEvent() {}

type CardDiscardedEvent struct {
	ByPlayerID PlayerID
	CardID     CardID
}

func (CardDiscardedEvent) isEvent() {}

type TimeFrozenEvent struct {
	ByPlayerID PlayerID
}

func (TimeFrozenEvent) isEvent() {}

type TimeUnfrozenEvent struct {
	ByPlayerID PlayerID
}

func (TimeUnfrozenEvent) isEvent() {}

type DoorOpenedEvent struct {
	CardID CardID
}

func (DoorOpenedEvent) isEvent() {}

type HeroAbilityUsedEvent struct {
	ByPlayerID PlayerID
}

func (HeroAbilityUsedEvent) isEvent() {}

type ArtifactUsedEvent struct {
	ByPlayerID PlayerID
}

func (ArtifactUsedEvent) isEvent() {}

type ArtifactReEnabledEvent struct {
	ArtifactID *ArtifactCard
}

func (ArtifactReEnabledEvent) isEvent() {}

type PlayerHealedEvent struct {
	PlayerID PlayerID
	CardIDs  []CardID
}

func (PlayerHealedEvent) isEvent() {}

type ExtensionToggledEvent struct {
	Enabled bool
}

func (ExtensionToggledEvent) isEvent() {}

type HandDonatedEvent struct {
	FromPlayerID PlayerID
	ToPlayerID   PlayerID
	CardIDs      []CardID
}

func (HandDonatedEvent) isEvent() {}

type HandStolenEvent struct {
	FromPlayerID PlayerID
	ToPlayerID   PlayerID
	CardIDs      []CardID
}

func (HandStolenEvent) isEvent() {}

type EventPromptOpenedEvent struct {
	Kind           InteractionKind
	RequiredCounts map[PlayerID]int
}

func (EventPromptOpenedEvent) isEvent() {}

type PlayerEventChoiceSubmittedEvent struct {
	PlayerID PlayerID
}

func (PlayerEventChoiceSubmittedEvent) isEvent() {}

type HeroMatFlippedEvent struct {
	PlayerID PlayerID
	From     *Hero
	To       *Hero
}

func (HeroMatFlippedEvent) isEvent() {}

type PlayerHandVoidedEvent struct {
	PlayerID      PlayerID
	VoidedCardIDs []CardID
}

func (PlayerHandVoidedEvent) isEvent() {}
