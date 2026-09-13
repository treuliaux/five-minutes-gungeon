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
	IsBeaten(playedCards []PlayerCard) bool
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
func (r *DoorCard) IsBeaten(playedCards []PlayerCard) bool {
	return checkResourcesFulfilled(playedCards, r.resources)
}

type EventCard struct {
	Type   ChallengeKind
	name   string
	Action EventAction
}

func (r *EventCard) isDungeonCard()   {}
func (r *EventCard) isChallengeCard() {}
func (r *EventCard) IsBeaten(_ []PlayerCard) bool {
	return true
}

type MiniBossCard struct {
	Type      ChallengeKind
	name      string
	resources []ResourceType
}

func (r *MiniBossCard) isDungeonCard()   {}
func (r *MiniBossCard) isChallengeCard() {}
func (r *MiniBossCard) IsBeaten(playedCards []PlayerCard) bool {
	return checkResourcesFulfilled(playedCards, r.resources)
}

type BossMat struct {
	name      string
	resources []ResourceType
}

func (r *BossMat) isDungeonCard() {}
func (r *BossMat) IsBeaten(playedCards []PlayerCard) bool {
	return checkResourcesFulfilled(playedCards, r.resources)
}

func checkResourcesFulfilled(playedCards []PlayerCard, required []ResourceType) bool {
	totalPlayed := make(map[ResourceType]int, 5)
	for _, playedCard := range playedCards {
		if rc, ok := playedCard.(*ResourceCard); ok {
			for _, r := range rc.resources {
				totalPlayed[r]++
			}
		}
	}
	totalRequired := make(map[ResourceType]int, 5)
	for _, r := range required {
		totalRequired[r]++
	}
	for res, count := range totalRequired {
		if totalPlayed[res] < count {
			return false
		}
	}
	return true
}
