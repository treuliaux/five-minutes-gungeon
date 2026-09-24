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
	if len(game.Players) != 0 {
		t.Errorf("expected 0 players initially, got %d", len(game.Players))
	}
	if game.HandSize != 0 {
		t.Errorf("expected 0 handSize initially, got %d", game.HandSize)
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
	if len(game.Players) != 1 {
		t.Fatalf("expected 1 player in game, got %d", len(game.Players))
	}
	if game.Players[0].Name != "Sir Lancelot" {
		t.Errorf("expected player name 'Sir Lancelot', got '%s'", game.Players[0].Name)
	}
	if game.Players[0].Hero.Class != Paladin {
		t.Errorf("expected player hero class Paladin, got %v", game.Players[0].Hero.Class)
	}

	// Verify PlayerAddedEvent
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	addedEvt, ok := events[0].(PlayerAddedEvent)
	if !ok {
		t.Fatalf("expected PlayerAddedEvent, got %T", events[0])
	}
	if addedEvt.Player != game.Players[0] {
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
	if len(game.Players) != 1 {
		t.Errorf("expected game to still have 1 player, got %d", len(game.Players))
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
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Barbarian, true)
	p3, _ := NewPlayer("P3", Gladiator, true)
	p4, _ := NewPlayer("P4", Valkyrie, true)
	p5, _ := NewPlayer("P5", Sorceress, true)
	p6, _ := NewPlayer("P6", Wizard, true)
	p7, _ := NewPlayer("P7", Huntress, true)

	game := &Game{
		Players: []*Player{p1, p2, p3, p4, p5, p6, p7},
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
	var startFound, doorOpenedFound bool
	for _, e := range events {
		if _, ok := e.(GameStartedEvent); ok {
			startFound = true
		}
		if _, ok := e.(DoorOpenedEvent); ok {
			doorOpenedFound = true
		}
	}
	if !startFound || !doorOpenedFound {
		t.Errorf("expected GameStartedEvent and DoorOpenedEvent in events: %v", events)
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
				p, _ := NewPlayer(fmt.Sprintf("P%d", i+1), class, true)
				// Ensure all players have a deck with enough cards for test
				p.Deck = NewYellowDeck().IncludeExtension()
				p.Deck.PutAtop(
					&ResourceCard{Resources: []ResourceType{Sword}},
					&ResourceCard{Resources: []ResourceType{Shield}},
					&ResourceCard{Resources: []ResourceType{Arrow}},
					&ResourceCard{Resources: []ResourceType{Jump}},
				)
				game.Players = append(game.Players, p)
			}

			_, err := game.Apply(StartCmd{})
			if err != nil {
				t.Fatalf("failed to start: %v", err)
			}
			if game.HandSize != tt.expectedHandSize {
				t.Errorf("expected handSize %d, got %d", tt.expectedHandSize, game.HandSize)
			}
			for i, p := range game.Players {
				if len(p.Hand) != tt.expectedHandSize {
					t.Errorf("player %d hand size expected %d, got %d", i, tt.expectedHandSize, len(p.Hand))
				}
			}
		})
	}
}

// --- Card Playing & Drawing Tests ---

func TestPlayCardSuccessAndAutoDraw(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	barbarian, _ := NewPlayer("Barbarian", Barbarian, true)

	// Custom decks with known cards
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	cDraw := &ResourceCard{Resources: []ResourceType{Arrow}}

	paladin.Hand = []PlayerCard{c1, c2}
	paladin.Deck = &Deck{Cards: []PlayerCard{cDraw}}

	game := &Game{
		Players:   []*Player{paladin, barbarian},
		HandSize:  2,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	events, err := game.Apply(PlayCardCmd{Player: paladin, Card: c1})
	if err != nil {
		t.Fatalf("failed to play card: %v", err)
	}

	// Verify card was played on field
	if len(game.PlayField.Field) != 1 || game.PlayField.Field[0] != c1 {
		t.Errorf("expected card to be on playfield")
	}
	// Verify card was removed from hand and auto-drawn back to handSize 2
	if len(paladin.Hand) != 2 {
		t.Fatalf("expected player hand to auto-draw back to 2, got %d", len(paladin.Hand))
	}
	if paladin.Hand[0] != c2 || paladin.Hand[1] != cDraw {
		t.Errorf("expected hand to contain [c2, cDraw], got %v", paladin.Hand)
	}
	if len(paladin.Deck.Cards) != 0 {
		t.Errorf("expected deck to be empty after draw, got %d", len(paladin.Deck.Cards))
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
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	cInHand := &ResourceCard{Resources: []ResourceType{Sword}}
	cNotInHand := &ResourceCard{Resources: []ResourceType{Shield}}

	paladin.Hand = []PlayerCard{cInHand}

	game := &Game{
		Players:   []*Player{paladin},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	_, err := game.Apply(PlayCardCmd{Player: paladin, Card: cNotInHand})
	if err == nil {
		t.Error("expected error when playing card not in hand, got nil")
	}
	if len(game.PlayField.Field) != 0 {
		t.Errorf("expected field to be empty")
	}
	if len(paladin.Hand) != 1 || paladin.Hand[0] != cInHand {
		t.Errorf("expected hand to remain unchanged")
	}
}

func TestPlayCardUnfreezesTime(t *testing.T) {
	wizard, _ := NewPlayer("Wizard", Wizard, true)
	card := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{card}

	game := &Game{
		Players:      []*Player{wizard},
		HandSize:     1,
		IsTimeFrozen: true,
		PlayField:    NewPlayfield(),
		Status:       Playing,
	}

	events, err := game.Apply(PlayCardCmd{Player: wizard, Card: card})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.IsTimeFrozen {
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
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}

	paladin.Hand = []PlayerCard{c1, c2}

	game := &Game{
		Players: []*Player{paladin},
		Status:  Playing,
	}

	events, err := game.Apply(DiscardCardCmd{Player: paladin, Card: c1})
	if err != nil {
		t.Fatalf("failed to discard: %v", err)
	}

	if len(paladin.Hand) != 1 || paladin.Hand[0] != c2 {
		t.Errorf("expected hand to only contain c2")
	}
	if paladin.Discard.Length() != 1 || paladin.Discard.Cards[0] != c1 {
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
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	cInHand := &ResourceCard{Resources: []ResourceType{Sword}}
	cNotInHand := &ResourceCard{Resources: []ResourceType{Shield}}

	paladin.Hand = []PlayerCard{cInHand}

	game := &Game{
		Players: []*Player{paladin},
		Status:  Playing,
	}

	_, err := game.Apply(DiscardCardCmd{Player: paladin, Card: cNotInHand})
	if err == nil {
		t.Error("expected error when discarding card not in hand, got nil")
	}
	if paladin.Discard.Length() != 0 {
		t.Errorf("expected discard pile to be empty")
	}
}

func TestPlayerHealRecyclesDiscardPile(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}

	paladin.Discard.Cards = []PlayerCard{c1, c2, c3}
	paladin.Deck = &Deck{Cards: []PlayerCard{}}

	// Partial heal (amount 2: takes c3, c2 to deck, leaving c1 in discard)
	events, err := paladin.Heal(2)
	if err != nil {
		t.Fatalf("unexpected error healing player: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	healedEvt, ok := events[0].(PlayerHealedEvent)
	if !ok || len(healedEvt.Cards) != 2 {
		t.Errorf("expected 2 healed, got %+v", events[0])
	}
	if paladin.Discard.Length() != 1 || paladin.Discard.Cards[0] != c1 {
		t.Errorf("expected discard to have [c1], got %v", paladin.Discard)
	}
	if len(paladin.Deck.Cards) != 2 || paladin.Deck.Cards[0] != c3 || paladin.Deck.Cards[1] != c2 {
		t.Errorf("expected deck to have [c3, c2], got %v", paladin.Deck.Cards)
	}

	// Full heal (amount 0 or amount >= len(discard) heals all remaining)
	events, err = paladin.Heal(0)
	if err != nil {
		t.Fatalf("unexpected error healing player: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	healedEvt, ok = events[0].(PlayerHealedEvent)
	if !ok || len(healedEvt.Cards) != 1 {
		t.Errorf("expected 1 healed, got %+v", events[0])
	}
	if paladin.Discard.Length() != 0 {
		t.Errorf("expected discard pile to be empty after heal, got %d", paladin.Discard.Length())
	}
	if len(paladin.Deck.Cards) != 3 {
		t.Fatalf("expected deck to have 3 recycled cards, got %d", len(paladin.Deck.Cards))
	}
}

func TestGameHealPlayerEmitsEvent(t *testing.T) {
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	paladin.Discard.Cards = []PlayerCard{c1}
	paladin.Deck = &Deck{Cards: []PlayerCard{}}

	events, err := paladin.Heal(1)
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
	if healedEvt.Player != paladin || len(healedEvt.Cards) != 1 {
		t.Errorf("unexpected healed event values: %+v", healedEvt)
	}
}

// --- Playfield Matching & Door Resolution Tests ---

func TestPlayfieldResourceRequirementResolution(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorMonster,
		Name:      "Orc",
		Resources: []ResourceType{Sword, Sword, Shield},
	}
	_, _ = pf.AddDungeonCard(door, 0)

	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten initially")
	}

	// Play 1 Sword - not beaten
	p, _ := NewPlayer("P", Paladin, true)
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Sword}})
	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with 1/2 swords")
	}

	// Play Dual card (Sword + Shield) - total: 2 Swords, 1 Shield -> Beaten!
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Sword, Shield}})
	if !pf.IsPlayfieldBeaten() {
		t.Error("expected playfield to be beaten with required 2 Swords and 1 Shield")
	}
}

