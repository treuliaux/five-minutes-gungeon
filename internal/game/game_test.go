package game

import (
	"fmt"
	"testing"
	"time"
)

// --- Game Initialization & Player Management Tests ---

func TestNewGameInitialization(t *testing.T) {
	game := NewGame()

	if game == nil {
		t.Fatal("expected NewGame() to return non-nil game instance")
	}
	if game.Status != Waiting {
		t.Errorf("expected initial status to be Waiting, got %v", game.Status)
	}
	if len(game.players) != 0 {
		t.Errorf("expected 0 players initially, got %d", len(game.players))
	}
	if game.handSize != 0 {
		t.Errorf("expected 0 handSize initially, got %d", game.handSize)
	}
}

func TestGameAddPlayerSuccess(t *testing.T) {
	game := NewGame()

	events, err := game.Apply(AddPlayerCmd{
		Name:  "Sir Lancelot",
		Class: Paladin,
	})
	if err != nil {
		t.Fatalf("expected player to be added successfully, got error: %v", err)
	}
	if len(game.players) != 1 {
		t.Fatalf("expected 1 player in game, got %d", len(game.players))
	}
	if game.players[0].name != "Sir Lancelot" {
		t.Errorf("expected player name 'Sir Lancelot', got '%s'", game.players[0].name)
	}
	if game.players[0].hero.class != Paladin {
		t.Errorf("expected player hero class Paladin, got %v", game.players[0].hero.class)
	}

	// Verify PlayerAddedEvent
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	addedEvt, ok := events[0].(PlayerAddedEvent)
	if !ok {
		t.Fatalf("expected PlayerAddedEvent, got %T", events[0])
	}
	if addedEvt.Player != game.players[0] {
		t.Errorf("expected event Player to match created player")
	}
}

func TestGameAddPlayerDuplicateClassRejected(t *testing.T) {
	game := NewGame()

	_, err := game.Apply(AddPlayerCmd{Name: "Paladin 1", Class: Paladin})
	if err != nil {
		t.Fatalf("failed adding first Paladin: %v", err)
	}

	_, err = game.Apply(AddPlayerCmd{Name: "Paladin 2", Class: Paladin})
	if err == nil {
		t.Error("expected error when adding duplicate class Paladin, got nil")
	}
	if len(game.players) != 1 {
		t.Errorf("expected game to still have 1 player, got %d", len(game.players))
	}
}

func TestGameAddPlayerWhenNotWaitingRejected(t *testing.T) {
	game := NewGame()
	_, _ = game.Apply(AddPlayerCmd{Name: "P1", Class: Paladin})
	_, _ = game.Apply(AddPlayerCmd{Name: "P2", Class: Barbarian})
	_, _ = game.Apply(StartCmd{})

	_, err := game.Apply(AddPlayerCmd{Name: "P3", Class: Valkyrie})
	if err == nil {
		t.Error("expected error when adding player while game is Playing, got nil")
	}
}

func TestGameAddPlayerUnknownClassRejected(t *testing.T) {
	game := NewGame()

	_, err := game.Apply(AddPlayerCmd{Name: "Invalid", Class: HeroClass(999)})
	if err == nil {
		t.Error("expected error for unknown hero class, got nil")
	}
}

// --- Game Start & Hand Size Determination Tests ---

func TestGameStartPlayerCountBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		playerCount int
		shouldError bool
	}{
		{"0 players", 0, true},
		{"1 player", 1, true},
		{"2 players", 2, false},
		{"3 players", 3, false},
		{"4 players", 4, false},
		{"5 players", 5, false},
		{"6 players", 6, false},
	}

	classes := []HeroClass{Paladin, Barbarian, Gladiator, Valkyrie, Sorceress, Wizard}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			for i := 0; i < tt.playerCount; i++ {
				_, err := game.Apply(AddPlayerCmd{
					Name:  fmt.Sprintf("Player %d", i+1),
					Class: classes[i],
				})
				if err != nil {
					t.Fatalf("unexpected error adding player: %v", err)
				}
			}

			_, err := game.Apply(StartCmd{})
			if tt.shouldError && err == nil {
				t.Errorf("expected error starting with %d players, got nil", tt.playerCount)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected success starting with %d players, got error: %v", tt.playerCount, err)
			}
		})
	}
}

func TestGameCannotStartWithMoreThanSixPlayers(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin)
	p2, _ := NewPlayer("P2", Barbarian)
	p3, _ := NewPlayer("P3", Gladiator)
	p4, _ := NewPlayer("P4", Valkyrie)
	p5, _ := NewPlayer("P5", Sorceress)
	p6, _ := NewPlayer("P6", Wizard)
	p7, _ := NewPlayer("P7", Huntress)

	game := &Game{
		players: []*Player{p1, p2, p3, p4, p5, p6, p7},
		Status:  Waiting,
	}

	_, err := game.Apply(StartCmd{})
	if err == nil {
		t.Error("expected error starting game with 7 players, got nil")
	}
}

func TestGameCannotStartTwice(t *testing.T) {
	game := NewGame()
	_, _ = game.Apply(AddPlayerCmd{Name: "P1", Class: Paladin})
	_, _ = game.Apply(AddPlayerCmd{Name: "P2", Class: Barbarian})

	events, err := game.Apply(StartCmd{})
	if err != nil {
		t.Fatalf("unexpected error on first start: %v", err)
	}
	if game.Status != Playing {
		t.Errorf("expected game status Playing, got %v", game.Status)
	}
	if len(events) < 2 {
		t.Fatalf("expected at least GameStartedEvent and DoorOpenedEvent, got %d events", len(events))
	}
	if _, ok := events[0].(GameStartedEvent); !ok {
		t.Errorf("expected first event to be GameStartedEvent, got %T", events[0])
	}
	if _, ok := events[1].(DoorOpenedEvent); !ok {
		t.Errorf("expected second event to be DoorOpenedEvent, got %T", events[1])
	}

	// Attempt second start
	_, err = game.Apply(StartCmd{})
	if err == nil {
		t.Error("expected error when starting already playing game, got nil")
	}
}

func TestGameHandSizePerPlayerCount(t *testing.T) {
	tests := []struct {
		playerCount      int
		expectedHandSize int
	}{
		{2, 5},
		{3, 4},
		{4, 3},
		{5, 3},
		{6, 3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d players hand size %d", tt.playerCount, tt.expectedHandSize), func(t *testing.T) {
			game := NewGame()
			classes := []HeroClass{Paladin, Barbarian, Gladiator, Valkyrie, Sorceress, Wizard}
			for i := 0; i < tt.playerCount; i++ {
				// Use Paladin and Barbarian which have non-nil preset decks
				class := classes[i]
				p, _ := NewPlayer(fmt.Sprintf("P%d", i+1), class)
				// Ensure all players have a deck with enough cards for test
				p.deck = NewYellowDeck()
				p.deck.putAtop(
					&ResourceCard{resources: []ResourceType{Sword}},
					&ResourceCard{resources: []ResourceType{Shield}},
					&ResourceCard{resources: []ResourceType{Arrow}},
					&ResourceCard{resources: []ResourceType{Jump}},
				)
				game.players = append(game.players, p)
			}

			_, err := game.Apply(StartCmd{})
			if err != nil {
				t.Fatalf("failed to start: %v", err)
			}
			if game.handSize != tt.expectedHandSize {
				t.Errorf("expected handSize %d, got %d", tt.expectedHandSize, game.handSize)
			}
			for i, p := range game.players {
				if len(p.hand) != tt.expectedHandSize {
					t.Errorf("player %d hand size expected %d, got %d", i, tt.expectedHandSize, len(p.hand))
				}
			}
		})
	}
}

