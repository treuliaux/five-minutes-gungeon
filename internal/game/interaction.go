package game

import (
	"errors"
	"math/rand/v2"
)

//go:generate go run github.com/fairjungle/enumer -type=InteractionKind -json
type InteractionKind uint8

const (
	InteractionNoInteraction InteractionKind = iota
	InteractionTeamChoicePlayer
	InteractionTeamChoiceResource
	InteractionTeamChoiceArtifact
	InteractionPlayerDiscardCards
	InteractionPlayerDonatesHand
)

type PendingInteraction interface {
	isInteraction()
	Init(ctx Context) ([]Event, error)
}

type TeamChoicePlayerInteraction struct {
	Kind             InteractionKind
	Card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*Player
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (TeamChoicePlayerInteraction) isInteraction() {}
func (i TeamChoicePlayerInteraction) Init(ctx Context) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}
	return i.init(ctx)
}
func (i TeamChoicePlayerInteraction) Target() (*Player, error) {
	return tallyMajorityVote[*Player](i.CollectedChoices, "no player target found")
}

type PlayerDiscardCardsInteraction struct {
	Kind             InteractionKind
	Card             DungeonCard
	requiredCounts   map[*Player]int
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player][]PlayerCard
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (PlayerDiscardCardsInteraction) isInteraction() {}
func (i PlayerDiscardCardsInteraction) Init(ctx Context) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}

	return i.init(ctx)
}

type TeamChoiceResourceInteraction struct {
	Kind             InteractionKind
	Card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]ResourceType
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (TeamChoiceResourceInteraction) isInteraction() {}
func (i TeamChoiceResourceInteraction) Init(ctx Context) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}

	return i.init(ctx)
}
func (i TeamChoiceResourceInteraction) Target() (ResourceType, error) {
	return tallyMajorityVote[ResourceType](i.CollectedChoices, "no resource type target found")
}

type PlayerDonatesHandInteraction struct {
	Kind             InteractionKind
	Card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*Player
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (PlayerDonatesHandInteraction) isInteraction() {}
func (i PlayerDonatesHandInteraction) Init(ctx Context) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}

	return i.init(ctx)
}

type TeamChoiceArtifactInteraction struct {
	Kind             InteractionKind
	Card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*ArtifactCard
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (TeamChoiceArtifactInteraction) isInteraction() {}
func (i TeamChoiceArtifactInteraction) Init(ctx Context) ([]Event, error) {
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