func TestPlayfieldWildCardRequirementResolution(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorMonster,
		Name:      "Chimera",
		Resources: []ResourceType{Sword, Shield, Arrow},
	}
	_, _ = pf.AddDungeonCard(door, 0)

	p, _ := NewPlayer("P", Paladin, true)

	// Play 1 Sword - missing Shield and Arrow
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Sword}})
	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with only 1 Sword")
	}

	// Play 1 WildCard - covers Shield or Arrow, but 1 resource still missing
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{WildCard}})
	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with 1 Sword + 1 WildCard for 3 requirements")
	}

	// Play 2nd WildCard - total: 1 Sword + 2 WildCards -> fulfills Sword + Shield + Arrow!
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{WildCard}})
	if !pf.IsPlayfieldBeaten() {
		t.Error("expected playfield to be beaten with 1 Sword and 2 WildCards")
	}
}

func TestPlayfieldWildCardMultipleDeficits(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorObstacle,
		Name:      "Heavy Gate",
		Resources: []ResourceType{Sword, Sword, Shield, Shield},
	}
	_, _ = pf.AddDungeonCard(door, 0)

	p, _ := NewPlayer("P", Paladin, true)

	// Play 1 Sword (deficit: 1 Sword, 2 Shields = 3 deficit)
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Sword}})

	// Play 2 WildCards (covers 2 out of 3 deficits, 1 deficit remains)
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{WildCard, WildCard}})
	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not to be beaten with 2 WildCards covering 3 deficits")
	}

	// Play 3rd WildCard (covers all 3 deficits) -> Beaten!
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{WildCard}})
	if !pf.IsPlayfieldBeaten() {
		t.Error("expected playfield to be beaten when WildCards cover all multiple resource deficits")
	}
}

func TestPlayfieldWildCardMultiDoorResolution(t *testing.T) {
	pf := NewPlayfield()

	d1 := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword, Jump}}
	d2 := &DoorCard{Type: DoorObstacle, Name: "Trap", Resources: []ResourceType{Shield, Scroll}}

	_, _ = pf.AddDungeonCard(d1, 0)
	_, _ = pf.AddDungeonCard(d2, 0)

	p, _ := NewPlayer("P", Paladin, true)

	// Play 1 Sword and 1 Shield (missing Jump and Scroll)
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Sword}})
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Shield}})
	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not beaten before WildCards")
	}

	// Play 2 WildCards -> covers Jump and Scroll across both doors
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{WildCard, WildCard}})
	if !pf.IsPlayfieldBeaten() {
		t.Error("expected playfield beaten when WildCards fulfill deficits across multiple doors")
	}
}

func TestPlayfieldWildCardExcess(t *testing.T) {
	pf := NewPlayfield()

	door := &DoorCard{
		Type:      DoorPerson,
		Name:      "Guard",
		Resources: []ResourceType{Sword},
	}
	_, _ = pf.AddDungeonCard(door, 0)

	p, _ := NewPlayer("P", Paladin, true)

	// Play 2 WildCards for 1 Sword requirement -> Beaten!
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{WildCard, WildCard}})
	if !pf.IsPlayfieldBeaten() {
		t.Error("expected playfield to be beaten with excess WildCards")
	}
}

func TestPlayfieldMultiDoorResolution(t *testing.T) {
	pf := NewPlayfield()

	d1 := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	d2 := &DoorCard{Type: DoorObstacle, Name: "Trap", Resources: []ResourceType{Jump}}

	_, _ = pf.AddDungeonCard(d1, 0)
	_, _ = pf.AddDungeonCard(d2, 0)

	p, _ := NewPlayer("P", Paladin, true)
	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Sword}})
	if pf.IsPlayfieldBeaten() {
		t.Error("expected playfield not beaten: Jump still missing")
	}

	_, _ = pf.AddPlayerCard(p, &ResourceCard{Resources: []ResourceType{Jump}})
	if !pf.IsPlayfieldBeaten() {
		t.Error("expected playfield beaten when both doors' requirements are fulfilled")
	}
}