// --- Card Playing & Drawing Tests ---

func TestPlayCardSuccessAndAutoDraw(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	barbarian, _ := NewPlayer("Barbarian", Barbarian)

	// Custom decks with known cards
	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	cDraw := &ResourceCard{resources: []ResourceType{Arrow}}

	paladin.hand = []PlayerCard{c1, c2}
	paladin.deck = &Deck{cards: []PlayerCard{cDraw}}

	game := &Game{
		players:   []*Player{paladin, barbarian},
		handSize:  2,
		playField: *NewPlayfield(),
		Status:    Playing,
	}

	events, err := game.Apply(PlayCardCmd{Player: paladin, Card: c1})
	if err != nil {
		t.Fatalf("failed to play card: %v", err)
	}

	// Verify card was played on field
	if len(game.playField.field) != 1 || game.playField.field[0] != c1 {
		t.Errorf("expected card to be on playfield")
	}
	// Verify card was removed from hand and auto-drawn back to handSize 2
	if len(paladin.hand) != 2 {
		t.Fatalf("expected player hand to auto-draw back to 2, got %d", len(paladin.hand))
	}
	if paladin.hand[0] != c2 || paladin.hand[1] != cDraw {
		t.Errorf("expected hand to contain [c2, cDraw], got %v", paladin.hand)
	}
	if len(paladin.deck.cards) != 0 {
		t.Errorf("expected deck to be empty after draw, got %d", len(paladin.deck.cards))
	}

	// Verify CardPlayedEvent
	if len(events) < 1 {
		t.Fatalf("expected at least CardPlayedEvent")
	}
	playedEvt, ok := events[0].(CardPlayedEvent)
	if !ok {
		t.Fatalf("expected CardPlayedEvent, got %T", events[0])
	}
	if playedEvt.ByPlayer != paladin || playedEvt.Card != c1 {
		t.Errorf("mismatch in CardPlayedEvent attributes")
	}
}

func TestPlayCardNotInHandRejected(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	cInHand := &ResourceCard{resources: []ResourceType{Sword}}
	cNotInHand := &ResourceCard{resources: []ResourceType{Shield}}

	paladin.hand = []PlayerCard{cInHand}

	game := &Game{
		players:   []*Player{paladin},
		handSize:  1,
		playField: *NewPlayfield(),
		Status:    Playing,
	}

	_, err := game.Apply(PlayCardCmd{Player: paladin, Card: cNotInHand})
	if err == nil {
		t.Error("expected error when playing card not in hand, got nil")
	}
	if len(game.playField.field) != 0 {
		t.Errorf("expected field to be empty")
	}
	if len(paladin.hand) != 1 || paladin.hand[0] != cInHand {
		t.Errorf("expected hand to remain unchanged")
	}
}

func TestPlayCardUnfreezesTime(t *testing.T) {
	wizard, _ := NewPlayer("Wizard", Wizard)
	card := &ResourceCard{resources: []ResourceType{Scroll}}
	wizard.hand = []PlayerCard{card}

	game := &Game{
		players:      []*Player{wizard},
		handSize:     1,
		isTimeFrozen: true,
		playField:    *NewPlayfield(),
		Status:       Playing,
	}

	events, err := game.Apply(PlayCardCmd{Player: wizard, Card: card})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.isTimeFrozen {
		t.Errorf("expected time to be unfrozen after card play")
	}

	// Verify TimeUnfrozenEvent
	foundUnfrozen := false
	for _, evt := range events {
		if unfreezeEvt, ok := evt.(TimeUnfrozenEvent); ok {
			foundUnfrozen = true
			if unfreezeEvt.ByPlayer != wizard {
				t.Errorf("expected TimeUnfrozenEvent ByPlayer to be wizard")
			}
		}
	}
	if !foundUnfrozen {
		t.Errorf("expected TimeUnfrozenEvent to be emitted")
	}
}

// --- Card Discarding & Player Healing Tests ---

func TestDiscardCardSuccess(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}

	paladin.hand = []PlayerCard{c1, c2}

	game := &Game{
		players: []*Player{paladin},
		Status:  Playing,
	}

	events, err := game.Apply(DiscardCardCmd{Player: paladin, Card: c1})
	if err != nil {
		t.Fatalf("failed to discard: %v", err)
	}

	if len(paladin.hand) != 1 || paladin.hand[0] != c2 {
		t.Errorf("expected hand to only contain c2")
	}
	if len(paladin.discard) != 1 || paladin.discard[0] != c1 {
		t.Errorf("expected discard pile to contain c1")
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	discardEvt, ok := events[0].(CardDiscardedEvent)
	if !ok {
		t.Fatalf("expected CardDiscardedEvent, got %T", events[0])
	}
	if discardEvt.ByPlayer != paladin || discardEvt.Card != c1 {
		t.Errorf("mismatch in CardDiscardedEvent attributes")
	}
}

func TestDiscardCardNotInHandRejected(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	cInHand := &ResourceCard{resources: []ResourceType{Sword}}
	cNotInHand := &ResourceCard{resources: []ResourceType{Shield}}

	paladin.hand = []PlayerCard{cInHand}

	game := &Game{
		players: []*Player{paladin},
		Status:  Playing,
	}

	_, err := game.Apply(DiscardCardCmd{Player: paladin, Card: cNotInHand})
	if err == nil {
		t.Error("expected error when discarding card not in hand, got nil")
	}
	if len(paladin.discard) != 0 {
		t.Errorf("expected discard pile to be empty")
	}
}

func TestPlayerHealRecyclesDiscardPile(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Arrow}}

	paladin.discard = []PlayerCard{c1, c2, c3}
	paladin.deck = &Deck{cards: []PlayerCard{}}

	// Partial heal (amount 2: takes c2, c3 to deck, leaving c1 in discard)
	healed := paladin.heal(2)
	if healed != 2 {
		t.Errorf("expected 2 healed, got %d", healed)
	}
	if len(paladin.discard) != 1 || paladin.discard[0] != c1 {
		t.Errorf("expected discard to have [c1], got %v", paladin.discard)
	}
	if len(paladin.deck.cards) != 2 || paladin.deck.cards[0] != c2 || paladin.deck.cards[1] != c3 {
		t.Errorf("expected deck to have [c2, c3], got %v", paladin.deck.cards)
	}

	// Full heal (amount 0 or amount >= len(discard) heals all remaining)
	healed = paladin.heal(0)
	if healed != 1 {
		t.Errorf("expected 1 healed, got %d", healed)
	}
	if len(paladin.discard) != 0 {
		t.Errorf("expected discard pile to be empty after heal, got %d", len(paladin.discard))
	}
	if len(paladin.deck.cards) != 3 {
		t.Fatalf("expected deck to have 3 recycled cards, got %d", len(paladin.deck.cards))
	}
}

