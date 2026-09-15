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

type TrickShot struct {
	Target DungeonCard
}

func (TrickShot) isAbility() {}

func (t TrickShot) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Ranger) {
		return nil, fmt.Errorf("player is not a Ranger")
	}

	return defeatDoorKindByAbility(ctx, t.Target, DoorPerson)
}

// Huntress

type AnimalCompanion struct {
	Target *Player
}

func (AnimalCompanion) isAbility() {}

func (a AnimalCompanion) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Huntress) {
		return nil, fmt.Errorf("player is not a Huntress")
	}

	if a.Target == nil {
		return nil, fmt.Errorf("target player cannot be nil")
	}

	return ctx.Engine.DrawCards(a.Target, 4)
}

// Valkyrie

type Inspire struct {
}

func (Inspire) isAbility() {}

func (Inspire) Execute(ctx AbilityContext) ([]Event, error) {
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

type Smite struct {
	Target DungeonCard
}

func (Smite) isAbility() {}

func (s Smite) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Paladin) {
		return nil, fmt.Errorf("player is not a Paladin")
	}

	return defeatDoorKindByAbility(ctx, s.Target, DoorMonster)
}

// Wizard

type StopTime struct {
}

func (StopTime) isAbility() {}

func (StopTime) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Wizard) {
		return nil, fmt.Errorf("player is not a Wizard")
	}

	return ctx.Engine.StopTime(ctx.Player), nil
}

// Sorceress

type Teleport struct {
	Target DungeonCard
}

func (Teleport) isAbility() {}

func (t Teleport) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Sorceress) {
		return nil, fmt.Errorf("player is not a Sorceress")
	}

	return defeatDoorKindByAbility(ctx, t.Target, DoorObstacle)
}

// Barbarian

type Slay struct {
	Target DungeonCard
}

func (Slay) isAbility() {}

func (s Slay) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Barbarian) {
		return nil, fmt.Errorf("player is not a Barbarian")
	}

	return defeatDoorKindByAbility(ctx, s.Target, DoorMonster)
}

// Gladiator

type Intimidate struct {
	Target DungeonCard
}

func (Intimidate) isAbility() {}

func (i Intimidate) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Gladiator) {
		return nil, fmt.Errorf("player is not a Gladiator")
	}

	return defeatDoorKindByAbility(ctx, i.Target, DoorPerson)
}

// Ninja

type Vault struct {
	Target DungeonCard
}

func (Vault) isAbility() {}

func (v Vault) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Ninja) {
		return nil, fmt.Errorf("player is not a Ninja")
	}

	return defeatDoorKindByAbility(ctx, v.Target, DoorObstacle)
}

// Thief

type Pickpocket struct {
}

func (Pickpocket) isAbility() {}

func (Pickpocket) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Thief) {
		return nil, fmt.Errorf("player is not a Thief")
	}

	return ctx.Engine.DrawCards(ctx.Player, 5)
}

// Druid

type ForestSpirits struct {
	Target DungeonCard
}

func (ForestSpirits) isAbility() {}

func (ForestSpirits) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Druid) {
		return nil, fmt.Errorf("player is not a Druid")
	}

	// TODO: move curse to the bottom of dungeon doors stack

	return nil, nil
}

// Shaman

type SpiritAnimal struct {
	Target *Player
}

func (SpiritAnimal) isAbility() {}

func (s SpiritAnimal) Execute(ctx AbilityContext) ([]Event, error) {
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
