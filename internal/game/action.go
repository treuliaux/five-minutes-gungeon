package game

import (
	"fmt"
)

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

type SnipeAction struct {
	Target DungeonCard
}

func (SnipeAction) isCardAction() {}
func (a SnipeAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorPerson)
}

type WildCardAction struct {
}

func (WildCardAction) isCardAction() {}
func (a WildCardAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	var err error
	events, err = ctx.Engine.PlayArbitraryCard(ctx.Player, &ResourceCard{Resources: []ResourceType{WildCard}})
	if err != nil {
		return events, err
	}
	var removeEvents []Event
	removeEvents, err = ctx.Engine.RemovePlayerCardFromPlayfield(ctx.Card)
	events = append(events, removeEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

type HealingHerbsAction struct {
	Target *Player
}

func (HealingHerbsAction) isCardAction() {}
func (a HealingHerbsAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := a.Target
	if target == nil {
		var err error
		target, err = smartPlayerTargeting(ctx.Card, ctx.Engine.ListOtherPlayers(ctx.Player))
		if err != nil {
			return nil, err
		}
	}
	if target.Discard == nil || target.Discard.Empty() {
		return nil, nil
	}
	count := min(4, target.Discard.Length())

	return ctx.Engine.DrawCardsFromDiscard(target, count)
}

type CriticalHitAction struct {
	Target DungeonCard
}

func (CriticalHitAction) isCardAction() {}
func (a CriticalHitAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorMonster)
}

type ExtraQuiverAction struct {
}

func (ExtraQuiverAction) isCardAction() {}
func (a ExtraQuiverAction) Execute(ctx CardActionContext) ([]Event, error) {
	return makePlayersDrawFromDeck(ctx.Engine, ctx.Engine.ListPlayers(), 2)
}

// Valkyrie / Paladin

type SmiteAction struct {
	Target DungeonCard
}

func (SmiteAction) isCardAction() {}
func (a SmiteAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorMonster)
}

type DivineShieldAction struct {
}