func TestGameHealPlayerEmitsEvent(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	paladin.discard = []PlayerCard{c1}
	paladin.deck = &Deck{cards: []PlayerCard{}}

	game := &Game{
		players: []*Player{paladin},
		Status:  Playing,
	}

	events, err := game.healPlayer(paladin, 1)
	if err != nil {
		t.Fatalf("unexpected error healing player: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	healedEvt, ok := events[0].(PlayerHealedEvent)
	if !ok {
		t.Fatalf("expected PlayerHealedEvent, got %T", events[0])
	}
	if healedEvt.Player != paladin || healedEvt.amount != 1 {
		t.Errorf("unexpected healed event values: %+v", healedEvt)
	}
}

// --- Playfield Matching & Door Resolution Tests ---

func TestPlayfieldResourceRequirementResolution(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorMonster,
		name:      "Orc",
		resources: []ResourceType{Sword, Sword, Shield},
	}
	_, _ = pf.addDungeonCard(door)

	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten initially")
	}

	// Play 1 Sword - not beaten
	p, _ := NewPlayer("P", Paladin)
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Sword}})
	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with 1/2 swords")
	}

	// Play Dual card (Sword + Shield) - total: 2 Swords, 1 Shield -> Beaten!
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Sword, Shield}})
	if !pf.isPlayfieldBeaten() {
		t.Error("expected playfield to be beaten with required 2 Swords and 1 Shield")
	}
}

func TestPlayfieldWildCardRequirementResolution(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorMonster,
		name:      "Chimera",
		resources: []ResourceType{Sword, Shield, Arrow},
	}
	_, _ = pf.addDungeonCard(door)

	p, _ := NewPlayer("P", Paladin)

	// Play 1 Sword - missing Shield and Arrow
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Sword}})
	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with only 1 Sword")
	}

	// Play 1 WildCard - covers Shield or Arrow, but 1 resource still missing
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{WildCard}})
	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with 1 Sword + 1 WildCard for 3 requirements")
	}

	// Play 2nd WildCard - total: 1 Sword + 2 WildCards -> fulfills Sword + Shield + Arrow!
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{WildCard}})
	if !pf.isPlayfieldBeaten() {
		t.Error("expected playfield to be beaten with 1 Sword and 2 WildCards")
	}
}

func TestPlayfieldWildCardMultipleDeficits(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorObstacle,
		name:      "Heavy Gate",
		resources: []ResourceType{Sword, Sword, Shield, Shield},
	}
	_, _ = pf.addDungeonCard(door)

	p, _ := NewPlayer("P", Paladin)

	// Play 1 Sword (deficit: 1 Sword, 2 Shields = 3 deficit)
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Sword}})

	// Play 2 WildCards (covers 2 out of 3 deficits, 1 deficit remains)
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{WildCard, WildCard}})
	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with 2 WildCards covering 3 deficits")
	}

	// Play 3rd WildCard (covers all 3 deficits) -> Beaten!
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{WildCard}})
	if !pf.isPlayfieldBeaten() {
		t.Error("expected playfield to be beaten when WildCards cover all multiple resource deficits")
	}
}

func TestPlayfieldWildCardMultiDoorResolution(t *testing.T) {
	pf := NewPlayfield()

	d1 := &DoorCard{Type: DoorMonster, name: "Goblin", resources: []ResourceType{Sword, Jump}}
	d2 := &DoorCard{Type: DoorObstacle, name: "Trap", resources: []ResourceType{Shield, Scroll}}

	_, _ = pf.addDungeonCard(d1)
	_, _ = pf.addDungeonCard(d2)

	p, _ := NewPlayer("P", Paladin)

	// Play 1 Sword and 1 Shield (missing Jump and Scroll)
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Sword}})
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Shield}})
	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not beaten before WildCards")
	}

	// Play 2 WildCards -> covers Jump and Scroll across both doors
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{WildCard, WildCard}})
	if !pf.isPlayfieldBeaten() {
		t.Error("expected playfield beaten when WildCards fulfill deficits across multiple doors")
	}
}

func TestPlayfieldWildCardExcess(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorPerson,
		name:      "Guard",
		resources: []ResourceType{Sword},
	}
	_, _ = pf.addDungeonCard(door)

	p, _ := NewPlayer("P", Paladin)

	// Play 2 WildCards for 1 Sword requirement -> Beaten!
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{WildCard, WildCard}})
	if !pf.isPlayfieldBeaten() {
		t.Error("expected playfield to be beaten with excess WildCards")
	}
}

func TestPlayfieldMultiDoorResolution(t *testing.T) {
	pf := NewPlayfield()

	d1 := &DoorCard{Type: DoorMonster, name: "Goblin", resources: []ResourceType{Sword}}
	d2 := &DoorCard{Type: DoorObstacle, name: "Trap", resources: []ResourceType{Jump}}

	_, _ = pf.addDungeonCard(d1)
	_, _ = pf.addDungeonCard(d2)

	p, _ := NewPlayer("P", Paladin)
	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Sword}})
	if pf.isPlayfieldBeaten() {
		t.Error("expected playfield not beaten: Jump still missing")
	}

	_, _ = pf.addPlayerCard(p, &ResourceCard{resources: []ResourceType{Jump}})
	if !pf.isPlayfieldBeaten() {
		t.Error("expected playfield beaten when both doors' requirements are fulfilled")
	}
}

func TestPlayfieldAddMoreThanTwoDoorsRejected(t *testing.T) {
	pf := NewPlayfield()
	d1 := &DoorCard{Type: DoorMonster, name: "D1"}
	d2 := &DoorCard{Type: DoorMonster, name: "D2"}
	d3 := &DoorCard{Type: DoorMonster, name: "D3"}

	_, err := pf.addDungeonCard(d1)
	if err != nil {
		t.Fatalf("unexpected error adding d1: %v", err)
	}
	_, err = pf.addDungeonCard(d2)
	if err != nil {
		t.Fatalf("unexpected error adding d2: %v", err)
	}
	_, err = pf.addDungeonCard(d3)
	if err == nil {
		t.Error("expected error when adding more than 2 doors, got nil")
	}
}

func TestPlayfieldDefeatDoorNotInPlayfield(t *testing.T) {
	pf := NewPlayfield()
	d1 := &DoorCard{Type: DoorMonster, name: "D1"}
	d2 := &DoorCard{Type: DoorMonster, name: "D2"}

	_, _ = pf.addDungeonCard(d1)

	_, err := pf.defeatDoor(d2)
	if err == nil {
		t.Error("expected error when defeating door not on playfield, got nil")
	}
}

