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
	RequiredCounts() map[*Player]int
	Card() DungeonCard
	Init(ctx Context) ([]Event, error)
}

type TeamChoicePlayerInteraction struct {
	card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*Player
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (TeamChoicePlayerInteraction) isInteraction()                  {}
func (TeamChoicePlayerInteraction) Kind() InteractionKind           { return InteractionTeamChoicePlayer }
func (TeamChoicePlayerInteraction) RequiredCounts() map[*Player]int { return nil }
func (i TeamChoicePlayerInteraction) Card() DungeonCard {
	return i.card
}
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
	card             DungeonCard
	requiredCounts   map[*Player]int
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player][]PlayerCard
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (PlayerDiscardCardsInteraction) isInteraction()                    {}
func (PlayerDiscardCardsInteraction) Kind() InteractionKind             { return InteractionPlayerDiscardCards }
func (i PlayerDiscardCardsInteraction) RequiredCounts() map[*Player]int { return i.requiredCounts }
func (i PlayerDiscardCardsInteraction) Card() DungeonCard               { return i.card }
func (i PlayerDiscardCardsInteraction) Init(ctx Context) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}

	return i.init(ctx)
}

type TeamChoiceResourceInteraction struct {
	card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]ResourceType
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (TeamChoiceResourceInteraction) isInteraction()                  {}
func (TeamChoiceResourceInteraction) Kind() InteractionKind           { return InteractionTeamChoiceResource }
func (TeamChoiceResourceInteraction) RequiredCounts() map[*Player]int { return nil }
func (i TeamChoiceResourceInteraction) Card() DungeonCard             { return i.card }
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
	card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*Player
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (PlayerDonatesHandInteraction) isInteraction()                  {}
func (PlayerDonatesHandInteraction) Kind() InteractionKind           { return InteractionPlayerDonatesHand }
func (PlayerDonatesHandInteraction) RequiredCounts() map[*Player]int { return nil }
func (i PlayerDonatesHandInteraction) Card() DungeonCard {
	return i.card
}
func (i PlayerDonatesHandInteraction) Init(ctx Context) ([]Event, error) {
	if i.init == nil {
		return nil, nil
	}

	return i.init(ctx)
}

type TeamChoiceArtifactInteraction struct {
	card             DungeonCard
	PendingPlayers   map[*Player]bool
	CollectedChoices map[*Player]*ArtifactCard
	OnComplete       func(ctx Context) ([]Event, error)
	init             func(ctx Context) ([]Event, error)
}

func (TeamChoiceArtifactInteraction) isInteraction()                  {}
func (TeamChoiceArtifactInteraction) Kind() InteractionKind           { return InteractionTeamChoiceArtifact }
func (TeamChoiceArtifactInteraction) RequiredCounts() map[*Player]int { return nil }
func (i TeamChoiceArtifactInteraction) Card() DungeonCard {
	return i.card
}
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
