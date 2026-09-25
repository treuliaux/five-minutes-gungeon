package game

import (
	"errors"
	"slices"
	"testing"
)

// --- Lifecycle & Zone Management Tests ---

func TestCurseActivationLifecycle(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Test Passive Curse",
		Effect: TimeCannotBeStopped,
	}

	game := &Game{
		Players:   []*Player{p1},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	events, err := game.PlayField.AddDungeonCard(curse, game)
	if err != nil {
		t.Fatalf("failed to add curse to playfield: %v", err)
	}

	// Should be placed in ActiveCurses, not OpenedDoors
	if !slices.Contains(game.PlayField.ActiveCurses, curse) {
		t.Errorf("expected curse to be in ActiveCurses")
	}
	if slices.Contains(game.PlayField.OpenedDoors, DungeonCard(curse)) {
		t.Errorf("curse should not be in OpenedDoors")
	}

	// Should emit DoorOpenedEvent and CurseActivatedEvent
	var doorOpenedFound, curseActivatedFound bool
	for _, e := range events {
		if doe, ok := e.(DoorOpenedEvent); ok && doe.DungeonCard == curse {
			doorOpenedFound = true
		}
		if cae, ok := e.(CurseActivatedEvent); ok && cae.Card == curse {
			curseActivatedFound = true
		}
	}
	if !doorOpenedFound {
		t.Error("expected DoorOpenedEvent for curse")
	}
	if !curseActivatedFound {
		t.Error("expected CurseActivatedEvent for curse")
	}

	// HasActiveCurseEffect should return true
	if !game.HasActiveCurseEffect(TimeCannotBeStopped) {
		t.Error("expected HasActiveCurseEffect to be true")
	}
}

func TestCurseRemovalLifecycle(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Test Passive Curse",
		Effect: TimeCannotBeStopped,
	}

	game := &Game{
		Players:   []*Player{p1},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)

	events, err := game.PlayField.RemoveDungeonCard(game, curse)
	if err != nil {
		t.Fatalf("failed to remove curse: %v", err)
	}

	if slices.Contains(game.PlayField.ActiveCurses, curse) {
		t.Errorf("expected curse to be removed from ActiveCurses")
	}
	if game.HasActiveCurseEffect(TimeCannotBeStopped) {
		t.Errorf("expected HasActiveCurseEffect to be false after removal")
	}

	var curseRemovedFound bool
	for _, e := range events {
		if cre, ok := e.(CurseRemovedEvent); ok && cre.Card == curse {
			curseRemovedFound = true
		}
	}
	if !curseRemovedFound {
		t.Error("expected CurseRemovedEvent upon curse removal")
	}
}

func TestOpenDoorContinuesOnCurse(t *testing.T) {
	curse := &CurseCard{Type: ChallengeCurse, Name: "A Curse", Effect: TimeCannotBeStopped}
	door := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{door, curse}, // curse is on top (end of slice)
	}

	game := &Game{
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	events, err := game.OpenDoor()
	if err != nil {
		t.Fatalf("OpenDoor failed: %v", err)
	}

	// Both curse and door should have been drawn
	if !slices.Contains(game.PlayField.ActiveCurses, curse) {
		t.Error("expected curse to be active")
	}
	if !game.PlayField.HasActiveDoor(door) {
		t.Error("expected door to be opened")
	}
	if len(game.PlayField.OpenedDoors) != 1 {
		t.Errorf("expected 1 opened door, got %d", len(game.PlayField.OpenedDoors))
	}

	var curseActivatedFound, doorOpenedCount int
	for _, e := range events {
		if _, ok := e.(CurseActivatedEvent); ok {
			curseActivatedFound++
		}
		if _, ok := e.(DoorOpenedEvent); ok {
			doorOpenedCount++
		}
	}
	if curseActivatedFound != 1 {
		t.Errorf("expected 1 CurseActivatedEvent, got %d", curseActivatedFound)
	}
	if doorOpenedCount != 2 {
		t.Errorf("expected 2 DoorOpenedEvent, got %d", doorOpenedCount)
	}
}

// --- Specific Curse Effects Tests ---

