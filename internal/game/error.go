package game

import (
	"errors"
	"fmt"
)

var (
	ErrPlayerNotFound            = errors.New("player not found")
	ErrCardNotFound              = errors.New("card not found")
	ErrCardNotInHand             = errors.New("card not in hand")
	ErrGameNotPlaying            = errors.New("game is not in playing state")
	ErrGameAlreadyStarted        = errors.New("game has already started")
	ErrGameIncorrectPlayerNumber = errors.New("game has incorrect number of players")
	ErrGameExtensionDisabled     = errors.New("extension is disabled")
	ErrInvalidTarget             = errors.New("invalid target")
	ErrExpectedTarget            = errors.New("expected target")
	ErrNoPendingInteraction      = errors.New("no pending player interaction")
	ErrPendingInteraction        = errors.New("pending player interaction")
	ErrSmartTargeting            = errors.New("no candidate found")
	ErrClassAlreadyPicked        = errors.New("class already picked")
	ErrPlayerAlreadyExists       = errors.New("player already exists")
	ErrArtifactAlreadyUsed       = errors.New("artifact already used")
)

type AmbiguousTargetError[T any] struct {
	Source       any
	ValidTargets []T
}

func (e *AmbiguousTargetError[T]) Error() string {
	return fmt.Sprintf("ambiguous target for source %T: %d matching targets available", e.Source, len(e.ValidTargets))
}

type CurseRuleViolatedError struct {
	Curse    GameCurseEffect
	ByPlayer *Player
}

func (e *CurseRuleViolatedError) Error() string {
	return fmt.Sprintf("curse rule %v violated by player %v ", e.Curse, e.ByPlayer)
}

type GameTerminatedError struct {
	Command Command
}

func (e *GameTerminatedError) Error() string {
	return fmt.Sprintf("game terminated, could not handle command %v", e.Command)
}
