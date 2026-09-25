package game

import (
	"context"
	"slices"
	"testing"
	"time"
)

// --- Unit Tests for All 20 Event Cards ---

func TestAmbushEvent(t *testing.T) {
	d1 := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	d2 := &DoorCard{Type: DoorObstacle, Name: "Pit", Resources: []ResourceType{Jump}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{d1, d2},
	}
	p1, _ := NewPlayer("P1", Paladin, true)
	game := &Game{
		Players:   []*Player{p1},
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Ambush!", Action: AmbushEvent{}}
	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
	}

	events, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("AmbushEvent execution failed: %v", err)
	}

	// Should have opened 2 doors onto the playfield
	if len(game.PlayField.OpenedDoors) != 2 {
		t.Fatalf("expected 2 opened doors, got %d", len(game.PlayField.OpenedDoors))
	}
	if !game.PlayField.HasActiveDoor(d1) || !game.PlayField.HasActiveDoor(d2) {
		t.Errorf("expected d1 and d2 to be active on playfield, got %v", game.PlayField.OpenedDoors)
	}

	var doorOpenCount int
	for _, e := range events {
		if _, ok := e.(DoorOpenedEvent); ok {
			doorOpenCount++
		}
	}
	if doorOpenCount != 2 {
		t.Errorf("expected 2 DoorOpenedEvent in events, got %d", doorOpenCount)
	}
}

func TestDungeonErrorInYourFavorEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	dCard := &ResourceCard{Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{}
	p1.Deck = &Deck{Cards: []PlayerCard{dCard, dCard, dCard, dCard, dCard, dCard}}
	p2.Hand = []PlayerCard{}
	p2.Deck = &Deck{Cards: []PlayerCard{dCard, dCard, dCard}} // only 3 cards in deck

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Dungeon Error in Your Favor", Action: DungeonErrorInYourFavorEvent{}}
	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("DungeonErrorInYourFavorEvent execution failed: %v", err)
	}

	// P1 should have drawn 5 cards
	if len(p1.Hand) != 5 {
		t.Errorf("expected P1 to have drawn 5 cards, got %d", len(p1.Hand))
	}
	// P2 should have drawn all available 3 cards without error
	if len(p2.Hand) != 3 {
		t.Errorf("expected P2 to have drawn 3 cards, got %d", len(p2.Hand))
	}
}

func TestSuddenIllnessEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}

	p1.Hand = []PlayerCard{c1, c2}
	p2.Hand = []PlayerCard{c1}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Sudden Illness", Action: SuddenIllnessEvent{}}
	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("SuddenIllnessEvent failed: %v", err)
	}

	if len(p1.Hand) != 5 || p1.Discard.Length() != 2 {
		t.Errorf("expected P1 hand to be refiled to 5 cards and discard to have 2 cards, got hand: %d, discard: %d", len(p1.Hand), p1.Discard.Length())
	}
	if len(p2.Hand) != 5 || p2.Discard.Length() != 1 {
		t.Errorf("expected P2 hand to be refiled to 5 cards to have 1 card, got hand: %d, discard: %d", len(p2.Hand), p2.Discard.Length())
	}
}

func TestCrowdFundingEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{c1}
	p2.Hand = []PlayerCard{c1}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Crowd Funding", Action: CrowdFundingEvent{}}
	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("CrowdFundingEvent failed: %v", err)
	}

	if len(p1.Hand) != 5 || len(p2.Hand) != 5 {
		t.Errorf("expected all players to discard hands and then refiled to 5 cards, got P1: %d, P2: %d", len(p1.Hand), len(p2.Hand))
	}
}

func TestYetMoreSpikesEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}

	p1.Hand = []PlayerCard{c1, c2}
	p2.Hand = []PlayerCard{c1}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Yet More Spikes!", Action: YetMoreSpikesEvent{}}
	interaction := &TeamChoicePlayerInteraction{
		card:             eventCard,
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player]*Player{p1: p1, p2: p1}, // Voted P1
		OnComplete:       YetMoreSpikesEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("YetMoreSpikesEvent failed: %v", err)
	}

	// P1 was targeted -> discarded entire hand
	if len(p1.Hand) != 5 || p1.Discard.Length() != 2 {
		t.Errorf("expected P1 hand to be refiled to 5 cards and discard length 5, got hand %d, discard %d", len(p1.Hand), p1.Discard.Length())
	}
	// P2 was not targeted -> kept their hand
	if len(p2.Hand) != 1 {
		t.Errorf("expected P2 to keep 1 card, got %d", len(p2.Hand))
	}
}

func TestGimmeAHandEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	p3, _ := NewPlayer("P3", Wizard, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Scroll}}

	p1.Hand = []PlayerCard{c1}
	p2.Hand = []PlayerCard{c2}
	p3.Hand = []PlayerCard{c3}

	game := &Game{
		Players:   []*Player{p1, p2, p3},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Gimme a Hand!", Action: GimmeAHandEvent{}}
	interaction := &TeamChoicePlayerInteraction{
		card:             eventCard,
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player]*Player{p1: p1, p2: p1, p3: p1}, // Target P1
		OnComplete:       GimmeAHandEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("GimmeAHandEvent failed: %v", err)
	}

	// P1 should now hold all 3 cards (c1 + c2 + c3)
	if len(p1.Hand) != 3 {
		t.Errorf("expected P1 to hold 3 cards, got %d", len(p1.Hand))
	}
	if !p1.HasCardInHand(c1) || !p1.HasCardInHand(c2) || !p1.HasCardInHand(c3) {
		t.Errorf("expected P1 to have c1, c2, c3, got %v", p1.Hand)
	}
	// Teammates should have 0 cards in hand
	if len(p2.Hand) != 0 || len(p3.Hand) != 0 {
		t.Errorf("expected P2 and P3 hands to be empty, got P2: %d, P3: %d", len(p2.Hand), len(p3.Hand))
	}
}

func TestLockedDoorEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}
	multiCard := &ResourceCard{Resources: []ResourceType{Sword, Jump}}

	p1.Hand = []PlayerCard{swordCard, shieldCard}
	p2.Hand = []PlayerCard{multiCard, shieldCard}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Locked Door!", Action: LockedDoorEvent{}}
	interaction := &TeamChoiceResourceInteraction{
		card:             eventCard,
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player]ResourceType{p1: Sword, p2: Sword}, // Voted Sword
		OnComplete:       LockedDoorEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("LockedDoorEvent failed: %v", err)
	}

	// P1 had [swordCard, shieldCard] -> swordCard discarded, shieldCard remains
	if len(p1.Hand) != 1 || p1.Hand[0] != shieldCard {
		t.Errorf("expected P1 hand to have [shieldCard], got %v", p1.Hand)
	}
	// P2 had [multiCard(Sword,Jump), shieldCard] -> multiCard discarded because it contains Sword
	if len(p2.Hand) != 1 || p2.Hand[0] != shieldCard {
		t.Errorf("expected P2 hand to have [shieldCard], got %v", p2.Hand)
	}
}

func TestABooBooEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}

	p1.Hand = []PlayerCard{c1, c2}
	p2.Hand = []PlayerCard{c3}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "A Boo-Boo", Action: ABooBooEvent{}}
	interaction := &PlayerDiscardCardsInteraction{
		card:             eventCard,
		requiredCounts:   map[*Player]int{p1: 1, p2: 1},
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player][]PlayerCard{p1: {c1}, p2: {c3}},
		OnComplete:       ABooBooEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("ABooBooEvent failed: %v", err)
	}

	// P1 discarded c1 (in discard), hand refilled to 5 and retains c2
	if p1.Discard.Length() != 1 {
		t.Errorf("expected P1 discard to have 1 card, got %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 5 || !p1.HasCardInHand(c2) {
		t.Errorf("expected P1 hand to be refilled to 5 cards and retain c2, got %v", p1.Hand)
	}
	// P2 discarded c3 (in discard), hand refilled to 5
	if p2.Discard.Length() != 1 {
		t.Errorf("expected P2 discard to have 1 card, got %d", p2.Discard.Length())
	}
	if len(p2.Hand) != 5 {
		t.Errorf("expected P2 hand to be refilled to 5 cards, got %d", len(p2.Hand))
	}
}

func TestTrapDoorEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c4 := &ResourceCard{Resources: []ResourceType{Jump}}

	p1.Hand = []PlayerCard{c1, c2, c3, c4}
	p2.Hand = []PlayerCard{c1, c2} // only 2 cards

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Trap Door", Action: TrapDoorEvent{}}
	interaction := &PlayerDiscardCardsInteraction{
		card:             eventCard,
		requiredCounts:   map[*Player]int{p1: 3, p2: 3},
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player][]PlayerCard{p1: {c1, c2, c3}, p2: {c1, c2}},
		OnComplete:       TrapDoorEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("TrapDoorEvent failed: %v", err)
	}

	// P1 discarded 3 cards (c1, c2, c3), kept c4 and refilled to 5
	if p1.Discard.Length() != 3 {
		t.Errorf("expected P1 discard to have 3 cards, got %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 5 || !p1.HasCardInHand(c4) {
		t.Errorf("expected P1 hand to be refilled to 5 cards and retain c4, got %v", p1.Hand)
	}
	// P2 had 2 cards and discarded both, refilled to 5
	if p2.Discard.Length() != 2 {
		t.Errorf("expected P2 discard to have 2 cards, got %d", p2.Discard.Length())
	}
	if len(p2.Hand) != 5 {
		t.Errorf("expected P2 hand to be refilled to 5 cards, got %d", len(p2.Hand))
	}
}

func TestConfusionEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}

	p1.Hand = []PlayerCard{c1, c2}
	p2.Hand = []PlayerCard{c3}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Confusion", Action: ConfusionEvent{}}
	// P1 passes hand to P2, P2 passes hand to P1
	interaction := &PlayerDonatesHandInteraction{
		card:             eventCard,
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player]*Player{p1: p2, p2: p1},
		OnComplete:       ConfusionEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	events, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("ConfusionEvent failed: %v", err)
	}

	// P1 should now have P2's previous hand [c3]
	if len(p1.Hand) != 1 || p1.Hand[0] != c3 {
		t.Errorf("expected P1 hand to have [c3], got %v", p1.Hand)
	}
	// P2 should now have P1's previous hand [c1, c2]
	if len(p2.Hand) != 2 || !p2.HasCardInHand(c1) || !p2.HasCardInHand(c2) {
		t.Errorf("expected P2 hand to have [c1, c2], got %v", p2.Hand)
	}

	var donatedCount int
	for _, e := range events {
		if _, ok := e.(HandDonatedEvent); ok {
			donatedCount++
		}
	}
	if donatedCount != 2 {
		t.Errorf("expected 2 HandDonatedEvent in events, got %d", donatedCount)
	}
}

func TestAnUngodlyAmountOfPorcupinesEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	dCard := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{}
	p1.Deck = &Deck{Cards: []PlayerCard{dCard, dCard, dCard, dCard}}

	game := &Game{
		Players:   []*Player{p1},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "An Ungodly Amount of Porcupines", Action: AnUngodlyAmountOfPorcupinesEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	// Interaction returns the interaction descriptor, Init triggers the 3-card draw phase
	interaction := eventCard.Action.Interaction(ctx)
	initEvents, err := interaction.Init(ctx)
	if err != nil {
		t.Fatalf("interaction.Init failed: %v", err)
	}
	if len(initEvents) != 3 {
		t.Errorf("expected 3 draw events from Init, got %d", len(initEvents))
	}
	if len(p1.Hand) != 3 {
		t.Fatalf("expected P1 to have drawn 3 cards during interaction init, got %d", len(p1.Hand))
	}

	// Submit discard of the 3 cards
	pdi := interaction.(*PlayerDiscardCardsInteraction)
	pdi.CollectedChoices[p1] = slices.Clone(p1.Hand)
	ctx.Input = pdi

	var execEvents []Event
	execEvents, err = eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("AnUngodlyAmountOfPorcupinesEvent execution failed: %v", err)
	}
	_ = execEvents

	if len(p1.Hand) != 0 || p1.Discard.Length() != 3 {
		t.Errorf("expected P1 hand to be empty after discarding 3 cards, got hand: %d, discard: %d", len(p1.Hand), p1.Discard.Length())
	}
}

func TestPoisonedMilkEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	p3, _ := NewPlayer("P3", Wizard, true)

	c := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{c, c, c, c} // size 4 (tied largest)
	p2.Hand = []PlayerCard{c, c, c, c} // size 4 (tied largest)
	p3.Hand = []PlayerCard{c, c}       // size 2

	game := &Game{
		Players:   []*Player{p1, p2, p3},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Poisoned Milk", Action: PoisonedMilkEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("PoisonedMilkEvent failed: %v", err)
	}

	// P1 and P2 (tied for largest hand) should discard entire hand and refill up to 4
	if len(p1.Hand) != 4 || p1.Discard.Length() != 4 {
		t.Errorf("expected P1 hand to be discarded and refilled to 4, got hand %d, discard %d", len(p1.Hand), p1.Discard.Length())
	}
	if len(p2.Hand) != 4 || p2.Discard.Length() != 4 {
		t.Errorf("expected P2 hand to be discarded and refilled to 4, got hand %d, discard %d", len(p2.Hand), p2.Discard.Length())
	}
	// P3 (smaller hand) should keep their 2 cards without discard
	if len(p3.Hand) != 2 || p3.Discard.Length() != 0 {
		t.Errorf("expected P3 to keep hand of 2 cards without discard, got hand %d, discard %d", len(p3.Hand), p3.Discard.Length())
	}
}

func TestAcidPolishEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{shieldCard, swordCard} // has Shield -> discards hand
	p2.Hand = []PlayerCard{swordCard, swordCard}  // no Shield -> keeps hand

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Acid Polish", Action: AcidPolishEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("AcidPolishEvent failed: %v", err)
	}

	if len(p1.Hand) != 5 || p1.Discard.Length() != 2 {
		t.Errorf("expected P1 (held Shield) to discard entire hand and refill to 5, got hand %d, discard %d", len(p1.Hand), p1.Discard.Length())
	}
	if len(p2.Hand) != 2 || p2.Discard.Length() != 0 {
		t.Errorf("expected P2 (no Shield) to keep hand of 2, got hand %d, discard %d", len(p2.Hand), p2.Discard.Length())
	}
}

func TestWaxedFloorEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	c := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{c, c, c, c, c, c} // 6 cards (> 5) -> discards hand
	p2.Hand = []PlayerCard{c, c, c, c, c}    // 5 cards (<= 5) -> keeps hand

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Waxed Floor", Action: WaxedFloorEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("WaxedFloorEvent failed: %v", err)
	}

	if len(p1.Hand) != 5 || p1.Discard.Length() != 6 {
		t.Errorf("expected P1 (6 cards) to discard entire hand and refill to 5, got hand %d, discard %d", len(p1.Hand), p1.Discard.Length())
	}
	if len(p2.Hand) != 5 || p2.Discard.Length() != 0 {
		t.Errorf("expected P2 (5 cards) to keep hand of 5, got hand %d, discard %d", len(p2.Hand), p2.Discard.Length())
	}
}

func TestCorrosiveSpitEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{shieldCard, swordCard}
	p2.Hand = []PlayerCard{swordCard}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Corrosive Spit", Action: CorrosiveSpitEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("CorrosiveSpitEvent failed: %v", err)
	}

	if p1.Discard.Length() != 1 {
		t.Errorf("expected P1 discard to have 1 card, got %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 5 || !p1.HasCardInHand(swordCard) {
		t.Errorf("expected P1 to keep swordCard and refill to 5, got %v", p1.Hand)
	}
}

func TestEnsnaredEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Ranger, true)
	p2, _ := NewPlayer("P2", Paladin, true)
	jumpCard := &ResourceCard{Resources: []ResourceType{Jump}}
	arrowCard := &ResourceCard{Resources: []ResourceType{Arrow}}

	p1.Hand = []PlayerCard{jumpCard, arrowCard}
	p2.Hand = []PlayerCard{arrowCard}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Ensnared!", Action: EnsnaredEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("EnsnaredEvent failed: %v", err)
	}

	if p1.Discard.Length() != 1 {
		t.Errorf("expected P1 discard to have 1 card, got %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 5 || !p1.HasCardInHand(arrowCard) {
		t.Errorf("expected P1 to keep arrowCard and refill to 5, got %v", p1.Hand)
	}
}

func TestMySwordsEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Barbarian, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}

	p1.Hand = []PlayerCard{swordCard, shieldCard}
	p2.Hand = []PlayerCard{shieldCard}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "My Swords!", Action: MySwordsEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("MySwordsEvent failed: %v", err)
	}

	if p1.Discard.Length() != 1 {
		t.Errorf("expected P1 discard to have 1 card, got %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 5 || !p1.HasCardInHand(shieldCard) {
		t.Errorf("expected P1 to keep shieldCard and refill to 5, got %v", p1.Hand)
	}
}

func TestFireBreathEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c := &ResourceCard{Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{c, c}
	p2.Hand = []PlayerCard{c, c}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Fire Breath", Action: FireBreathEvent{}}
	// Team chooses P1 to spare
	interaction := &TeamChoicePlayerInteraction{
		card:             eventCard,
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player]*Player{p1: p1, p2: p1},
		OnComplete:       FireBreathEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("FireBreathEvent failed: %v", err)
	}

	// P1 spared -> keeps hand
	if len(p1.Hand) != 2 || p1.Discard.Length() != 0 {
		t.Errorf("expected P1 to keep hand of 2 without discard, got hand %d, discard %d", len(p1.Hand), p1.Discard.Length())
	}
	// P2 not spared -> discards hand and refills to 5
	if len(p2.Hand) != 5 || p2.Discard.Length() != 2 {
		t.Errorf("expected P2 to discard entire hand (2 cards) and refill to 5, got hand %d, discard %d", len(p2.Hand), p2.Discard.Length())
	}
}

func TestTailSwipeEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Arrow}}

	p1.Hand = []PlayerCard{c1}
	p2.Hand = []PlayerCard{c2}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "Tail Swipe", Action: TailSwipeEvent{}}
	interaction := &PlayerDonatesHandInteraction{
		card:             eventCard,
		PendingPlayers:   map[*Player]bool{},
		CollectedChoices: map[*Player]*Player{p1: p2, p2: p1},
		OnComplete:       TailSwipeEvent{}.Execute,
	}

	ctx := &CardEventContext{
		engine: game,
		Card:   eventCard,
		Input:  interaction,
	}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("TailSwipeEvent failed: %v", err)
	}

	if len(p1.Hand) != 1 || p1.Hand[0] != c2 {
		t.Errorf("expected P1 to hold c2, got %v", p1.Hand)
	}
	if len(p2.Hand) != 1 || p2.Hand[0] != c1 {
		t.Errorf("expected P2 to hold c1, got %v", p2.Hand)
	}
}

func TestATwentySidedBoulderEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c := &ResourceCard{Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{c, c}
	p2.Hand = []PlayerCard{c}

	game := &Game{
		Players:   []*Player{p1, p2},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}

	eventCard := &EventCard{Name: "A 20-Sided Boulder", Action: ATwentySidedBoulderEvent{}}
	ctx := &CardEventContext{engine: game, Card: eventCard}

	_, err := eventCard.Action.Execute(ctx)
	if err != nil {
		t.Fatalf("ATwentySidedBoulderEvent failed: %v", err)
	}

	if p1.Discard.Length() != 2 || len(p1.Hand) != 5 {
		t.Errorf("expected P1 to discard 2 cards and refill to 5, got hand: %d, discard: %d", len(p1.Hand), p1.Discard.Length())
	}
	if p2.Discard.Length() != 1 || len(p2.Hand) != 5 {
		t.Errorf("expected P2 to discard 1 card and refill to 5, got hand: %d, discard: %d", len(p2.Hand), p2.Discard.Length())
	}
}

// --- Variations of Full Event Resolution Cycles ---

func TestEventResolutionCycle_ImmediateAutomaticEvent(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{c, c}
	p2.Hand = []PlayerCard{c}

	nextDoor := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{nextDoor},
	}

	eventCard := &EventCard{Name: "Sudden Illness", Action: SuddenIllnessEvent{}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  2,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	// Tick before 2 seconds elapsed -> event should NOT resolve yet
	game.InGameTimer = time.Second * 1
	events, err := game.Tick(time.Millisecond * 50)
	if err != nil {
		t.Fatalf("tick failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected no events before 2s elapsed, got %d events", len(events))
	}
	if !game.PlayField.HasActiveDoor(eventCard) {
		t.Error("expected eventCard to remain active on playfield before 2s")
	}

	// Advance timer past 2 seconds -> event auto-resolves, executes effect, and defeats eventCard
	game.InGameTimer = time.Second * 3
	events, err = game.Tick(time.Millisecond * 50)
	if err != nil {
		t.Fatalf("tick after 2s failed: %v", err)
	}

	// Hands should be discarded and refilled up to hand size (2)
	if p1.Discard.Length() != 2 || len(p1.Hand) != 2 {
		t.Errorf("expected P1 to discard 2 and refill to 2, got hand %d, discard %d", len(p1.Hand), p1.Discard.Length())
	}
	if p2.Discard.Length() != 1 || len(p2.Hand) != 2 {
		t.Errorf("expected P2 to discard 1 and refill to 2, got hand %d, discard %d", len(p2.Hand), p2.Discard.Length())
	}

	// Event card defeated and next door opened
	if game.PlayField.HasActiveDoor(eventCard) {
		t.Error("expected eventCard to be defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened after event resolution")
	}

	var doorDefeatedFound, doorOpenedFound bool
	for _, e := range events {
		if de, ok := e.(DoorDefeatedEvent); ok && de.DungeonCard == eventCard {
			doorDefeatedFound = true
		}
		if de, ok := e.(DoorOpenedEvent); ok && de.DungeonCard == nextDoor {
			doorOpenedFound = true
		}
	}
	if !doorDefeatedFound || !doorOpenedFound {
		t.Errorf("expected DoorDefeatedEvent and DoorOpenedEvent in events: %v", events)
	}
}