func TestCurseTimeCannotBeStopped(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	p1.Hand = []PlayerCard{c1, c2}
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2, c1, c2, c1}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Clock Blocked",
		Effect: TimeCannotBeStopped,
	}

	game := &Game{
		Players:   []*Player{p1, p1},
		HandSize:  5,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)

	events, err := game.StopTime(p1)
	if err == nil {
		t.Fatal("expected error when stopping time while TimeCannotBeStopped is active")
	}

	var curseErr *CurseRuleViolatedError
	if !errors.As(err, &curseErr) {
		t.Fatalf("expected *CurseRuleViolatedError, got %T: %v", err, err)
	}
	if curseErr.Curse != TimeCannotBeStopped {
		t.Errorf("expected TimeCannotBeStopped curse error, got %v", curseErr.Curse)
	}

	if game.IsTimeFrozen {
		t.Error("expected time not to be frozen")
	}

	// Voided hand penalty verification
	if p1.Discard.Length() != 0 {
		t.Errorf("expected voided cards to be destroyed, got discard length %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 5 {
		t.Errorf("expected hand to be refilled to 5, got %d", len(p1.Hand))
	}

	var voidEventFound bool
	for _, e := range events {
		if ve, ok := e.(PlayerHandVoidedEvent); ok && ve.Player == p1 {
			voidEventFound = true
			if len(ve.VoidedCards) != 2 {
				t.Errorf("expected 2 voided cards in event, got %d", len(ve.VoidedCards))
			}
		}
	}
	if !voidEventFound {
		t.Error("expected PlayerHandVoidedEvent in events")
	}
}

func TestCurseActionsCannotBePlayed(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	actionCard := &ActionCard{Name: "Holy Hand Grenade", Action: HolyHandGrenadeAction{}}
	resourceCard := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{actionCard, resourceCard}
	p1.Deck = &Deck{Cards: []PlayerCard{resourceCard, resourceCard, resourceCard, resourceCard, resourceCard}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Sheepified!",
		Effect: ActionsCannotBePlayed,
	}
	door := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{p1, p1},
		HandSize:  5,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door, game)

	// Playing action card should fail with curse violation
	events, err := game.Apply(PlayCardCmd{Player: p1, Card: actionCard})
	if err == nil {
		t.Fatal("expected error when playing action card under ActionsCannotBePlayed")
	}

	var curseErr *CurseRuleViolatedError
	if !errors.As(err, &curseErr) {
		t.Fatalf("expected *CurseRuleViolatedError, got %T: %v", err, err)
	}
	if curseErr.Curse != ActionsCannotBePlayed {
		t.Errorf("expected ActionsCannotBePlayed curse error, got %v", curseErr.Curse)
	}

	// Action card must not be leaked into PlayField.Field
	if slices.Contains(game.PlayField.Field, PlayerCard(actionCard)) {
		t.Error("illegal action card must not be added to playfield")
	}

	// Door must still be alive
	if !game.PlayField.HasActiveDoor(door) {
		t.Error("door should not be defeated")
	}

	// Hand voided and refilled to 5
	if len(p1.Hand) != 5 {
		t.Errorf("expected hand to be refilled to 5, got %d", len(p1.Hand))
	}

	var voidFound bool
	for _, e := range events {
		if _, ok := e.(PlayerHandVoidedEvent); ok {
			voidFound = true
		}
	}
	if !voidFound {
		t.Error("expected PlayerHandVoidedEvent")
	}

	// Playing a resource card should succeed normally
	resCard := p1.Hand[0]
	_, err = game.Apply(PlayCardCmd{Player: p1, Card: resCard})
	if err != nil {
		t.Fatalf("playing resource card should succeed: %v", err)
	}
}

func TestCurseAbilitiesCannotBePlayed(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	p1.Hand = []PlayerCard{c1, c2, c3}
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c1, c2}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Gorgon's Gaze",
		Effect: AbilitiesCannotBePlayed,
	}
	door := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{p1, p1},
		HandSize:  5,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door, game)

	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       p1,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SmiteAbility{Target: door},
	})
	if err == nil {
		t.Fatal("expected error when using ability under AbilitiesCannotBePlayed")
	}

	var curseErr *CurseRuleViolatedError
	if !errors.As(err, &curseErr) {
		t.Fatalf("expected *CurseRuleViolatedError, got %T: %v", err, err)
	}
	if curseErr.Curse != AbilitiesCannotBePlayed {
		t.Errorf("expected AbilitiesCannotBePlayed curse error, got %v", curseErr.Curse)
	}

	// Door should still be active
	if !game.PlayField.HasActiveDoor(door) {
		t.Error("door should not be defeated by blocked ability")
	}

	// Discard should be 0 (cards destroyed in void, not discarded for ability)
	if p1.Discard.Length() != 0 {
		t.Errorf("expected 0 cards in discard, got %d", p1.Discard.Length())
	}

	// Hand refilled to 5
	if len(p1.Hand) != 5 {
		t.Errorf("expected hand refilled to 5, got %d", len(p1.Hand))
	}

	var voidFound bool
	for _, e := range events {
		if _, ok := e.(PlayerHandVoidedEvent); ok {
			voidFound = true
		}
	}
	if !voidFound {
		t.Error("expected PlayerHandVoidedEvent")
	}
}

