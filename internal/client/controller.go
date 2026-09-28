package client

import (
	"context"

	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type GameController interface {
	GetSnapshot() (game.GameSnapshotDTO, error)
	Dispatch(ctx context.Context, cmd game.Command) error
	Events() <-chan game.Event
	Close() error
}