func TestPlayfieldAddMoreThanTwoDoorsRejected(t *testing.T) {
	pf := NewPlayfield()
	d1 := &DoorCard{Type: DoorMonster, Name: "D1"}
	d2 := &DoorCard{Type: DoorMonster, Name: "D2"}
	d3 := &DoorCard{Type: DoorMonster, Name: "D3"}

	_, err := pf.AddDungeonCard(d1, 0)
	if err != nil {
		t.Fatalf("unexpected error adding d1: %v", err)
	}
	_, err = pf.AddDungeonCard(d2, 0)
	if err != nil {
		t.Fatalf("unexpected error adding d2: %v", err)
	}
	_, err = pf.AddDungeonCard(d3, 0)
	if err == nil {
		t.Error("expected error when adding more than 2 doors, got nil")
	}
}

func TestPlayfieldDefeatDoorNotInPlayfield(t *testing.T) {
	pf := NewPlayfield()
	d1 := &DoorCard{Type: DoorMonster, Name: "D1"}
	d2 := &DoorCard{Type: DoorMonster, Name: "D2"}

	_, _ = pf.AddDungeonCard(d1, 0)

	_, err := pf.DefeatDoor(d2)
	if err == nil {
		t.Error("expected error when defeating door not on playfield, got nil")
	}
}

// --- Tick & Deterministic Game Flow Tests ---

func TestTickTimeoutCausesDefeat(t *testing.T) {
	paladin, _ := NewPlayer("P1", Paladin, true)
	barbarian, _ := NewPlayer("P2", Barbarian, true)

	game := &Game{
		Players:   []*Player{paladin, barbarian},
		Dungeon:   NewBaseDungeon(1, 2),
		PlayField: NewPlayfield(),
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
	paladin, _ := NewPlayer("P1", Paladin, true)
	barbarian, _ := NewPlayer("P2", Barbarian, true)

	game := &Game{
		Players:      []*Player{paladin, barbarian},
		Dungeon:      NewBaseDungeon(1, 2),
		PlayField:    NewPlayfield(),
		IsTimeFrozen: true,
		Status:       Playing,
	}

	_, err := game.Tick(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.Status != Playing {
		t.Errorf("expected game to still be Playing because time is frozen, got %v", game.Status)
	}
	if game.InGameTimer != 0 {
		t.Errorf("expected inGameTimer to remain 0, got %v", game.InGameTimer)
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
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	barbarian, _ := NewPlayer("Barbarian", Barbarian, true)

	// Dungeon with 1 door and 1 boss
	door1 := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	boss := &BossMat{Name: "Final Boss", Resources: []ResourceType{Shield}}

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

	// 1. Start Game
	startEvents, err := game.Apply(StartCmd{})
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}
	if len(startEvents) < 2 {
		t.Fatalf("expected Start and Door events, got %d", len(startEvents))
	}
	if !game.PlayField.HasActiveDoor(door1) {
		t.Fatal("expected door1 to be active on playfield")
	}

	// 2. Play Sword to match door1
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	paladin.Hand = []PlayerCard{swordCard}
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
	if !game.IsFightingBoss {
		t.Error("expected game.IsFightingBoss to be true")
	}

	// 4. Play Shield to match Boss requirements
	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}
	barbarian.Hand = []PlayerCard{shieldCard}
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
	paladin, _ := NewPlayer("Paladin", Paladin, true)
	barbarian, _ := NewPlayer("Barbarian", Barbarian, true)

	door1 := &DoorCard{Type: DoorMonster, Name: "Chimera", Resources: []ResourceType{Sword, Arrow, Shield}}
	door2 := &DoorCard{Type: DoorObstacle, Name: "Spike Wall", Resources: []ResourceType{Jump}}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss", Resources: []ResourceType{Scroll}},
		Doors: []DungeonCard{door2, door1},
	}

	game := &Game{
		Players:   []*Player{paladin, barbarian},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Waiting,
	}

	_, _ = game.Apply(StartCmd{})

	// Paladin plays 1 Sword
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	paladin.Hand = []PlayerCard{swordCard}
	_, err := game.Apply(PlayCardCmd{Player: paladin, Card: swordCard})
	if err != nil {
		t.Fatalf("paladin failed to play sword: %v", err)
	}

	// Barbarian plays 2 WildCards to satisfy remaining Arrow and Shield
	wildCard1 := &ResourceCard{Resources: []ResourceType{WildCard}}
	wildCard2 := &ResourceCard{Resources: []ResourceType{WildCard}}
	barbarian.Hand = []PlayerCard{wildCard1, wildCard2}

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
	if !game.PlayField.HasActiveDoor(door2) {
		t.Error("expected door2 to be the active door on playfield")
	}
}

// --- Hero Abilities Tests ---

func TestHeroAbilityTrickShotSuccess(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	paladin, _ := NewPlayer("Arthur", Paladin, true)

	doorPerson := &DoorCard{Type: DoorPerson, Name: "Evil Guard", Resources: []ResourceType{Sword, Shield, Jump}}
	nextDoor := &DoorCard{Type: DoorMonster, Name: "Dragon", Resources: []ResourceType{Sword}}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{nextDoor, doorPerson},
	}

	game := &Game{
		Players:   []*Player{ranger, paladin},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Waiting,
	}

	_, _ = game.Apply(StartCmd{})

	if len(ranger.Hand) < 3 {
		t.Fatalf("expected ranger to have at least 3 cards in hand, got %d", len(ranger.Hand))
	}
	discards := []PlayerCard{ranger.Hand[0], ranger.Hand[1], ranger.Hand[2]}

	// Ranger uses TrickShot on doorPerson
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: discards,
		Ability:      TrickShotAbility{Target: doorPerson},
	})
	if err != nil {
		t.Fatalf("failed to use TrickShot: %v", err)
	}

	if ranger.Discard.Length() != 3 {
		t.Errorf("expected ranger discard to have 3 cards, got %d", ranger.Discard.Length())
	}

	// Verify doorPerson was defeated and nextDoor opened
	if game.PlayField.HasActiveDoor(doorPerson) {
		t.Error("expected doorPerson to no longer be active")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be active")
	}

	// Verify event order: HeroAbilityUsedEvent -> DiscardEvents -> Defeat/Clear/Open events
	if len(events) < 5 {
		t.Fatalf("expected at least 5 events, got %d", len(events))
	}
	if _, ok := events[3].(HeroAbilityUsedEvent); !ok {
		t.Errorf("expected third event to be HeroAbilityUsedEvent, got %T", events[0])
	}
}

