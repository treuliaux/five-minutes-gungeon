package game

type ResourceType int

const (
	Sword ResourceType = iota
	Arrow
	Shield
	Jump
	Scroll
)

// Deck types

type PlayerCard interface {
	isPlayerCard()
}

type ResourceCard struct {
	Resources []ResourceType
}

func (r ResourceCard) isPlayerCard() {}

type ActionCard struct {
	Name   string
	Action CardAction
}

func (r ActionCard) isPlayerCard() {}

// Dungeon types

type DungeonCard interface {
	isDungeonCard()
}

type ChallengeCard interface {
	isChallengeCard()
}

type DoorKind int

const (
	Monster DoorKind = iota
	Obstacle
	Person
)

type ChallengeKind int

const (
	MiniBoss ChallengeKind = iota
	Event
)

type DoorCard struct {
	Type      DoorKind
	Name      string
	Resources []ResourceType
}

func (r DoorCard) isDungeonCard() {}

type EventCard struct {
	Type   ChallengeKind
	Name   string
	Action EventAction
}

func (r EventCard) isDungeonCard()   {}
func (r EventCard) isChallengeCard() {}

type MiniBossCard struct {
	d interface {
		isDoorCard()
	}
	Type      ChallengeKind
	Name      string
	Resources []ResourceType
}

func (r MiniBossCard) isDungeonCard()   {}
func (r MiniBossCard) isChallengeCard() {}

type BossMat struct {
	Name      string
	Resources []ResourceType
}

func (r BossMat) isDungeonCard() {}
