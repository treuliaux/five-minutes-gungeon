package game

import (
	"fmt"
	"slices"
)

type CardAction interface {
	isCardAction()
	Execute(ctx Context) ([]Event, error)
}

type SnipeAction struct{}

func (SnipeAction) isCardAction() {}
func (a SnipeAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorPerson))
}

type WildCardAction struct {
}

func (WildCardAction) isCardAction() {}
func (a WildCardAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var events []Event
	var err error
	events, err = ctx.Engine().PlayArbitraryCard(actCtx.Player, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{WildCard}})
	if err != nil {
		return events, err
	}
	var removeEvents []Event
	removeEvents, err = ctx.Engine().RemovePlayerCard(actCtx.Card)
	events = append(events, removeEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

type HealingHerbsAction struct{}

func (HealingHerbsAction) isCardAction() {}
func (a HealingHerbsAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var target *Player
	if len(actCtx.TargetPlayers) != 0 {
		target = actCtx.TargetPlayers[0]
	}
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, otherPlayers(ctx.Engine().ListPlayers(), actCtx.Player))
		if err != nil {
			return nil, err
		}
	}
	if target.Discard == nil || target.Discard.Empty() {
		return nil, nil
	}
	count := min(4, target.Discard.Length())

	return target.DrawCardsFromDiscard(count)
}

type CriticalHitAction struct{}

func (CriticalHitAction) isCardAction() {}
func (a CriticalHitAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorMonster))
}

type ExtraQuiverAction struct {
}

func (ExtraQuiverAction) isCardAction() {}
func (a ExtraQuiverAction) Execute(ctx Context) ([]Event, error) {
	return makePlayersDrawFromDeck(ctx.Engine().ListPlayers(), 2)
}

// Valkyrie / Paladin

type SmiteAction struct{}

func (SmiteAction) isCardAction() {}
func (a SmiteAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorMonster))
}

type DivineShieldAction struct {
}

func (DivineShieldAction) isCardAction() {}
func (a DivineShieldAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	events, err := ctx.Engine().StopTime(actCtx.Player)
	if err != nil {
		return nil, err
	}
	drawEvents, err := makePlayersDrawFromDeck(ctx.Engine().ListPlayers(), 1)
	events = append(events, drawEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

type HolyHandGrenadeAction struct{}

func (HolyHandGrenadeAction) isCardAction() {}
func (a HolyHandGrenadeAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddEverything())
}

type HealAction struct{}

func (HealAction) isCardAction() {}
func (a HealAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var target *Player
	if len(actCtx.TargetPlayers) != 0 {
		target = actCtx.TargetPlayers[0]
	}
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, otherPlayers(ctx.Engine().ListPlayers(), actCtx.Player))
		if err != nil {
			return nil, err
		}
	}

	return target.Heal(target.Discard.Length())
}

type HealthPotionAction struct {
}

func (HealthPotionAction) isCardAction() {}
func (a HealthPotionAction) Execute(ctx Context) ([]Event, error) {
	return makePlayersDrawFromDiscard(ctx.Engine().ListPlayers(), 3)
}

type MysticRuneAction struct {
}

func (MysticRuneAction) isCardAction() {}
func (a MysticRuneAction) Execute(ctx Context) ([]Event, error) {
	var events []Event
	for _, player := range ctx.Engine().ListPlayers() {
		drawEvents, err := player.DrawCardsFromDeck(len(ctx.Engine().ActiveDoors(NewDoorsFilter().AddCurses())))
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
func (a RallyAction) Execute(ctx Context) ([]Event, error) {
	var events []Event
	for _, player := range ctx.Engine().ListPlayers() {
		drawEvents, err := player.DrawResourceCardsFromDiscard([]ResourceType{Sword, Shield})
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

type FireballAction struct{}

func (FireballAction) isCardAction() {}
func (a FireballAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorMonster))
}

type MagicBombAction struct {
}

func (MagicBombAction) isCardAction() {}
func (a MagicBombAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var events []Event
	var err error
	events, err = ctx.Engine().PlayArbitraryCard(actCtx.Player, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword, Shield, Arrow, Scroll, Jump}})
	if err != nil {
		return events, err
	}
	var removeEvents []Event
	removeEvents, err = ctx.Engine().RemovePlayerCard(actCtx.Card)
	events = append(events, removeEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

type CancelAction struct{}

func (CancelAction) isCardAction() {}
func (a CancelAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	target := actCtx.TargetCard
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, ctx.Engine().ActiveDoors(NewDoorsFilter().AddEvents()))
		if err != nil {
			return nil, err
		}
	}
	actionEvents, err := defeatDoorByFilter(ctx.Engine(), actCtx.Card, target, NewDoorsFilter().AddEvents())
	if err != nil {
		return actionEvents, err
	}

	return append([]Event{EventCounteredEvent{
		ByPlayerID: actCtx.Player.Id,
		CardID:     target.ID(),
		WithCardID: actCtx.Card.ID(),
	}}, actionEvents...), nil
}

type PortalAction struct{}

func (PortalAction) isCardAction() {}
func (a PortalAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	target := actCtx.TargetCard
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, ctx.Engine().ActiveDoors(NewDoorsFilter().AddAllDoors().AddMiniBoss()))
		if err != nil {
			return nil, err
		}
	}
	switch target.(type) {
	case *EventCard, *CurseCard, *BossMat:
		return nil, fmt.Errorf("curses, events, and boss mat cannot be targeted")
	default:
		return ctx.Engine().SendDungeonCardBottomDungeon(target)
	}
}

