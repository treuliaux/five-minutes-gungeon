package game

import (
	"fmt"
	"slices"
)

type Playfield struct {
	field       []PlayerCard
	openedDoors []DungeonCard
}

func NewPlayfield() *Playfield {
	return &Playfield{
		field:       make([]PlayerCard, 0, 16),
		openedDoors: make([]DungeonCard, 0, 2),
	}
}

func (p *Playfield) addPlayerCard(player *Player, card PlayerCard) ([]Event, error) {
	p.field = append(p.field, card)

	return []Event{CardPlayedEvent{
		ByPlayer: player,
		Card:     card,
	}}, nil
}

func (p *Playfield) addDungeonCard(card DungeonCard) ([]Event, error) {
	if p.isDoorsFull() {
		return []Event{}, fmt.Errorf("cannot add more than two dungeon cards")
	}
	p.openedDoors = append(p.openedDoors, card)

	return []Event{DoorOpenedEvent{DungeonCard: card}}, nil
}

func (p *Playfield) clearField() ([]Event, error) {
	var events []Event
	p.openedDoors = make([]DungeonCard, 0, 2)
	p.field = make([]PlayerCard, 0, 16)

	return append(events, FieldClearedEvent{}), nil
}

func (p *Playfield) isDoorsFull() bool {
	return len(p.openedDoors) == 2
}

func (p *Playfield) hasDoorsOpened() bool {
	return len(p.openedDoors) != 0
}

func (p *Playfield) hasActiveDoor(target DungeonCard) bool {
	return slices.Contains(p.openedDoors, target)
}

func (p *Playfield) defeatDoor(target DungeonCard) ([]Event, error) {
	var events []Event
	if !p.hasActiveDoor(target) {
		return events, fmt.Errorf("target is not the current dungeon card")
	}

	p.openedDoors = slices.DeleteFunc(p.openedDoors, func(dungeonCard DungeonCard) bool {
		return dungeonCard == target
	})

	return append(events, DoorDefeatedEvent{DungeonCard: target}), nil
}

func (p *Playfield) defeatAllDoors() ([]Event, error) {
	var events []Event

	if len(p.openedDoors) == 0 {
		return events, fmt.Errorf("no active doors")
	}

	for _, card := range p.openedDoors {
		events = append(events, DoorDefeatedEvent{DungeonCard: card})
	}
	clearFieldEvent, err := p.clearField()
	events = append(events, clearFieldEvent...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (p *Playfield) isPlayfieldBeaten() bool {
	if !p.hasDoorsOpened() {
		return true
	}

	totalPlayed := make(map[ResourceType]int, 6)
	for _, playedCard := range p.field {
		if rc, ok := playedCard.(*ResourceCard); ok {
			for _, r := range rc.resources {
				totalPlayed[r]++
			}
		}
	}
	totalRequired := make(map[ResourceType]int, 6)
	for _, door := range p.openedDoors {
		for _, r := range door.require() {
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

func (p *Playfield) isFightingBoss(dungeon *Dungeon) bool {
	return len(p.openedDoors) == 1 && p.openedDoors[0] == dungeon.boss
}