func (DivineShieldAction) isCardAction() {}
func (a DivineShieldAction) Execute(ctx CardActionContext) ([]Event, error) {
	events, err := ctx.Engine.StopTime(ctx.Player)
	if err != nil {
		return nil, err
	}
	drawEvents, err := makePlayersDrawFromDeck(ctx.Engine, ctx.Engine.ListPlayers(), 1)
	events = append(events, drawEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

type HolyHandGrenadeAction struct {
	Target DungeonCard
}

func (HolyHandGrenadeAction) isCardAction() {}
func (a HolyHandGrenadeAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorByAction(ctx, a.Target)
}

type HealAction struct {
	Target *Player
}

func (HealAction) isCardAction() {}
func (a HealAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := a.Target
	if target == nil {
		var err error
		target, err = smartPlayerTargeting(ctx.Card, ctx.Engine.ListOtherPlayers(ctx.Player))
		if err != nil {
			return nil, err
		}
	}

	return ctx.Engine.HealPlayer(target, target.Discard.Length())
}

type HealthPotionAction struct {
}

func (HealthPotionAction) isCardAction() {}
func (a HealthPotionAction) Execute(ctx CardActionContext) ([]Event, error) {
	return makePlayersDrawFromDiscard(ctx.Engine, ctx.Engine.ListPlayers(), 3)
}

type MysticRuneAction struct {
}

func (MysticRuneAction) isCardAction() {}
func (a MysticRuneAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	for _, player := range ctx.Engine.ListPlayers() {
		drawEvents, err := ctx.Engine.DrawCardsFromDeck(player, len(ctx.Engine.GetActiveCurses()))
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

type RallyAction struct {
}

func (RallyAction) isCardAction() {}
func (a RallyAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	for _, player := range ctx.Engine.ListPlayers() {
		drawEvents, err := ctx.Engine.DrawResourceCardsFromDiscard(player, []ResourceType{Sword, Shield})
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

type FireballAction struct {
	Target DungeonCard
}

func (FireballAction) isCardAction() {}
func (a FireballAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorMonster)
}

type MagicBombAction struct {
}

func (MagicBombAction) isCardAction() {}
func (a MagicBombAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	var err error
	events, err = ctx.Engine.PlayArbitraryCard(ctx.Player, &ResourceCard{Resources: []ResourceType{Sword, Shield, Arrow, Scroll, Jump}})
	if err != nil {
		return events, err
	}
	var removeEvents []Event
	removeEvents, err = ctx.Engine.RemovePlayerCardFromPlayfield(ctx.Card)
	events = append(events, removeEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

type CancelAction struct {
	Target DungeonCard
}

func (CancelAction) isCardAction() {}
func (a CancelAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := a.Target
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{},
			[]ChallengeKind{ChallengeEvent},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	actionEvents, err := defeatEventDoorByAction(ctx, target)
	if err != nil {
		return actionEvents, err
	}

	return append([]Event{EventCounteredEvent{
		ByPlayer:  ctx.Player,
		EventCard: target,
		WithCard:  ctx.Card,
	}}, actionEvents...), nil
}

type PortalAction struct {
	Target DungeonCard
}

func (PortalAction) isCardAction() {}
func (a PortalAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := a.Target
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{DoorPerson, DoorObstacle, DoorMonster},
			[]ChallengeKind{ChallengeMiniBoss},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	switch target.(type) {
	case *EventCard, *CurseCard, *BossMat:
		return nil, fmt.Errorf("curses, events, and boss mat cannot be targeted")
	default:
		return ctx.Engine.SendDungeonCardBottomDungeon(target)
	}
}

type TimeWarpAction struct {
}

func (TimeWarpAction) isCardAction() {}

func (TimeWarpAction) Execute(ctx CardActionContext) ([]Event, error) {
	return ctx.Engine.StopTime(ctx.Player)
}

type MightyLeapAction struct {
	Target DungeonCard
}

func (MightyLeapAction) isCardAction() {}
func (a MightyLeapAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorObstacle)
}

type EnrageAction struct {
	Targets []*Player
}

func (EnrageAction) isCardAction() {}
func (a EnrageAction) Execute(ctx CardActionContext) ([]Event, error) {
	targets, err := smartTwoPlayersTargeting(ctx, a.Targets)
	if err != nil {
		return nil, err
	}

	return makePlayersDrawFromDeck(ctx.Engine, targets, 3)
}

type CrushAction struct {
	Target DungeonCard
}

func (CrushAction) isCardAction() {}
func (a CrushAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatMiniBossByAction(ctx, a.Target)
}

type BattleRageAction struct {
	Target DungeonCard
}

func (BattleRageAction) isCardAction() {}
func (a BattleRageAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := a.Target
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{},
			[]ChallengeKind{ChallengeCurse},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*CurseCard); !ok {
		return nil, fmt.Errorf("target is not a curse")
	}

	return ctx.Engine.SendDungeonCardBottomDungeon(target)
}

type BackstabAction struct {
	Target DungeonCard
}

func (BackstabAction) isCardAction() {}
func (a BackstabAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorPerson)
}

type SprintAction struct {
	Target DungeonCard
}

func (SprintAction) isCardAction() {}
func (a SprintAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorObstacle)
}

type StealAction struct {
	Target *Player
}

func (StealAction) isCardAction() {}
func (a StealAction) Execute(ctx CardActionContext) ([]Event, error) {
	// TODO: Steal another's player's whole hand
	return nil, nil
}

type DonateAction struct {
	Target *Player
}

func (DonateAction) isCardAction() {}
func (a DonateAction) Execute(ctx CardActionContext) ([]Event, error) {
	// TODO: Donate whole hand to another player
	return nil, nil
}

type ThrowingKnivesAction struct {
}

func (ThrowingKnivesAction) isCardAction() {}
func (a ThrowingKnivesAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	var err error
	events, err = ctx.Engine.PlayArbitraryCard(ctx.Player, &ResourceCard{Resources: []ResourceType{WildCard, WildCard, WildCard}})
	if err != nil {
		return events, err
	}
	var removeEvents []Event
	removeEvents, err = ctx.Engine.RemovePlayerCardFromPlayfield(ctx.Card)
	events = append(events, removeEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

type TameCreatureAction struct {
	Target DungeonCard
}

func (TameCreatureAction) isCardAction() {}
func (a TameCreatureAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorMonster)
}

type TrueSightAction struct {
	Target DungeonCard
}

func (TrueSightAction) isCardAction() {}
func (a TrueSightAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorObstacle)
}

type LivingVinesAction struct {
	Target DungeonCard
}

func (LivingVinesAction) isCardAction() {}
func (a LivingVinesAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, a.Target, DoorPerson)
}

type CleanseAction struct {
	Target DungeonCard
}

func (CleanseAction) isCardAction() {}
func (a CleanseAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := a.Target
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{},
			[]ChallengeKind{ChallengeCurse},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	targetCurse, ok := target.(*CurseCard)
	if !ok {
		return nil, fmt.Errorf("target is not a curse")
	}

	return ctx.Engine.RemoveCurseFromPlayfield(targetCurse)
}

type AncientHealingAction struct {
	Targets []*Player
}

func (AncientHealingAction) isCardAction() {}
func (a AncientHealingAction) Execute(ctx CardActionContext) ([]Event, error) {
	targets, err := smartTwoPlayersTargeting(ctx, a.Targets)
	if err != nil {
		return nil, err
	}

	return makePlayersDrawFromDiscard(ctx.Engine, targets, 2)
}

type EventAction interface {
	isCardEvent()
	Execute(ctx CardEventContext) ([]Event, error)
}

type CardEventContext struct {
	Engine GameEngine
	Card   DungeonCard
}

func defeatDoorKindByAction(ctx CardActionContext, target DungeonCard, doorKind DoorKind) ([]Event, error) {
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{doorKind},
			[]ChallengeKind{},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*BossMat); ok {
		return nil, fmt.Errorf("boss mat cannot be targeted")
	}

	return ctx.Engine.DefeatDoor(target)
}

func defeatDoorByAction(ctx CardActionContext, target DungeonCard) ([]Event, error) {
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{DoorMonster, DoorObstacle, DoorPerson},
			[]ChallengeKind{ChallengeCurse, ChallengeMiniBoss, ChallengeEvent},
			true,
		))
		if err != nil {
			return nil, err
		}
	}

	return ctx.Engine.DefeatDoor(target)
}

