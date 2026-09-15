package game

import "fmt"

type Ability interface {
	isAbility()
	Execute(ctx AbilityContext) ([]Event, error)
}

type AbilityContext struct {
	Engine GameEngine
	Player *Player
}

// Ranger

type TrickShotAbility struct {
	Target DungeonCard
}

func (TrickShotAbility) isAbility() {}

func (t TrickShotAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Ranger) {
		return nil, fmt.Errorf("player is not a Ranger")
	}

	return defeatDoorKindByAbility(ctx, t.Target, DoorPerson)
}

// Huntress

type AnimalCompanionAbility struct {
	Target *Player
}

func (AnimalCompanionAbility) isAbility() {}

func (a AnimalCompanionAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Huntress) {
		return nil, fmt.Errorf("player is not a Huntress")
	}

	if a.Target == nil {
		return nil, fmt.Errorf("target player cannot be nil")
	}

	return ctx.Engine.DrawCards(a.Target, 4)
}

// Valkyrie

type InspireAbility struct {
}

func (InspireAbility) isAbility() {}

func (InspireAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Valkyrie) {
		return nil, fmt.Errorf("player is not a Valkyrie")
	}

	var events []Event
	for _, player := range ctx.Engine.ListOtherPlayers(ctx.Player) {
		drawnEvents, err := ctx.Engine.DrawCards(player, 2)
		events = append(events, drawnEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

// Paladin

type SmiteAbility struct {
	Target DungeonCard
}

func (SmiteAbility) isAbility() {}

func (s SmiteAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Paladin) {
		return nil, fmt.Errorf("player is not a Paladin")
	}

	return defeatDoorKindByAbility(ctx, s.Target, DoorMonster)
}

// Wizard

type StopTimeAbility struct {
}

func (StopTimeAbility) isAbility() {}

func (StopTimeAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Wizard) {
		return nil, fmt.Errorf("player is not a Wizard")
	}

	return ctx.Engine.StopTime(ctx.Player)
}

// Sorceress

type TeleportAbility struct {
	Target DungeonCard
}

func (TeleportAbility) isAbility() {}

func (t TeleportAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Sorceress) {
		return nil, fmt.Errorf("player is not a Sorceress")
	}

	return defeatDoorKindByAbility(ctx, t.Target, DoorObstacle)
}

// Barbarian

type SlayAbility struct {
	Target DungeonCard
}

func (SlayAbility) isAbility() {}

func (s SlayAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Barbarian) {
		return nil, fmt.Errorf("player is not a Barbarian")
	}

	return defeatDoorKindByAbility(ctx, s.Target, DoorMonster)
}

// Gladiator

type IntimidateAbility struct {
	Target DungeonCard
}

func (IntimidateAbility) isAbility() {}

func (i IntimidateAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Gladiator) {
		return nil, fmt.Errorf("player is not a Gladiator")
	}

	return defeatDoorKindByAbility(ctx, i.Target, DoorPerson)
}

// Ninja

type VaultAbility struct {
	Target DungeonCard
}

func (VaultAbility) isAbility() {}

func (v VaultAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Ninja) {
		return nil, fmt.Errorf("player is not a Ninja")
	}

	return defeatDoorKindByAbility(ctx, v.Target, DoorObstacle)
}

// Thief

type PickpocketAbility struct {
}

func (PickpocketAbility) isAbility() {}

func (PickpocketAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Thief) {
		return nil, fmt.Errorf("player is not a Thief")
	}

	return ctx.Engine.DrawCards(ctx.Player, 5)
}

// Druid

type ForestSpiritsAbility struct {
	Target DungeonCard
}

func (ForestSpiritsAbility) isAbility() {}

func (ForestSpiritsAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Druid) {
		return nil, fmt.Errorf("player is not a Druid")
	}

	// TODO: move curse to the bottom of dungeon doors stack

	return nil, nil
}

// Shaman

type SpiritAnimalAbility struct {
	Target *Player
}

func (SpiritAnimalAbility) isAbility() {}

func (s SpiritAnimalAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Shaman) {
		return nil, fmt.Errorf("player is not a Shaman")
	}

	if s.Target == nil {
		return nil, fmt.Errorf("target player cannot be nil")
	}

	return ctx.Engine.HealPlayer(s.Target, 3)
}

func checkPlayerClass(player *Player, class HeroClass) bool {
	return player.Hero.Class == class
}

func defeatDoorKindByAbility(ctx AbilityContext, target DungeonCard, doorKind DoorKind) ([]Event, error) {
	if !ctx.Engine.HasActiveDoor(target) {
		return nil, fmt.Errorf("target not found")
	}
	door, ok := target.(*DoorCard)
	if !ok || door.Type != doorKind {
		return nil, fmt.Errorf("target is not a door of kind %v", doorKind)
	}

	return ctx.Engine.DefeatDoor(target)
}
