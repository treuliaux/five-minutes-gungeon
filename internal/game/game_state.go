package game

import "time"

type LevelState struct {
	// Active Level & Playfield LevelState (Reset between levels)
	Status             Status
	PendingInteraction PendingInteraction
	Playfield          *Playfield
	Dungeon            *Dungeon
	IsFightingBoss     bool
	IsTimeFrozen       bool

	// Timing & Debounce
	InGameTimer         time.Duration
	RealElapsedTime     time.Duration
	LastPlayedCardTimer time.Duration

	// Fast Index & Lookup Caches
	CurseExpectingDiscards map[*Player]int
	PlayerCardsMap         map[CardID]PlayerCard
	DungeonCardsMap        map[CardID]DungeonCard
}

func newRoundState(status Status) *LevelState {
	return &LevelState{
		Status:                 status,
		Playfield:              NewPlayfield(),
		CurseExpectingDiscards: make(map[*Player]int),
	}
}

func (r *LevelState) ClearStopTimeCurse() {
	r.CurseExpectingDiscards = make(map[*Player]int)
}
