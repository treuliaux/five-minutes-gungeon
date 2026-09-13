package game

import (
	"context"
	"time"
)

const tickDuration = time.Second / 20

type Runner struct {
	game   *Game
	cmd    chan Command
	ticker *time.Ticker
}

func NewRunner(game *Game) *Runner {
	return &Runner{
		game: game,
		cmd:  make(chan Command),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	r.ticker = time.NewTicker(tickDuration)
	defer r.ticker.Stop()

	for {
		select {
		case cmd := <-r.cmd:
			r.game.Apply(cmd)
		case <-r.ticker.C:
			r.game.Tick(tickDuration)
		case <-ctx.Done():
			return ctx.Err()
		}
		if r.game.Status == Victory || r.game.Status == Defeat {
			return nil
		}
	}
}

func (r *Runner) AddPlayer(name string, class HeroClass) error {
	reply := make(chan error, 1)
	r.cmd <- AddPlayerCmd{Name: name, Class: class, reply: reply}

	return <-reply
}

func (r *Runner) PlayCard(p *Player, c PlayerCard) error {
	reply := make(chan error, 1)
	r.cmd <- PlayCardCmd{Player: p, Card: c, reply: reply}

	return <-reply
}

func (r *Runner) DiscardCard(p *Player, c PlayerCard) error {
	reply := make(chan error, 1)
	r.cmd <- DiscardCardCmd{Player: p, Card: c, reply: reply}

	return <-reply
}