func TestHeroAbilityTrickShotInvalidClassRejected(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	doorPerson := &DoorCard{Type: DoorPerson, Name: "Guard"}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	paladin.Hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		Players:   []*Player{paladin},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorPerson, 0)

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TrickShotAbility{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when non-Ranger attempts TrickShot, got nil")
	}
	// Hand must not be discarded on failure
	if len(paladin.Hand) != 3 {
		t.Errorf("expected hand to remain intact on failure, got %d", len(paladin.Hand))
	}
}

func TestHeroAbilityTrickShotInvalidTargetTypeRejected(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	doorMonster := &DoorCard{Type: DoorMonster, Name: "Orc"} // Monster, not Person

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	ranger.Hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		Players:   []*Player{ranger},
		PlayField: NewPlayfield(),
		Dungeon:   &Dungeon{Doors: []DungeonCard{}},
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, 0)

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TrickShotAbility{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when targeting non-Person door with TrickShot, got nil")
	}
	if len(ranger.Hand) != 3 {
		t.Errorf("expected hand to remain intact, got %d", len(ranger.Hand))
	}
}

func TestHeroAbilityRequiresExactlyThreeCards(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	ranger.Hand = []PlayerCard{c1, c2}

	game := &Game{
		Players: []*Player{ranger},
		Status:  Playing,
	}

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2}, // only 2 cards
		Ability:      TrickShotAbility{},
	})
	if err == nil {
		t.Error("expected error when providing less than 3 discard cards, got nil")
	}
}

func TestHeroAbilityStopTime(t *testing.T) {
	wizard, _ := NewPlayer("Gandalf", Wizard, true)
	c1 := &ResourceCard{Resources: []ResourceType{Scroll}}
	c2 := &ResourceCard{Resources: []ResourceType{Scroll}}
	c3 := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		Players:      []*Player{wizard},
		PlayField:    NewPlayfield(),
		IsTimeFrozen: false,
		Status:       Playing,
	}

	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       wizard,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      StopTimeAbility{},
	})
	if err != nil {
		t.Fatalf("failed StopTime: %v", err)
	}
	if !game.IsTimeFrozen {
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
	cPlay := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{cPlay}
	playEvents, err := game.Apply(PlayCardCmd{Player: wizard, Card: cPlay})
	if err != nil {
		t.Fatalf("failed to play card: %v", err)
	}
	if game.IsTimeFrozen {
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
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	doorMonster := &DoorCard{Type: DoorMonster, Name: "Dragon"}
	doorPerson := &DoorCard{Type: DoorPerson, Name: "Thief"}
	nextDoor := &DoorCard{Type: DoorObstacle, Name: "Wall"}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Shield}}
	paladin.Hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{doorMonster, nextDoor},
	}

	game := &Game{
		Players:   []*Player{paladin},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, 0)

	// 1. Success on DoorMonster
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SmiteAbility{Target: doorMonster},
	})
	if err != nil {
		t.Fatalf("failed Smite: %v", err)
	}
	if len(events) < 4 { // HeroAbilityUsedEvent, 3x CardDiscardedEvent, DoorDefeatedEvent, etc.
		t.Errorf("expected at least 4 events, got %d", len(events))
	}

	// 2. Reject non-Monster
	paladin.Hand = []PlayerCard{c1, c2, c3}
	_, _ = game.PlayField.AddDungeonCard(doorPerson, 0)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SmiteAbility{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when Smite targets non-Monster, got nil")
	}

	// 3. Reject non-Paladin
	barbarian, _ := NewPlayer("Conan", Barbarian, true)
	barbarian.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       barbarian,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SmiteAbility{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when non-Paladin uses Smite, got nil")
	}
}

func TestHeroAbilitySlay(t *testing.T) {
	barbarian, _ := NewPlayer("Conan", Barbarian, true)
	doorMonster := &DoorCard{Type: DoorMonster, Name: "Beast"}
	doorObstacle := &DoorCard{Type: DoorObstacle, Name: "Pit"}
	nextDoor := &DoorCard{Type: DoorPerson, Name: "Guard"}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Sword}}
	c3 := &ResourceCard{Resources: []ResourceType{Sword}}
	barbarian.Hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{doorMonster, nextDoor},
	}

	game := &Game{
		Players:   []*Player{barbarian},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, 0)

	// Success on DoorMonster
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       barbarian,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SlayAbility{Target: doorMonster},
	})
	if err != nil {
		t.Fatalf("failed Slay: %v", err)
	}

	// Reject non-Monster
	barbarian.Hand = []PlayerCard{c1, c2, c3}
	_, _ = game.PlayField.AddDungeonCard(doorObstacle, 0)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       barbarian,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SlayAbility{Target: doorObstacle},
	})
	if err == nil {
		t.Error("expected error when Slay targets non-Monster, got nil")
	}

	// Reject non-Barbarian
	sorceress, _ := NewPlayer("Sorceress", Sorceress, true)
	sorceress.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       sorceress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SlayAbility{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when non-Barbarian uses Slay, got nil")
	}
}

func TestHeroAbilityTeleport(t *testing.T) {
	sorceress, _ := NewPlayer("Jaina", Sorceress, true)
	doorObstacle := &DoorCard{Type: DoorObstacle, Name: "Wall"}
	doorPerson := &DoorCard{Type: DoorPerson, Name: "Bandit"}
	nextDoor := &DoorCard{Type: DoorMonster, Name: "Dragon"}

	c1 := &ResourceCard{Resources: []ResourceType{Scroll}}
	c2 := &ResourceCard{Resources: []ResourceType{Scroll}}
	c3 := &ResourceCard{Resources: []ResourceType{Scroll}}
	sorceress.Hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{doorObstacle, nextDoor},
	}

	game := &Game{
		Players:   []*Player{sorceress},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorObstacle, 0)

	// Success on DoorObstacle
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       sorceress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TeleportAbility{Target: doorObstacle},
	})
	if err != nil {
		t.Fatalf("failed Teleport: %v", err)
	}

	// Reject non-Obstacle
	sorceress.Hand = []PlayerCard{c1, c2, c3}
	_, _ = game.PlayField.AddDungeonCard(doorPerson, 0)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       sorceress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TeleportAbility{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when Teleport targets non-Obstacle, got nil")
	}

	// Reject non-Sorceress
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	paladin.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      TeleportAbility{Target: doorObstacle},
	})
	if err == nil {
		t.Error("expected error when non-Sorceress uses Teleport, got nil")
	}
}