// --- Tick & Deterministic Game Flow Tests ---

func TestTickTimeoutCausesDefeat(t *testing.T) {
	paladin, _ := NewPlayer("P1", Paladin)
	barbarian, _ := NewPlayer("P2", Barbarian)

	game := &Game{
		players:   []*Player{paladin, barbarian},
		dungeon:   NewDungeon(),
		playField: *NewPlayfield(),
		Status:    Waiting,
	}

	_, err := game.Apply(StartCmd{})
	if err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Tick 4 minutes 59 seconds - should still be playing
	events, err := game.Tick(4*time.Minute + 59*time.Second)
	if err != nil {
		t.Fatalf("unexpected tick error: %v", err)
	}
	if game.Status != Playing {
		t.Errorf("expected game to still be Playing, got %v", game.Status)
	}
	if len(events) != 0 {
		t.Errorf("expected no events before timeout")
	}

	// Advance remaining 1 second (total: 5m0s) - should defeat
	events, err = game.Tick(1 * time.Second)
	if err != nil {
		t.Fatalf("unexpected tick error: %v", err)
	}
	if game.Status != Defeat {
		t.Errorf("expected game status Defeat, got %v", game.Status)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if _, ok := events[0].(GameLostEvent); !ok {
		t.Errorf("expected GameLostEvent, got %T", events[0])
	}

	// Subsequent ticks on ended game should be no-ops
	events, err = game.Tick(1 * time.Second)
	if err != nil || len(events) != 0 {
		t.Errorf("expected no error and no events after game end")
	}
}

func TestTickFrozenTimeDoesNotAdvanceGameTimer(t *testing.T) {
	paladin, _ := NewPlayer("P1", Paladin)
	barbarian, _ := NewPlayer("P2", Barbarian)

	game := &Game{
		players:      []*Player{paladin, barbarian},
		dungeon:      NewDungeon(),
		playField:    *NewPlayfield(),
		isTimeFrozen: true,
		Status:       Playing,
	}

	_, err := game.Tick(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.Status != Playing {
		t.Errorf("expected game to still be Playing because time is frozen, got %v", game.Status)
	}
	if game.inGameTimer != 0 {
		t.Errorf("expected inGameTimer to remain 0, got %v", game.inGameTimer)
	}
}

func TestTickWhenNotPlayingReturnsError(t *testing.T) {
	game := NewGame()

	_, err := game.Tick(time.Second)
	if err == nil {
		t.Error("expected error when ticking game in Waiting state, got nil")
	}
}

func TestFullDungeonProgressionToVictory(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	barbarian, _ := NewPlayer("Barbarian", Barbarian)

	// Dungeon with 1 door and 1 boss
	door1 := &DoorCard{Type: DoorMonster, name: "Goblin", resources: []ResourceType{Sword}}
	boss := &BossMat{name: "Final Boss", resources: []ResourceType{Shield}}

	dungeon := &Dungeon{
		boss:  boss,
		doors: []DungeonCard{door1},
	}

	game := &Game{
		players:   []*Player{paladin, barbarian},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Waiting,
	}

	// 1. Start Game
	startEvents, err := game.Apply(StartCmd{})
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	if len(startEvents) < 2 {
		t.Fatalf("expected Start and Door events, got %d", len(startEvents))
	}
	if !game.playField.hasActiveDoor(door1) {
		t.Fatal("expected door1 to be active on playfield")
	}

	// 2. Play Sword to match door1
	swordCard := &ResourceCard{resources: []ResourceType{Sword}}
	paladin.hand = []PlayerCard{swordCard}
	_, err = game.Apply(PlayCardCmd{Player: paladin, Card: swordCard})
	if err != nil {
		t.Fatalf("failed to play sword: %v", err)
	}

	// 3. Tick past debounce duration (0.5s) to resolve door1
	tickEvents, err := game.Tick(cardPlayDebounceDuration + 10*time.Millisecond)
	if err != nil {
		t.Fatalf("failed tick: %v", err)
	}

	// Expect door1 defeated, field cleared, boss door opened
	foundDoorDefeated := false
	foundFieldCleared := false
	foundBossOpened := false
	for _, evt := range tickEvents {
		switch e := evt.(type) {
		case DoorDefeatedEvent:
			if e.DungeonCard == door1 {
				foundDoorDefeated = true
			}
		case FieldClearedEvent:
			foundFieldCleared = true
		case DoorOpenedEvent:
			if e.DungeonCard == boss {
				foundBossOpened = true
			}
		}
	}
	if !foundDoorDefeated || !foundFieldCleared || !foundBossOpened {
		t.Errorf("missing expected events in door resolution: defeated=%v, cleared=%v, bossOpened=%v",
			foundDoorDefeated, foundFieldCleared, foundBossOpened)
	}
	if !game.isFightingBoss {
		t.Error("expected game.isFightingBoss to be true")
	}

	// 4. Play Shield to match Boss requirements
	shieldCard := &ResourceCard{resources: []ResourceType{Shield}}
	barbarian.hand = []PlayerCard{shieldCard}
	_, err = game.Apply(PlayCardCmd{Player: barbarian, Card: shieldCard})
	if err != nil {
		t.Fatalf("failed to play shield: %v", err)
	}

	// 5. Tick to resolve Boss
	bossTickEvents, err := game.Tick(1 * time.Millisecond)
	if err != nil {
		t.Fatalf("failed boss tick: %v", err)
	}

	if game.Status != Victory {
		t.Errorf("expected game status Victory, got %v", game.Status)
	}

	foundGameWon := false
	for _, evt := range bossTickEvents {
		if _, ok := evt.(GameWonEvent); ok {
			foundGameWon = true
		}
	}
	if !foundGameWon {
		t.Error("expected GameWonEvent to be emitted upon boss defeat")
	}
}

func TestGameWildCardResolvesDoor(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin)
	barbarian, _ := NewPlayer("Barbarian", Barbarian)

	door1 := &DoorCard{Type: DoorMonster, name: "Chimera", resources: []ResourceType{Sword, Arrow, Shield}}
	door2 := &DoorCard{Type: DoorObstacle, name: "Spike Wall", resources: []ResourceType{Jump}}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss", resources: []ResourceType{Scroll}},
		doors: []DungeonCard{door1, door2},
	}

	game := &Game{
		players:   []*Player{paladin, barbarian},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Waiting,
	}

	_, _ = game.Apply(StartCmd{})

	// Paladin plays 1 Sword
	swordCard := &ResourceCard{resources: []ResourceType{Sword}}
	paladin.hand = []PlayerCard{swordCard}
	_, err := game.Apply(PlayCardCmd{Player: paladin, Card: swordCard})
	if err != nil {
		t.Fatalf("paladin failed to play sword: %v", err)
	}

	// Barbarian plays 2 WildCards to satisfy remaining Arrow and Shield
	wildCard1 := &ResourceCard{resources: []ResourceType{WildCard}}
	wildCard2 := &ResourceCard{resources: []ResourceType{WildCard}}
	barbarian.hand = []PlayerCard{wildCard1, wildCard2}

	_, err = game.Apply(PlayCardCmd{Player: barbarian, Card: wildCard1})
	if err != nil {
		t.Fatalf("barbarian failed to play first wildcard: %v", err)
	}
	_, err = game.Apply(PlayCardCmd{Player: barbarian, Card: wildCard2})
	if err != nil {
		t.Fatalf("barbarian failed to play second wildcard: %v", err)
	}

	// Tick past debounce duration to resolve door1
	tickEvents, err := game.Tick(cardPlayDebounceDuration + 10*time.Millisecond)
	if err != nil {
		t.Fatalf("failed tick: %v", err)
	}

	foundDoorDefeated := false
	foundDoorOpened := false
	for _, evt := range tickEvents {
		if dDefeat, ok := evt.(DoorDefeatedEvent); ok && dDefeat.DungeonCard == door1 {
			foundDoorDefeated = true
		}
		if dOpen, ok := evt.(DoorOpenedEvent); ok && dOpen.DungeonCard == door2 {
			foundDoorOpened = true
		}
	}

	if !foundDoorDefeated {
		t.Error("expected door1 to be defeated via WildCards")
	}
	if !foundDoorOpened {
		t.Error("expected door2 to be opened after door1 defeated")
	}
	if !game.playField.hasActiveDoor(door2) {
		t.Error("expected door2 to be the active door on playfield")
	}
}

