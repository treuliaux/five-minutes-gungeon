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
	Resources []ResourceType
}

func (rc *ResourceCard) isPlayerCard() {}

type ActionCard struct {
	Name   string
	Action CardAction
}

func (ac *ActionCard) isPlayerCard() {}

// Dungeon types

type DungeonCard interface {
	isDungeonCard()
	Require() []ResourceType
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
	Name      string
	Resources []ResourceType
}

func (dc *DoorCard) isDungeonCard() {}
func (dc *DoorCard) Require() []ResourceType {
	return dc.Resources
}

type EventCard struct {
	Type   ChallengeKind
	Name   string
	Action EventAction
}

func (ec *EventCard) isDungeonCard()   {}
func (ec *EventCard) isChallengeCard() {}
func (ec *EventCard) Require() []ResourceType {
	return nil
}

type MiniBossCard struct {
	Type      ChallengeKind
	Name      string
	Resources []ResourceType
}

func (mc *MiniBossCard) isDungeonCard()   {}
func (mc *MiniBossCard) isChallengeCard() {}
func (mc *MiniBossCard) Require() []ResourceType {
	return mc.Resources
}

type BossMat struct {
	Name      string
	Resources []ResourceType
}

func (bm *BossMat) isDungeonCard() {}
func (bm *BossMat) Require() []ResourceType {
	return bm.Resources
}
