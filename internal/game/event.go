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

type ActionCardPlayedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (ActionCardPlayedEvent) isEvent() {}

type EventCounteredEvent struct {
	ByPlayer  *Player
	EventCard DungeonCard
	WithCard  PlayerCard
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

type PlayerHealedEvent struct {
	Player *Player
	Cards  []PlayerCard
}

func (PlayerHealedEvent) isEvent() {}