func TestHeroAbilityVault(t *testing.T) {
	ninja, _ := NewPlayer("Shadow", Ninja, true)
	doorObstacle := &DoorCard{Type: DoorObstacle, Name: "Trap"}
	doorMonster := &DoorCard{Type: DoorMonster, Name: "Ogre"}
	nextDoor := &DoorCard{Type: DoorPerson, Name: "Guard"}

	c1 := &ResourceCard{Resources: []ResourceType{Jump}}
	c2 := &ResourceCard{Resources: []ResourceType{Jump}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	ninja.Hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{doorObstacle, nextDoor},
	}

	game := &Game{
		Players:   []*Player{ninja},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorObstacle, 0)

	// Success on DoorObstacle
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ninja,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      VaultAbility{Target: doorObstacle},
	})
	if err != nil {
		t.Fatalf("failed Vault: %v", err)
	}

	// Reject non-Obstacle
	ninja.Hand = []PlayerCard{c1, c2, c3}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, 0)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       ninja,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      VaultAbility{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when Vault targets non-Obstacle, got nil")
	}

	// Reject non-Ninja
	thief, _ := NewPlayer("Rogue", Thief, true)
	thief.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       thief,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      VaultAbility{Target: doorObstacle},
	})
	if err == nil {
		t.Error("expected error when non-Ninja uses Vault, got nil")
	}
}

func TestHeroAbilityIntimidate(t *testing.T) {
	gladiator, _ := NewPlayer("Spartacus", Gladiator, true)
	doorPerson := &DoorCard{Type: DoorPerson, Name: "Guard"}
	doorMonster := &DoorCard{Type: DoorMonster, Name: "Wolf"}
	nextDoor := &DoorCard{Type: DoorObstacle, Name: "Wall"}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Sword}}
	c3 := &ResourceCard{Resources: []ResourceType{Sword}}
	gladiator.Hand = []PlayerCard{c1, c2, c3}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{doorPerson, nextDoor},
	}

	game := &Game{
		Players:   []*Player{gladiator},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorPerson, 0)

	// Success on DoorPerson
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       gladiator,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      IntimidateAbility{Target: doorPerson},
	})
	if err != nil {
		t.Fatalf("failed Intimidate: %v", err)
	}

	// Reject non-Person
	gladiator.Hand = []PlayerCard{c1, c2, c3}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, 0)
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       gladiator,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      IntimidateAbility{Target: doorMonster},
	})
	if err == nil {
		t.Error("expected error when Intimidate targets non-Person, got nil")
	}

	// Reject non-Gladiator
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	paladin.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      IntimidateAbility{Target: doorPerson},
	})
	if err == nil {
		t.Error("expected error when non-Gladiator uses Intimidate, got nil")
	}
}

func TestHeroAbilityAnimalCompanion(t *testing.T) {
	huntress, _ := NewPlayer("Artemis", Huntress, true)
	paladin, _ := NewPlayer("Arthur", Paladin, true)

	c1 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c2 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	huntress.Hand = []PlayerCard{c1, c2, c3}

	paladin.Hand = []PlayerCard{}
	paladin.Deck = &Deck{
		Cards: []PlayerCard{
			&ResourceCard{Resources: []ResourceType{Sword}},
			&ResourceCard{Resources: []ResourceType{Shield}},
			&ResourceCard{Resources: []ResourceType{Jump}},
			&ResourceCard{Resources: []ResourceType{Scroll}},
		},
	}

	game := &Game{
		Players: []*Player{huntress, paladin},
		Status:  Playing,
	}

	// 1. Success: target player draws 4 cards
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       huntress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanionAbility{Target: paladin},
	})
	if err != nil {
		t.Fatalf("failed AnimalCompanion: %v", err)
	}
	if len(paladin.Hand) != 4 {
		t.Errorf("expected target player to have 4 cards, got %d", len(paladin.Hand))
	}
	foundDrawnEvents := 0
	for _, evt := range events {
		if _, ok := evt.(CardDrawnFromDeckEvent); ok {
			foundDrawnEvents++
		}
	}
	if foundDrawnEvents != 9 {
		t.Errorf("expected 9 CardDrawnEvent (4 target + 5 refilled), got %d", foundDrawnEvents)
	}

	// 2. nil target should smart-target the other player
	huntress.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       huntress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanionAbility{Target: nil},
	})
	if err != nil {
		t.Error("expected no error when target is nil, and auto-targeting is not ambiguous")
	}

	// 3. Reject non-Huntress
	ranger, _ := NewPlayer("Robin", Ranger, true)
	ranger.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanionAbility{Target: paladin},
	})
	if err == nil {
		t.Error("expected error when non-Huntress uses AnimalCompanion, got nil")
	}

	// 4. Smart-targeting should return an ambiguous error
	game.Players = append(game.Players, ranger)
	huntress.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       huntress,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanionAbility{Target: nil},
	})
	if err == nil {
		t.Error("expected error when target is nil, got nil")
	}
}

func TestHeroAbilityInspire(t *testing.T) {
	valkyrie, _ := NewPlayer("Freya", Valkyrie, true)
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	barbarian, _ := NewPlayer("Conan", Barbarian, true)

	c1 := &ResourceCard{Resources: []ResourceType{Shield}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Shield}}
	valkyrie.Hand = []PlayerCard{c1, c2, c3}

	paladin.Hand = []PlayerCard{}
	paladin.Deck = &Deck{
		Cards: []PlayerCard{
			&ResourceCard{Resources: []ResourceType{Sword}},
			&ResourceCard{Resources: []ResourceType{Sword}},
		},
	}
	barbarian.Hand = []PlayerCard{}
	barbarian.Deck = &Deck{
		Cards: []PlayerCard{
			&ResourceCard{Resources: []ResourceType{Sword}},
			&ResourceCard{Resources: []ResourceType{Sword}},
		},
	}

	game := &Game{
		Players: []*Player{valkyrie, paladin, barbarian},
		Status:  Playing,
	}

	// Success: other players draw 2 cards each, Valkyrie draws 0
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       valkyrie,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      InspireAbility{},
	})
	if err != nil {
		t.Fatalf("failed Inspire: %v", err)
	}
	if len(paladin.Hand) != 2 {
		t.Errorf("expected paladin to have drawn 2 cards, got %d", len(paladin.Hand))
	}
	if len(barbarian.Hand) != 2 {
		t.Errorf("expected barbarian to have drawn 2 cards, got %d", len(barbarian.Hand))
	}
	if len(valkyrie.Hand) != 4 {
		t.Errorf("expected valkyrie to have 4 cards (after hand refill), got %d", len(valkyrie.Hand))
	}

	foundDrawnEvents := 0
	for _, evt := range events {
		if _, ok := evt.(CardDrawnFromDeckEvent); ok {
			foundDrawnEvents++
		}
	}
	if foundDrawnEvents != 8 {
		t.Errorf("expected 8 CardDrawnEvent (2 per player + 4 refilled for valkyrie), got %d", foundDrawnEvents)
	}

	// Reject non-Valkyrie
	paladin.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       paladin,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      InspireAbility{},
	})
	if err == nil {
		t.Error("expected error when non-Valkyrie uses Inspire, got nil")
	}
}

