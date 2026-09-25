package game

import "fmt"

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
