package main

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/client"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

func startSession(cfg game.Config) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())

		gameController, err := client.StartLocalSession(ctx, cfg)
		if err != nil {
			cancel()
			return errMsg{err: err}
		}

		return sessionStartedMsg{
			controller: gameController,
			ctx:        ctx,
			cancel:     cancel,
		}
	}
}

func waitForEvent(events <-chan game.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-events
		if !ok {
			// TODO: Handle session ended
			return nil
		}

		return gameEventMsg(ev)
	}
}

func askForSnapshot(controller client.GameController) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := controller.GetSnapshot()
		if err != nil {
			return forwardError(err)
		}

		return snapshotMsg(snapshot)
	}
}

func forwardError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err: err}
	}
}
