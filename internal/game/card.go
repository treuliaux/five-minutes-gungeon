package game

import "time"

type ResourceType int

const (
	Sword ResourceType = iota
	Arrow
	Shield
	Jump
	Scroll
	WildCard
	InfiniteSword
	InfiniteArrow
	InfiniteShield
	InfiniteJump
	InfiniteScroll
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
	Name      string
	Action    CardAction
	Extension bool
}

func (ac *ActionCard) isPlayerCard() {}

type ArtifactCard struct {
	Color  DeckColor
	Name   string
	Action ArtifactAction
	Used   bool
}

// Dungeon types

type DungeonCard interface {
	isDungeonCard()
	Require() []ResourceType
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
	ChallengeCurse
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
	Type       ChallengeKind
	Name       string
	Action     EventAction
	OpenedTime time.Duration
	Extension  bool
}

func (ec *EventCard) isDungeonCard() {}
func (ec *EventCard) Require() []ResourceType {
	return nil
}

type MiniBossCard struct {
	Type      ChallengeKind
	Name      string
	Resources []ResourceType
	Extension bool
}

func (mc *MiniBossCard) isDungeonCard() {}
func (mc *MiniBossCard) Require() []ResourceType {
	return mc.Resources
}

type CurseCard struct {
	Type   ChallengeKind
	Name   string
	Apply  CurseHook
	Cure   CurseHook
	Effect GameCurseEffect
}

func (cc *CurseCard) isDungeonCard() {}
func (cc *CurseCard) Require() []ResourceType {
	return nil
}

type BossMat struct {
	Name                 string
	Resources            []ResourceType
	DeckSize             int
	AdditionalChallenges int
	SpecialAbilities     []DungeonCard
}

func (bm *BossMat) isDungeonCard() {}
func (bm *BossMat) Require() []ResourceType {
	return bm.Resources
}

func IsInfiniteVersionOf(inf ResourceType, base ResourceType) bool {
	switch inf {
	case InfiniteSword:
		return base == Sword
	case InfiniteArrow:
		return base == Arrow
	case InfiniteShield:
		return base == Shield
	case InfiniteJump:
		return base == Jump
	case InfiniteScroll:
		return base == Scroll
	default:
		return false
	}
}

func IsBaseResource(base ResourceType) bool {
	switch base {
	case Sword:
		return true
	case Arrow:
		return true
	case Shield:
		return true
	case Jump:
		return true
	case Scroll:
		return true
	default:
		return false
	}
}
