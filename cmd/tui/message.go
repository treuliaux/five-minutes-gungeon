package main

import (
	"context"

	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type sessionStartedMsg struct {
	controller          client.GameController
	ctx                 context.Context
	cancel              context.CancelFunc
	config              game.Config
	estimatedDifficulty uint8
}

type gameEventMsg game.Event

type snapshotMsg game.GameSnapshot
