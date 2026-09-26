package game

import "time"

type ArtifactID uint32
type CardID uint32
type IdentifiableCard interface {
	ID() CardID
}

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
	ID() CardID
}

type ResourceCard struct {
	Id        CardID
	Resources []ResourceType
}

func (rc *ResourceCard) isPlayerCard() {}
func (rc *ResourceCard) ID() CardID {
	return rc.Id
}

type ActionCard struct {
	Id        CardID
	Name      string
	Action    CardAction
	Extension bool
}

func (ac *ActionCard) isPlayerCard() {}
func (ac *ActionCard) ID() CardID {
	return ac.Id
}

type ArtifactCard struct {
	Id     ArtifactID
	Color  DeckColor
	Name   string
	Action ArtifactAction
	Used   bool
}

func (ac *ArtifactCard) ID() ArtifactID {
	return ac.Id
}

// Dungeon types

type DungeonCard interface {
	isDungeonCard()
	Require() []ResourceType
	ID() CardID
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
	Id        CardID
	Type      DoorKind
	Name      string
	Resources []ResourceType
}

func (dc *DoorCard) isDungeonCard() {}
func (dc *DoorCard) Require() []ResourceType {
	return dc.Resources
}
func (dc *DoorCard) ID() CardID {
	return dc.Id
}

type EventCard struct {
	Id         CardID
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
func (ec *EventCard) ID() CardID {
	return ec.Id
}

type MiniBossCard struct {
	Id        CardID
	Type      ChallengeKind
	Name      string
	Resources []ResourceType
	Extension bool
}

func (mc *MiniBossCard) isDungeonCard() {}
func (mc *MiniBossCard) Require() []ResourceType {
	return mc.Resources
}
func (mc *MiniBossCard) ID() CardID {
	return mc.Id
}

type CurseCard struct {
	Id     CardID
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
func (cc *CurseCard) ID() CardID {
	return cc.Id
}

type BossMat struct {
	Id                   CardID
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
func (bm *BossMat) ID() CardID {
	return bm.Id
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