func TestEventResolutionCycle_TeamChoicePlayerInteraction(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{c, c}
	p2.Hand = []PlayerCard{c, c}

	nextDoor := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	eventCard := &EventCard{Name: "Yet More Spikes!", Action: YetMoreSpikesEvent{}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  2,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	// Advance timer to trigger prompt opening
	game.InGameTimer = time.Second * 3
	events, err := game.Tick(time.Millisecond * 50)
	if err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	// Verify EventPromptOpenedEvent emitted and PendingInteraction set
	if game.PendingInteraction == nil {
		t.Fatal("expected PendingInteraction to be set")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	promptEvent, ok := events[0].(EventPromptOpenedEvent)
	if !ok || promptEvent.Kind != InteractionTeamChoicePlayer {
		t.Fatalf("expected EventPromptOpenedEvent with InteractionTeamChoicePlayer, got %v", events[0])
	}

	// Normal card playing must be blocked while interaction is pending
	_, err = game.Apply(PlayCardCmd{Player: p1, Card: c})
	if err == nil {
		t.Error("expected error when playing card during pending interaction, got nil")
	}

	// Player 1 submits choice (votes P1) -> partial submission
	events, err = game.Apply(SubmitPromptChoiceCmd{
		Player:       p1,
		TargetPlayer: p1,
	})
	if err != nil {
		t.Fatalf("P1 SubmitEventChoice failed: %v", err)
	}
	if len(events) != 1 || events[0] != (PlayerEventChoiceSubmittedEvent{Player: p1}) {
		t.Errorf("expected PlayerEventChoiceSubmittedEvent for P1, got %v", events)
	}
	// Interaction is still pending (waiting for P2)
	if game.PendingInteraction == nil {
		t.Fatal("expected PendingInteraction to still be active")
	}

	// Player 2 submits choice (votes P1) -> interaction finishes
	events, err = game.Apply(SubmitPromptChoiceCmd{
		Player:       p2,
		TargetPlayer: p1,
	})
	if err != nil {
		t.Fatalf("P2 SubmitEventChoice failed: %v", err)
	}

	// PendingInteraction cleared
	if game.PendingInteraction != nil {
		t.Error("expected PendingInteraction to be cleared after completion")
	}

	// Effect executed: P1 discarded initial hand into Discard, and was refilled from deck
	if p1.Discard.Length() != 2 {
		t.Errorf("expected targeted P1 to have 2 cards in discard pile, got %d", p1.Discard.Length())
	}
	if len(p1.Hand) != 2 {
		t.Errorf("expected targeted P1 hand to be refilled to handSize (2), got %d", len(p1.Hand))
	}
	if len(p2.Hand) != 2 {
		t.Errorf("expected untargeted P2 to keep hand, got %d", len(p2.Hand))
	}

	// Event card defeated and next door opened
	if game.PlayField.HasActiveDoor(eventCard) {
		t.Error("expected eventCard to be defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened")
	}
}

func TestEventResolutionCycle_MultiPlayerDiscardBarrier(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c4 := &ResourceCard{Resources: []ResourceType{Jump}}

	p1.Hand = []PlayerCard{c1, c2}
	p2.Hand = []PlayerCard{c3, c4}

	nextDoor := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	eventCard := &EventCard{Name: "A Boo-Boo", Action: ABooBooEvent{}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  2,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	// Trigger prompt opening
	game.InGameTimer = time.Second * 3
	_, err := game.Tick(time.Millisecond * 50)
	if err != nil {
		t.Fatalf("tick failed: %v", err)
	}
	if game.PendingInteraction == nil {
		t.Fatal("expected PendingInteraction for A Boo-Boo")
	}

	// P1 submits invalid card (not in hand) -> rejected
	notInHand := &ResourceCard{Resources: []ResourceType{Scroll}}
	_, err = game.Apply(SubmitPromptChoiceCmd{
		Player: p1,
		Cards:  []PlayerCard{notInHand},
	})
	if err == nil {
		t.Error("expected error for submitting card not in hand, got nil")
	}

	// P1 submits valid card (c1)
	_, err = game.Apply(SubmitPromptChoiceCmd{
		Player: p1,
		Cards:  []PlayerCard{c1},
	})
	if err != nil {
		t.Fatalf("P1 valid submit failed: %v", err)
	}

	// P2 submits valid card (c3)
	_, err = game.Apply(SubmitPromptChoiceCmd{
		Player: p2,
		Cards:  []PlayerCard{c3},
	})
	if err != nil {
		t.Fatalf("P2 valid submit failed: %v", err)
	}

	// Interaction cleared and both players discarded their chosen card
	if game.PendingInteraction != nil {
		t.Error("expected PendingInteraction to be cleared")
	}
	if p1.Discard.Length() != 1 || p1.Discard.Cards[0] != c1 {
		t.Errorf("expected P1 to have discarded c1, got %v", p1.Discard.Cards)
	}
	if p2.Discard.Length() != 1 || p2.Discard.Cards[0] != c3 {
		t.Errorf("expected P2 to have discarded c3, got %v", p2.Discard.Cards)
	}
	if !p1.HasCardInHand(c2) || len(p1.Hand) != 2 {
		t.Errorf("expected P1 to keep c2 and be refilled to 2 cards, got %v", p1.Hand)
	}
	if !p2.HasCardInHand(c4) || len(p2.Hand) != 2 {
		t.Errorf("expected P2 to keep c4 and be refilled to 2 cards, got %v", p2.Hand)
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be active")
	}
}

func TestEventResolutionCycle_TeamChoiceResourceInteraction(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	shieldCard := &ResourceCard{Resources: []ResourceType{Shield}}
	p1.Hand = []PlayerCard{swordCard, shieldCard}
	p2.Hand = []PlayerCard{swordCard}

	nextDoor := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	eventCard := &EventCard{Name: "Locked Door!", Action: LockedDoorEvent{}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  2,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	// Trigger prompt
	game.InGameTimer = time.Second * 3
	_, _ = game.Tick(time.Millisecond * 50)

	res := Sword
	// P1 votes Sword
	_, err := game.Apply(SubmitPromptChoiceCmd{
		Player:   p1,
		Resource: &res,
	})
	if err != nil {
		t.Fatalf("P1 vote failed: %v", err)
	}

	// P2 votes Sword -> finishes interaction
	_, err = game.Apply(SubmitPromptChoiceCmd{
		Player:   p2,
		Resource: &res,
	})
	if err != nil {
		t.Fatalf("P2 vote failed: %v", err)
	}

	// P1 discarded swordCard, P2 discarded swordCard
	if p1.Discard.Length() != 1 || p1.Discard.Cards[0] != swordCard {
		t.Errorf("expected P1 to have discarded swordCard, got %v", p1.Discard.Cards)
	}
	if p2.Discard.Length() != 1 || p2.Discard.Cards[0] != swordCard {
		t.Errorf("expected P2 to have discarded swordCard, got %v", p2.Discard.Cards)
	}
	if !p1.HasCardInHand(shieldCard) {
		t.Errorf("expected P1 to keep shieldCard, got %v", p1.Hand)
	}
	if p2.HasCardInHand(swordCard) {
		t.Errorf("expected P2 to have discarded swordCard, got %v", p2.Hand)
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened")
	}
}

func TestEventResolutionCycle_CanceledByCancelAction(t *testing.T) {
	wizard, _ := NewPlayer("Gandalf", Wizard, true)
	cancelCard := &ActionCard{Name: "Cancel", Action: CancelAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{cancelCard}
	wizard.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	eventCard := &EventCard{Name: "Sudden Illness", Action: SuddenIllnessEvent{}}
	nextDoor := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{wizard},
		HandSize:  1,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	// Before 2 seconds elapses, Wizard plays Cancel card
	events, err := game.Apply(PlayCardCmd{Player: wizard, Card: cancelCard})
	if err != nil {
		t.Fatalf("Cancel play failed: %v", err)
	}

	// Event card defeated, effect did not run (hand refilled, not discarded)
	if game.PlayField.HasActiveDoor(eventCard) {
		t.Error("expected eventCard to be defeated by Cancel")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be active")
	}

	var counterEvtFound bool
	for _, e := range events {
		if ce, ok := e.(EventCounteredEvent); ok && ce.ByPlayer == wizard && ce.EventCard == eventCard {
			counterEvtFound = true
		}
	}
	if !counterEvtFound {
		t.Errorf("expected EventCounteredEvent in events: %v", events)
	}
}

func TestEventResolutionCycle_RunnerIntegration(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)
	c := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{c, c}
	p2.Hand = []PlayerCard{c}

	eventCard := &EventCard{Name: "Yet More Spikes!", Action: YetMoreSpikesEvent{}, OpenedTime: -2 * time.Second}
	boss := &BossMat{Name: "Boss", Resources: []ResourceType{Scroll}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  2,
		Dungeon:   &Dungeon{Boss: boss, Doors: []DungeonCard{boss}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	runner := NewRunner(game)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- runner.Run(ctx)
	}()

	sub := runner.Subscribe()

	// Wait until prompt opened event is received
	var promptOpened bool
	timeout := time.After(3 * time.Second)
	for !promptOpened {
		select {
		case evt, ok := <-sub:
			if !ok {
				t.Fatal("subscriber channel closed unexpectedly")
			}
			if pe, ok := evt.(EventPromptOpenedEvent); ok && pe.Kind == InteractionTeamChoicePlayer {
				promptOpened = true
			}
		case <-timeout:
			t.Fatal("timed out waiting for EventPromptOpenedEvent")
		}
	}

	// Submit event choices through Runner
	err := runner.SubmitPromptChoice(ctx, p1, p2, nil, nil) // P1 votes P2
	if err != nil {
		t.Fatalf("P1 SubmitEventChoice failed: %v", err)
	}
	err = runner.SubmitPromptChoice(ctx, p2, p2, nil, nil) // P2 votes P2
	if err != nil {
		t.Fatalf("P2 SubmitEventChoice failed: %v", err)
	}

	// Give runner brief tick to complete door transition
	time.Sleep(100 * time.Millisecond)

	// P2 was chosen by team -> discarded hand into discard pile, refilled from deck
	if p2.Discard.Length() != 1 {
		t.Errorf("expected P2 to have 1 discarded card, got %d", p2.Discard.Length())
	}
	if len(p2.Hand) != 2 {
		t.Errorf("expected P2 hand to be refilled to 2, got %d", len(p2.Hand))
	}

	cancel()
	<-runErrCh
}

func TestEventResolutionCycle_HybridDrawAndDiscardPrompt(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c4 := &ResourceCard{Resources: []ResourceType{Jump}}

	p1.Hand = []PlayerCard{}
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c4}}

	p2.Hand = []PlayerCard{}
	p2.Deck = &Deck{Cards: []PlayerCard{c1, c2, c3, c4}}

	eventCard := &EventCard{Name: "An Ungodly Amount of Porcupines", Action: AnUngodlyAmountOfPorcupinesEvent{}}
	nextDoor := &DoorCard{Type: DoorObstacle, Name: "Trap", Resources: []ResourceType{Jump}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{nextDoor},
	}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  2,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	// Tick after 2s triggers event resolution -> runs Init() which draws 3 cards per player
	game.InGameTimer = 3 * time.Second
	events, err := game.Tick(50 * time.Millisecond)
	if err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	// Verify both players drew 3 cards during Init()
	if len(p1.Hand) != 3 || len(p2.Hand) != 3 {
		t.Fatalf("expected both players to have 3 cards in hand, got P1: %d, P2: %d", len(p1.Hand), len(p2.Hand))
	}

	// Verify CardDrawnFromDeckEvent and EventPromptOpenedEvent were emitted
	var drawEventCount int
	var promptOpenedFound bool
	for _, e := range events {
		if _, ok := e.(CardDrawnFromDeckEvent); ok {
			drawEventCount++
		}
		if pe, ok := e.(EventPromptOpenedEvent); ok && pe.Kind == InteractionPlayerDiscardCards {
			promptOpenedFound = true
		}
	}
	if drawEventCount != 6 {
		t.Errorf("expected 6 CardDrawnFromDeckEvent (3 per player), got %d", drawEventCount)
	}
	if !promptOpenedFound {
		t.Error("expected EventPromptOpenedEvent in events")
	}

	// Discard choices submitted
	p1CardsToDiscard := slices.Clone(p1.Hand)
	_, err = game.Apply(SubmitPromptChoiceCmd{
		Player: p1,
		Cards:  p1CardsToDiscard,
	})
	if err != nil {
		t.Fatalf("P1 SubmitEventChoice failed: %v", err)
	}

	p2CardsToDiscard := slices.Clone(p2.Hand)
	events, err = game.Apply(SubmitPromptChoiceCmd{
		Player: p2,
		Cards:  p2CardsToDiscard,
	})
	if err != nil {
		t.Fatalf("P2 SubmitEventChoice failed: %v", err)
	}

	// Both hands discarded 3 cards, and drew remaining deck cards (1 each)
	if p1.Discard.Length() != 3 || p2.Discard.Length() != 3 {
		t.Errorf("expected both players' discard to have 3 cards, got P1: %d, P2: %d", p1.Discard.Length(), p2.Discard.Length())
	}
	if len(p1.Hand) != 1 || len(p2.Hand) != 1 {
		t.Errorf("expected both players' hands to be 1 (refilled from remaining deck), got P1: %d, P2: %d", len(p1.Hand), len(p2.Hand))
	}

	// Event door defeated and next door opened
	if game.PlayField.HasActiveDoor(eventCard) {
		t.Error("expected eventCard to be defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened")
	}
	if game.PendingInteraction != nil {
		t.Error("expected PendingInteraction to be cleared")
	}
}