func TestHeroAbilityPickpocket(t *testing.T) {
	thief, _ := NewPlayer("Lupin", Thief, true)
	c1 := &ResourceCard{Resources: []ResourceType{Jump}}
	c2 := &ResourceCard{Resources: []ResourceType{Jump}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	thief.Hand = []PlayerCard{c1, c2, c3}

	thief.Deck = &Deck{
		Cards: []PlayerCard{
			&ResourceCard{Resources: []ResourceType{Jump}},
			&ResourceCard{Resources: []ResourceType{Jump}},
			&ResourceCard{Resources: []ResourceType{Jump}},
			&ResourceCard{Resources: []ResourceType{Jump}},
			&ResourceCard{Resources: []ResourceType{Jump}},
		},
	}

	game := &Game{
		Players: []*Player{thief},
		Status:  Playing,
	}

	// Success: draws 5 cards
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       thief,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      PickpocketAbility{},
	})
	if err != nil {
		t.Fatalf("failed Pickpocket: %v", err)
	}
	if len(thief.Hand) != 5 {
		t.Errorf("expected thief to have 5 cards in hand, got %d", len(thief.Hand))
	}

	foundDrawnEvents := 0
	for _, evt := range events {
		if _, ok := evt.(CardDrawnFromDeckEvent); ok {
			foundDrawnEvents++
		}
	}
	if foundDrawnEvents != 5 {
		t.Errorf("expected 5 CardDrawnEvent, got %d", foundDrawnEvents)
	}

	// Reject non-Thief
	ninja, _ := NewPlayer("Ninja", Ninja, true)
	ninja.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       ninja,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      PickpocketAbility{},
	})
	if err == nil {
		t.Error("expected error when non-Thief uses Pickpocket, got nil")
	}
}

func TestHeroAbilityForestSpirits(t *testing.T) {
	druid, _ := NewPlayer("Malfurion", Druid, true)
	c1 := &ResourceCard{Resources: []ResourceType{Shield}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Shield}}
	druid.Hand = []PlayerCard{c1, c2, c3}

	game := &Game{
		Players: []*Player{druid},
		Status:  Playing,
		Dungeon: &Dungeon{
			Boss:  nil,
			Doors: make([]DungeonCard, 0),
		},
		PlayField: NewPlayfield(),
	}
	_, _ = game.PlayField.AddDungeonCard(&CurseCard{
		Type: ChallengeCurse,
		Name: "Tourbillon de Wazaa",
	}, 0)

	// Success on Druid
	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       druid,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      ForestSpiritsAbility{},
	})
	if err != nil {
		t.Fatalf("failed ForestSpirits: %v", err)
	}

	// Reject non-Druid
	shaman, _ := NewPlayer("Thrall", Shaman, true)
	shaman.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      ForestSpiritsAbility{},
	})
	if err == nil {
		t.Error("expected error when non-Druid uses ForestSpirits, got nil")
	}
}

func TestHeroAbilitySpiritAnimal(t *testing.T) {
	shaman, _ := NewPlayer("Thrall", Shaman, true)
	paladin, _ := NewPlayer("Arthur", Paladin, true)

	c1 := &ResourceCard{Resources: []ResourceType{Shield}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Shield}}
	shaman.Hand = []PlayerCard{c1, c2, c3}

	paladin.Discard.Cards = []PlayerCard{
		&ResourceCard{Resources: []ResourceType{Sword}},
		&ResourceCard{Resources: []ResourceType{Shield}},
		&ResourceCard{Resources: []ResourceType{Jump}},
	}
	paladin.Deck = &Deck{Cards: []PlayerCard{}}

	game := &Game{
		Players: []*Player{shaman, paladin},
		Status:  Playing,
	}

	// 1. Success: heals target player 3 cards
	events, err := game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SpiritAnimalAbility{Target: paladin},
	})
	if err != nil {
		t.Fatalf("failed SpiritAnimal: %v", err)
	}
	if paladin.Discard.Length() != 0 {
		t.Errorf("expected paladin discard to be empty, got %d", paladin.Discard.Length())
	}
	if len(paladin.Deck.Cards) != 3 {
		t.Errorf("expected paladin deck to have 3 healed cards, got %d", len(paladin.Deck.Cards))
	}

	foundHealedEvent := false
	for _, evt := range events {
		if h, ok := evt.(PlayerHealedEvent); ok && h.Player == paladin && len(h.Cards) == 3 {
			foundHealedEvent = true
		}
	}
	if !foundHealedEvent {
		t.Error("expected PlayerHealedEvent with amount 3 in emitted events")
	}

	// 2. nil target should smart-target the other player
	shaman.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SpiritAnimalAbility{Target: nil},
	})
	if err != nil {
		t.Error("expected no error when target is nil, and auto-targeting is not ambiguous")
	}

	// 3. Reject non-Shaman
	druid, _ := NewPlayer("Druid", Druid, true)
	druid.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       druid,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      SpiritAnimalAbility{Target: paladin},
	})
	if err == nil {
		t.Error("expected error when non-Shaman uses SpiritAnimal, got nil")
	}

	// 4. Smart-targeting should return an ambiguous error
	game.Players = append(game.Players, druid)
	shaman.Hand = []PlayerCard{c1, c2, c3}
	_, err = game.Apply(UseHeroAbilityCmd{
		Player:       shaman,
		DiscardCards: []PlayerCard{c1, c2, c3},
		Ability:      AnimalCompanionAbility{Target: nil},
	})
	if err == nil {
		t.Error("expected error when target is nil, got nil")
	}
}

func TestHeroAbilityDiscardCardsMustBeInHand(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	cNotInHand := &ResourceCard{Resources: []ResourceType{Arrow}}

	ranger.Hand = []PlayerCard{c1, c2}

	game := &Game{
		Players: []*Player{ranger},
		Status:  Playing,
	}

	_, err := game.Apply(UseHeroAbilityCmd{
		Player:       ranger,
		DiscardCards: []PlayerCard{c1, c2, cNotInHand},
		Ability:      TrickShotAbility{},
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
			if hero.Class != tc.class {
				t.Errorf("expected class %v, got %v", tc.class, hero.Class)
			}
			if hero.Color != tc.expectedColor {
				t.Errorf("expected color %v, got %v", tc.expectedColor, hero.Color)
			}
			if hero.Name != tc.expectedName {
				t.Errorf("expected name %s, got %s", tc.expectedName, hero.Name)
			}

			deck := hero.NewBaseDeck()
			if deck == nil {
				t.Fatalf("expected non-nil deck for %s", tc.expectedName)
			}
			if deck.Color != tc.expectedColor {
				t.Errorf("expected deck color %v, got %v", tc.expectedColor, deck.Color)
			}
			if len(deck.Cards) == 0 {
				t.Errorf("expected non-empty deck for %s", tc.expectedName)
			}

			player, err := NewPlayer("TestPlayer", tc.class, true)
			if err != nil {
				t.Fatalf("unexpected error creating player with hero %s: %v", tc.expectedName, err)
			}
			if player.Hero.Class != tc.class {
				t.Errorf("expected player hero class %v, got %v", tc.class, player.Hero.Class)
			}
			if player.Deck == nil || len(player.Deck.Cards) == 0 {
				t.Errorf("expected player to have populated deck for %s", tc.expectedName)
			}
		})
	}
}

