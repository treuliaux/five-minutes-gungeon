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
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	barbarian, _ := NewPlayer("Conan", Barbarian, true)

	door1 := &DoorCard{Id: nextCardId(), Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	boss := &BossMat{Id: nextCardId(), Name: "Piti Amenou", Resources: []ResourceType{Shield}}
	dungeon := &Dungeon{
		Boss:  boss,
		Doors: []DungeonCard{door1},
	}

	game := &Game{
		Players:   []*Player{paladin, barbarian},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
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
	swordCard := &ResourceCard{Id: 6666, Resources: []ResourceType{Sword}}
	paladin.Hand = []PlayerCard{swordCard}
	registerCardsInTestGame(game)

	if err := runner.PlayCardSimple(ctx, paladin.Id, swordCard.ID()); err != nil {
		t.Fatalf("paladin failed to play sword: %v", err)
	}

	expectEvent(t, sub, func(e Event) bool { _, ok := e.(CardPlayedEvent); return ok })
	// Wait for debounce and door progression events
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(DoorDefeatedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(FieldClearedEvent); return ok })
	expectEvent(t, sub, func(e Event) bool {
		doorEvt, ok := e.(DoorOpenedEvent)
		return ok && doorEvt.CardID == boss.ID()
	})

	// 3. Barbarian plays Shield to beat Boss
	shieldCard := &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Shield}}
	barbarian.Hand = []PlayerCard{shieldCard}
	registerCardsInTestGame(game)

	if err := runner.PlayCardSimple(ctx, barbarian.Id, shieldCard.ID()); err != nil {
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

	// sub2 should be closed
	_, ok := <-sub2
	if ok {
		t.Errorf("unsubscribed sub2 channel should be closed")
	}
}

func TestRunnerCommandErrorPropagation(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	cInHand := &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Sword}}
	cNotInHand := &ResourceCard{Id: 6666, Resources: []ResourceType{Shield}}
	paladin.Hand = []PlayerCard{cInHand}

	game := &Game{
		Players:   []*Player{paladin},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	runner := NewRunner(game)
	ctx := t.Context()

	go func() {
		_ = runner.Run(ctx)
	}()
	_ = runner.Start(ctx)

	// Attempt to play card not in hand
	err := runner.PlayCardSimple(ctx, paladin.Id, cNotInHand.ID())
	if err == nil {
		t.Fatal("expected error playing card not in hand, got nil")
	}
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunnerConcurrentPlayerCommandsRace(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	barbarian, _ := NewPlayer("Barbarian", Barbarian, true)
	valkyrie, _ := NewPlayer("Valkyrie", Valkyrie, true)
	gladiator, _ := NewPlayer("Gladiator", Gladiator, true)

	game := &Game{
		Players:   []*Player{paladin, barbarian, valkyrie, gladiator},
		PlayField: NewPlayfield(),
		Status:    Waiting,
	}

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	if err := runner.Start(ctx); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Concurrently send commands and query snapshots from multiple goroutines
	var wg sync.WaitGroup
	players := []*Player{paladin, barbarian, valkyrie, gladiator}

	// Add artifact
	artAxe := &ArtifactCard{Id: 1, Color: Red, Name: "Battle Axe", MultiAction: true, Action: BattleAxeArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{artAxe}

	// Goroutines for players sending card plays, discards, and abilities
	for _, p := range players {
		wg.Go(func() {
			for i := range 30 {
				if len(p.Hand) > 0 {
					card := p.Hand[0]
					switch i % 4 {
					case 0:
						_ = runner.PlayCardSimple(ctx, p.Id, card.ID())
					case 1:
						_ = runner.DiscardCards(ctx, p.Id, []CardID{card.ID()})
					case 2:
						_ = runner.UseHeroAbilitySimple(ctx, p.Id, []CardID{card.ID()})
					case 3:
						_ = runner.UseArtifact(ctx, p.Id, artAxe.Id, SecondArtifactAction, 0)
					}
				}
				time.Sleep(1 * time.Millisecond)
			}
		})
	}

	// Concurrent snapshot query goroutines
	for range 4 {
		wg.Go(func() {
			for range 20 {
				snap, err := runner.Snapshot(ctx)
				if err == nil {
					if snap.Status != Playing && snap.Status != Waiting && snap.Status != Victory && snap.Status != Defeat {
						t.Errorf("unexpected snapshot status: %v", snap.Status)
					}
				}
				time.Sleep(2 * time.Millisecond)
			}
		})
	}

	wg.Wait()
	cancel()
}

func TestRunnerUseArtifact(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	doorMonster := &DoorCard{Id: 100, Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	nextDoor := &DoorCard{Id: 101, Type: DoorObstacle, Name: "Wall", Resources: []ResourceType{Jump}}

	dungeon := &Dungeon{
		Boss:  &BossMat{Id: 999, Name: "Boss"},
		Doors: []DungeonCard{nextDoor},
	}

	artAxe := &ArtifactCard{
		Id:          1,
		Color:       Red,
		Name:        "Battle Axe",
		Action:      BattleAxeArtifact{},
		MultiAction: true,
	}

	game := &Game{
		Players:      []*Player{p1},
		Dungeon:      dungeon,
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, game)
	game.PlayField.Artifacts = []*ArtifactCard{artAxe}
	registerCardsInTestGame(game)

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	// Use Battle Axe First Action (defeat monster)
	err := runner.UseArtifact(ctx, p1.Id, artAxe.Id, FirstArtifactAction, doorMonster.ID())
	if err != nil {
		t.Fatalf("failed to use artifact via runner: %v", err)
	}

	expectEvent(t, sub, func(e Event) bool {
		evt, ok := e.(DoorDefeatedEvent)
		return ok && evt.CardID == doorMonster.ID()
	})

	if game.PlayField.HasActiveDoor(doorMonster) {
		t.Error("expected doorMonster to be defeated")
	}
}

func TestRunnerPlayCardTargeting(t *testing.T) {
	p1, _ := NewPlayer("Robin", Ranger, true)
	p2, _ := NewPlayer("Arthur", Paladin, true)

	snipeCard := &ActionCard{Id: 501, Name: "Snipe", Action: SnipeAction{}}
	p1.Hand = []PlayerCard{snipeCard}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Id: 502, Resources: []ResourceType{Sword}}}}

	doorPerson1 := &DoorCard{Id: 101, Type: DoorPerson, Name: "Guard 1"}
	doorPerson2 := &DoorCard{Id: 102, Type: DoorPerson, Name: "Guard 2"}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	game.PlayField.OpenedDoors = []DungeonCard{doorPerson1, doorPerson2}
	registerCardsInTestGame(game)

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	// Play Snipe explicitly targeting doorPerson2
	err := runner.PlayCardWithCardTarget(ctx, p1.Id, snipeCard.ID(), doorPerson2.ID())
	if err != nil {
		t.Fatalf("failed to play Snipe with target: %v", err)
	}

	expectEvent(t, sub, func(e Event) bool {
		evt, ok := e.(DoorDefeatedEvent)
		return ok && evt.CardID == doorPerson2.ID()
	})

	if game.PlayField.HasActiveDoor(doorPerson2) {
		t.Error("expected doorPerson2 to be defeated")
	}
	if !game.PlayField.HasActiveDoor(doorPerson1) {
		t.Error("expected doorPerson1 to remain active")
	}
}

