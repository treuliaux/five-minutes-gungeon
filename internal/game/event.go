package game

type Event interface {
	isEvent()
}

type GameStartedEvent struct {
}

func (evt GameStartedEvent) isEvent() {}

type GameLostEvent struct {
}

func (evt GameLostEvent) isEvent() {}

type GameWonEvent struct {
}

func (evt GameWonEvent) isEvent() {}

type DoorDefeatedEvent struct {
	DungeonCard DungeonCard
}

func (evt DoorDefeatedEvent) isEvent() {}

type FieldClearedEvent struct {
}

func (evt FieldClearedEvent) isEvent() {}

type PlayerAddedEvent struct {
	Player *Player
}

func (evt PlayerAddedEvent) isEvent() {}

type CardPlayedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (evt CardPlayedEvent) isEvent() {}

type CardDiscardedEvent struct {
	ByPlayer *Player
	Card     PlayerCard
}

func (evt CardDiscardedEvent) isEvent() {}

type TimeFrozenEvent struct {
	ByPlayer *Player
}

func (evt TimeFrozenEvent) isEvent() {}

type TimeUnfrozenEvent struct {
	ByPlayer *Player
}

func (evt TimeUnfrozenEvent) isEvent() {}

type DoorOpenedEvent struct {
	DungeonCard DungeonCard
}

func (evt DoorOpenedEvent) isEvent() {}

type HeroAbilityUsedEvent struct {
	ByPlayer *Player
}

func (evt HeroAbilityUsedEvent) isEvent() {}
