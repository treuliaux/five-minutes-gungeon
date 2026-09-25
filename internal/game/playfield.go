package game

import (
	"fmt"
	"slices"
)

type Playfield struct {
	Field        []PlayerCard
	OpenedDoors  []DungeonCard
	ActiveCurses []*CurseCard
	Artifacts    []*ArtifactCard
}

func NewPlayfield() *Playfield {
	return &Playfield{
		Field:        make([]PlayerCard, 0, 16),
		OpenedDoors:  make([]DungeonCard, 0, 2),
		ActiveCurses: make([]*CurseCard, 0, 5),
		Artifacts:    make([]*ArtifactCard, 0, 6),
	}
}

func (p *Playfield) AddPlayerCard(player *Player, card PlayerCard) ([]Event, error) {
	p.Field = append(p.Field, card)

	return []Event{CardPlayedEvent{
		ByPlayer: player,
		Card:     card,
	}}, nil
}

func (p *Playfield) RemovePlayerCard(card PlayerCard) ([]Event, error) {
	if !slices.Contains(p.Field, card) {
		return nil, fmt.Errorf("target card is not in play")
	}
	p.Field = slices.DeleteFunc(p.Field, func(c PlayerCard) bool { return c == card })

	return []Event{CardRemovedEvent{
		Card: card,
	}}, nil
}

func (p *Playfield) RemoveDungeonCard(game *Game, card DungeonCard) ([]Event, error) {
	switch c := card.(type) {
	case *BossMat:
		return nil, fmt.Errorf("boss mat cannot be removed")
	case *CurseCard:
		if !slices.Contains(p.ActiveCurses, c) {
			return nil, fmt.Errorf("target card is not in play")
		}
		p.ActiveCurses = slices.DeleteFunc(p.ActiveCurses, func(e *CurseCard) bool { return e == c })
		var events []Event
		if c.Cure != nil {
			cureEvents, err := c.Cure(&CardCurseContext{
				engine: game,
				Card:   c,
			})
			events = append(events, cureEvents...)
			if err != nil {
				return events, err
			}
		}

		return append([]Event{CurseRemovedEvent{Card: c}}, events...), nil
	case *DoorCard, *MiniBossCard:
		if !slices.Contains(p.OpenedDoors, c) {
			return nil, fmt.Errorf("target card is not in play")
		}
		p.OpenedDoors = slices.DeleteFunc(p.OpenedDoors, func(e DungeonCard) bool { return e == c })

		return []Event{DoorCardRemovedEvent{Card: c}}, nil
	case *EventCard:
		if !slices.Contains(p.OpenedDoors, card) {
			return nil, fmt.Errorf("target card is not in play")
		}
		p.OpenedDoors = slices.DeleteFunc(p.OpenedDoors, func(e DungeonCard) bool { return e == c })

		return []Event{DoorCardRemovedEvent{Card: c}}, nil
	default:
		return nil, fmt.Errorf("unexpected dungeon card type: %v", c)
	}
}

func (p *Playfield) AddDungeonCard(card DungeonCard, g *Game) ([]Event, error) {
	switch c := card.(type) {
	case *CurseCard:
		p.ActiveCurses = append(p.ActiveCurses, c)
		var events []Event
		if c.Apply != nil {
			applyEvents, err := c.Apply(&CardCurseContext{
				engine: g,
				Card:   c,
			})
			events = append(events, applyEvents...)
			if err != nil {
				return events, err
			}
		}

		return append([]Event{DoorOpenedEvent{DungeonCard: c}, CurseActivatedEvent{Card: c}}, events...), nil
	case *DoorCard, *MiniBossCard, *BossMat:
		if p.IsDoorsFull() {
			return nil, fmt.Errorf("cannot add more than two dungeon cards")
		}
		p.OpenedDoors = append(p.OpenedDoors, c)

		return []Event{DoorOpenedEvent{DungeonCard: c}}, nil
	case *EventCard:
		if p.IsDoorsFull() {
			return nil, fmt.Errorf("cannot add more than two dungeon cards")
		}
		c.OpenedTime = g.InGameTimer
		p.OpenedDoors = append(p.OpenedDoors, c)

		return []Event{DoorOpenedEvent{DungeonCard: c}}, nil
	default:
		return nil, fmt.Errorf("unexpected dungeon card type: %v", c)
	}
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
		return nil, fmt.Errorf("target is not an active dungeon card")
	}
	p.OpenedDoors = slices.DeleteFunc(p.OpenedDoors, func(c DungeonCard) bool { return c == target })

	return []Event{DoorDefeatedEvent{DungeonCard: target}}, nil
}