// --- Hero Abilities Tests ---

func TestHeroAbilityTrickShotSuccess(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger)
	paladin, _ := NewPlayer("Arthur", Paladin)

	doorPerson := &DoorCard{Type: DoorPerson, name: "Evil Guard", resources: []ResourceType{Sword, Shield, Jump}}
	nextDoor := &DoorCard{Type: DoorMonster, name: "Dragon", resources: []ResourceType{Sword}}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss"},
		doors: []DungeonCard{doorPerson, nextDoor},
	}

	game := &Game{
		players:   []*Player{ranger, paladin},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Waiting,
	}

	_, _ = game.Apply(StartCmd{})

	if len(ranger.hand) < 3 {
		t.Fatalf("expected ranger to have at least 3 cards in hand, got %d", len(ranger.hand))
	}
	discards := []PlayerCard{ranger.hand[0], ranger.hand[1], ranger.hand[2]}

	// Ranger uses TrickShot on doorPerson
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: discards,
		Ability:      TrickShot{Target: doorPerson},
	})
	if err != nil {
		t.Fatalf("failed to use TrickShot: %v", err)
	}

	// Verify hand had 3 cards discarded (hand size was 5, now 2)
	if len(ranger.hand) != 2 {
		t.Errorf("expected ranger hand to have 2 cards remaining after discarding 3, got %d", len(ranger.hand))
	}
	if len(ranger.discard) != 3 {
		t.Errorf("expected ranger discard to have 3 cards, got %d", len(ranger.discard))
	}

	// Verify doorPerson was defeated and nextDoor opened
	if game.playField.hasActiveDoor(doorPerson) {
		t.Error("expected doorPerson to no longer be active")
	}
	if !game.playField.hasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be active")
	}

	// Verify event order: HeroAbilityUsedEvent -> DiscardEvents -> Defeat/Clear/Open events
	if len(events) < 5 {
		t.Fatalf("expected at least 5 events, got %d", len(events))
	}
	if _, ok := events[0].(HeroAbilityUsedEvent); !ok {
		t.Errorf("expected first event to be HeroAbilityUsedEvent, got %T", events[0])
	}
}

func TestHeroAbilityTrickShotInvalidClassRejected(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin)
	doorPerson := &DoorCard{Type: DoorPerson, name: "Guard"}

	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Arrow}}
	paladin.hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		players:   []*Player{paladin},
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorPerson)

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TrickShot{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when non-Ranger attempts TrickShot, got nil")
	}
	// Hand must not be discarded on failure
	if len(paladin.hand) != 3 {
		t.Errorf("expected hand to remain intact on failure, got %d", len(paladin.hand))
	}
}

func TestHeroAbilityTrickShotInvalidTargetTypeRejected(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger)
	doorMonster := &DoorCard{Type: DoorMonster, name: "Orc"} // Monster, not Person

	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Arrow}}
	ranger.hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		players:   []*Player{ranger},
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorMonster)

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TrickShot{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when targeting non-Person door with TrickShot, got nil")
	}
	if len(ranger.hand) != 3 {
		t.Errorf("expected hand to remain intact, got %d", len(ranger.hand))
	}
}

func TestHeroAbilityRequiresExactlyThreeCards(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger)
	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	ranger.hand = []PlayerCard{c1, c2}

	game := &Game{
		players: []*Player{ranger},
		Status:  Playing,
	}

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2}, // only 2 cards
		Ability:      TrickShot{},
	})
	if err == nil {
		t.Error("expected error when providing less than 3 discard cards, got nil")
	}
}

func TestHeroAbilityStopTime(t *testing.T) {
	wizard, _ := NewPlayer("Gandalf", Wizard)
	c1 := &ResourceCard{resources: []ResourceType{Scroll}}
	c2 := &ResourceCard{resources: []ResourceType{Scroll}}
	c3 := &ResourceCard{resources: []ResourceType{Scroll}}
	wizard.hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		players:      []*Player{wizard},
		isTimeFrozen: false,
		Status:       Playing,
	}

	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       wizard,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      StopTime{},
	})
	if err != nil {
		t.Fatalf("failed StopTime: %v", err)
	}
	if !game.isTimeFrozen {
		t.Error("expected isTimeFrozen to be true")
	}

	foundTimeFrozen := false
	for _, evt := range events {
		if _, ok := evt.(TimeFrozenEvent); ok {
			foundTimeFrozen = true
		}
	}
	if !foundTimeFrozen {
		t.Error("expected TimeFrozenEvent in emitted events")
	}

	// Playing a card unfreezes time
	cPlay := &ResourceCard{resources: []ResourceType{Scroll}}
	wizard.hand = []PlayerCard{cPlay}
	playEvents, err := game.Apply(PlayCardCmd{Player: wizard, Card: cPlay})
	if err != nil {
		t.Fatalf("failed to play card: %v", err)
	}
	if game.isTimeFrozen {
		t.Error("expected isTimeFrozen to be false after playing card")
	}
	foundUnfrozen := false
	for _, evt := range playEvents {
		if _, ok := evt.(TimeUnfrozenEvent); ok {
			foundUnfrozen = true
		}
	}
	if !foundUnfrozen {
		t.Error("expected TimeUnfrozenEvent when card is played while time is frozen")
	}
}

