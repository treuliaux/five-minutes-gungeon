package game

import (
	"context"
	"slices"
	"sync"
	"time"
)

const tickDuration = time.Second / 20

type Runner struct {
	game              *Game
	cmd               chan Command
	done              chan struct{}
	ticker            *time.Ticker
	subscribers       []*Subscriber
	subscriptionsLock sync.RWMutex
	canSubscribe      bool
}

func NewRunner(game *Game) *Runner {
	return &Runner{
		game:         game,
		cmd:          make(chan Command, 64),
		done:         make(chan struct{}),
		canSubscribe: true,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	r.ticker = time.NewTicker(tickDuration)
	resetGlobalCardIndex()
	defer func() {
		close(r.done)
		r.drainCommandsChanel()
		r.closeAndClearSubscribers()
		r.ticker.Stop()
		resetGlobalCardIndex()
	}()

	var events []Event
	for {
		select {
		case cmd := <-r.cmd:
			events, _ = r.game.Apply(cmd)
		case <-r.ticker.C:
			events, _ = r.game.Tick(tickDuration)
		case <-ctx.Done():
			r.subscriptionsLock.Lock()
			r.canSubscribe = false
			r.subscriptionsLock.Unlock()

			return ctx.Err()
		}
		for _, event := range events {
			r.broadcast(event)
		}
		if r.game.Status == Victory || r.game.Status == Defeat {
			r.subscriptionsLock.Lock()
			r.canSubscribe = false
			r.subscriptionsLock.Unlock()

			return nil
		}
	}
}

func (r *Runner) Subscribe() <-chan Event {
	r.subscriptionsLock.Lock()
	defer r.subscriptionsLock.Unlock()

	sub := NewSubscriber()
	if !r.canSubscribe {
		sub.CloseImmediately()

		return sub.Out
	}
	r.subscribers = append(r.subscribers, sub)

	return sub.Out
}

func (r *Runner) Unsubscribe(sub <-chan Event) {
	r.subscriptionsLock.Lock()
	defer r.subscriptionsLock.Unlock()

	for i, s := range r.subscribers {
		if s.Out == sub {
			s.CloseImmediately()
			r.subscribers[i] = nil
			r.subscribers = append(r.subscribers[:i], r.subscribers[i+1:]...)
			if len(r.subscribers) == 0 {
				r.subscribers = nil
			}

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

func (r *Runner) PlayCardSimple(ctx context.Context, actorID PlayerID, cardID CardID) error {
	return r.PlayCard(ctx, actorID, cardID, 0, nil)
}

func (r *Runner) PlayCardWithPlayersTarget(ctx context.Context, actorID PlayerID, cardID CardID, targetPlayerIDs []PlayerID) error {
	return r.PlayCard(ctx, actorID, cardID, 0, targetPlayerIDs)
}

func (r *Runner) PlayCardWithCardTarget(ctx context.Context, actorID PlayerID, cardID CardID, targetCardID CardID) error {
	return r.PlayCard(ctx, actorID, cardID, targetCardID, nil)
}

func (r *Runner) PlayCard(ctx context.Context, actorID PlayerID, cardID CardID, targetCardID CardID, targetPlayerIDs []PlayerID) error {
	reply := make(chan error, 1)
	cmd := PlayCardCmd{
		PlayerID:        actorID,
		CardID:          cardID,
		TargetCardID:    targetCardID,
		TargetPlayerIDs: targetPlayerIDs,
		reply:           reply,
	}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) DiscardCards(ctx context.Context, actorID PlayerID, cardIDs []CardID) error {
	reply := make(chan error, 1)
	cmd := DiscardCardsCmd{
		PlayerID: actorID,
		CardIDs:  cardIDs,
		reply:    reply,
	}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) UseHeroAbilitySimple(ctx context.Context, pID PlayerID, discardIDs []CardID) error {
	return r.UseHeroAbility(ctx, pID, discardIDs, 0, "")
}

func (r *Runner) UseHeroAbilityWithCardTarget(ctx context.Context, pID PlayerID, discardIDs []CardID, targetCard CardID) error {
	return r.UseHeroAbility(ctx, pID, discardIDs, targetCard, "")
}

func (r *Runner) UseHeroAbilityWithPlayerTarget(ctx context.Context, pID PlayerID, discardIDs []CardID, targetPlayer PlayerID) error {
	return r.UseHeroAbility(ctx, pID, discardIDs, 0, targetPlayer)
}

func (r *Runner) UseHeroAbility(ctx context.Context, pID PlayerID, discardIDs []CardID, targetCard CardID, targetPlayer PlayerID) error {
	reply := make(chan error, 1)
	cmd := UseHeroAbilityCmd{
		PlayerID:       pID,
		DiscardCardIDs: discardIDs,
		TargetCardID:   targetCard,
		TargetPlayerID: targetPlayer,
		reply:          reply,
	}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) SubmitPromptChoice(ctx context.Context, actorID PlayerID, targetID PlayerID, cardIDs []CardID, res *ResourceType) error {
	reply := make(chan error, 1)
	cmd := SubmitPromptChoiceCmd{
		PlayerID:       actorID,
		TargetPlayerID: targetID,
		CardIDs:        cardIDs,
		Resource:       res,
		reply:          reply,
	}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) UseArtifact(ctx context.Context, actorID PlayerID, artifactID ArtifactID, actionIndex ArtifactActionIndex, targetID CardID) error {
	reply := make(chan error, 1)
	cmd := UseArtifactCmd{
		PlayerID:    actorID,
		ArtifactID:  artifactID,
		ActionIndex: actionIndex,
		TargetID:    targetID,
		reply:       reply,
	}

	return guardedCmdCallAndReply(ctx, r, cmd, reply)
}

func (r *Runner) broadcast(event Event) {
	r.subscriptionsLock.RLock()
	defer r.subscriptionsLock.RUnlock()
	for _, sub := range slices.Clone(r.subscribers) {
		sub.In <- event
	}
}

func guardedCmdCallAndReply(ctx context.Context, r *Runner, command Command, reply chan error) error {
	select {
	case <-r.done:
		return &GameTerminatedError{Command: command}
	default:
	}

	select {
	case r.cmd <- command:
	case <-r.done:
		return &GameTerminatedError{Command: command}
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-reply:
		return err
	case <-r.done:
		return &GameTerminatedError{Command: command}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) closeAndClearSubscribers() {
	r.subscriptionsLock.Lock()
	defer r.subscriptionsLock.Unlock()

	var wg sync.WaitGroup
	for _, sub := range r.subscribers {
		sub.CloseGracefully()
		wg.Add(1)
		go func(s *Subscriber) {
			defer wg.Done()
			<-s.Done
		}(sub)
	}
	wg.Wait()
	r.subscribers = nil
}

func (r *Runner) drainCommandsChanel() {
	for {
		select {
		case cmd := <-r.cmd:
			if cmd.Reply() != nil {
				cmd.Reply() <- &GameTerminatedError{Command: cmd}
			}
		default:
			return
		}
	}
}
