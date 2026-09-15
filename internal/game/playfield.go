package game

import (
	"fmt"
	"slices"
)

type Playfield struct {
	Field       []PlayerCard
	OpenedDoors []DungeonCard
}

func NewPlayfield() *Playfield {
	return &Playfield{
		Field:       make([]PlayerCard, 0, 16),
		OpenedDoors: make([]DungeonCard, 0, 2),
	}
}

func (p *Playfield) AddPlayerCard(player *Player, card PlayerCard) ([]Event, error) {
	p.Field = append(p.Field, card)

	return []Event{CardPlayedEvent{
		ByPlayer: player,
		Card:     card,
	}}, nil
}

func (p *Playfield) AddDungeonCard(card DungeonCard) ([]Event, error) {
	if p.IsDoorsFull() {
		return nil, fmt.Errorf("cannot add more than two dungeon cards")
	}
	p.OpenedDoors = append(p.OpenedDoors, card)

	return []Event{DoorOpenedEvent{DungeonCard: card}}, nil
}

func (p *Playfield) ClearField() ([]Event, error) {
	p.OpenedDoors = make([]DungeonCard, 0, 2)
	p.Field = make([]PlayerCard, 0, 16)

	return []Event{FieldClearedEvent{}}, nil
}

func (p *Playfield) IsDoorsFull() bool {
	return len(p.OpenedDoors) == 2
}

func (p *Playfield) HasDoorsOpened() bool {
	return len(p.OpenedDoors) != 0
}

func (p *Playfield) HasActiveDoor(target DungeonCard) bool {
	return slices.Contains(p.OpenedDoors, target)
}

func (p *Playfield) DefeatDoor(target DungeonCard) ([]Event, error) {
	if !p.HasActiveDoor(target) {
		return nil, fmt.Errorf("target is not the current dungeon card")
	}

	p.OpenedDoors = slices.DeleteFunc(p.OpenedDoors, func(dungeonCard DungeonCard) bool {
		return dungeonCard == target
	})

	return []Event{DoorDefeatedEvent{DungeonCard: target}}, nil
}

func (p *Playfield) DefeatAllDoors() ([]Event, error) {
	if len(p.OpenedDoors) == 0 {
		return nil, fmt.Errorf("no active doors")
	}

	var events []Event
	for _, card := range p.OpenedDoors {
		events = append(events, DoorDefeatedEvent{DungeonCard: card})
	}
	clearEvents, err := p.ClearField()
	events = append(events, clearEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (p *Playfield) IsPlayfieldBeaten() bool {
	if !p.HasDoorsOpened() {
		return true
	}

	totalPlayed := make(map[ResourceType]int, 6)
	for _, playedCard := range p.Field {
		if rc, ok := playedCard.(*ResourceCard); ok {
			for _, r := range rc.Resources {
				totalPlayed[r]++
			}
		}
	}
	totalRequired := make(map[ResourceType]int, 6)
	for _, door := range p.OpenedDoors {
		for _, r := range door.Require() {
			totalRequired[r]++
		}
	}
	usedWildCard := 0
	for res, count := range totalRequired {
		if totalPlayed[res] < count {
			deficit := count - totalPlayed[res]
			availableWild := totalPlayed[WildCard] - usedWildCard
			if availableWild < deficit {
				return false
			}
			usedWildCard += deficit
		}
	}

	return true
}

func (p *Playfield) IsFightingBoss(dungeon *Dungeon) bool {
	return len(p.OpenedDoors) == 1 && p.OpenedDoors[0] == dungeon.Boss
}