type TimeWarpAction struct {
}

func (TimeWarpAction) isCardAction() {}

func (TimeWarpAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return ctx.Engine().StopTime(actCtx.Player)
}

type MightyLeapAction struct{}

func (MightyLeapAction) isCardAction() {}
func (a MightyLeapAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorObstacle))
}

type EnrageAction struct{}

func (EnrageAction) isCardAction() {}
func (a EnrageAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	targets, err := smartTwoPlayersTargeting(actCtx, actCtx.TargetPlayers)
	if err != nil {
		return nil, err
	}

	return makePlayersDrawFromDeck(targets, 3)
}

type CrushAction struct{}

func (CrushAction) isCardAction() {}
func (a CrushAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddMiniBoss())
}

type BattleRageAction struct{}

func (BattleRageAction) isCardAction() {}
func (a BattleRageAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	target := actCtx.TargetCard
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, ctx.Engine().ActiveDoors(NewDoorsFilter().AddCurses()))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*CurseCard); !ok {
		return nil, fmt.Errorf("target is not a curse")
	}

	return ctx.Engine().SendDungeonCardBottomDungeon(target)
}

type BackstabAction struct{}

func (BackstabAction) isCardAction() {}
func (a BackstabAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorPerson))
}

type SprintAction struct{}

func (SprintAction) isCardAction() {}
func (a SprintAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorObstacle))
}

type StealAction struct{}

func (StealAction) isCardAction() {}
func (a StealAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var target *Player
	if len(actCtx.TargetPlayers) != 0 {
		target = actCtx.TargetPlayers[0]
	}
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, otherPlayers(ctx.Engine().ListPlayers(), actCtx.Player))
		if err != nil {
			return nil, err
		}
	}

	stolenCards := slices.Clone(target.Hand)
	actCtx.Player.Hand = append(actCtx.Player.Hand, target.Hand...)
	target.Hand = make([]PlayerCard, 0)

	return []Event{HandStolenEvent{FromPlayerID: target.Id, ToPlayerID: actCtx.Player.Id, CardIDs: pluckCardIDs(stolenCards)}}, nil
}

type DonateAction struct{}

func (DonateAction) isCardAction() {}
func (a DonateAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var target *Player
	if len(actCtx.TargetPlayers) != 0 {
		target = actCtx.TargetPlayers[0]
	}
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, otherPlayers(ctx.Engine().ListPlayers(), actCtx.Player))
		if err != nil {
			return nil, err
		}
	}

	transferredCards := slices.Clone(actCtx.Player.Hand)
	target.Hand = append(target.Hand, actCtx.Player.Hand...)
	actCtx.Player.Hand = make([]PlayerCard, 0)

	return []Event{HandDonatedEvent{FromPlayerID: actCtx.Player.Id, ToPlayerID: target.Id, CardIDs: pluckCardIDs(transferredCards)}}, nil
}

type ThrowingKnivesAction struct {
}

func (ThrowingKnivesAction) isCardAction() {}
func (a ThrowingKnivesAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	var events []Event
	var err error
	events, err = ctx.Engine().PlayArbitraryCard(actCtx.Player, &ResourceCard{Id: nextCardId(), Resources: []ResourceType{WildCard, WildCard, WildCard}})
	if err != nil {
		return events, err
	}
	var removeEvents []Event
	removeEvents, err = ctx.Engine().RemovePlayerCard(actCtx.Card)
	events = append(events, removeEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

type TameCreatureAction struct{}

func (TameCreatureAction) isCardAction() {}
func (a TameCreatureAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorMonster))
}

type TrueSightAction struct{}

func (TrueSightAction) isCardAction() {}
func (a TrueSightAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorObstacle))
}

type LivingVinesAction struct{}

func (LivingVinesAction) isCardAction() {}
func (a LivingVinesAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	return defeatDoorByFilter(ctx.Engine(), actCtx.Card, actCtx.TargetCard, NewDoorsFilter().AddDoors(DoorPerson))
}

type CleanseAction struct{}

func (CleanseAction) isCardAction() {}
func (a CleanseAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	target := actCtx.TargetCard
	if target == nil {
		var err error
		target, err = smartTargeting(actCtx.Card, ctx.Engine().ActiveDoors(NewDoorsFilter().AddCurses()))
		if err != nil {
			return nil, err
		}
	}
	targetCurse, ok := target.(*CurseCard)
	if !ok {
		return nil, fmt.Errorf("target is not a curse")
	}

	return ctx.Engine().RemoveDungeonCard(targetCurse)
}

type AncientHealingAction struct{}

func (AncientHealingAction) isCardAction() {}
func (a AncientHealingAction) Execute(ctx Context) ([]Event, error) {
	actCtx, ok := ctx.(*CardActionContext)
	if !ok {
		return nil, fmt.Errorf("invalid context provided")
	}

	targets, err := smartTwoPlayersTargeting(actCtx, actCtx.TargetPlayers)
	if err != nil {
		return nil, err
	}

	return makePlayersDrawFromDiscard(targets, 2)
}

func makePlayersDrawFromDeck(players []*Player, count int) ([]Event, error) {
	var events []Event
	for _, player := range players {
		toDraw := min(count, player.Deck.Length())
		drawEvents, err := player.DrawCardsFromDeck(toDraw)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func makePlayersDrawFromDiscard(players []*Player, count int) ([]Event, error) {
	var events []Event
	for _, player := range players {
		toDraw := min(count, player.Discard.Length())
		drawEvents, err := player.DrawCardsFromDiscard(toDraw)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