func TestCurseHandSizeLimitedToThree(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	c4 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c5 := &ResourceCard{Resources: []ResourceType{Scroll}}
	p1.Hand = []PlayerCard{c1, c2, c3, c4, c5}
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c4, c5}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Cursed Blocks",
		Effect: HandSizeLimitedToThree,
	}
	door := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{p1, p1},
		HandSize:  5,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door, game)

	// TargetHandSize should now be 3
	if game.TargetHandSize() != 3 {
		t.Errorf("expected TargetHandSize to be 3, got %d", game.TargetHandSize())
	}

	// Playing a card while having >3 cards (5) should trigger HandSizeLimitedToThree violation
	events, err := game.Apply(PlayCardCmd{Player: p1, Card: c1})
	if err == nil {
		t.Fatal("expected error when playing card with >3 cards under HandSizeLimitedToThree")
	}

	var curseErr *CurseRuleViolatedError
	if !errors.As(err, &curseErr) {
		t.Fatalf("expected *CurseRuleViolatedError, got %T: %v", err, err)
	}
	if curseErr.Curse != HandSizeLimitedToThree {
		t.Errorf("expected HandSizeLimitedToThree curse error, got %v", curseErr.Curse)
	}

	// Hand must be refilled up to TargetHandSize (3)
	if len(p1.Hand) != 3 {
		t.Fatalf("expected hand to be refilled to 3, got %d", len(p1.Hand))
	}

	var voidFound bool
	for _, e := range events {
		if _, ok := e.(PlayerHandVoidedEvent); ok {
			voidFound = true
		}
	}
	if !voidFound {
		t.Error("expected PlayerHandVoidedEvent")
	}

	// Now holding 3 cards, playing a card should succeed and refill up to 3 cards
	cardToPlay := p1.Hand[0]
	_, err = game.Apply(PlayCardCmd{Player: p1, Card: cardToPlay})
	if err != nil {
		t.Fatalf("expected playing card with 3 cards in hand to succeed: %v", err)
	}
	if len(p1.Hand) != 3 {
		t.Errorf("expected hand to stay refilled to 3 after play, got %d", len(p1.Hand))
	}
}

func TestCurseHandSizeLimitedToThreeBlocksHeroAbility(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	c4 := &ResourceCard{Resources: []ResourceType{Arrow}}
	p1.Hand = []PlayerCard{c1, c2, c3, c4} // 4 cards > 3
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c4, c1}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Cursed Blocks",
		Effect: HandSizeLimitedToThree,
	}
	door := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{p1, p1},
		HandSize:  5,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door, game)

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       p1,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SmiteAbility{Target: door},
	})
	if err == nil {
		t.Fatal("expected error when using ability with >3 cards under HandSizeLimitedToThree")
	}

	var curseErr *CurseRuleViolatedError
	if !errors.As(err, &curseErr) {
		t.Fatalf("expected *CurseRuleViolatedError, got %T: %v", err, err)
	}
	if curseErr.Curse != HandSizeLimitedToThree {
		t.Errorf("expected HandSizeLimitedToThree curse error, got %v", curseErr.Curse)
	}

	if len(p1.Hand) != 3 {
		t.Errorf("expected hand refilled to 3, got %d", len(p1.Hand))
	}
}

