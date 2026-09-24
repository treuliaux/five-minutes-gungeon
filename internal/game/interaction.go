package game

import (
	"errors"
	"math/rand/v2"
)

type InteractionKind string

const (
	InteractionTeamChoicePlayer   InteractionKind = "TeamChoicePlayer"
	InteractionTeamChoiceResource InteractionKind = "TeamChoiceResource"
	InteractionTeamChoiceArtifact InteractionKind = "TeamChoiceArtifact"
	InteractionPlayerDiscardCards InteractionKind = "PlayerDiscardCards"
	InteractionPlayerDonatesHand  InteractionKind = "PlayerDonatesHand"
)

type PendingInteraction interface {
	isInteraction()
	Kind() InteractionKind
	RequiredCount() int
	EventCard() *EventCard
	Init(ctx CardEventContext) ([]Event, error)
}

type TeamChoicePlayerInteraction struct {
	eventCard        *EventCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*Player
	OnComplete       func(ctx CardEventContext) ([]Event, error)
	init             func(ctx CardEventContext) ([]Event, error)
}

func (TeamChoicePlayerInteraction) isInteraction()        {}
func (TeamChoicePlayerInteraction) Kind() InteractionKind { return InteractionTeamChoicePlayer }
func (TeamChoicePlayerInteraction) RequiredCount() int    { return 1 }
func (i TeamChoicePlayerInteraction) EventCard() *EventCard {
	return i.eventCard
}
func (i TeamChoicePlayerInteraction) Init(ctx CardEventContext) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}
	return i.init(ctx)
}
func (i TeamChoicePlayerInteraction) Target() (*Player, error) {
	return tallyMajorityVote[*Player](i.CollectedChoices, "no player target found")
}

type PlayerDiscardCardsInteraction struct {
	eventCard        *EventCard
	requiredCount    int
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player][]PlayerCard
	OnComplete       func(ctx CardEventContext) ([]Event, error)
	init             func(ctx CardEventContext) ([]Event, error)
}

func (PlayerDiscardCardsInteraction) isInteraction()          {}
func (PlayerDiscardCardsInteraction) Kind() InteractionKind   { return InteractionPlayerDiscardCards }
func (i PlayerDiscardCardsInteraction) RequiredCount() int    { return i.requiredCount }
func (i PlayerDiscardCardsInteraction) EventCard() *EventCard { return i.eventCard }
func (i PlayerDiscardCardsInteraction) Init(ctx CardEventContext) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}
	return i.init(ctx)
}

type TeamChoiceResourceInteraction struct {
	eventCard        *EventCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]ResourceType
	OnComplete       func(ctx CardEventContext) ([]Event, error)
	init             func(ctx CardEventContext) ([]Event, error)
}

func (TeamChoiceResourceInteraction) isInteraction()          {}
func (TeamChoiceResourceInteraction) Kind() InteractionKind   { return InteractionTeamChoiceResource }
func (TeamChoiceResourceInteraction) RequiredCount() int      { return 1 }
func (i TeamChoiceResourceInteraction) EventCard() *EventCard { return i.eventCard }
func (i TeamChoiceResourceInteraction) Init(ctx CardEventContext) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}
	return i.init(ctx)
}
func (i TeamChoiceResourceInteraction) Target() (ResourceType, error) {
	return tallyMajorityVote[ResourceType](i.CollectedChoices, "no resource type target found")
}

type PlayerDonatesHandInteraction struct {
	eventCard        *EventCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*Player
	OnComplete       func(ctx CardEventContext) ([]Event, error)
	init             func(ctx CardEventContext) ([]Event, error)
}

func (PlayerDonatesHandInteraction) isInteraction()        {}
func (PlayerDonatesHandInteraction) Kind() InteractionKind { return InteractionPlayerDonatesHand }
func (PlayerDonatesHandInteraction) RequiredCount() int    { return 1 }
func (i PlayerDonatesHandInteraction) EventCard() *EventCard {
	return i.eventCard
}
func (i PlayerDonatesHandInteraction) Init(ctx CardEventContext) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}
	return i.init(ctx)
}

type TeamChoiceArtifactInteraction struct {
	eventCard        *EventCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*ArtifactCard
	OnComplete       func(ctx CardEventContext) ([]Event, error)
	init             func(ctx CardEventContext) ([]Event, error)
}

func (TeamChoiceArtifactInteraction) isInteraction()        {}
func (TeamChoiceArtifactInteraction) Kind() InteractionKind { return InteractionTeamChoiceArtifact }
func (TeamChoiceArtifactInteraction) RequiredCount() int    { return 1 }
func (i TeamChoiceArtifactInteraction) EventCard() *EventCard {
	return i.eventCard
}
func (i TeamChoiceArtifactInteraction) Init(ctx CardEventContext) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}
	return i.init(ctx)
}
func (i TeamChoiceArtifactInteraction) Target() (*ArtifactCard, error) {
	return tallyMajorityVote[*ArtifactCard](i.CollectedChoices, "no artifact target found")
}

func tallyMajorityVote[T comparable](choices map[*Player]T, notFoundMsg string) (T, error) {
	var zero T
	counts := make(map[T]int, len(choices))
	for _, choice := range choices {
		counts[choice]++
	}
	maxVotes := 0
	var candidates []T
	for item, votes := range counts {
		switch {
		case votes > maxVotes:
			maxVotes = votes
			candidates = []T{item}
		case votes == maxVotes:
			candidates = append(candidates, item)
		}
	}

	if len(candidates) == 0 {
		return zero, errors.New(notFoundMsg)
	}

	return candidates[rand.IntN(len(candidates))], nil
}