func TestHeroAbilitySmite(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin)
	doorMonster := &DoorCard{Type: DoorMonster, name: "Dragon"}
	doorPerson := &DoorCard{Type: DoorPerson, name: "Thief"}
	nextDoor := &DoorCard{Type: DoorObstacle, name: "Wall"}

	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Shield}}
	paladin.hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss"},
		doors: []DungeonCard{doorMonster, nextDoor},
	}

	game := &Game{
		players:   []*Player{paladin},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorMonster)

	// 1. Success on DoorMonster
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Smite{Target: doorMonster},
	})
	if err != nil {
		t.Fatalf("failed Smite: %v", err)
	}
	if len(events) < 4 { // HeroAbilityUsedEvent, 3x CardDiscardedEvent, DoorDefeatedEvent, etc.
		t.Errorf("expected at least 4 events, got %d", len(events))
	}

	// 2. Reject non-Monster
	paladin.hand = []PlayerCard{c1, c2, c3}
	_, _ = game.playField.addDungeonCard(doorPerson)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Smite{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when Smite targets non-Monster, got nil")
	}

	// 3. Reject non-Paladin
	barbarian, _ := NewPlayer("Conan", Barbarian)
	barbarian.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       barbarian,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Smite{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when non-Paladin uses Smite, got nil")
	}
}

func TestHeroAbilitySlay(t *testing.T) {
	barbarian, _ := NewPlayer("Conan", Barbarian)
	doorMonster := &DoorCard{Type: DoorMonster, name: "Beast"}
	doorObstacle := &DoorCard{Type: DoorObstacle, name: "Pit"}
	nextDoor := &DoorCard{Type: DoorPerson, name: "Guard"}

	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Sword}}
	c3 := &ResourceCard{resources: []ResourceType{Sword}}
	barbarian.hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss"},
		doors: []DungeonCard{doorMonster, nextDoor},
	}

	game := &Game{
		players:   []*Player{barbarian},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorMonster)

	// Success on DoorMonster
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       barbarian,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Slay{Target: doorMonster},
	})
	if err != nil {
		t.Fatalf("failed Slay: %v", err)
	}

	// Reject non-Monster
	barbarian.hand = []PlayerCard{c1, c2, c3}
	_, _ = game.playField.addDungeonCard(doorObstacle)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       barbarian,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Slay{Target: doorObstacle},
	})
	if err == nil {
		t.Error("expected error when Slay targets non-Monster, got nil")
	}

	// Reject non-Barbarian
	sorceress, _ := NewPlayer("Sorceress", Sorceress)
	sorceress.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       sorceress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Slay{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when non-Barbarian uses Slay, got nil")
	}
}

func TestHeroAbilityTeleport(t *testing.T) {
	sorceress, _ := NewPlayer("Jaina", Sorceress)
	doorObstacle := &DoorCard{Type: DoorObstacle, name: "Wall"}
	doorPerson := &DoorCard{Type: DoorPerson, name: "Bandit"}
	nextDoor := &DoorCard{Type: DoorMonster, name: "Dragon"}

	c1 := &ResourceCard{resources: []ResourceType{Scroll}}
	c2 := &ResourceCard{resources: []ResourceType{Scroll}}
	c3 := &ResourceCard{resources: []ResourceType{Scroll}}
	sorceress.hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss"},
		doors: []DungeonCard{doorObstacle, nextDoor},
	}

	game := &Game{
		players:   []*Player{sorceress},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorObstacle)

	// Success on DoorObstacle
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       sorceress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Teleport{Target: doorObstacle},
	})
	if err != nil {
		t.Fatalf("failed Teleport: %v", err)
	}

	// Reject non-Obstacle
	sorceress.hand = []PlayerCard{c1, c2, c3}
	_, _ = game.playField.addDungeonCard(doorPerson)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       sorceress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Teleport{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when Teleport targets non-Obstacle, got nil")
	}

	// Reject non-Sorceress
	paladin, _ := NewPlayer("Arthur", Paladin)
	paladin.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Teleport{Target: doorObstacle},
	})
	if err == nil {
		t.Error("expected error when non-Sorceress uses Teleport, got nil")
	}
}

func TestHeroAbilityVault(t *testing.T) {
	ninja, _ := NewPlayer("Shadow", Ninja)
	doorObstacle := &DoorCard{Type: DoorObstacle, name: "Trap"}
	doorMonster := &DoorCard{Type: DoorMonster, name: "Ogre"}
	nextDoor := &DoorCard{Type: DoorPerson, name: "Guard"}

	c1 := &ResourceCard{resources: []ResourceType{Jump}}
	c2 := &ResourceCard{resources: []ResourceType{Jump}}
	c3 := &ResourceCard{resources: []ResourceType{Jump}}
	ninja.hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss"},
		doors: []DungeonCard{doorObstacle, nextDoor},
	}

	game := &Game{
		players:   []*Player{ninja},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorObstacle)

	// Success on DoorObstacle
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ninja,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Vault{Target: doorObstacle},
	})
	if err != nil {
		t.Fatalf("failed Vault: %v", err)
	}

	// Reject non-Obstacle
	ninja.hand = []PlayerCard{c1, c2, c3}
	_, _ = game.playField.addDungeonCard(doorMonster)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       ninja,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Vault{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when Vault targets non-Obstacle, got nil")
	}

	// Reject non-Ninja
	thief, _ := NewPlayer("Rogue", Thief)
	thief.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       thief,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Vault{Target: doorObstacle},
	})
	if err == nil {
		t.Error("expected error when non-Ninja uses Vault, got nil")
	}
}

func TestHeroAbilityIntimidate(t *testing.T) {
	gladiator, _ := NewPlayer("Spartacus", Gladiator)
	doorPerson := &DoorCard{Type: DoorPerson, name: "Guard"}
	doorMonster := &DoorCard{Type: DoorMonster, name: "Wolf"}
	nextDoor := &DoorCard{Type: DoorObstacle, name: "Wall"}

	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Sword}}
	c3 := &ResourceCard{resources: []ResourceType{Sword}}
	gladiator.hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		boss:  &BossMat{name: "Boss"},
		doors: []DungeonCard{doorPerson, nextDoor},
	}

	game := &Game{
		players:   []*Player{gladiator},
		dungeon:   dungeon,
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(doorPerson)

	// Success on DoorPerson
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       gladiator,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Intimidate{Target: doorPerson},
	})
	if err != nil {
		t.Fatalf("failed Intimidate: %v", err)
	}

	// Reject non-Person
	gladiator.hand = []PlayerCard{c1, c2, c3}
	_, _ = game.playField.addDungeonCard(doorMonster)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       gladiator,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Intimidate{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when Intimidate targets non-Person, got nil")
	}

	// Reject non-Gladiator
	paladin, _ := NewPlayer("Arthur", Paladin)
	paladin.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Intimidate{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when non-Gladiator uses Intimidate, got nil")
	}
}