func TestCurseFlippedHeroMat(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	p2, _ := NewPlayer("Robin", Ranger, true)

	if p1.Hero.Class != Paladin || p2.Hero.Class != Ranger {
		t.Fatal("heroes should have initial classes")
	}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Cursed Cosplayers",
		Effect: FlippedHeroMat,
		Apply: func(ctx Context) ([]Event, error) {
			var events []Event
			for _, p := range ctx.Engine().ListPlayers() {
				events = append(events, p.FlipHeroMat()...)
			}
			return events, nil
		},
		Cure: func(ctx Context) ([]Event, error) {
			var events []Event
			for _, p := range ctx.Engine().ListPlayers() {
				events = append(events, p.FlipHeroMat()...)
			}
			return events, nil
		},
	}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	// Add curse -> triggers Apply hook -> flips hero mats
	events, err := game.PlayField.AddDungeonCard(curse, game)
	if err != nil {
		t.Fatalf("failed to add curse: %v", err)
	}

	if p1.Hero.Class != Valkyrie || p2.Hero.Class != Huntress {
		t.Errorf("expected hero classes to be Valkyrie and Huntress after flip, got P1: %v, P2: %v",
			p1.Hero.Class, p2.Hero.Class)
	}

	var heroMatFlippedCount int
	for _, e := range events {
		if _, ok := e.(HeroMatFlippedEvent); ok {
			heroMatFlippedCount++
		}
	}
	if heroMatFlippedCount != 2 {
		t.Errorf("expected 2 HeroMatFlippedEvent, got %d", heroMatFlippedCount)
	}

	// Remove curse -> triggers Cure hook -> reverts hero mats
	cureEvents, err := game.PlayField.RemoveDungeonCard(game, curse)
	if err != nil {
		t.Fatalf("failed to remove curse: %v", err)
	}

	if p1.Hero.Class != Paladin || p2.Hero.Class != Ranger {
		t.Errorf("expected hero classes to revert to Paladin and Ranger after cure, got P1: %v, P2: %v",
			p1.Hero.Class, p2.Hero.Class)
	}

	heroMatFlippedCount = 0
	for _, e := range cureEvents {
		if _, ok := e.(HeroMatFlippedEvent); ok {
			heroMatFlippedCount++
		}
	}
	if heroMatFlippedCount != 2 {
		t.Errorf("expected 2 HeroMatFlippedEvent on cure, got %d", heroMatFlippedCount)
	}
}

func TestCurseThreeDiscardsWhenTimeStops(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	p2, _ := NewPlayer("Robin", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	c4 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c5 := &ResourceCard{Resources: []ResourceType{Scroll}}

	p1.Hand = []PlayerCard{c1, c2, c3, c4, c5}
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c4, c5}}
	p2.Hand = []PlayerCard{c1, c2, c3, c4, c5}
	p2.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c4, c5}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Conga-Rats!",
		Effect: ThreeDiscardsWhenTimeStops,
		Cure: func(ctx Context) ([]Event, error) {
			ctx.Engine().ClearStopTimeCurse()
			return nil, nil
		},
	}
	door := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:                []*Player{p1, p2},
		HandSize:               5,
		PlayField:              NewPlayfield(),
		Status:                 Playing,
		CurseExpectingDiscards: make(map[*Player]int),
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door, game)

	// Stopping time activates debt of 3 discards per player
	_, err := game.StopTime(p1)
	if err != nil {
		t.Fatalf("expected StopTime to succeed: %v", err)
	}

	if game.CurseExpectingDiscards[p1] != 3 || game.CurseExpectingDiscards[p2] != 3 {
		t.Fatalf("expected 3 discards debt for both players, got P1: %d, P2: %d",
			game.CurseExpectingDiscards[p1], game.CurseExpectingDiscards[p2])
	}

	// P1 attempts to play card without fulfilling discard debt -> triggers penalty
	events, err := game.Apply(PlayCardCmd{Player: p1, Card: p1.Hand[0]})
	if err == nil {
		t.Fatal("expected error when playing card while discard debt is active")
	}

	var curseErr *CurseRuleViolatedError
	if !errors.As(err, &curseErr) {
		t.Fatalf("expected *CurseRuleViolatedError, got %T: %v", err, err)
	}
	if curseErr.Curse != ThreeDiscardsWhenTimeStops {
		t.Errorf("expected ThreeDiscardsWhenTimeStops error, got %v", curseErr.Curse)
	}

	// P1 debt should be cleared after voiding
	if _, exists := game.CurseExpectingDiscards[p1]; exists {
		t.Error("expected P1 debt to be cleared after voiding")
	}

	// P2 satisfies debt progressively via DiscardCardsCmd
	_, err = game.Apply(DiscardCardsCmd{Player: p2, Cards: []PlayerCard{p2.Hand[0]}})
	if err != nil {
		t.Fatalf("discard 1 card failed: %v", err)
	}
	if game.CurseExpectingDiscards[p2] != 2 {
		t.Errorf("expected P2 debt to be 2, got %d", game.CurseExpectingDiscards[p2])
	}

	_, err = game.Apply(DiscardCardsCmd{Player: p2, Cards: []PlayerCard{p2.Hand[0], p2.Hand[1]}})
	if err != nil {
		t.Fatalf("discard 2 cards failed: %v", err)
	}
	if _, exists := game.CurseExpectingDiscards[p2]; exists {
		t.Errorf("expected P2 debt to be cleared after discarding 3 cards in total")
	}

	// P2 can now play cards without penalty
	_, err = game.Apply(PlayCardCmd{Player: p2, Card: p2.Hand[0]})
	if err != nil {
		t.Fatalf("P2 should be able to play card after debt cleared: %v", err)
	}

	var voidFound bool
	for _, e := range events {
		if _, ok := e.(PlayerHandVoidedEvent); ok {
			voidFound = true
		}
	}
	if !voidFound {
		t.Error("expected PlayerHandVoidedEvent for P1")
	}
}

