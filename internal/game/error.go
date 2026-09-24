package game

import "fmt"

type AmbiguousTargetError[T any] struct {
	Source       any
	ValidTargets []T
}

func (e *AmbiguousTargetError[T]) Error() string {
	return fmt.Sprintf("ambiguous target for source %T: %d matching targets available", e.Source, len(e.ValidTargets))
}