func TestAbilityEngineMethods(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Barbarian, true)
	p3, _ := NewPlayer("P3", Wizard, true)

	door := &DoorCard{Type: DoorMonster, Name: "Goblin"}

	game := &Game{
		Players:   []*Player{p1, p2, p3},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(door, 0)

	// 1. listOtherPlayers
	others := otherPlayers(game.ListPlayers(), p1)
	if len(others) != 2 || others[0] != p2 || others[1] != p3 {
		t.Errorf("unexpected listOtherPlayers result: %v", others)
	}

	// 2. hasActiveDoor
	if !game.PlayField.HasActiveDoor(door) {
		t.Error("expected door to be active")
	}
	fakeDoor := &DoorCard{Type: DoorObstacle, Name: "Fake"}
	if game.PlayField.HasActiveDoor(fakeDoor) {
		t.Error("expected fake door not to be active")
	}

	// 3. drawCards
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Sword}}}}
	p1.Hand = []PlayerCard{}
	drawnEvents, err := p1.DrawCardsFromDeck(1)
	if err != nil {
		t.Fatalf("failed drawCards: %v", err)
	}
	if len(drawnEvents) != 1 || len(p1.Hand) != 1 {
		t.Errorf("expected 1 card drawn, got hand len %d", len(p1.Hand))
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

// --- Deck & Discard Mechanics Tests ---

func TestDeckDrawAndPutAtopOrder(t *testing.T) {
	deck := &Deck{Cards: []PlayerCard{}}

	if deck.Draw() != nil {
		t.Error("expected Draw() on empty deck to return nil")
	}
	if !deck.Empty() || deck.Length() != 0 {
		t.Errorf("expected empty deck, got length %d", deck.Length())
	}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}

	// PutAtop adds to the top of the pile
	deck.PutAtop(c1, c2)
	deck.PutAtop(c3)

	if deck.Length() != 3 {
		t.Fatalf("expected length 3, got %d", deck.Length())
	}

	// Draw should draw from the top (c3 first, then c2, then c1)
	drawn1 := deck.Draw()
	if drawn1 != c3 {
		t.Errorf("expected first drawn card to be top card c3, got %v", drawn1)
	}

	drawn2 := deck.Draw()
	if drawn2 != c2 {
		t.Errorf("expected second drawn card to be c2, got %v", drawn2)
	}

	drawn3 := deck.Draw()
	if drawn3 != c1 {
		t.Errorf("expected third drawn card to be bottom card c1, got %v", drawn3)
	}

	if deck.Draw() != nil {
		t.Error("expected Draw() on exhausted deck to return nil")
	}
}

func TestDiscardDrawAndPutAtopOrder(t *testing.T) {
	discard := NewDiscard()

	if discard.Draw() != nil {
		t.Error("expected Draw() on empty discard to return nil")
	}
	if !discard.Empty() || discard.Length() != 0 {
		t.Errorf("expected empty discard, got length %d", discard.Length())
	}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}

	// PutAtop adds to the top of the discard pile
	discard.PutAtop(c1, c2)
	discard.PutAtop(c3)

	if discard.Length() != 3 {
		t.Fatalf("expected length 3, got %d", discard.Length())
	}

	// Draw should draw from the top (c3 first, then c2, then c1)
	drawn1 := discard.Draw()
	if drawn1 != c3 {
		t.Errorf("expected first drawn card to be top card c3, got %v", drawn1)
	}

	drawn2 := discard.Draw()
	if drawn2 != c2 {
		t.Errorf("expected second drawn card to be c2, got %v", drawn2)
	}

	drawn3 := discard.Draw()
	if drawn3 != c1 {
		t.Errorf("expected third drawn card to be bottom card c1, got %v", drawn3)
	}

	if discard.Draw() != nil {
		t.Error("expected Draw() on exhausted discard to return nil")
	}
}

func TestPlayerDrawCardsFromDiscardOrder(t *testing.T) {
	player, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}

	player.Discard.PutAtop(c1, c2)
	player.Hand = []PlayerCard{}

	// Draw 1 card from discard -> should draw c2 (top card)
	events, err := player.DrawCardsFromDiscard(1)
	if err != nil {
		t.Fatalf("unexpected error drawing from discard: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	drawnEvt, ok := events[0].(CardDrawnFromDiscardEvent)
	if !ok || drawnEvt.Card != c2 {
		t.Errorf("expected CardDrawnFromDiscardEvent with c2, got %+v", events[0])
	}
	if len(player.Hand) != 1 || player.Hand[0] != c2 {
		t.Errorf("expected hand to contain [c2], got %v", player.Hand)
	}

	// Draw remaining card -> should draw c1
	events, err = player.DrawCardsFromDiscard(1)
	if err != nil {
		t.Fatalf("unexpected error drawing from discard: %v", err)
	}
	if len(player.Hand) != 2 || player.Hand[1] != c1 {
		t.Errorf("expected hand to contain [c2, c1], got %v", player.Hand)
	}

	// Draw from now-empty discard -> should return with no errors
	events, err = player.DrawCardsFromDiscard(1)
	if events != nil || err != nil {
		t.Errorf("expected no errors when drawing from empty discard, got %v", err)
	}
}

func TestHeroDecksContainActionCards(t *testing.T) {
	decks := []*Deck{
		NewYellowDeck().IncludeExtension(),
		NewRedDeck().IncludeExtension(),
		NewGreenDeck().IncludeExtension(),
		NewBlueDeck().IncludeExtension(),
		NewPurpleDeck().IncludeExtension(),
		NewBlackDeck(),
	}

	for _, d := range decks {
		var actionCount, resourceCount int
		for _, card := range d.Cards {
			switch card.(type) {
			case *ActionCard:
				actionCount++
			case *ResourceCard:
				resourceCount++
			}
		}

		if actionCount == 0 {
			t.Errorf("deck color %v has 0 action cards", d.Color)
		}
		if resourceCount == 0 {
			t.Errorf("deck color %v has 0 resource cards", d.Color)
		}
	}
}

func TestHeroDeckSizes(t *testing.T) {
	standardDecks := []*Deck{
		NewYellowDeck().IncludeExtension(),
		NewRedDeck().IncludeExtension(),
		NewGreenDeck().IncludeExtension(),
		NewBlueDeck().IncludeExtension(),
		NewPurpleDeck().IncludeExtension(),
		NewBlackDeck(),
	}
	for _, d := range standardDecks {
		if d.Length() != 42 {
			t.Errorf("expected standard deck %v to have 42 cards, got %d", d.Color, d.Length())
		}
	}
}

func TestBossList(t *testing.T) {
	bosses := BossList()
	expectedNames := []string{
		"Baby Barbarian",
		"The Grime Reaper",
		"Zola the Gorgon",
		"A Freakin' Dragon!!!",
		"The Dungeon Master",
		"The Dungeon Master (Final Form)",
		"The K.I.C.K. 9000",
	}
	for i, boss := range bosses {
		if boss.Name != expectedNames[i] {
			t.Errorf("boss %d expected name '%s', got '%s'", i, expectedNames[i], boss.Name)
		}
		if len(boss.Require()) == 0 {
			t.Errorf("boss %s has no resource requirements", boss.Name)
		}
	}
}

func TestPlayfieldInfiniteResources(t *testing.T) {
	tests := []struct {
		name          string
		doorResources []ResourceType
		playedCard    PlayerCard
		expectedBeat  bool
	}{
		{
			name:          "InfiniteSword beats 10 Swords requirement",
			doorResources: []ResourceType{Sword, Sword, Sword, Sword, Sword, Sword, Sword, Sword, Sword, Sword},
			playedCard:    &ResourceCard{Resources: []ResourceType{InfiniteSword}},
			expectedBeat:  true,
		},
		{
			name:          "InfiniteShield beats 3 Shields requirement",
			doorResources: []ResourceType{Shield, Shield, Shield},
			playedCard:    &ResourceCard{Resources: []ResourceType{InfiniteShield}},
			expectedBeat:  true,
		},
		{
			name:          "InfiniteArrow does not satisfy missing Jump requirement",
			doorResources: []ResourceType{Arrow, Jump},
			playedCard:    &ResourceCard{Resources: []ResourceType{InfiniteArrow}},
			expectedBeat:  false,
		},
		{
			name:          "InfiniteScroll + Jump beats Scroll and Jump",
			doorResources: []ResourceType{Scroll, Scroll, Jump},
			playedCard:    &ResourceCard{Resources: []ResourceType{InfiniteScroll, Jump}},
			expectedBeat:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewPlayfield()
			door := &DoorCard{Type: DoorMonster, Resources: tt.doorResources}
			_, _ = pf.AddDungeonCard(door, 0)
			_, _ = pf.AddPlayerCard(&Player{}, tt.playedCard)

			beaten := pf.IsPlayfieldBeaten()
			if beaten != tt.expectedBeat {
				t.Errorf("expected IsPlayfieldBeaten=%v, got %v", tt.expectedBeat, beaten)
			}
		})
	}
}

