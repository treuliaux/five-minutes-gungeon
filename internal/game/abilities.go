package game

import "fmt"

type Ability interface {
	isAbility()
	Execute(ctx Context) ([]Event, error)
	Name() string
}

// Ranger

type TrickShotAbility struct{}

func (TrickShotAbility) isAbility() {}

func (a TrickShotAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Ranger) {
		return nil, fmt.Errorf("player is not a Ranger")
	}

	return defeatDoorByFilter(ctx.Engine(), aCtx.Player, aCtx.TargetCard, NewDoorsFilter().AddDoors(DoorPerson))
}
func (a TrickShotAbility) Name() string {
	return "Trick Shot"
}

// Huntress

type AnimalCompanionAbility struct{}

func (AnimalCompanionAbility) isAbility() {}

func (a AnimalCompanionAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Huntress) {
		return nil, fmt.Errorf("player is not a Huntress")
	}

	target := aCtx.TargetPlayer
	if target == nil {
		var err error
		target, err = smartTargeting(aCtx.Player, otherPlayers(ctx.Engine().ListPlayers(), aCtx.Player))
		if err != nil {
			return nil, err
		}
	}

	return target.DrawCardsFromDeck(4)
}
func (a AnimalCompanionAbility) Name() string {
	return "Animal Companion"
}

// Valkyrie

type InspireAbility struct {
}

func (InspireAbility) isAbility() {}

func (InspireAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Valkyrie) {
		return nil, fmt.Errorf("player is not a Valkyrie")
	}

	var events []Event
	for _, player := range otherPlayers(ctx.Engine().ListPlayers(), aCtx.Player) {
		drawnEvents, err := player.DrawCardsFromDeck(2)
		events = append(events, drawnEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (a InspireAbility) Name() string {
	return "Inspire"
}

// Paladin

type SmiteAbility struct{}

func (SmiteAbility) isAbility() {}

func (a SmiteAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Paladin) {
		return nil, fmt.Errorf("player is not a Paladin")
	}

	return defeatDoorByFilter(ctx.Engine(), aCtx.Player, aCtx.TargetCard, NewDoorsFilter().AddDoors(DoorMonster))
}
func (a SmiteAbility) Name() string {
	return "Smite"
}

// Wizard

type StopTimeAbility struct {
}

func (StopTimeAbility) isAbility() {}

func (StopTimeAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Wizard) {
		return nil, fmt.Errorf("player is not a Wizard")
	}

	return ctx.Engine().StopTime(aCtx.Player)
}
func (a StopTimeAbility) Name() string {
	return "Stop Time"
}

// Sorceress

type TeleportAbility struct{}

func (TeleportAbility) isAbility() {}

func (a TeleportAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Sorceress) {
		return nil, fmt.Errorf("player is not a Sorceress")
	}

	return defeatDoorByFilter(ctx.Engine(), aCtx.Player, aCtx.TargetCard, NewDoorsFilter().AddDoors(DoorObstacle))
}
func (a TeleportAbility) Name() string {
	return "Teleport"
}

// Barbarian

type SlayAbility struct{}

func (SlayAbility) isAbility() {}

func (a SlayAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Barbarian) {
		return nil, fmt.Errorf("player is not a Barbarian")
	}

	return defeatDoorByFilter(ctx.Engine(), aCtx.Player, aCtx.TargetCard, NewDoorsFilter().AddDoors(DoorMonster))
}
func (a SlayAbility) Name() string {
	return "Slay"
}

// Gladiator

type IntimidateAbility struct{}

func (IntimidateAbility) isAbility() {}

func (a IntimidateAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Gladiator) {
		return nil, fmt.Errorf("player is not a Gladiator")
	}

	return defeatDoorByFilter(ctx.Engine(), aCtx.Player, aCtx.TargetCard, NewDoorsFilter().AddDoors(DoorPerson))
}
func (a IntimidateAbility) Name() string {
	return "Intimidate"
}

// Ninja

type VaultAbility struct{}

func (VaultAbility) isAbility() {}

func (a VaultAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Ninja) {
		return nil, fmt.Errorf("player is not a Ninja")
	}

	return defeatDoorByFilter(ctx.Engine(), aCtx.Player, aCtx.TargetCard, NewDoorsFilter().AddDoors(DoorObstacle))
}
func (a VaultAbility) Name() string {
	return "Vault"
}

// Thief

type PickpocketAbility struct {
}

func (PickpocketAbility) isAbility() {}

func (PickpocketAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Thief) {
		return nil, fmt.Errorf("player is not a Thief")
	}

	return aCtx.Player.DrawCardsFromDeck(5)
}
func (a PickpocketAbility) Name() string {
	return "Pickpocket"
}

// Druid

type ForestSpiritsAbility struct{}

func (ForestSpiritsAbility) isAbility() {}

func (a ForestSpiritsAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Druid) {
		return nil, fmt.Errorf("player is not a Druid")
	}

	target := aCtx.TargetCard
	if target == nil {
		var err error
		target, err = smartTargeting(aCtx.Player, aCtx.Engine().ActiveDoors(NewDoorsFilter().AddCurses()))
		if err != nil {
			return nil, err
		}
	}

	if _, ok := target.(*CurseCard); !ok {
		return nil, fmt.Errorf("target is not a curse")
	}

	return ctx.Engine().SendDungeonCardBottomDungeon(target)
}
func (a ForestSpiritsAbility) Name() string {
	return "Forest Spirits"
}

// Shaman

type SpiritAnimalAbility struct{}

func (SpiritAnimalAbility) isAbility() {}

func (a SpiritAnimalAbility) Execute(ctx Context) ([]Event, error) {
	aCtx, ok := ctx.(*AbilityContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	if !checkPlayerClass(aCtx.Player, Shaman) {
		return nil, fmt.Errorf("player is not a Shaman")
	}

	target := aCtx.TargetPlayer
	if target == nil {
		var err error
		target, err = smartTargeting(aCtx.Player, otherPlayers(ctx.Engine().ListPlayers(), aCtx.Player))
		if err != nil {
			return nil, err
		}
	}
	if target == aCtx.Player {
		return nil, fmt.Errorf("player cannot self-target")
	}

	return target.Heal(3)
}
func (a SpiritAnimalAbility) Name() string {
	return "Spirit Animal"
}

func checkPlayerClass(player *Player, class HeroClass) bool {
	return player.Hero.Class == class
}