func TestCurseThreeDiscardsWhenTimeStopsCured(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	p2, _ := NewPlayer("Robin", Ranger, true)

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Conga-Rats!",
		Effect: ThreeDiscardsWhenTimeStops,
		Cure: func(ctx Context) ([]Event, error) {
			ctx.Engine().ClearStopTimeCurse()
			return nil, nil
		},
	}

	game := &Game{
		Players:                []*Player{p1, p2},
		HandSize:               5,
		PlayField:              NewPlayfield(),
		Status:                 Playing,
		CurseExpectingDiscards: make(map[*Player]int),
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)

	_, err := game.StopTime(p1)
	if err != nil {
		t.Fatalf("StopTime failed: %v", err)
	}
	if len(game.CurseExpectingDiscards) != 2 {
		t.Fatalf("expected 2 players with debts, got %d", len(game.CurseExpectingDiscards))
	}

	// Cure curse -> Cure hook calls ClearStopTimeCurse()
	_, err = game.PlayField.RemoveDungeonCard(game, curse)
	if err != nil {
		t.Fatalf("RemoveDungeonCard failed: %v", err)
	}

	if len(game.CurseExpectingDiscards) != 0 {
		t.Errorf("expected debts map to be empty after cure, got %d", len(game.CurseExpectingDiscards))
	}
}

func TestCurseDoorsOpenInPairs(t *testing.T) {
	d1 := &DoorCard{Type: DoorMonster, Name: "Monster 1", Resources: []ResourceType{Sword}}
	d2 := &DoorCard{Type: DoorObstacle, Name: "Obstacle 2", Resources: []ResourceType{Jump}}
	d3 := &DoorCard{Type: DoorPerson, Name: "Person 3", Resources: []ResourceType{Arrow}}
	boss := &BossMat{Name: "Boss"}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Endless Ambush",
		Effect: DoorsOpenInPairs,
	}

	dungeon := &Dungeon{
		Boss:  boss,
		Doors: []DungeonCard{d3, d2, d1}, // d1 is on top (end of slice), then d2
	}

	game := &Game{
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)

	// OpenDoor under DoorsOpenInPairs should open 2 doors
	_, err := game.OpenDoor()
	if err != nil {
		t.Fatalf("OpenDoor failed: %v", err)
	}

	if len(game.PlayField.OpenedDoors) != 2 {
		t.Fatalf("expected 2 opened doors on playfield, got %d", len(game.PlayField.OpenedDoors))
	}
	if !game.PlayField.HasActiveDoor(d1) || !game.PlayField.HasActiveDoor(d2) {
		t.Errorf("expected d1 and d2 to be opened, got %v", game.PlayField.OpenedDoors)
	}
}

