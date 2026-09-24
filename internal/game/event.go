package game

import (
	"fmt"
	"slices"
)

type EventAction interface {
	isCardEvent()
	Execute(ctx CardEventContext) ([]Event, error)
	Interaction(ctx CardEventContext) PendingInteraction
}

type CardEventContext struct {
	Engine GameEngine
	Card   *EventCard
	Input  PendingInteraction
}

type AmbushEvent struct {
}

func (AmbushEvent) isCardEvent() {}
func (e AmbushEvent) Execute(ctx CardEventContext) ([]Event, error) {
	events, err := ctx.Engine.OpenDoor()
	if err != nil {
		return events, err
	}
	secondDoorEvents, err := ctx.Engine.OpenDoor()
	events = append(events, secondDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}
func (e AmbushEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type DungeonErrorInYourFavorEvent struct {
}

func (DungeonErrorInYourFavorEvent) isCardEvent() {}
func (e DungeonErrorInYourFavorEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event

	for _, p := range ctx.Engine.ListPlayers() {
		drawEvents, err := p.DrawCardsFromDeck(5)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e DungeonErrorInYourFavorEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type SuddenIllnessEvent struct {
}

func (SuddenIllnessEvent) isCardEvent() {}
func (e SuddenIllnessEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event

	for _, p := range ctx.Engine.ListPlayers() {
		drawEvents, err := p.DiscardCards(p.Hand)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e SuddenIllnessEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type CrowdFundingEvent struct {
}

func (CrowdFundingEvent) isCardEvent() {}
func (e CrowdFundingEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	if ctx.Input != nil {
		input, ok := ctx.Input.(*TeamChoiceArtifactInteraction)
		if !ok {
			return nil, fmt.Errorf("invalid interaction type, received: %v", input)
		}

		target, err := input.Target()
		if err != nil {
			return nil, err
		}

		if !slices.Contains(ctx.Engine.ListArtifacts(), target) {
			return nil, fmt.Errorf("artifact is not in play")
		}
		target.Used = false
		events = append(events, ArtifactReEnabledEvent{Artifact: target})
	}

	for _, p := range ctx.Engine.ListPlayers() {
		discardEvents, err := p.DiscardCards(p.Hand)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e CrowdFundingEvent) Interaction(ctx CardEventContext) PendingInteraction {
	artifacts := ctx.Engine.ListArtifacts()
	if len(artifacts) == 0 {
		return nil
	}

	return &TeamChoiceArtifactInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]*ArtifactCard, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type GimmeAHandEvent struct {
}

func (GimmeAHandEvent) isCardEvent() {}
func (e GimmeAHandEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*TeamChoicePlayerInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	target, err := input.Target()
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, p := range otherPlayers(ctx.Engine.ListPlayers(), target) {
		transferredCards := slices.Clone(p.Hand)
		target.Hand = append(target.Hand, p.Hand...)
		p.Hand = make([]PlayerCard, 0)

		events = append(events, HandDonatedEvent{From: p, To: target, Cards: transferredCards})
	}

	return events, nil
}
func (e GimmeAHandEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &TeamChoicePlayerInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]*Player, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type YetMoreSpikesEvent struct {
}

func (YetMoreSpikesEvent) isCardEvent() {}
func (e YetMoreSpikesEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*TeamChoicePlayerInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	target, err := input.Target()
	if err != nil {
		return nil, err
	}

	events, err := target.DiscardCards(target.Hand)
	if err != nil {
		return events, err
	}

	return events, nil
}
func (e YetMoreSpikesEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &TeamChoicePlayerInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]*Player, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type ABooBooEvent struct {
}

func (ABooBooEvent) isCardEvent() {}
func (e ABooBooEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*PlayerDiscardCardsInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	var events []Event
	for p, selectedCards := range input.CollectedChoices {
		discardEvents, err := p.DiscardCards(selectedCards)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}

	}

	return events, nil
}
func (e ABooBooEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &PlayerDiscardCardsInteraction{
		eventCard:        ctx.Card,
		requiredCount:    1,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player][]PlayerCard, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type TrapDoorEvent struct {
}

func (TrapDoorEvent) isCardEvent() {}
func (e TrapDoorEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*PlayerDiscardCardsInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	var events []Event
	for p, selectedCards := range input.CollectedChoices {
		discardEvents, err := p.DiscardCards(selectedCards)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}

	}

	return events, nil
}
func (e TrapDoorEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &PlayerDiscardCardsInteraction{
		eventCard:        ctx.Card,
		requiredCount:    3,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player][]PlayerCard, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type ConfusionEvent struct {
}

func (ConfusionEvent) isCardEvent() {}
func (e ConfusionEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*PlayerDonatesHandInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	return allPlayerDonateHands(input), nil
}
func (e ConfusionEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &PlayerDonatesHandInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]*Player, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type LockedDoorEvent struct {
}

func (LockedDoorEvent) isCardEvent() {}
func (e LockedDoorEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*TeamChoiceResourceInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	target, err := input.Target()
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		discardEvents, err := discardAllResourceCardsContaining(p, target)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e LockedDoorEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &TeamChoiceResourceInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]ResourceType, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type AnUngodlyAmountOfPorcupinesEvent struct {
}

func (AnUngodlyAmountOfPorcupinesEvent) isCardEvent() {}
func (e AnUngodlyAmountOfPorcupinesEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*PlayerDiscardCardsInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	var events []Event
	for p, selectedCards := range input.CollectedChoices {
		discardEvents, err := p.DiscardCards(selectedCards)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}

	}

	return events, nil
}
func (e AnUngodlyAmountOfPorcupinesEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &PlayerDiscardCardsInteraction{
		eventCard:        ctx.Card,
		requiredCount:    3,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player][]PlayerCard, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
		init: func(ctx CardEventContext) ([]Event, error) {
			var events []Event
			for _, p := range ctx.Engine.ListPlayers() {
				drawCards, err := p.DrawCardsFromDeck(3)
				events = append(events, drawCards...)
				if err != nil {
					return events, err
				}
			}

			return events, nil
		},
	}
}

type PoisonedMilkEvent struct {
}

func (PoisonedMilkEvent) isCardEvent() {}
func (e PoisonedMilkEvent) Execute(ctx CardEventContext) ([]Event, error) {
	handSizes := make(map[int][]*Player, 6)
	biggestHandSize := 0
	for _, p := range ctx.Engine.ListPlayers() {
		handSizes[len(p.Hand)] = append(handSizes[len(p.Hand)], p)
		biggestHandSize = max(len(p.Hand), biggestHandSize)
	}

	var events []Event
	for _, p := range handSizes[biggestHandSize] {
		discardEvents, err := p.DiscardCards(p.Hand)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e PoisonedMilkEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type AcidPolishEvent struct {
}

func (AcidPolishEvent) isCardEvent() {}
func (e AcidPolishEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		for _, card := range slices.Clone(p.Hand) {
			if c, ok := card.(*ResourceCard); ok {
				if slices.Contains(c.Resources, Shield) {
					discardEvents, err := p.DiscardCards(p.Hand)
					events = append(events, discardEvents...)
					if err != nil {
						return events, err
					}
					break
				}
			}
		}
	}

	return events, nil
}
func (e AcidPolishEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type WaxedFloorEvent struct {
}

func (WaxedFloorEvent) isCardEvent() {}
func (e WaxedFloorEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		if len(p.Hand) > 5 {
			discardEvents, err := p.DiscardCards(p.Hand)
			events = append(events, discardEvents...)
			if err != nil {
				return events, err
			}
		}
	}

	return events, nil
}
func (e WaxedFloorEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type CorrosiveSpitEvent struct {
}

func (CorrosiveSpitEvent) isCardEvent() {}
func (e CorrosiveSpitEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		discardEvents, err := discardAllResourceCardsContaining(p, Shield)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e CorrosiveSpitEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type EnsnaredEvent struct {
}

func (EnsnaredEvent) isCardEvent() {}
func (e EnsnaredEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		discardEvents, err := discardAllResourceCardsContaining(p, Jump)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e EnsnaredEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type MySwordsEvent struct {
}

func (MySwordsEvent) isCardEvent() {}
func (e MySwordsEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		discardEvents, err := discardAllResourceCardsContaining(p, Sword)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e MySwordsEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

type FireBreathEvent struct {
}

func (FireBreathEvent) isCardEvent() {}
func (e FireBreathEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*TeamChoicePlayerInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	target, err := input.Target()
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, p := range otherPlayers(ctx.Engine.ListPlayers(), target) {
		discardEvents, err := p.DiscardCards(p.Hand)
		events = append(events, discardEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e FireBreathEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &TeamChoicePlayerInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]*Player, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type TailSwipeEvent struct {
}

func (TailSwipeEvent) isCardEvent() {}
func (e TailSwipeEvent) Execute(ctx CardEventContext) ([]Event, error) {
	input, ok := ctx.Input.(*PlayerDonatesHandInteraction)
	if !ok {
		return nil, fmt.Errorf("invalid interaction type, received: %v", input)
	}

	return allPlayerDonateHands(input), nil
}
func (e TailSwipeEvent) Interaction(ctx CardEventContext) PendingInteraction {
	return &PlayerDonatesHandInteraction{
		eventCard:        ctx.Card,
		PendingPlayers:   populatePendingPlayers(ctx),
		CollectedChoices: make(map[*Player]*Player, len(ctx.Engine.ListPlayers())),
		OnComplete:       e.Execute,
	}
}

type ATwentySidedBoulderEvent struct {
}

func (ATwentySidedBoulderEvent) isCardEvent() {}
func (e ATwentySidedBoulderEvent) Execute(ctx CardEventContext) ([]Event, error) {
	var events []Event
	for _, p := range ctx.Engine.ListPlayers() {
		drawEvents, err := p.DiscardCards(p.Hand)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
func (e ATwentySidedBoulderEvent) Interaction(_ CardEventContext) PendingInteraction {
	return nil
}

func discardAllResourceCardsContaining(player *Player, resource ResourceType) ([]Event, error) {
	var events []Event
	for _, card := range slices.Clone(player.Hand) {
		if c, ok := card.(*ResourceCard); ok {
			if slices.Contains(c.Resources, resource) {
				discardEvents, err := player.DiscardCard(c)
				events = append(events, discardEvents...)
				if err != nil {
					return events, err
				}
			}
		}
	}

	return events, nil
}

func allPlayerDonateHands(input *PlayerDonatesHandInteraction) []Event {
	var events []Event
	newHands := make(map[*Player][]PlayerCard, 6)
	for sourcePlayer, targetPlayer := range input.CollectedChoices {
		transferredCards := slices.Clone(sourcePlayer.Hand)
		sourcePlayer.Hand = make([]PlayerCard, 0)
		newHands[targetPlayer] = append(newHands[targetPlayer], transferredCards...)

		events = append(events, HandDonatedEvent{From: sourcePlayer, To: targetPlayer, Cards: transferredCards})
	}

	for targetPlayer, cards := range newHands {
		targetPlayer.Hand = append(targetPlayer.Hand, cards...)
	}
	return events
}

func populatePendingPlayers(ctx CardEventContext) map[*Player]bool {
	pendingPlayers := make(map[*Player]bool, len(ctx.Engine.ListPlayers()))
	for _, p := range ctx.Engine.ListPlayers() {
		pendingPlayers[p] = true
	}

	return pendingPlayers
}
