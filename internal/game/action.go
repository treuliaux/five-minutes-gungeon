package game

import "fmt"

type CardAction interface {
	isCardAction()
	Execute(ctx CardActionContext) ([]Event, error)
}

type CardActionContext struct {
	Engine GameEngine
	Player *Player
	Card   PlayerCard
}

type AmbiguousCardTargetError struct {
	Card         PlayerCard
	ValidTargets []DungeonCard
}

func (e *AmbiguousCardTargetError) Error() string {
	return fmt.Sprintf("ambiguous target for card %T: %d matching targets available", e.Card, len(e.ValidTargets))
}

type AmbiguousPlayerTargetError struct {
	Card         PlayerCard
	ValidTargets []*Player
}

func (e *AmbiguousPlayerTargetError) Error() string {
	return fmt.Sprintf("ambiguous target for card %T: %d matching targets available", e.Card, len(e.ValidTargets))
}

// Ranger / Huntress

type SnipeAction struct {
	Target DungeonCard
}

func (SnipeAction) isCardAction() {}
func (s SnipeAction) Execute(ctx CardActionContext) ([]Event, error) {
	const targetType = DoorPerson
	target := s.Target
	if target == nil {
		var err error
		target, err = smartCardTargeting(
			ctx.Card,
			func() []DungeonCard { return ctx.Engine.GetActiveDoorsOfKind(targetType) },
		)
		if err != nil {
			return nil, err
		}
	}

	return defeatDoorKindByAction(ctx, target, targetType)
}

type HealingHerbsAction struct {
	Target *Player
}

func (HealingHerbsAction) isCardAction() {}
func (h HealingHerbsAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := h.Target
	if target == nil {
		var err error
		target, err = smartPlayerTargeting(
			ctx.Card,
			func() []*Player { return ctx.Engine.ListOtherPlayers(ctx.Player) },
		)
		if err != nil {
			return nil, err
		}
	}

	return ctx.Engine.DrawCards(target, 4)
}

type CriticalHitAction struct {
	Target DungeonCard
}

func (CriticalHitAction) isCardAction() {}
func (c CriticalHitAction) Execute(ctx CardActionContext) ([]Event, error) {
	const targetType = DoorMonster
	target := c.Target
	if target == nil {
		var err error
		target, err = smartCardTargeting(
			ctx.Card,
			func() []DungeonCard { return ctx.Engine.GetActiveDoorsOfKind(targetType) },
		)
		if err != nil {
			return nil, err
		}
	}

	return defeatDoorKindByAction(ctx, target, targetType)
}

type ExtraQuiverAction struct {
}

func (ExtraQuiverAction) isCardAction() {}
func (e ExtraQuiverAction) Execute(ctx CardActionContext) ([]Event, error) {
	return makePlayersDraw(ctx.Engine, ctx.Engine.ListPlayers(), 2)
}

// Valkyrie / Paladin

type SmiteAction struct {
	Target DungeonCard
}

func (SmiteAction) isCardAction() {}
func (s SmiteAction) Execute(ctx CardActionContext) ([]Event, error) {
	const targetType = DoorMonster
	target := s.Target
	if target == nil {
		var err error
		target, err = smartCardTargeting(
			ctx.Card,
			func() []DungeonCard { return ctx.Engine.GetActiveDoorsOfKind(targetType) },
		)
		if err != nil {
			return nil, err
		}
	}

	return defeatDoorKindByAction(ctx, target, targetType)
}

type DivineShieldAction struct {
}

func (DivineShieldAction) isCardAction() {}
func (d DivineShieldAction) Execute(ctx CardActionContext) ([]Event, error) {
	events, err := ctx.Engine.StopTime(ctx.Player)
	if err != nil {
		return nil, err
	}
	drawEvents, err := makePlayersDraw(ctx.Engine, ctx.Engine.ListPlayers(), 1)
	events = append(events, drawEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

type EventAction func()

func defeatDoorKindByAction(ctx CardActionContext, target DungeonCard, doorKind DoorKind) ([]Event, error) {
	if !ctx.Engine.HasActiveDoor(target) {
		return nil, fmt.Errorf("target not found")
	}
	door, ok := target.(*DoorCard)
	if !ok || door.Type != doorKind {
		return nil, fmt.Errorf("target is not a door of kind %v", doorKind)
	}

	return ctx.Engine.DefeatDoor(target)
}

func makePlayersDraw(gameEngine GameEngine, players []*Player, count int) ([]Event, error) {
	var events []Event
	for _, player := range players {
		drawEvents, err := gameEngine.DrawCards(player, count)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func smartCardTargeting(card PlayerCard, candidatesFunc func() []DungeonCard) (DungeonCard, error) {
	candidates := candidatesFunc()
	switch len(candidates) {
	case 0:
		return nil, fmt.Errorf("no valid door active")
	case 1:
		return candidates[0], nil
	default:
		return nil, &AmbiguousCardTargetError{
			Card:         card,
			ValidTargets: candidates,
		}
	}
}

func smartPlayerTargeting(card PlayerCard, candidatesFunc func() []*Player) (*Player, error) {
	candidates := candidatesFunc()
	switch len(candidates) {
	case 0:
		return nil, fmt.Errorf("no valid target player")
	case 1:
		return candidates[0], nil
	default:
		return nil, &AmbiguousPlayerTargetError{
			Card:         card,
			ValidTargets: candidates,
		}
	}
}