func TestDruidForestSpiritsCyclesCurseToBottom(t *testing.T) {
	druid, _ := NewPlayer("Malfurion", Druid, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	druid.Hand = []PlayerCard{c1, c2, c3}
	druid.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c1, c2}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Cursed Blocks",
		Effect: HandSizeLimitedToThree,
	}
	door1 := &DoorCard{Type: DoorMonster, Name: "Monster 1", Resources: []ResourceType{Sword}}
	door2 := &DoorCard{Type: DoorMonster, Name: "Monster 2", Resources: []ResourceType{Sword}}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{door1, door2},
	}

	game := &Game{
		Players:   []*Player{druid, druid},
		HandSize:  5,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door1, game)

	// Druid uses ForestSpiritsAbility targeting curse
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       druid,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      ForestSpiritsAbility{Target: curse},
	})
	if err != nil {
		t.Fatalf("ForestSpiritsAbility failed: %v", err)
	}

	// Curse should be removed from ActiveCurses
	if slices.Contains(game.PlayField.ActiveCurses, curse) {
		t.Error("expected curse to be removed from ActiveCurses")
	}

	// Curse should be at the bottom of dungeon deck
	if len(dungeon.Doors) == 0 || dungeon.Doors[0] != curse {
		t.Errorf("expected curse to be at the bottom of dungeon doors, got %v", dungeon.Doors)
	}

	// CurseRemovedEvent and DungeonCardSentToBottomEvent should be emitted
	var curseRemovedFound, sentToBottomFound bool
	for _, e := range events {
		if cre, ok := e.(CurseRemovedEvent); ok && cre.Card == curse {
			curseRemovedFound = true
		}
		if sbe, ok := e.(DungeonCardSentToBottomEvent); ok && sbe.Card == curse {
			sentToBottomFound = true
		}
	}
	if !curseRemovedFound {
		t.Error("expected CurseRemovedEvent in events")
	}
	if !sentToBottomFound {
		t.Error("expected DungeonCardSentToBottomEvent in events")
	}
}

func TestCurseQueryWithDoorsFilter(t *testing.T) {
	curse := &CurseCard{Type: ChallengeCurse, Name: "A Curse", Effect: AbilitiesCannotBePlayed}
	door := &DoorCard{Type: DoorMonster, Name: "A Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	_, _ = game.PlayField.AddDungeonCard(door, game)

	// Filter with AddCurses
	curseOnly := game.ActiveDoors(NewDoorsFilter().AddCurses())
	if len(curseOnly) != 1 || curseOnly[0] != curse {
		t.Errorf("expected 1 curse in ActiveDoors with AddCurses, got %v", curseOnly)
	}

	// Filter with AddDoors
	doorsOnly := game.ActiveDoors(NewDoorsFilter().AddDoors(DoorMonster))
	if len(doorsOnly) != 1 || doorsOnly[0] != door {
		t.Errorf("expected 1 door in ActiveDoors with AddDoors, got %v", doorsOnly)
	}
}

func TestPassiveSocialCurses(t *testing.T) {
	socialCurses := []*CurseCard{
		{Type: ChallengeCurse, Name: "Blinding Light", Effect: HandsHidden},
		{Type: ChallengeCurse, Name: "House Rules!", Effect: PlayersHandFacingAway},
		{Type: ChallengeCurse, Name: "Cursed Blanket", Effect: PlayersCanOnlyUseOneHandToPlay},
		{Type: ChallengeCurse, Name: "Waffles Waffles !", Effect: PlayersMustOnlySayWaffles},
		{Type: ChallengeCurse, Name: "A Boot-Alion of Kittens", Effect: PlayersMustOnlySayMeow},
	}

	for _, curse := range socialCurses {
		t.Run(curse.Name, func(t *testing.T) {
			game := &Game{
				PlayField: NewPlayfield(),
				Status:    Playing,
			}

			addEvents, err := game.PlayField.AddDungeonCard(curse, game)
			if err != nil {
				t.Fatalf("failed to add %s: %v", curse.Name, err)
			}
			if !slices.Contains(game.PlayField.ActiveCurses, curse) {
				t.Errorf("expected %s to be in ActiveCurses", curse.Name)
			}
			if !game.HasActiveCurseEffect(curse.Effect) {
				t.Errorf("expected HasActiveCurseEffect(%v) to be true", curse.Effect)
			}

			var activatedFound bool
			for _, e := range addEvents {
				if cae, ok := e.(CurseActivatedEvent); ok && cae.Card == curse {
					activatedFound = true
				}
			}
			if !activatedFound {
				t.Errorf("expected CurseActivatedEvent for %s", curse.Name)
			}

			remEvents, err := game.PlayField.RemoveDungeonCard(game, curse)
			if err != nil {
				t.Fatalf("failed to remove %s: %v", curse.Name, err)
			}
			if slices.Contains(game.PlayField.ActiveCurses, curse) {
				t.Errorf("expected %s to be removed from ActiveCurses", curse.Name)
			}

			var removedFound bool
			for _, e := range remEvents {
				if cre, ok := e.(CurseRemovedEvent); ok && cre.Card == curse {
					removedFound = true
				}
			}
			if !removedFound {
				t.Errorf("expected CurseRemovedEvent for %s", curse.Name)
			}
		})
	}
}
