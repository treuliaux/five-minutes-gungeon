package game

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// --- Integration Tests for Runner ---

func TestRunnerFullMatchLifecycle(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin)
	barbarian, _ := NewPlayer("Conan", Barbarian)

	door1 := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	boss := &BossMat{Name: "Piti Amenou", Resources: []ResourceType{Shield}}
	dungeon := &Dungeon{
		Boss:  boss,
		Doors: []DungeonCard{door1},
	}

	game := &Game{
		Players:   []*Player{paladin, barbarian},
		Dungeon:   dungeon,
		PlayField: *NewPlayfield(),
		Status:    Waiting,
	}

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Run(ctx)
	}()

	// 1. Start the game via Runner
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("failed to start game via runner: %v", err)
	}

	// Drain initial events (GameStartedEvent, DoorOpenedEvent)
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(GameStartedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(DoorOpenedEvent); return ok })

	// 2. Paladin plays Sword to beat door1
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	paladin.Hand = []PlayerCard{swordCard}

	if err := runner.PlayCard(ctx, paladin, swordCard); err != nil {
		t.Fatalf("paladin failed to play sword: %v", err)
	}

	expectEvent(t, sub, func(e Event) bool { _, ok := e.(CardPlayedEvent); return ok })
	// Wait for debounce and door progression events
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(DoorDefeatedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(FieldClearedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool {
		doorEvt, ok := e.(DoorOpenedEvent)
		return ok && doorEvt.DungeonCard == boss
	})

	// 3. Barbarian plays Shield to beat Boss
	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}
	barbarian.Hand = []PlayerCard{shieldCard}

	if err := runner.PlayCard(ctx, barbarian, shieldCard); err != nil {
		t.Fatalf("barbarian failed to play shield: %v", err)
	}

	expectEvent(t, sub, func(e Event) bool { _, ok := e.(CardPlayedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(DoorDefeatedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(FieldClearedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(GameWonEvent); return ok })

	// 4. Verify runner completes cleanly
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected runner to exit with nil, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not exit within timeout after victory")
	}

	if game.Status != Victory {
		t.Errorf("expected game status Victory, got %v", game.Status)
	}
}

func TestRunnerGracefulCancellation(t *testing.T) {
	game := NewGame()
	runner := NewRunner(game)

	ctx, cancel := context.WithCancel(context.Background())
	sub := runner.Subscribe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Run(ctx)
	}()

	// Allow runner to start
	time.Sleep(20 * time.Millisecond)

	// Cancel context
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runner failed to stop promptly upon cancellation")
	}

	// Verify subscriber channel is closed
	select {
	case _, ok := <-sub:
		if ok {
			t.Error("expected subscriber channel to be closed upon runner teardown")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("subscriber channel was not closed")
	}
}

