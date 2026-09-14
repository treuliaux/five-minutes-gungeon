package game

import "fmt"

type AbilityParams interface {
	isAbilityParams()
}

type AbilityContext struct {
	game   *Game
	player *Player
}

// Ranger

type TrickShotParams struct {
	Target DungeonCard
}

func (TrickShotParams) isAbilityParams() {}

func applyTrickShot(ctx AbilityContext, params TrickShotParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Ranger) {
		return []Event{}, fmt.Errorf("player is not a Ranger")
	}
	if ctx.game.currentDungeonCard != params.Target {
		return []Event{}, fmt.Errorf("target not found")
	}
	door, ok := ctx.game.currentDungeonCard.(*DoorCard)
	if !ok || door.Type != DoorPerson {
		return []Event{}, fmt.Errorf("target is not a Person door")
	}

	events, err := ctx.game.defeatDoor(params.Target)
	if err != nil {
		return []Event{}, err
	}

	return events, nil
}

// Huntress

type AnimalCompanionParams struct {
	Target *Player
}

func (AnimalCompanionParams) isAbilityParams() {}

func applyAnimalCompanion(ctx AbilityContext, params AnimalCompanionParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Huntress) {
		return []Event{}, fmt.Errorf("player is not a Huntress")
	}

	return []Event{}, nil
}

// Valkyrie

type InspireParams struct {
}

func (InspireParams) isAbilityParams() {}

func applyInspire(ctx AbilityContext, _ InspireParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Valkyrie) {
		return []Event{}, fmt.Errorf("player is not a Valkyrie")
	}

	return []Event{}, nil
}

// Paladin

type SmiteParams struct {
	Target DungeonCard
}

func (SmiteParams) isAbilityParams() {}

func applySmite(ctx AbilityContext, params SmiteParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Paladin) {
		return []Event{}, fmt.Errorf("player is not a Paladin")
	}

	return []Event{}, nil
}

// Wizard

type StopTimeParams struct {
}

func (StopTimeParams) isAbilityParams() {}

func applyStopTime(ctx AbilityContext, _ StopTimeParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Wizard) {
		return []Event{}, fmt.Errorf("player is not a Wizard")
	}

	return ctx.game.stopTime(ctx.player), nil
}

// Sorceress

type TeleportParams struct {
	Target DungeonCard
}

func (TeleportParams) isAbilityParams() {}

func applyTeleport(ctx AbilityContext, params TeleportParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Sorceress) {
		return []Event{}, fmt.Errorf("player is not a Sorceress")
	}

	return []Event{}, nil
}

// Barbarian

type SlayParams struct {
	Target DungeonCard
}

func (SlayParams) isAbilityParams() {}

func applySlay(ctx AbilityContext, params SlayParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Barbarian) {
		return []Event{}, fmt.Errorf("player is not a Barbarian")
	}

	return []Event{}, nil
}

// Gladiator

type IntimidateParams struct {
	Target DungeonCard
}

func (IntimidateParams) isAbilityParams() {}

func applyIntimidate(ctx AbilityContext, params IntimidateParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Gladiator) {
		return []Event{}, fmt.Errorf("player is not a Gladiator")
	}

	return []Event{}, nil
}

// Ninja

type VaultParams struct {
	Target DungeonCard
}

func (VaultParams) isAbilityParams() {}

func applyVault(ctx AbilityContext, params VaultParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Ninja) {
		return []Event{}, fmt.Errorf("player is not a Ninja")
	}

	return []Event{}, nil
}

// Thief

type PickpocketParams struct {
}

func (PickpocketParams) isAbilityParams() {}

func applyPickpocket(ctx AbilityContext, _ PickpocketParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Thief) {
		return []Event{}, fmt.Errorf("player is not a Thief")
	}

	return []Event{}, nil
}

// Druid

type ForestSpiritsParams struct {
	Target DungeonCard
}

func (ForestSpiritsParams) isAbilityParams() {}

func applyForestSpirits(ctx AbilityContext, params ForestSpiritsParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Druid) {
		return []Event{}, fmt.Errorf("player is not a Druid")
	}

	return []Event{}, nil
}

// Shaman

type SpiritAnimalParams struct {
	Target *Player
}

func (SpiritAnimalParams) isAbilityParams() {}

func applySpiritAnimal(ctx AbilityContext, params SpiritAnimalParams) ([]Event, error) {
	if !checkPlayerClass(ctx.player, Shaman) {
		return []Event{}, fmt.Errorf("player is not a Shaman")
	}

	return []Event{}, nil
}

func checkPlayerClass(player *Player, class HeroClass) bool {
	return player.hero.class == class
}
