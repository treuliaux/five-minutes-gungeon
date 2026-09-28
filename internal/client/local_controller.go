package client

import (
	"context"
	"time"

	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type LocalController struct {
	runner   *game.Runner
	eventsCh <-chan game.Event
}

func NewLocalController(runner *game.Runner) *LocalController {
	return &LocalController{
		runner:   runner,
		eventsCh: runner.Subscribe(),
	}
}

func StartLocalSession(ctx context.Context, cfg game.Config) (*LocalController, error) {
	g := game.NewGame(cfg)

	runner := game.NewRunner(g)
	go func() {
		_ = runner.Run(ctx)
	}()

	return NewLocalController(runner), nil
}

func (c *LocalController) GetSnapshot() (game.GameSnapshotDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return c.runner.Snapshot(ctx)
}

func (c *LocalController) Dispatch(ctx context.Context, command game.Command) error {
	switch cmd := command.(type) {
	case game.StartCmd:
		return c.runner.Start(ctx)
	case game.AddPlayerCmd:
		return c.runner.AddPlayer(ctx, cmd.Name, cmd.Class)
	case game.ChangeHeroCmd:
		return c.runner.ChangeHero(ctx, cmd.PlayerID, cmd.Class)
	case game.PlayCardCmd:
		return c.runner.PlayCard(ctx, cmd.PlayerID, cmd.CardID, cmd.TargetCardID, cmd.TargetPlayerIDs)
	case game.DiscardCardsCmd:
		return c.runner.DiscardCards(ctx, cmd.PlayerID, cmd.CardIDs)
	case game.UseHeroAbilityCmd:
		return c.runner.UseHeroAbility(ctx, cmd.PlayerID, cmd.DiscardCardIDs, cmd.TargetCardID, cmd.TargetPlayerID)
	case game.SubmitPromptChoiceCmd:
		return c.runner.SubmitPromptChoice(ctx, cmd.PlayerID, cmd.TargetPlayerID, cmd.CardIDs, cmd.Resource, cmd.TargetArtifactID)
	case game.UseArtifactCmd:
		return c.runner.UseArtifact(ctx, cmd.PlayerID, cmd.ArtifactID, cmd.ActionIndex, cmd.TargetID)
	}

	return ErrUnknownCommand
}

func (c *LocalController) Events() <-chan game.Event {
	if c.eventsCh == nil {
		c.eventsCh = c.runner.Subscribe()
	}

	return c.eventsCh
}

func (c *LocalController) Close() error {
	c.runner.Unsubscribe(c.eventsCh)

	return nil
}
