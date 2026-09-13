package game

import (
	"context"
	"sync"
	"time"
)

const tickDuration = time.Second / 20

type Runner struct {
	game        *Game
	cmd         chan Command
	ticker      *time.Ticker
	subscribers []chan Event
	subLock     sync.RWMutex
}

func NewRunner(game *Game) *Runner {
	return &Runner{
		game: game,
		cmd:  make(chan Command, 64),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	r.ticker = time.NewTicker(tickDuration)
	defer r.ticker.Stop()
	defer func() {
		r.subLock.Lock()
		defer r.subLock.Unlock()

		for _, sub := range r.subscribers {
			close(sub)
		}
	}()

	var events []Event
	for {
		select {
		case cmd := <-r.cmd:
			events, _ = r.game.Apply(cmd)
		case <-r.ticker.C:
			events, _ = r.game.Tick(tickDuration)
		case <-ctx.Done():
			return ctx.Err()
		}
		for _, event := range events {
			r.broadcast(event)
		}
		if r.game.Status == Victory || r.game.Status == Defeat {
			return nil
		}
	}
}

func (r *Runner) Subscribe() <-chan Event {
	r.subLock.Lock()
	defer r.subLock.Unlock()

	sub := make(chan Event, 64)
	r.subscribers = append(r.subscribers, sub)

	return sub
}

func (r *Runner) Unsubscribe(sub <-chan Event) {
	r.subLock.Lock()
	defer r.subLock.Unlock()

	for i, s := range r.subscribers {
		if s == sub {
			r.subscribers = append(r.subscribers[:i], r.subscribers[i+1:]...)

			return
		}
	}
}

func (r *Runner) AddPlayer(ctx context.Context, name string, class HeroClass) error {
	reply := make(chan error, 1)
	cmd := AddPlayerCmd{Name: name, Class: class, reply: reply}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) Start(ctx context.Context) error {
	reply := make(chan error, 1)
	cmd := StartCmd{reply: reply}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) PlayCard(ctx context.Context, p *Player, c PlayerCard) error {
	reply := make(chan error, 1)
	cmd := PlayCardCmd{Player: p, Card: c, reply: reply}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) DiscardCard(ctx context.Context, p *Player, c PlayerCard) error {
	reply := make(chan error, 1)
	cmd := DiscardCardCmd{Player: p, Card: c, reply: reply}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) broadcast(event Event) {
	r.subLock.RLock()
	for _, sub := range r.subscribers {
		select {
		case sub <- event:
		default:
			// Buffer full: drop non-critical event, log warning, or disconnect slow consumer
		}
	}
	r.subLock.RUnlock()
}

// Waiting for channels' processing won't block if the runner (or context) has ended.
func guardedCmdCallAndReply(ctx context.Context, r *Runner, cmd Command, reply chan error) error {
	select {
	case r.cmd <- cmd:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
