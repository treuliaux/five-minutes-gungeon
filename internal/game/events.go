package game

import "time"

type Event interface {
	isEvent()
}

type TimerTickEvent struct {
	TimeLeftDuration time.Duration
}

func (TimerTickEvent) isEvent() {}

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
	CardID CardID `json:"cardId"`
}

func (DoorDefeatedEvent) isEvent() {}

type FieldClearedEvent struct {
}

func (FieldClearedEvent) isEvent() {}

type PlayerAddedEvent struct {
	PlayerID PlayerID `json:"playerId"`
}

func (PlayerAddedEvent) isEvent() {}

type PlayerChangedHeroEvent struct {
	Player PlayerDTO `json:"player"`
}

func (PlayerChangedHeroEvent) isEvent() {}

type CardPlayedEvent struct {
	ByPlayer PlayerDTO     `json:"byPlayer"`
	Card     PlayerCardDTO `json:"card"`
}

func (CardPlayedEvent) isEvent() {}

type CardRemovedEvent struct {
	CardID CardID `json:"cardId"`
}

func (CardRemovedEvent) isEvent() {}

type DoorCardRemovedEvent struct {
	CardID CardID `json:"cardId"`
}

func (DoorCardRemovedEvent) isEvent() {}

type DungeonCardSentToBottomEvent struct {
	CardID CardID `json:"cardId"`
}

func (DungeonCardSentToBottomEvent) isEvent() {}

type DungeonCardDiscardedEvent struct {
	CardID CardID `json:"cardId"`
}

func (DungeonCardDiscardedEvent) isEvent() {}

type CurseActivatedEvent struct {
	CardID CardID `json:"cardId"`
}

func (CurseActivatedEvent) isEvent() {}

type CurseRemovedEvent struct {
	CardID CardID `json:"cardId"`
}

func (CurseRemovedEvent) isEvent() {}

type EventCounteredEvent struct {
	ByPlayerID     PlayerID   `json:"byPlayerId"`
	CardID         CardID     `json:"cardId"`
	WithCardID     CardID     `json:"withCardId,omitempty"`
	WithArtifactID ArtifactID `json:"withArtifactId,omitempty"`
}

func (EventCounteredEvent) isEvent() {}

type CardDrawnFromDeckEvent struct {
	ByPlayerID PlayerID `json:"byPlayerId"`
	CardID     CardID   `json:"cardId"`
}

func (CardDrawnFromDeckEvent) isEvent() {}

type CardDrawnFromDiscardEvent struct {
	ByPlayerID PlayerID `json:"byPlayerId"`
	CardID     CardID   `json:"cardId"`
}

func (CardDrawnFromDiscardEvent) isEvent() {}

type CardDiscardedEvent struct {
	ByPlayerID PlayerID `json:"byPlayerId"`
	CardID     CardID   `json:"cardId"`
}

func (CardDiscardedEvent) isEvent() {}

type TimeFrozenEvent struct {
	ByPlayerID PlayerID `json:"byPlayerId"`
}

func (TimeFrozenEvent) isEvent() {}

type TimeUnfrozenEvent struct {
	ByPlayerID PlayerID `json:"byPlayerId"`
}

func (TimeUnfrozenEvent) isEvent() {}

type DoorOpenedEvent struct {
	CardID CardID `json:"cardId"`
}

func (DoorOpenedEvent) isEvent() {}

type HeroAbilityUsedEvent struct {
	ByPlayer PlayerDTO `json:"byPlayer"`
}

func (HeroAbilityUsedEvent) isEvent() {}

type ArtifactUsedEvent struct {
	ByPlayerID PlayerID `json:"byPlayerId"`
}

func (ArtifactUsedEvent) isEvent() {}

type ArtifactReEnabledEvent struct {
	ArtifactID ArtifactID `json:"artifactID"`
}

func (ArtifactReEnabledEvent) isEvent() {}

type PlayerHealedEvent struct {
	PlayerID PlayerID `json:"playerId"`
	CardIDs  []CardID `json:"cardIds"`
}

func (PlayerHealedEvent) isEvent() {}

type ExtensionToggledEvent struct {
	Enabled bool `json:"enabled"`
}

func (ExtensionToggledEvent) isEvent() {}

type HandDonatedEvent struct {
	FromPlayerID PlayerID `json:"fromPlayerId"`
	ToPlayerID   PlayerID `json:"toPlayerId"`
	CardIDs      []CardID `json:"cardIds"`
}

func (HandDonatedEvent) isEvent() {}

type HandStolenEvent struct {
	FromPlayerID PlayerID `json:"fromPlayerId"`
	ToPlayerID   PlayerID `json:"toPlayerId"`
	CardIDs      []CardID `json:"cardIds"`
}

func (HandStolenEvent) isEvent() {}

type EventPromptOpenedEvent struct {
	Kind           InteractionKind  `json:"kind"`
	RequiredCounts map[PlayerID]int `json:"requiredCounts,omitempty"`
}

func (EventPromptOpenedEvent) isEvent() {}

type PlayerEventChoiceSubmittedEvent struct {
	PlayerID PlayerID `json:"playerId"`
}

func (PlayerEventChoiceSubmittedEvent) isEvent() {}

type HeroMatFlippedEvent struct {
	PlayerID PlayerID  `json:"playerId"`
	From     HeroClass `json:"from"`
	To       HeroClass `json:"to"`
}

func (HeroMatFlippedEvent) isEvent() {}

type PlayerHandVoidedEvent struct {
	PlayerID      PlayerID `json:"playerId"`
	VoidedCardIDs []CardID `json:"voidedCardIDs"`
}

func (PlayerHandVoidedEvent) isEvent() {}

type DungeonDefeated struct {
	Boss     DungeonCardDTO `json:"boss"`
	NextBoss DungeonCardDTO `json:"nextBoss"`
}

func (DungeonDefeated) isEvent() {}

type DungeonPrevailed struct {
	Boss     DungeonCardDTO `json:"boss"`
	NextBoss DungeonCardDTO `json:"nextBoss"`
}

func (DungeonPrevailed) isEvent() {}

type CampaignEnded struct {
	Boss DungeonCardDTO `json:"boss"`
}

func (CampaignEnded) isEvent() {}
