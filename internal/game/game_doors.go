package game

import (
	"fmt"
	"slices"
)

func (g *Game) OpenDoor() ([]Event, error) {
	if g.LevelState.Dungeon == nil {
		return nil, nil
	}
	var events []Event
	var err error
	loop := true
	for loop || (g.HasActiveCurseEffect(DoorsOpenInPairs) && len(g.LevelState.Playfield.OpenedDoors) != 2) {
		loop = len(g.LevelState.Playfield.OpenedDoors) == 0
		card := g.LevelState.Dungeon.OpenDoor()
		if card == nil {
			break
		}
		var addDungeonCardEvents []Event
		addDungeonCardEvents, err = g.LevelState.Playfield.AddDungeonCard(card, g)
		events = append(events, addDungeonCardEvents...)
		if err != nil {
			return events, err
		}
		switch card.(type) {
		case *DoorCard, *MiniBossCard, *EventCard:
			loop = false
		case *BossMat:
			g.LevelState.IsFightingBoss = true
			loop = false
		case *CurseCard:
		}
	}

	return events, nil
}

func (g *Game) DefeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event

	if _, ok := target.(*CurseCard); ok {
		return nil, fmt.Errorf("curses cannot be defeated")
	}

	defeatDoorEvents, err := g.LevelState.Playfield.DefeatDoor(target)
	events = append(events, defeatDoorEvents...)
	if err != nil {
		return events, err
	}

	if _, ok := target.(*BossMat); ok || g.LevelState.IsFightingBoss {
		events = append(events, GameWonEvent{})
		events = append(events, g.victory()...)

		return events, nil
	}

	if g.LevelState.Playfield.HasDoorsOpened() {
		return events, nil
	}

	events = append(events, g.clearField()...)

	openDoorEvents, err := g.OpenDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) ActiveDoors(f *DoorsFilter) []DungeonCard {
	if f == nil {
		f = NewDoorsFilter().AddEverything()
	}
	var picked []DungeonCard
	var playfieldHaystack []DungeonCard
	for _, door := range g.LevelState.Playfield.OpenedDoors {
		playfieldHaystack = append(playfieldHaystack, door)
	}
	for _, curse := range g.LevelState.Playfield.ActiveCurses {
		playfieldHaystack = append(playfieldHaystack, curse)
	}
	for _, door := range playfieldHaystack {
		switch d := door.(type) {
		case *BossMat:
			if f.IncludeBossMat {
				picked = append(picked, d)
			}
		case *CurseCard:
			if slices.Contains(f.ChallengeKinds, ChallengeCurse) {
				picked = append(picked, d)
			}
		case *EventCard:
			if slices.Contains(f.ChallengeKinds, ChallengeEvent) {
				picked = append(picked, d)
			}
		case *MiniBossCard:
			if slices.Contains(f.ChallengeKinds, ChallengeMiniBoss) {
				picked = append(picked, d)
			}
		case *DoorCard:
			if slices.Contains(f.DoorKinds, d.Type) {
				picked = append(picked, d)
			}
		}
	}

	return picked
}

func (g *Game) SendDungeonCardBottomDungeon(card DungeonCard) ([]Event, error) {
	switch card.(type) {
	case *BossMat, *EventCard:
		return nil, fmt.Errorf("card cannot be sent back to dungeon")
	}
	events, err := g.LevelState.Playfield.RemoveDungeonCard(g, card)
	if err != nil {
		return events, err
	}
	g.LevelState.Dungeon.PutDoorBelowDeck(card)

	if len(g.LevelState.Playfield.OpenedDoors) == 0 {
		openDoorEvents, err := g.OpenDoor()
		events = append(events, openDoorEvents...)
		if err != nil {
			return events, err
		}
	}

	return append(events, DungeonCardSentToBottomEvent{CardID: card.ID()}), nil
}

func (g *Game) DiscardTopCardFromDungeon() ([]Event, error) {
	if len(g.LevelState.Dungeon.Doors) == 0 {
		return nil, nil
	}

	return []Event{DungeonCardDiscardedEvent{CardID: g.LevelState.Dungeon.OpenDoor().ID()}}, nil
}
