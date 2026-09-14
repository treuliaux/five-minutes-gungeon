package game

import "fmt"

type Ability interface {
	isAbility()
	execute(ctx AbilityContext) ([]Event, error)
}

type AbilityContext struct {
	engine GameEngine
	player *Player
}

// Ranger

type TrickShot struct {
	Target DungeonCard
}

func (TrickShot) isAbility() {}

func (p TrickShot) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Ranger) {
		return []Event{}, fmt.Errorf("player is not a Ranger")
	}

	return defeatDoorKindByAbility(ctx, p.Target, DoorPerson)
}

// Huntress

type AnimalCompanion struct {
	Target *Player
}

func (AnimalCompanion) isAbility() {}

func (p AnimalCompanion) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Huntress) {
		return []Event{}, fmt.Errorf("player is not a Huntress")
	}

	if p.Target == nil {
		return []Event{}, fmt.Errorf("target player cannot be nil")
	}

	events, err := ctx.engine.drawCards(p.Target, 4)
	if err != nil {
		return []Event{}, err
	}

	return events, nil
}

// Valkyrie

type Inspire struct {
}

func (Inspire) isAbility() {}

func (p Inspire) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Valkyrie) {
		return []Event{}, fmt.Errorf("player is not a Valkyrie")
	}

	var cardDrawnEvents []Event
	for _, player := range ctx.engine.listOtherPlayers(ctx.player) {
		events, err := ctx.engine.drawCards(player, 2)
		cardDrawnEvents = append(cardDrawnEvents, events...)
		if err != nil {
			return cardDrawnEvents, err
		}
	}

	return cardDrawnEvents, nil
}

// Paladin

type Smite struct {
	Target DungeonCard
}

func (Smite) isAbility() {}

func (p Smite) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Paladin) {
		return []Event{}, fmt.Errorf("player is not a Paladin")
	}

	return defeatDoorKindByAbility(ctx, p.Target, DoorMonster)
}

// Wizard

type StopTime struct {
}

func (StopTime) isAbility() {}

func (p StopTime) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Wizard) {
		return []Event{}, fmt.Errorf("player is not a Wizard")
	}

	return ctx.engine.stopTime(ctx.player), nil
}

// Sorceress

type Teleport struct {
	Target DungeonCard
}

func (Teleport) isAbility() {}

func (p Teleport) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Sorceress) {
		return []Event{}, fmt.Errorf("player is not a Sorceress")
	}

	return defeatDoorKindByAbility(ctx, p.Target, DoorObstacle)
}

// Barbarian

type Slay struct {
	Target DungeonCard
}

func (Slay) isAbility() {}

func (p Slay) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Barbarian) {
		return []Event{}, fmt.Errorf("player is not a Barbarian")
	}

	return defeatDoorKindByAbility(ctx, p.Target, DoorMonster)
}

// Gladiator

type Intimidate struct {
	Target DungeonCard
}

func (Intimidate) isAbility() {}

func (p Intimidate) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Gladiator) {
		return []Event{}, fmt.Errorf("player is not a Gladiator")
	}

	return defeatDoorKindByAbility(ctx, p.Target, DoorPerson)
}

// Ninja

type Vault struct {
	Target DungeonCard
}

func (Vault) isAbility() {}

func (p Vault) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Ninja) {
		return []Event{}, fmt.Errorf("player is not a Ninja")
	}

	return defeatDoorKindByAbility(ctx, p.Target, DoorObstacle)
}

// Thief

type Pickpocket struct {
}

func (Pickpocket) isAbility() {}

func (p Pickpocket) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Thief) {
		return []Event{}, fmt.Errorf("player is not a Thief")
	}

	events, err := ctx.engine.drawCards(ctx.player, 5)
	if err != nil {
		return []Event{}, err
	}

	return events, nil
}

// Druid

type ForestSpirits struct {
	Target DungeonCard
}

func (ForestSpirits) isAbility() {}

func (p ForestSpirits) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Druid) {
		return []Event{}, fmt.Errorf("player is not a Druid")
	}

	// TODO: move curse to the bottom of dungeon doors stack

	return []Event{}, nil
}

// Shaman

type SpiritAnimal struct {
	Target *Player
}

func (SpiritAnimal) isAbility() {}

func (p SpiritAnimal) execute(ctx AbilityContext) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Shaman) {
		return []Event{}, fmt.Errorf("player is not a Shaman")
	}

	if p.Target == nil {
		return []Event{}, fmt.Errorf("target player cannot be nil")
	}

	return ctx.engine.healPlayer(p.Target, 3)
}

func checkPlayerClass(player *Player, class HeroClass) bool {
	return player.hero.class == class
}

func defeatDoorKindByAbility(ctx AbilityContext, target DungeonCard, doorKind DoorKind) ([]Event, error) {
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
