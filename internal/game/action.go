package game

import "fmt"

type CardAction interface {
	isCardAction()
	Execute(ctx CardActionContext) ([]Event, error)
}

type CardActionContext struct {
	Engine GameEngine
	Player *Player
}

// Ranger

type Snipe struct {
	Target DungeonCard
}

func (Snipe) isCardAction() {}

func (s Snipe) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, s.Target, DoorPerson)
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