func TestHeroAbilityAnimalCompanion(t *testing.T) {
	huntress, _ := NewPlayer("Artemis", Huntress)
	paladin, _ := NewPlayer("Arthur", Paladin)

	c1 := &ResourceCard{resources: []ResourceType{Arrow}}
	c2 := &ResourceCard{resources: []ResourceType{Arrow}}
	c3 := &ResourceCard{resources: []ResourceType{Arrow}}
	huntress.hand = []PlayerCard{c1, c2, c3}

	paladin.hand = []PlayerCard{}
	paladin.deck = &Deck{
		cards: []PlayerCard{
			&ResourceCard{resources: []ResourceType{Sword}},
			&ResourceCard{resources: []ResourceType{Shield}},
			&ResourceCard{resources: []ResourceType{Jump}},
			&ResourceCard{resources: []ResourceType{Scroll}},
		},
	}

	game := &Game{
		players: []*Player{huntress, paladin},
		Status:  Playing,
	}

	// 1. Success: target player draws 4 cards
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       huntress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanion{Target: paladin},
	})
	if err != nil {
		t.Fatalf("failed AnimalCompanion: %v", err)
	}
	if len(paladin.hand) != 4 {
		t.Errorf("expected target player to have 4 cards, got %d", len(paladin.hand))
	}
	foundDrawnEvents := 0
	for _, evt := range events {
		if _, ok := evt.(CardDrawnEvent); ok {
			foundDrawnEvents++
		}
	}
	if foundDrawnEvents != 4 {
		t.Errorf("expected 4 CardDrawnEvent, got %d", foundDrawnEvents)
	}

	// 2. Reject nil target
	huntress.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       huntress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanion{Target: nil},
	})
	if err == nil {
		t.Error("expected error when target is nil, got nil")
	}

	// 3. Reject non-Huntress
	ranger, _ := NewPlayer("Robin", Ranger)
	ranger.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanion{Target: paladin},
	})
	if err == nil {
		t.Error("expected error when non-Huntress uses AnimalCompanion, got nil")
	}
}

func TestHeroAbilityInspire(t *testing.T) {
	valkyrie, _ := NewPlayer("Freya", Valkyrie)
	paladin, _ := NewPlayer("Arthur", Paladin)
	barbarian, _ := NewPlayer("Conan", Barbarian)

	c1 := &ResourceCard{resources: []ResourceType{Shield}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Shield}}
	valkyrie.hand = []PlayerCard{c1, c2, c3}

	paladin.hand = []PlayerCard{}
	paladin.deck = &Deck{
		cards: []PlayerCard{
			&ResourceCard{resources: []ResourceType{Sword}},
			&ResourceCard{resources: []ResourceType{Sword}},
		},
	}
	barbarian.hand = []PlayerCard{}
	barbarian.deck = &Deck{
		cards: []PlayerCard{
			&ResourceCard{resources: []ResourceType{Sword}},
			&ResourceCard{resources: []ResourceType{Sword}},
		},
	}

	game := &Game{
		players: []*Player{valkyrie, paladin, barbarian},
		Status:  Playing,
	}

	// Success: other players draw 2 cards each, Valkyrie draws 0
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       valkyrie,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Inspire{},
	})
	if err != nil {
		t.Fatalf("failed Inspire: %v", err)
	}
	if len(paladin.hand) != 2 {
		t.Errorf("expected paladin to have drawn 2 cards, got %d", len(paladin.hand))
	}
	if len(barbarian.hand) != 2 {
		t.Errorf("expected barbarian to have drawn 2 cards, got %d", len(barbarian.hand))
	}
	if len(valkyrie.hand) != 0 {
		t.Errorf("expected valkyrie to have 0 cards (only discards), got %d", len(valkyrie.hand))
	}

	foundDrawnEvents := 0
	for _, evt := range events {
		if _, ok := evt.(CardDrawnEvent); ok {
			foundDrawnEvents++
		}
	}
	if foundDrawnEvents != 4 {
		t.Errorf("expected 4 CardDrawnEvent (2 per player), got %d", foundDrawnEvents)
	}

	// Reject non-Valkyrie
	paladin.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Inspire{},
	})
	if err == nil {
		t.Error("expected error when non-Valkyrie uses Inspire, got nil")
	}
}

func TestHeroAbilityPickpocket(t *testing.T) {
	thief, _ := NewPlayer("Lupin", Thief)
	c1 := &ResourceCard{resources: []ResourceType{Jump}}
	c2 := &ResourceCard{resources: []ResourceType{Jump}}
	c3 := &ResourceCard{resources: []ResourceType{Jump}}
	thief.hand = []PlayerCard{c1, c2, c3}

	thief.deck = &Deck{
		cards: []PlayerCard{
			&ResourceCard{resources: []ResourceType{Jump}},
			&ResourceCard{resources: []ResourceType{Jump}},
			&ResourceCard{resources: []ResourceType{Jump}},
			&ResourceCard{resources: []ResourceType{Jump}},
			&ResourceCard{resources: []ResourceType{Jump}},
		},
	}

	game := &Game{
		players: []*Player{thief},
		Status:  Playing,
	}

	// Success: draws 5 cards
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       thief,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Pickpocket{},
	})
	if err != nil {
		t.Fatalf("failed Pickpocket: %v", err)
	}
	if len(thief.hand) != 5 {
		t.Errorf("expected thief to have 5 cards in hand, got %d", len(thief.hand))
	}

	foundDrawnEvents := 0
	for _, evt := range events {
		if _, ok := evt.(CardDrawnEvent); ok {
			foundDrawnEvents++
		}
	}
	if foundDrawnEvents != 5 {
		t.Errorf("expected 5 CardDrawnEvent, got %d", foundDrawnEvents)
	}

	// Reject non-Thief
	ninja, _ := NewPlayer("Ninja", Ninja)
	ninja.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       ninja,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      Pickpocket{},
	})
	if err == nil {
		t.Error("expected error when non-Thief uses Pickpocket, got nil")
	}
}

func TestHeroAbilityForestSpirits(t *testing.T) {
	druid, _ := NewPlayer("Malfurion", Druid)
	c1 := &ResourceCard{resources: []ResourceType{Shield}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Shield}}
	druid.hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		players: []*Player{druid},
		Status:  Playing,
	}

	// Success on Druid
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       druid,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      ForestSpirits{},
	})
	if err != nil {
		t.Fatalf("failed ForestSpirits: %v", err)
	}

	// Reject non-Druid
	shaman, _ := NewPlayer("Thrall", Shaman)
	shaman.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      ForestSpirits{},
	})
	if err == nil {
		t.Error("expected error when non-Druid uses ForestSpirits, got nil")
	}
}

