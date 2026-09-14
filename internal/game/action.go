package game

import "fmt"

type CardAction interface {
	isCardAction()
	execute(ctx CardActionContext) ([]Event, error)
}

type CardActionContext struct {
	engine GameEngine
	player *Player
}

// Ranger

type Snipe struct {
	Target DungeonCard
}

func (Snipe) isCardAction() {}

func (s Snipe) execute(ctx CardActionContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Ranger) {
		return []Event{}, fmt.Errorf("player is not a Ranger")
	}

	return defeatDoorKindByAction(ctx, s.Target, DoorPerson)
}

type EventAction func()

func defeatDoorKindByAction(ctx CardActionContext, target DungeonCard, doorKind DoorKind) ([]Event, error) {
	if !ctx.engine.hasActiveDoor(target) {
		return []Event{}, fmt.Errorf("target not found")
	}
	door, ok := target.(*DoorCard)
	if !ok || door.Type != doorKind {
		return []Event{}, fmt.Errorf("target is not a door of kind %v", doorKind)
	}

	events, err := ctx.engine.defeatDoor(target)
	if err != nil {
		return []Event{}, err
	}

	return events, nil
}