func TestRunnerSubmitPromptChoice(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	p2, _ := NewPlayer("Conan", Barbarian, true)

	c1 := &ResourceCard{Id: 101, Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Id: 102, Resources: []ResourceType{Shield}}
	p1.Hand = []PlayerCard{c1, c2}

	eventCard := &EventCard{Id: 901, Name: "A Boo-Boo", Action: ABooBooEvent{}}
	interaction := &PlayerDiscardCardsInteraction{
		Kind:             InteractionPlayerDiscardCards,
		Card:             eventCard,
		requiredCounts:   map[*Player]int{p1: 1},
		PendingPlayers:   map[*Player]bool{p1: true},
		CollectedChoices: make(map[*Player][]PlayerCard),
		OnComplete:       ABooBooEvent{}.Execute,
	}

	nextDoor := &DoorCard{Id: 902, Type: DoorMonster, Name: "Goblin"}
	dungeon := &Dungeon{
		Boss:  &BossMat{Id: 999, Name: "Boss"},
		Doors: []DungeonCard{nextDoor},
	}

	game := &Game{
		Players:            []*Player{p1, p2},
		Dungeon:            dungeon,
		PlayField:          NewPlayfield(),
		Status:             Playing,
		PendingInteraction: interaction,
	}
	game.PlayField.OpenedDoors = []DungeonCard{eventCard}
	registerCardsInTestGame(game)

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	// Submit discard prompt choice
	err := runner.SubmitPromptChoiceCards(ctx, p1.Id, []CardID{c1.ID()})
	if err != nil {
		t.Fatalf("failed to submit prompt choice via runner: %v", err)
	}

	expectEvent(t, sub, func(e Event) bool {
		evt, ok := e.(CardDiscardedEvent)
		return ok && evt.ByPlayerID == p1.Id && evt.CardID == c1.ID()
	})

	if game.PendingInteraction != nil {
		t.Error("expected pending interaction to be cleared")
	}
	if p1.HasCardInHand(c1) {
		t.Error("expected c1 to be discarded from hand")
	}
}

