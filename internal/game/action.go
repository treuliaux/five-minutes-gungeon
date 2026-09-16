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
func (s SnipeAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, s.Target, DoorPerson)
}

type HealingHerbsAction struct {
	Target *Player
}

func (HealingHerbsAction) isCardAction() {}
func (h HealingHerbsAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := h.Target
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
func (c CriticalHitAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, c.Target, DoorMonster)
}

type ExtraQuiverAction struct {
}

func (ExtraQuiverAction) isCardAction() {}
func (e ExtraQuiverAction) Execute(ctx CardActionContext) ([]Event, error) {
	return makePlayersDraw(ctx.Engine, ctx.Engine.ListPlayers(), 2)
}

// Valkyrie / Paladin

type SmiteAction struct {
	Target DungeonCard
}

func (SmiteAction) isCardAction() {}
func (s SmiteAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, s.Target, DoorMonster)
}

type DivineShieldAction struct {
}

func (DivineShieldAction) isCardAction() {}
func (d DivineShieldAction) Execute(ctx CardActionContext) ([]Event, error) {
	events, err := ctx.Engine.StopTime(ctx.Player)
	if err != nil {
		return nil, err
	}
	drawEvents, err := makePlayersDraw(ctx.Engine, ctx.Engine.ListPlayers(), 1)
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
func (h HolyHandGrenadeAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorByAction(ctx, h.Target)
}

type HealAction struct {
	Target *Player
}

func (HealAction) isCardAction() {}
func (h HealAction) Execute(ctx CardActionContext) ([]Event, error) {
	target := h.Target
	if target == nil {
		var err error
		target, err = smartPlayerTargeting(ctx.Card, ctx.Engine.ListOtherPlayers(ctx.Player))
		if err != nil {
			return nil, err
		}
	}

	return ctx.Engine.HealPlayer(target, 0)
}

type HealthPotionAction struct {
}

func (HealthPotionAction) isCardAction() {}
func (h HealthPotionAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	for _, player := range ctx.Engine.ListPlayers() {
		if player.Discard == nil || player.Discard.Empty() {
			continue
		}
		count := min(3, player.Discard.Length())
		drawFromDiscardEvents, err := ctx.Engine.DrawCardsFromDiscard(player, count)
		events = append(events, drawFromDiscardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

type MysticRuneAction struct {
}

func (MysticRuneAction) isCardAction() {}
func (m MysticRuneAction) Execute(ctx CardActionContext) ([]Event, error) {
	var events []Event
	for _, player := range ctx.Engine.ListPlayers() {
		drawEvents, err := ctx.Engine.DrawCardsFromDeck(player, 1) // TODO: Replace `1` by active curses count
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
func (r RallyAction) Execute(ctx CardActionContext) ([]Event, error) {
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
func (f FireballAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, f.Target, DoorMonster)
}

type CancelAction struct {
	Target DungeonCard
}

func (CancelAction) isCardAction() {}
func (c CancelAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatEventDoorByAction(ctx, c.Target)
}

type PortalAction struct {
	Target DungeonCard
}

func (PortalAction) isCardAction() {}
func (p PortalAction) Execute(ctx CardActionContext) ([]Event, error) {
	// TODO: Move an opened door at the bottom of dungeon pile (except Events and Curses)
	return nil, nil
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
func (m MightyLeapAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, m.Target, DoorObstacle)
}

type EnrageAction struct {
	Targets []*Player
}

func (EnrageAction) isCardAction() {}
func (m EnrageAction) Execute(ctx CardActionContext) ([]Event, error) {
	targets := m.Targets
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

	return makePlayersDraw(ctx.Engine, targets, 3)
}

type CrushAction struct {
	Target DungeonCard
}

func (CrushAction) isCardAction() {}
func (c CrushAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatMiniBossByAction(ctx, c.Target)
}

type BattleRageAction struct {
	Target DungeonCard
}

func (BattleRageAction) isCardAction() {}
func (b BattleRageAction) Execute(ctx CardActionContext) ([]Event, error) {
	// TODO: Move a curse door at the bottom of dungeon pile
	return nil, nil
}

type BackstabAction struct {
	Target DungeonCard
}

func (BackstabAction) isCardAction() {}
func (b BackstabAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, b.Target, DoorPerson)
}

type SprintAction struct {
	Target DungeonCard
}

func (SprintAction) isCardAction() {}
func (s SprintAction) Execute(ctx CardActionContext) ([]Event, error) {
	return defeatDoorKindByAction(ctx, s.Target, DoorObstacle)
}

type StealAction struct {
	Target DungeonCard
}

func (StealAction) isCardAction() {}
func (s StealAction) Execute(ctx CardActionContext) ([]Event, error) {
	// TODO: Steal another's player's whole hand
	return nil, nil
}

type DonateAction struct {
	Target DungeonCard
}

func (DonateAction) isCardAction() {}
func (d DonateAction) Execute(ctx CardActionContext) ([]Event, error) {
	// TODO: Donate whole hand to another player
	return nil, nil
}

type EventAction func()

func defeatDoorKindByAction(ctx CardActionContext, target DungeonCard, doorKind DoorKind) ([]Event, error) {
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoorsOfKind(doorKind))
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
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveDoors())
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*BossMat); ok {
		return nil, fmt.Errorf("boss mat cannot be targeted")
	}

	return ctx.Engine.DefeatDoor(target)
}

func defeatMiniBossByAction(ctx CardActionContext, target DungeonCard) ([]Event, error) {
	if target == nil {
		var err error
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveMiniBossDoors())
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
		target, err = smartDungeonCardTargeting(ctx.Card, ctx.Engine.GetActiveEventDoors())
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*BossMat); ok {
		return nil, fmt.Errorf("boss mat cannot be targeted")
	}

	return ctx.Engine.DefeatDoor(target)
}

func makePlayersDraw(gameEngine GameEngine, players []*Player, count int) ([]Event, error) {
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
