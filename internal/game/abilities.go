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

func (a TrickShotAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Ranger) {
		return nil, fmt.Errorf("player is not a Ranger")
	}

	return defeatDoorByFilter(ctx.Engine, ctx.Player, a.Target, NewDoorsFilter().AddDoors(DoorPerson))
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

	target := a.Target
	if target == nil {
		var err error
		target, err = smartTargeting(nil, otherPlayers(ctx.Engine.ListPlayers(), ctx.Player))
		if err != nil {
			return nil, err
		}
	}

	return target.DrawCardsFromDeck(4)
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
	for _, player := range otherPlayers(ctx.Engine.ListPlayers(), ctx.Player) {
		drawnEvents, err := player.DrawCardsFromDeck(2)
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

func (a SmiteAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Paladin) {
		return nil, fmt.Errorf("player is not a Paladin")
	}

	return defeatDoorByFilter(ctx.Engine, ctx.Player, a.Target, NewDoorsFilter().AddDoors(DoorMonster))
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

func (a TeleportAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Sorceress) {
		return nil, fmt.Errorf("player is not a Sorceress")
	}

	return defeatDoorByFilter(ctx.Engine, ctx.Player, a.Target, NewDoorsFilter().AddDoors(DoorObstacle))
}

// Barbarian

type SlayAbility struct {
	Target DungeonCard
}

func (SlayAbility) isAbility() {}

func (a SlayAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Barbarian) {
		return nil, fmt.Errorf("player is not a Barbarian")
	}

	return defeatDoorByFilter(ctx.Engine, ctx.Player, a.Target, NewDoorsFilter().AddDoors(DoorMonster))
}

// Gladiator

type IntimidateAbility struct {
	Target DungeonCard
}

func (IntimidateAbility) isAbility() {}

func (a IntimidateAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Gladiator) {
		return nil, fmt.Errorf("player is not a Gladiator")
	}

	return defeatDoorByFilter(ctx.Engine, ctx.Player, a.Target, NewDoorsFilter().AddDoors(DoorPerson))
}

// Ninja

type VaultAbility struct {
	Target DungeonCard
}

func (VaultAbility) isAbility() {}

func (a VaultAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Ninja) {
		return nil, fmt.Errorf("player is not a Ninja")
	}

	return defeatDoorByFilter(ctx.Engine, ctx.Player, a.Target, NewDoorsFilter().AddDoors(DoorObstacle))
}

// Thief

type PickpocketAbility struct {
}

func (PickpocketAbility) isAbility() {}

func (PickpocketAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Thief) {
		return nil, fmt.Errorf("player is not a Thief")
	}

	return ctx.Player.DrawCardsFromDeck(5)
}

// Druid

type ForestSpiritsAbility struct {
	Target DungeonCard
}

func (ForestSpiritsAbility) isAbility() {}

func (a ForestSpiritsAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Druid) {
		return nil, fmt.Errorf("player is not a Druid")
	}

	target := a.Target
	if target == nil {
		var err error
		target, err = smartTargeting(nil, ctx.Engine.ActiveDoors(NewDoorsFilter().AddCurses()))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*CurseCard); !ok {
		return nil, fmt.Errorf("target is not a curse")
	}

	return ctx.Engine.SendDungeonCardBottomDungeon(target)
}

// Shaman

type SpiritAnimalAbility struct {
	Target *Player
}

func (SpiritAnimalAbility) isAbility() {}

func (a SpiritAnimalAbility) Execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.Player, Shaman) {
		return nil, fmt.Errorf("player is not a Shaman")
	}

	target := a.Target
	if target == nil {
		var err error
		target, err = smartTargeting(nil, otherPlayers(ctx.Engine.ListPlayers(), ctx.Player))
		if err != nil {
			return nil, err
		}
	}

	return target.Heal(3)
}

func checkPlayerClass(player *Player, class HeroClass) bool {
	return player.Hero.Class == class
}