func defeatMiniBossByAction(ctx CardActionContext, target DungeonCard) ([]Event, error) {
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{},
			[]ChallengeKind{ChallengeMiniBoss},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*BossMat); ok {
		return nil, fmt.Errorf("boss mat cannot be targeted")
	}

	return ctx.Engine.DefeatDoor(target)
}

func defeatEventDoorByAction(ctx CardActionContext, target DungeonCard) ([]Event, error) {
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfType(
			[]DoorKind{},
			[]ChallengeKind{ChallengeEvent},
			false,
		))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*BossMat); ok {
		return nil, fmt.Errorf("boss mat cannot be targeted")
	}

	return ctx.Engine.DefeatDoor(target)
}

func makePlayersDrawFromDeck(gameEngine GameEngine, players []*Player, count int) ([]Event, error) {
	var events []Event
	for _, player := range players {
		drawEvents, err := gameEngine.DrawCardsFromDeck(player, count)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func makePlayersDrawFromDiscard(gameEngine GameEngine, players []*Player, count int) ([]Event, error) {
	var events []Event
	for _, player := range players {
		if player.Discard == nil || player.Discard.Empty() {
			continue
		}
		toDraw := min(count, player.Discard.Length())
		drawEvents, err := gameEngine.DrawCardsFromDiscard(player, toDraw)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func smartDungeonCardTargeting(card PlayerCard, candidates []DungeonCard) (DungeonCard, error) {
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

func smartPlayerTargeting(card PlayerCard, candidates []*Player) (*Player, error) {
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

func smartTwoPlayersTargeting(ctx CardActionContext, targets []*Player) ([]*Player, error) {
	if len(targets) > 2 {
		return nil, fmt.Errorf("cannot target more than 2 players")
	}
	if len(targets) == 0 {
		candidates := ctx.Engine.ListPlayers()
		switch len(candidates) {
		case 0:
			return nil, fmt.Errorf("no valid target player")
		case 1, 2:
			targets = candidates
		default:
			return nil, &AmbiguousPlayerTargetError{
				Card:         ctx.Card,
				ValidTargets: candidates,
			}
		}
	}
	return targets, nil
}