func TestRunnerSnapshotQuery(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Id: 101, Resources: []ResourceType{Sword}}
	paladin.Hand = []PlayerCard{c1}

	door := &DoorCard{Id: 201, Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	boss := &BossMat{Id: 202, Name: "Boss"}

	game := &Game{
		Players:   []*Player{paladin},
		Dungeon:   &Dungeon{Boss: boss, Doors: []DungeonCard{}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	game.PlayField.OpenedDoors = []DungeonCard{door}
	registerCardsInTestGame(game)

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = runner.Run(ctx)
	}()

	snapshot, err := runner.Snapshot(ctx)
	if err != nil {
		t.Fatalf("failed to get snapshot via runner: %v", err)
	}

	if snapshot.Status != Playing {
		t.Errorf("expected Playing, got %v", snapshot.Status)
	}
	if len(snapshot.Players) != 1 || snapshot.Players[0].Id != paladin.Id {
		t.Errorf("unexpected players in snapshot: %+v", snapshot.Players)
	}
	if len(snapshot.OpenedDoors) != 1 || snapshot.OpenedDoors[0].Id != door.ID() {
		t.Errorf("unexpected opened doors in snapshot: %+v", snapshot.OpenedDoors)
	}
	if snapshot.Boss == nil || snapshot.Boss.Id != boss.ID() {
		t.Errorf("unexpected boss in snapshot: %+v", snapshot.Boss)
	}
}

func TestRunnerInvalidIDsErrorHandling(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Id: 101, Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{c1}

	game := &Game{
		Players:      []*Player{p1},
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	registerCardsInTestGame(game)

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = runner.Run(ctx)
	}()

	// 1. Invalid player ID in PlayCard
	err := runner.PlayCardSimple(ctx, "UnknownPlayer", c1.ID())
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("expected player not found error, got: %v", err)
	}

	// 2. Invalid card ID in DiscardCards
	err = runner.DiscardCards(ctx, p1.Id, []CardID{9999})
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("expected card not found error, got: %v", err)
	}

	// 3. Invalid player ID in UseHeroAbility
	err = runner.UseHeroAbilitySimple(ctx, "UnknownPlayer", []CardID{c1.ID()})
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("expected player not found error, got: %v", err)
	}

	// 4. Invalid artifact ID in UseArtifact
	err = runner.UseArtifact(ctx, p1.Id, 9999, FirstArtifactAction, 0)
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("expected artifact not found error, got: %v", err)
	}
}

func TestRunnerCommandsOnTerminatedRunner(t *testing.T) {
	game := NewGame()
	runner := NewRunner(game)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Run(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-errCh

	// Any command after termination should return GameTerminatedError
	err := runner.PlayCardSimple(context.Background(), "Arthur", 1)
	var termErr *GameTerminatedError
	if !errors.As(err, &termErr) {
		t.Errorf("expected GameTerminatedError for PlayCard, got: %v", err)
	}

	err = runner.DiscardCards(context.Background(), "Arthur", []CardID{1})
	if !errors.As(err, &termErr) {
		t.Errorf("expected GameTerminatedError for DiscardCards, got: %v", err)
	}

	err = runner.UseHeroAbilitySimple(context.Background(), "Arthur", []CardID{1})
	if !errors.As(err, &termErr) {
		t.Errorf("expected GameTerminatedError for UseHeroAbility, got: %v", err)
	}

	err = runner.UseArtifact(context.Background(), "Arthur", 1, FirstArtifactAction, 0)
	if !errors.As(err, &termErr) {
		t.Errorf("expected GameTerminatedError for UseArtifact, got: %v", err)
	}

	err = runner.SubmitPromptChoiceCards(context.Background(), "Arthur", []CardID{1})
	if !errors.As(err, &termErr) {
		t.Errorf("expected GameTerminatedError for SubmitPromptChoice, got: %v", err)
	}

	_, err = runner.Snapshot(context.Background())
	if !errors.As(err, &termErr) {
		t.Errorf("expected GameTerminatedError for Snapshot, got: %v", err)
	}
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
	wizard, _ := NewPlayer("Gandalf", Wizard, true)
	paladin, _ := NewPlayer("Arthur", Paladin, true)

	game := &Game{
		Players:   []*Player{wizard, paladin},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := runner.Subscribe()

	go func() {
		_ = runner.Run(ctx)
	}()

	c1 := &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}}
	c2 := &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}}
	c3 := &ResourceCard{Id: nextCardId(), Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{c1, c2, c3}
	registerCardsInTestGame(game)

	err := runner.UseHeroAbilitySimple(ctx, wizard.Id, []CardID{c1.ID(), c2.ID(), c3.ID()})
	if err != nil {
		t.Fatalf("failed to use ability via runner: %v", err)
	}

	// Expect HeroAbilityUsedEvent, 3x CardDiscardedEvent, TimeFrozenEvent broadcasted
	expectEvent(t, sub, func(e Event) bool {
		evt, ok := e.(HeroAbilityUsedEvent)
		return ok && evt.ByPlayerID == wizard.Id
	})
	expectEvent(t, sub, func(e Event) bool { _, ok := e.(TimeFrozenEvent); return ok })

	if !game.IsTimeFrozen {
		t.Error("expected game.isTimeFrozen to be true")
	}
}

// --- Test Helpers ---

func expectEvent(t *testing.T, sub <-chan Event, predicate func(Event) bool) {
	t.Helper()
	timeout := time.After(2 * time.Minute)
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
