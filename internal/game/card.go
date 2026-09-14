package game

type ResourceType int

const (
	Sword ResourceType = iota
	Arrow
	Shield
	Jump
	Scroll
	WildCard
)

// Deck types

type PlayerCard interface {
	isPlayerCard()
}

type ResourceCard struct {
	resources []ResourceType
}

func (r *ResourceCard) isPlayerCard() {}

type ActionCard struct {
	name   string
	action CardAction
}

func (r *ActionCard) isPlayerCard() {}

// Dungeon types

type DungeonCard interface {
	isDungeonCard()
	require() []ResourceType
}

type ChallengeCard interface {
	isChallengeCard()
}

type DoorKind int

const (
	DoorMonster DoorKind = iota
	DoorObstacle
	DoorPerson
)

type ChallengeKind int

const (
	ChallengeMiniBoss ChallengeKind = iota
	ChallengeEvent
)

type DoorCard struct {
	Type      DoorKind
	name      string
	resources []ResourceType
}

func (r *DoorCard) isDungeonCard() {}
func (r *DoorCard) require() []ResourceType {
	return r.resources
}

type EventCard struct {
	Type   ChallengeKind
	name   string
	Action EventAction
}

func (r *EventCard) isDungeonCard()   {}
func (r *EventCard) isChallengeCard() {}
func (r *EventCard) require() []ResourceType {
	return []ResourceType{}
}

type MiniBossCard struct {
	Type      ChallengeKind
	name      string
	resources []ResourceType
}

func (r *MiniBossCard) isDungeonCard()   {}
func (r *MiniBossCard) isChallengeCard() {}
func (r *MiniBossCard) require() []ResourceType {
	return r.resources
}

type BossMat struct {
	name      string
	resources []ResourceType
}

func (r *BossMat) isDungeonCard() {}
func (r *BossMat) require() []ResourceType {
	return r.resources
}
