package game

import (
	"fmt"
	"slices"
)

type ArtifactActionIndex int

const (
	FirstArtifactAction ArtifactActionIndex = iota
	SecondArtifactAction
)

type ArtifactAction interface {
	isArtifactAction()
	Execute(ctx ArtifactActionContext) ([]Event, error)
}

type BattleAxeArtifact struct {
}

func (BattleAxeArtifact) isArtifactAction() {}
func (a BattleAxeArtifact) Execute(ctx ArtifactActionContext) ([]Event, error) {
	var events []Event
	switch ctx.ChosenAction {
	case FirstArtifactAction:
		return defeatDoorByFilter(ctx.Engine(), ctx.Artifact, ctx.Target, NewDoorsFilter().AddDoors(DoorMonster))
	case SecondArtifactAction:
		for _, p := range ctx.Engine().ListPlayers() {
			drawEvents, err := p.DrawCardsFromDeck(2)
			events = append(events, drawEvents...)
			if err != nil {
				return events, err
			}
		}
	}

	return events, nil
}

type TheInfinityScrollArtifact struct {
}

func (TheInfinityScrollArtifact) isArtifactAction() {}
func (a TheInfinityScrollArtifact) Execute(ctx ArtifactActionContext) ([]Event, error) {
	switch ctx.ChosenAction {
	case FirstArtifactAction:
		return defeatDoorByFilter(ctx.Engine(), ctx.Artifact, ctx.Target, NewDoorsFilter().AddMiniBoss())
	case SecondArtifactAction:
		if ctx.Target == nil {
			var err error
			ctx.Target, err = smartTargeting(ctx.Artifact, ctx.Engine().ActiveDoors(NewDoorsFilter().AddEvents()))
			if err != nil {
				return nil, err
			}
		}
		actionEvents, err := defeatDoorByFilter(ctx.Engine(), ctx.Artifact, ctx.Target, NewDoorsFilter().AddEvents())
		if err != nil {
			return actionEvents, err
		}

		return append([]Event{EventCounteredEvent{
			ByPlayer:  ctx.Player,
			EventCard: ctx.Target,
			WithCard:  ctx.Artifact,
		}}, actionEvents...), nil
	default:
		return nil, fmt.Errorf("unexpected artifact action index value: %v", ctx.ChosenAction)
	}
}

type RainbowHerbsArtifact struct {
}

func (RainbowHerbsArtifact) isArtifactAction() {}
func (a RainbowHerbsArtifact) Execute(ctx ArtifactActionContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine().ListPlayers() {
		healEvents, err := p.ArtifactHeal()
		events = append(events, healEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

type MJhoilnorArtifact struct {
}

func (MJhoilnorArtifact) isArtifactAction() {}
func (a MJhoilnorArtifact) Execute(ctx ArtifactActionContext) ([]Event, error) {
	var events []Event
	for range 2 {
		discardEvents, err := ctx.Engine().DiscardTopCardFromDungeon()
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

type SundialWatchArtifact struct {
}

func (SundialWatchArtifact) isArtifactAction() {}
func (a SundialWatchArtifact) Execute(ctx ArtifactActionContext) ([]Event, error) {
	return ctx.Engine().StopTime(ctx.Player)
}

type CurseZapperArtifact struct {
}

func (CurseZapperArtifact) isArtifactAction() {}
func (a CurseZapperArtifact) Execute(ctx ArtifactActionContext) ([]Event, error) {
	var events []Event
	for _, curse := range slices.Clone(ctx.Engine().ActiveDoors(NewDoorsFilter().AddCurses())) {
		cureEvents, err := ctx.Engine().RemoveDungeonCard(curse)
		events = append(events, cureEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