func TestRunnerContextTimeout(t *testing.T) {
	game := NewGame()
	runner := NewRunner(game)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := runner.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestRunnerSubscriberFanOutAndUnsubscribe(t *testing.T) {
	game := NewGame()
	runner := NewRunner(game)

	ctx := t.Context()

	sub1 := runner.Subscribe()
	sub2 := runner.Subscribe()
	sub3 := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	// Unsubscribe sub2
	runner.Unsubscribe(sub2)

	// Add player via runner
	if err := runner.AddPlayer(ctx, "Arthur", Paladin); err != nil {
		t.Fatalf("failed to add player: %v", err)
	}

	// sub1 and sub3 must receive PlayerAddedEvent
	expectEvent(t, sub1, func(e Event) bool { _, ok := e.(PlayerAddedEvent); return ok })
	expectEvent(t, sub3, func(e Event) bool { _, ok := e.(PlayerAddedEvent); return ok })

	// sub2 should receive nothing
	select {
	case evt := <-sub2:
		t.Errorf("unsubscribed sub2 received unexpected event: %v", evt)
	case <-time.After(50 * time.Millisecond):
		// Expected: no event on unsubscribed channel
	}
}

func TestRunnerCommandErrorPropagation(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin)
	cInHand := &ResourceCard{Resources: []ResourceType{Sword}}
	cNotInHand := &ResourceCard{Resources: []ResourceType{Shield}}
	paladin.Hand = []PlayerCard{cInHand}

	game := &Game{
		Players:   []*Player{paladin},
		PlayField: *NewPlayfield(),
		Status:    Playing,
	}

	runner := NewRunner(game)
	ctx := t.Context()

	go func() {
		_ = runner.Run(ctx)
	}()

	// Attempt to play card not in hand
	err := runner.PlayCard(ctx, paladin, cNotInHand)
	if err == nil {
		t.Fatal("expected error playing card not in hand, got nil")
	}
	if err.Error() != "player does not have card in hand" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunnerConcurrentPlayerCommandsRace(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	barbarian, _ := NewPlayer("Barbarian", Barbarian)
	valkyrie, _ := NewPlayer("Valkyrie", Valkyrie)
	gladiator, _ := NewPlayer("Gladiator", Gladiator)

	game := &Game{
		Players:   []*Player{paladin, barbarian, valkyrie, gladiator},
		PlayField: *NewPlayfield(),
		Status:    Waiting,
	}

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	if err := runner.Start(ctx); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Concurrently send commands from multiple goroutines
	var wg sync.WaitGroup
	players := []*Player{paladin, barbarian, valkyrie, gladiator}

	for _, p := range players {
		wg.Go(func() {
			for i := range 20 {
				// Try playing or discarding cards concurrently
				if len(p.Hand) > 0 {
					card := p.Hand[0]
					if i%2 == 0 {
						_ = runner.PlayCard(ctx, p, card)
					} else {
						_ = runner.DiscardCard(ctx, p, card)
					}
				}
				time.Sleep(1 * time.Millisecond)
			}
		})
	}

	// Drain subscriber events in background to prevent buffer stall
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		for {
			select {
			case _, ok := <-sub:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	cancel()
	<-drainDone
}

func TestRunnerHelperRespectsCallerContextCancellation(t *testing.T) {
	game := NewGame()
	runner := NewRunner(game)

	// Context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.AddPlayer(ctx, "Arthur", Paladin)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled for cancelled caller context, got: %v", err)
	}
}

func TestRunnerUseHeroAbility(t *testing.T) {
	wizard, _ := NewPlayer("Gandalf", Wizard)
	paladin, _ := NewPlayer("Arthur", Paladin)

	game := &Game{
		Players:   []*Player{wizard, paladin},
		PlayField: *NewPlayfield(),
		Status:    Playing,
	}

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	c1 := &ResourceCard{Resources: []ResourceType{Scroll}}
	c2 := &ResourceCard{Resources: []ResourceType{Scroll}}
	c3 := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{c1, c2, c3}

	err := runner.UseHeroAbility(ctx, wizard, []PlayerCard{c1, c2, c3}, StopTimeAbility{})
	if err != nil {
		t.Fatalf("failed to use ability via runner: %v", err)
	}

	// Expect HeroAbilityUsedEvent, 3x CardDiscardedEvent, TimeFrozenEvent broadcasted
	expectEvent(t, sub, func(e Event) bool {
		evt, ok := e.(HeroAbilityUsedEvent)
		return ok && evt.ByPlayer == wizard
	})
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(TimeFrozenEvent); return ok })

	if !game.IsTimeFrozen {
		t.Error("expected game.isTimeFrozen to be true")
	}
}

// --- Test Helpers ---

func expectEvent(t *testing.T, sub <-chan Event, predicate func(Event) bool) {
	t.Helper()
	timeout := time.After(2 * time.Second)
	for {
		select {
		case evt, ok := <-sub:
			if !ok {
				t.Fatal("subscriber channel closed while waiting for event")
			}
			if predicate(evt) {
				return
			}
		case <-timeout:
			t.Fatal("timed out waiting for expected event")
		}
	}
}