func (p *Playfield) DefeatAllDoors() ([]Event, error) {
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

	return !p.HasActiveEvents() && p.hasAllRequiredResources()
}

func (p *Playfield) IsFightingBoss(dungeon *Dungeon) bool {
	return len(p.OpenedDoors) == 1 && p.OpenedDoors[0] == dungeon.Boss
}

func (p *Playfield) DoorsOnly() []DungeonCard {
	var openedDoors []DungeonCard
	for _, c := range p.OpenedDoors {
		if _, ok := c.(*BossMat); !ok {
			openedDoors = append(openedDoors, c)
		}
	}

	return openedDoors
}

func (p *Playfield) HasActiveEvents() bool {
	for _, openedDoor := range p.OpenedDoors {
		switch openedDoor.(type) {
		case *EventCard:
			return true
		default:
		}
	}

	return false
}

func (p *Playfield) ResolveEvent(ctx *CardEventContext) ([]Event, error) {
	eventCard, ok := ctx.Card.(*EventCard)
	if !ok {
		return nil, fmt.Errorf("expected event card")
	}

	if eventCard.Action == nil {
		return nil, nil
	}
	events, err := eventCard.Action.Execute(ctx)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (p *Playfield) hasAllRequiredResources() bool {
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
	for requiredResource, count := range totalRequired {
		if IsBaseResource(requiredResource) {
			hasInfiniteVersionPlayed := false
			for playedResource := range totalPlayed {
				if IsInfiniteVersionOf(playedResource, requiredResource) {
					hasInfiniteVersionPlayed = true
					break
				}
			}
			if hasInfiniteVersionPlayed {
				continue
			}
		}

		if totalPlayed[requiredResource] < count {
			deficit := count - totalPlayed[requiredResource]
			availableWild := totalPlayed[WildCard] - usedWildCard
			if availableWild < deficit {
				return false
			}
			usedWildCard += deficit
		}
	}

	return true
}

func (p *Playfield) SetupArtifacts(deckColors []DeckColor) {
	for _, a := range Artifacts() {
		if !slices.Contains(deckColors, a.Color) {
			p.Artifacts = append(p.Artifacts, a)
		}
	}
}

func Artifacts() []*ArtifactCard {
	return []*ArtifactCard{
		{
			Color:  Green,
			Name:   "Rainbow Herbs",
			Action: &RainbowHerbsArtifact{},
			Used:   false,
		},
		{
			Color:  Yellow,
			Name:   "M-jh'öilnør",
			Action: &MJhoilnorArtifact{},
			Used:   false,
		},
		{
			Color:  Purple,
			Name:   "Sundial Watch",
			Action: &SundialWatchArtifact{},
			Used:   false,
		},
		{
			Color:  Red,
			Name:   "Battle Axe",
			Action: &BattleAxeArtifact{},
			Used:   false,
		},
		{
			Color:  Blue,
			Name:   "The Infinity Scroll",
			Action: &TheInfinityScrollArtifact{},
			Used:   false,
		},
		{
			Color:  Black,
			Name:   "Curse Zapper",
			Action: &CurseZapperArtifact{},
			Used:   false,
		},
	}
}

func (p *Playfield) ArtifactCanBePlayed(artifact *ArtifactCard) bool {
	return slices.Contains(p.Artifacts, artifact) && !artifact.Used
}