// --- Dungeon & Deck Construction Tests ---

func TestDungeonConstruction(t *testing.T) {
	// 1. Level 1 (Baby Barbarian), 2 players, no extension
	d1 := NewBaseDungeon(1, 2)
	if d1.Boss.Name != "Baby Barbarian" {
		t.Errorf("expected boss 'Baby Barbarian', got %q", d1.Boss.Name)
	}
	// Deck size: 20 doors + 2*2 challenges = 24
	if len(d1.Doors) != 24 {
		t.Errorf("expected 24 door cards in dungeon, got %d", len(d1.Doors))
	}

	// 2. Level 6 (The Dungeon Master Final Form), 4 players, extension enabled
	d6 := NewExtensionDungeon(6, 4)
	if d6.Boss.Name != "The Dungeon Master (Final Form)" {
		t.Errorf("expected boss 'The Dungeon Master (Final Form)', got %q", d6.Boss.Name)
	}
	// Deck size: 50 total boss doors/abilities (45 doors + 5 boss abilities) + 4*2 challenges = 58
	if len(d6.Doors) != 58 {
		t.Errorf("expected 58 doors in Final Form dungeon deck, got %d", len(d6.Doors))
	}

	// 3. Level clamping (e.g. lvl 99 clamps to BossList len)
	dMax := NewBaseDungeon(99, 2)
	if dMax.Boss == nil || dMax.Boss.Name != "The K.I.C.K. 9000" {
		t.Errorf("expected valid clamped boss for level 99, got %v", dMax.Boss)
	}

	// 4. PutDoorBelowDeck mechanics
	d := &Dungeon{
		Doors: []DungeonCard{
			&DoorCard{Name: "Door1"},
			&DoorCard{Name: "Door2"},
		},
	}
	bottomDoor := &DoorCard{Name: "BottomDoor"}
	d.PutDoorBelowDeck(bottomDoor)
	if len(d.Doors) != 3 || d.Doors[0] != bottomDoor {
		t.Errorf("expected bottomDoor at index 0, got %v", d.Doors[0])
	}

	// PutDoorBelowDeck ignores BossMat and EventCard
	d.PutDoorBelowDeck(&BossMat{Name: "Boss"})
	d.PutDoorBelowDeck(&EventCard{Name: "Event"})
	if len(d.Doors) != 3 {
		t.Errorf("expected BossMat and EventCard to not be added below deck, got len %d", len(d.Doors))
	}
}

// --- Player Edge Cases & Discard Mechanics ---

func TestPlayerNewValidation(t *testing.T) {
	// Druid & Shaman require extension
	_, err := NewPlayer("DruidPlayer", Druid, false)
	if err == nil {
		t.Error("expected error when creating Druid without extension, got nil")
	}
	_, err = NewPlayer("ShamanPlayer", Shaman, false)
	if err == nil {
		t.Error("expected error when creating Shaman without extension, got nil")
	}

	// Valid with extension
	pDruid, err := NewPlayer("DruidPlayer", Druid, true)
	if err != nil || pDruid == nil {
		t.Errorf("expected Druid created with extension, got error: %v", err)
	}
	pShaman, err := NewPlayer("ShamanPlayer", Shaman, true)
	if err != nil || pShaman == nil {
		t.Errorf("expected Shaman created with extension, got error: %v", err)
	}
}

func TestPlayerDrawResourceCardsFromDiscard(t *testing.T) {
	p, _ := NewPlayer("Arthur", Paladin, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Sword, Arrow}}
	c4 := &ResourceCard{Resources: []ResourceType{Jump}}

	p.Discard = &Discard{Cards: []PlayerCard{c1, c2, c3, c4}}
	p.Hand = []PlayerCard{}

	// Request Sword: c3 and c1 provide Sword
	events, err := p.DrawResourceCardsFromDiscard([]ResourceType{Sword})
	if err != nil {
		t.Fatalf("failed DrawResourceCardsFromDiscard: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 cards drawn providing Sword, got %d", len(events))
	}
	if len(p.Hand) != 2 {
		t.Errorf("expected 2 cards in hand, got %d", len(p.Hand))
	}
	if p.Discard.Length() != 2 {
		t.Errorf("expected 2 cards remaining in discard (c2, c4), got %d", p.Discard.Length())
	}
}