func TestHeroAbilitySpiritAnimal(t *testing.T) {
	shaman, _ := NewPlayer("Thrall", Shaman)
	paladin, _ := NewPlayer("Arthur", Paladin)

	c1 := &ResourceCard{resources: []ResourceType{Shield}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	c3 := &ResourceCard{resources: []ResourceType{Shield}}
	shaman.hand = []PlayerCard{c1, c2, c3}

	paladin.discard = []PlayerCard{
		&ResourceCard{resources: []ResourceType{Sword}},
		&ResourceCard{resources: []ResourceType{Shield}},
		&ResourceCard{resources: []ResourceType{Jump}},
	}
	paladin.deck = &Deck{cards: []PlayerCard{}}

	game := &Game{
		players: []*Player{shaman, paladin},
		Status:  Playing,
	}

	// 1. Success: heals target player 3 cards
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SpiritAnimal{Target: paladin},
	})
	if err != nil {
		t.Fatalf("failed SpiritAnimal: %v", err)
	}
	if len(paladin.discard) != 0 {
		t.Errorf("expected paladin discard to be empty, got %d", len(paladin.discard))
	}
	if len(paladin.deck.cards) != 3 {
		t.Errorf("expected paladin deck to have 3 healed cards, got %d", len(paladin.deck.cards))
	}

	foundHealedEvent := false
	for _, evt := range events {
		if h, ok := evt.(PlayerHealedEvent); ok && h.Player == paladin && h.amount == 3 {
			foundHealedEvent = true
		}
	}
	if !foundHealedEvent {
		t.Error("expected PlayerHealedEvent with amount 3 in emitted events")
	}

	// 2. Reject nil target
	shaman.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SpiritAnimal{Target: nil},
	})
	if err == nil {
		t.Error("expected error when target is nil, got nil")
	}

	// 3. Reject non-Shaman
	druid, _ := NewPlayer("Druid", Druid)
	druid.hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       druid,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SpiritAnimal{Target: paladin},
	})
	if err == nil {
		t.Error("expected error when non-Shaman uses SpiritAnimal, got nil")
	}
}

func TestHeroAbilityDiscardCardsMustBeInHand(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger)
	c1 := &ResourceCard{resources: []ResourceType{Sword}}
	c2 := &ResourceCard{resources: []ResourceType{Shield}}
	cNotInHand := &ResourceCard{resources: []ResourceType{Arrow}}

	ranger.hand = []PlayerCard{c1, c2}

	game := &Game{
		players: []*Player{ranger},
		Status:  Playing,
	}

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2, cNotInHand},
		Ability:      TrickShot{},
	})
	if err == nil {
		t.Error("expected error when discard cards are not in hand, got nil")
	}
}

func TestAllHeroClassesAndDecks(t *testing.T) {
	classes := []struct {
		class         HeroClass
		expectedColor DeckColor
		expectedName  string
	}{
		{Sorceress, Blue, "Sorceress"},
		{Wizard, Blue, "Wizard"},
		{Huntress, Green, "Huntress"},
		{Ranger, Green, "Ranger"},
		{Ninja, Purple, "Ninja"},
		{Thief, Purple, "Thief"},
		{Paladin, Yellow, "Paladin"},
		{Valkyrie, Yellow, "Valkyrie"},
		{Barbarian, Red, "Barbarian"},
		{Gladiator, Red, "Gladiator"},
		{Druid, Black, "Druid"},
		{Shaman, Black, "Shaman"},
	}

	for _, tc := range classes {
		t.Run(tc.expectedName, func(t *testing.T) {
			hero, err := NewHeroFromHeroClass(tc.class)
			if err != nil {
				t.Fatalf("unexpected error creating hero %s: %v", tc.expectedName, err)
			}
			if hero.class != tc.class {
				t.Errorf("expected class %v, got %v", tc.class, hero.class)
			}
			if hero.color != tc.expectedColor {
				t.Errorf("expected color %v, got %v", tc.expectedColor, hero.color)
			}
			if hero.name != tc.expectedName {
				t.Errorf("expected name %s, got %s", tc.expectedName, hero.name)
			}

			deck := hero.NewDeck()
			if deck == nil {
				t.Fatalf("expected non-nil deck for %s", tc.expectedName)
			}
			if deck.Color != tc.expectedColor {
				t.Errorf("expected deck color %v, got %v", tc.expectedColor, deck.Color)
			}
			if len(deck.cards) == 0 {
				t.Errorf("expected non-empty deck for %s", tc.expectedName)
			}

			player, err := NewPlayer("TestPlayer", tc.class)
			if err != nil {
				t.Fatalf("unexpected error creating player with hero %s: %v", tc.expectedName, err)
			}
			if player.hero.class != tc.class {
				t.Errorf("expected player hero class %v, got %v", tc.class, player.hero.class)
			}
			if player.deck == nil || len(player.deck.cards) == 0 {
				t.Errorf("expected player to have populated deck for %s", tc.expectedName)
			}
		})
	}
}

func TestAbilityEngineMethods(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin)
	p2, _ := NewPlayer("P2", Barbarian)
	p3, _ := NewPlayer("P3", Wizard)

	door := &DoorCard{Type: DoorMonster, name: "Goblin"}

	game := &Game{
		players:   []*Player{p1, p2, p3},
		playField: *NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.playField.addDungeonCard(door)

	// 1. listOtherPlayers
	others := game.listOtherPlayers(p1)
	if len(others) != 2 || others[0] != p2 || others[1] != p3 {
		t.Errorf("unexpected listOtherPlayers result: %v", others)
	}

	// 2. hasActiveDoor
	if !game.hasActiveDoor(door) {
		t.Error("expected door to be active")
	}
	fakeDoor := &DoorCard{Type: DoorObstacle, name: "Fake"}
	if game.hasActiveDoor(fakeDoor) {
		t.Error("expected fake door not to be active")
	}

	// 3. drawCards
	p1.deck = &Deck{cards: []PlayerCard{&ResourceCard{resources: []ResourceType{Sword}}}}
	p1.hand = []PlayerCard{}
	drawnEvents, err := game.drawCards(p1, 1)
	if err != nil {
		t.Fatalf("failed drawCards: %v", err)
	}
	if len(drawnEvents) != 1 || len(p1.hand) != 1 {
		t.Errorf("expected 1 card drawn, got hand len %d", len(p1.hand))
	}
}

// --- Reply Channel & Error Propagation Tests ---

func TestApplyWithReplyChannel(t *testing.T) {
	game := NewGame()
	reply := make(chan error, 1)

	// Valid command
	_, _ = game.Apply(AddPlayerCmd{
		Name:  "Arthur",
		Class: Paladin,
		reply: reply,
	})

	select {
	case err := <-reply:
		if err != nil {
			t.Errorf("expected nil error on reply channel, got %v", err)
		}
	default:
		t.Fatal("expected reply channel to have received a response")
	}

	// Invalid command
	_, _ = game.Apply(AddPlayerCmd{
		Name:  "Duplicate Arthur",
		Class: Paladin,
		reply: reply,
	})

	select {
	case err := <-reply:
		if err == nil {
			t.Error("expected error on reply channel for duplicate player, got nil")
		}
	default:
		t.Fatal("expected reply channel to have received an error response")
	}
}

func TestApplyWithNilReplyDoesNotPanicOrBlock(t *testing.T) {
	game := NewGame()

	// Should not block or panic even if reply is nil
	_, err := game.Apply(AddPlayerCmd{
		Name:  "Arthur",
		Class: Paladin,
		reply: nil,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
