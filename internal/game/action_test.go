package game

import (
	"slices"
	"testing"
	"time"
)

func TestActionCardHolyHandGrenade(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	doorMonster := &DoorCard{Id: 10, Type: DoorMonster, Name: "Dragon", Resources: []ResourceType{Sword, Shield}}
	doorObstacle := &DoorCard{Id: 20, Type: DoorObstacle, Name: "Wall", Resources: []ResourceType{Jump}}
	bossMat := &BossMat{Id: 30, Name: "Boss", Resources: []ResourceType{Scroll}}

	hhgCard := &ActionCard{Id: 666, Name: "Holy Hand Grenade", Action: HolyHandGrenadeAction{}}
	drawCard := &ResourceCard{Id: 1, Resources: []ResourceType{Sword}}

	paladin.Hand = []PlayerCard{hhgCard}
	paladin.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	dungeon := &Dungeon{
		Boss:  bossMat,
		Doors: []DungeonCard{doorObstacle},
	}

	game := &Game{
		Players:   []*Player{paladin},
		HandSize:  1,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, game)
	registerCardsInTestGame(game)

	// 1. Play HHG auto-targets the single active door
	events, err := game.Apply(PlayCardCmd{PlayerID: paladin.Id, CardID: hhgCard.ID()})
	if err != nil {
		t.Fatalf("failed to play Holy Hand Grenade: %v", err)
	}

	// Verify doorMonster was defeated and next door opened
	if game.PlayField.HasActiveDoor(doorMonster) {
		t.Error("expected doorMonster to be defeated")
	}
	if !game.PlayField.HasActiveDoor(doorObstacle) {
		t.Error("expected doorObstacle to be opened")
	}

	// Verify HHG was played on playfield and cleared upon door defeat, hand auto-refilled
	if len(paladin.Hand) != 1 || paladin.Hand[0] != drawCard {
		t.Errorf("expected hand to contain [drawCard], got %v", paladin.Hand)
	}
	if paladin.Discard.Length() != 0 {
		t.Errorf("expected discard to be empty (action card cleared with playfield), got %v", paladin.Discard.Cards)
	}
	if len(game.PlayField.Field) != 0 {
		t.Errorf("expected playfield to be empty after door defeat, got %v", game.PlayField.Field)
	}

	// Verify events: CardPlayedEvent, DoorDefeatedEvent, etc.
	var playedEvtFound, defeatEvtFound bool
	for _, e := range events {
		if _, ok := e.(CardPlayedEvent); ok {
			playedEvtFound = true
		}
		if dev, ok := e.(DoorDefeatedEvent); ok && dev.CardID == doorMonster.ID() {
			defeatEvtFound = true
		}
	}
	if !playedEvtFound || !defeatEvtFound {
		t.Errorf("expected CardPlayedEvent and DoorDefeatedEvent in %v", events)
	}

	// 2. Targeting BossMat while it is not yet active on the playfield must be rejected
	paladin.Hand = []PlayerCard{hhgCard}
	registerCardsInTestGame(game)
	_, err = game.Apply(PlayCardCmd{
		PlayerID:     paladin.Id,
		CardID:       hhgCard.ID(),
		TargetCardID: bossMat.ID(),
	})
	if err == nil {
		t.Error("expected error when targeting BossMat with Holy Hand Grenade, got nil")
	}

	// 3. Play HHG when only BossMat is active should work
	game.PlayField.OpenedDoors = []DungeonCard{bossMat}
	paladin.Hand = []PlayerCard{hhgCard}
	registerCardsInTestGame(game)
	_, err = game.Apply(PlayCardCmd{PlayerID: paladin.Id, CardID: hhgCard.ID()})
	if err != nil {
		t.Errorf("expected no error when only BossMat is active, got: %v", err)
	}
}

func TestActionCardHeal(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	healCard := &ActionCard{Id: 666, Name: "Heal", Action: HealAction{}}
	drawCard := &ResourceCard{Id: 1, Resources: []ResourceType{Sword}}

	p1.Hand = []PlayerCard{healCard}
	p1.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	c1 := &ResourceCard{Id: 2, Resources: []ResourceType{Shield}}
	c2 := &ResourceCard{Id: 3, Resources: []ResourceType{Arrow}}
	p2.Discard.PutAtop(c1, c2)
	p2.Deck = &Deck{Cards: []PlayerCard{}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: healCard.ID()})
	if err != nil {
		t.Fatalf("failed to play Heal action: %v", err)
	}

	// Verify P2 discard moved to top of P2 deck
	if p2.Discard.Length() != 0 {
		t.Errorf("expected P2 discard to be empty, got %d", p2.Discard.Length())
	}
	if len(p2.Deck.Cards) != 2 {
		t.Fatalf("expected P2 deck to have 2 recycled cards, got %d", len(p2.Deck.Cards))
	}

	// Verify PlayerHealedEvent emitted
	var healFound bool
	for _, e := range events {
		if he, ok := e.(PlayerHealedEvent); ok && he.PlayerID == p2.Id {
			healFound = true
			if len(he.CardIDs) != 2 {
				t.Errorf("expected 2 healed cards, got %d", len(he.CardIDs))
			}
		}
	}
	if !healFound {
		t.Errorf("expected PlayerHealedEvent in events: %v", events)
	}

	// Verify P1 hand and playfield/discard
	if len(p1.Hand) != 1 || p1.Hand[0] != drawCard {
		t.Errorf("expected P1 hand to refill to [drawCard]")
	}
	if p1.Discard.Length() != 0 {
		t.Errorf("expected P1 discard to be empty, got %d", p1.Discard.Length())
	}
	if len(game.PlayField.Field) != 1 || game.PlayField.Field[0] != healCard {
		t.Errorf("expected healCard to be on playfield, got %v", game.PlayField.Field)
	}
}

func TestActionCardHealthPotion(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	potCard := &ActionCard{Id: 666, Name: "Health Potion", Action: HealthPotionAction{}}
	p1.Hand = []PlayerCard{potCard}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Sword}}}}

	// P1 has 4 cards in discard -> should draw 3 (top 3)
	c1 := &ResourceCard{Resources: []ResourceType{Shield}}
	c2 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	c4 := &ResourceCard{Resources: []ResourceType{Scroll}}
	p1.Discard.PutAtop(c1, c2, c3, c4)

	// P2 has 1 card in discard -> should draw 1 without error
	p2c1 := &ResourceCard{Resources: []ResourceType{WildCard}}
	p2.Discard.PutAtop(p2c1)
	p2.Hand = []PlayerCard{}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: potCard.Id})
	if err != nil {
		t.Fatalf("failed to play Health Potion: %v", err)
	}

	// P1: played potCard (placed on playfield), drew 3 from discard (c4, c3, c2), and auto-drew 1 from deck
	// P1 discard had [c1, c2, c3, c4] -> drew top 3 (c4, c3, c2), leaving [c1]
	if p1.Discard.Length() != 1 || p1.Discard.Cards[0] != c1 {
		t.Errorf("expected P1 discard to have 1 card [c1] remaining, got %v", p1.Discard.Cards)
	}
	if len(game.PlayField.Field) != 1 || game.PlayField.Field[0] != potCard {
		t.Errorf("expected potCard to be on playfield, got %v", game.PlayField.Field)
	}

	// P2: drew 1 card from discard, leaving 0
	if p2.Discard.Length() != 0 {
		t.Errorf("expected P2 discard to be empty, got %d", p2.Discard.Length())
	}
	if len(p2.Hand) != 1 || p2.Hand[0] != p2c1 {
		t.Errorf("expected P2 hand to have [p2c1], got %v", p2.Hand)
	}

	// Verify CardDrawnFromDiscardEvent emitted
	discardDrawnCount := 0
	for _, e := range events {
		if _, ok := e.(CardDrawnFromDiscardEvent); ok {
			discardDrawnCount++
		}
	}
	if discardDrawnCount != 4 { // 3 for P1 + 1 for P2
		t.Errorf("expected 4 CardDrawnFromDiscardEvent, got %d", discardDrawnCount)
	}
}

func TestActionCardMysticRune(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	runeCard := &ActionCard{Id: 666, Name: "Mystic Rune", Action: MysticRuneAction{}}
	draw1 := &ResourceCard{Resources: []ResourceType{Sword}}
	draw2 := &ResourceCard{Resources: []ResourceType{Shield}}
	drawDeck := &ResourceCard{Resources: []ResourceType{Arrow}}

	p1.Hand = []PlayerCard{runeCard}
	p1.Deck = &Deck{Cards: []PlayerCard{draw1, drawDeck}}
	p2.Hand = []PlayerCard{}
	p2.Deck = &Deck{Cards: []PlayerCard{draw2}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(&CurseCard{
		Type:   ChallengeCurse,
		Name:   "Tourbillon de Wazaa",
		Effect: TimeCannotBeStopped,
	}, game)
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: runeCard.Id})
	if err != nil {
		t.Fatalf("failed to play Mystic Rune: %v", err)
	}

	// P2 drew 1 card from deck
	if len(p2.Hand) != 1 || p2.Hand[0] != draw2 {
		t.Errorf("expected P2 hand to have [draw2], got %v", p2.Hand)
	}

	// Verify CardDrawnFromDeckEvent
	var p2DrawnFound bool
	for _, e := range events {
		if de, ok := e.(CardDrawnFromDeckEvent); ok && de.ByPlayerID == p2.Id && de.CardID == draw2.Id {
			p2DrawnFound = true
		}
	}
	if !p2DrawnFound {
		t.Errorf("expected P2 CardDrawnFromDeckEvent in %v", events)
	}
}

func TestActionCardRally(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	rallyCard := &ActionCard{Id: 666, Name: "Rally", Action: RallyAction{}}
	p1.Hand = []PlayerCard{rallyCard}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Sword}}}}

	// P1 Discard: bottom [Scroll, Sword, Jump, Shield, Arrow] top
	// Matching cards: Sword, Shield -> should be drawn into hand in LIFO order (Shield then Sword)
	// Remaining in discard: Scroll, Jump, Arrow (plus rallyCard if put in discard after, or before)
	cScroll := &ResourceCard{Resources: []ResourceType{Scroll}}
	cSword := &ResourceCard{Resources: []ResourceType{Sword}}
	cJump := &ResourceCard{Resources: []ResourceType{Jump}}
	cShield := &ResourceCard{Resources: []ResourceType{Shield}}
	cArrow := &ResourceCard{Resources: []ResourceType{Arrow}}

	p1.Discard.PutAtop(cScroll, cSword, cJump, cShield, cArrow)

	// P2 has empty discard -> should not error
	p2.Discard = NewDiscard()

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: rallyCard.Id})
	if err != nil {
		t.Fatalf("failed to play Rally action: %v", err)
	}

	// Verify matching cards drawn into P1 hand
	if !p1.HasCardInHand(cSword) || !p1.HasCardInHand(cShield) {
		t.Errorf("expected P1 hand to contain Sword and Shield cards, got %v", p1.Hand)
	}

	// Verify non-matching cards remained in discard
	if p1.Discard.Length() < 3 {
		t.Errorf("expected non-matching cards to remain in discard, got %d", p1.Discard.Length())
	}

	// Verify events
	var drawnSword, drawnShield bool
	for _, e := range events {
		if de, ok := e.(CardDrawnFromDiscardEvent); ok {
			if de.CardID == cSword.Id {
				drawnSword = true
			}
			if de.CardID == cShield.Id {
				drawnShield = true
			}
		}
	}
	if !drawnSword || !drawnShield {
		t.Errorf("expected CardDrawnFromDiscardEvent for both Sword and Shield, got events: %v", events)
	}
}

func TestActionCardFailedPlayKeepsCardInHand(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	snipeCard := &ActionCard{Id: 666, Name: "Snipe", Action: SnipeAction{}} // Needs DoorPerson
	ranger.Hand = []PlayerCard{snipeCard}

	doorMonster := &DoorCard{Type: DoorMonster, Name: "Goblin"}

	game := &Game{
		Players:   []*Player{ranger},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Dungeon: &Dungeon{
			Boss:  nil,
			Doors: make([]DungeonCard, 0),
		},
		Status: Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, game)
	registerCardsInTestGame(game)

	// Snipe should fail because no Person door is active
	_, err := game.Apply(PlayCardCmd{PlayerID: ranger.Id, CardID: snipeCard.Id})
	if err == nil {
		t.Fatal("expected error when playing Snipe with no active Person door")
	}

	// Card must remain in hand and not be discarded
	if len(ranger.Hand) != 1 || ranger.Hand[0] != snipeCard {
		t.Errorf("expected snipeCard to remain in hand, got %v", ranger.Hand)
	}
	if ranger.Discard.Length() != 0 {
		t.Errorf("expected discard to be empty, got %d", ranger.Discard.Length())
	}
}

func TestActionCardPersistsOnPlayfieldUntilRoomCleared(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	healCard := &ActionCard{Id: 666, Name: "Heal", Action: HealAction{}}
	swordCard := &ResourceCard{Resources: []ResourceType{Sword}}
	drawCard1 := &ResourceCard{Resources: []ResourceType{Arrow}}
	drawCard2 := &ResourceCard{Resources: []ResourceType{Shield}}

	p1.Hand = []PlayerCard{healCard, swordCard}
	p1.Deck = &Deck{Cards: []PlayerCard{drawCard1, drawCard2}}

	doorMonster := &DoorCard{Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
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
	_, _ = game.PlayField.AddDungeonCard(doorMonster, game)
	registerCardsInTestGame(game)

	// 1. Play Heal Action: door is NOT defeated yet, healCard must be on playfield
	_, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: healCard.Id})
	if err != nil {
		t.Fatalf("failed to play heal: %v", err)
	}

	if len(game.PlayField.Field) != 1 || game.PlayField.Field[0] != healCard {
		t.Fatalf("expected healCard to remain on playfield, got %v", game.PlayField.Field)
	}
	if p1.Discard.Length() != 0 {
		t.Errorf("expected P1 discard to be empty, got %d", p1.Discard.Length())
	}

	// 2. Play Sword Resource: now door requirements are satisfied
	_, err = game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: swordCard.ID()})
	if err != nil {
		t.Fatalf("failed to play sword: %v", err)
	}

	if len(game.PlayField.Field) != 2 {
		t.Fatalf("expected 2 cards on playfield before resolution, got %v", game.PlayField.Field)
	}

	// 3. Tick advances and resolves the beaten room
	game.LastPlayedCardTimer = -time.Second // bypass debounce
	events, err := game.Tick(time.Millisecond * 100)
	if err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	// Verify doorMonster was defeated and field was cleared
	if game.PlayField.HasActiveDoor(doorMonster) {
		t.Error("expected doorMonster to be defeated")
	}
	if len(game.PlayField.Field) != 0 {
		t.Errorf("expected playfield to be cleared after room resolution, got %v", game.PlayField.Field)
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be active")
	}

	var fieldClearedFound bool
	for _, e := range events {
		if _, ok := e.(FieldClearedEvent); ok {
			fieldClearedFound = true
		}
	}
	if !fieldClearedFound {
		t.Errorf("expected FieldClearedEvent in events: %v", events)
	}
}

func TestActionCardTimeWarp(t *testing.T) {
	wizard, _ := NewPlayer("Gandalf", Wizard, true)
	timeWarpCard := &ActionCard{Id: 666, Name: "Time Warp", Action: TimeWarpAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Scroll}}

	wizard.Hand = []PlayerCard{timeWarpCard}
	wizard.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	doorMonster := &DoorCard{Type: DoorMonster, Name: "Dragon", Resources: []ResourceType{Sword}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{doorMonster},
	}

	game := &Game{
		Players:   []*Player{wizard},
		HandSize:  1,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(doorMonster, game)
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: wizard.Id, CardID: timeWarpCard.Id})
	if err != nil {
		t.Fatalf("failed to play Time Warp: %v", err)
	}

	if !game.IsTimeFrozen {
		t.Error("expected game time to be frozen")
	}

	var frozenEvtFound bool
	for _, e := range events {
		if _, ok := e.(TimeFrozenEvent); ok {
			frozenEvtFound = true
		}
	}
	if !frozenEvtFound {
		t.Errorf("expected TimeFrozenEvent in events: %v", events)
	}
}

func TestActionCardSnipeAndDefeatDoorKinds(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	snipeCard := &ActionCard{Id: 666, Name: "Snipe", Action: SnipeAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Arrow}}

	ranger.Hand = []PlayerCard{snipeCard}
	ranger.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	personDoor := &DoorCard{Type: DoorPerson, Name: "Guard", Resources: []ResourceType{Arrow}}
	nextDoor := &DoorCard{Type: DoorObstacle, Name: "Wall", Resources: []ResourceType{Jump}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{nextDoor},
	}

	game := &Game{
		Players:   []*Player{ranger},
		HandSize:  1,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(personDoor, game)
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: ranger.Id, CardID: snipeCard.Id})
	if err != nil {
		t.Fatalf("failed to play Snipe: %v", err)
	}

	if game.PlayField.HasActiveDoor(personDoor) {
		t.Error("expected personDoor to be defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be active")
	}

	var defeatFound bool
	for _, e := range events {
		if de, ok := e.(DoorDefeatedEvent); ok && de.CardID == personDoor.Id {
			defeatFound = true
		}
	}
	if !defeatFound {
		t.Error("expected DoorDefeatedEvent in events")
	}
}

func TestActionCardEnrage(t *testing.T) {
	p1, _ := NewPlayer("P1", Barbarian, true)
	p2, _ := NewPlayer("P2", Gladiator, true)

	enrageCard := &ActionCard{Id: 666, Name: "Enrage", Action: EnrageAction{}}
	d1 := &ResourceCard{Resources: []ResourceType{Sword}}
	d2 := &ResourceCard{Resources: []ResourceType{Shield}}
	d3 := &ResourceCard{Resources: []ResourceType{Jump}}

	p1.Hand = []PlayerCard{enrageCard}
	p1.Deck = &Deck{Cards: []PlayerCard{d1, d2, d3, d1}}
	p2.Deck = &Deck{Cards: []PlayerCard{d1, d2, d3}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: enrageCard.Id})
	if err != nil {
		t.Fatalf("failed to play Enrage: %v", err)
	}

	// 2 players: auto-targets both players, each draws 3 cards from deck (+ P1 auto-draws 1 to hand size)
	if len(p2.Hand) != 3 {
		t.Errorf("expected P2 to have 3 cards in hand, got %d", len(p2.Hand))
	}
	if len(events) < 6 {
		t.Errorf("expected draw events for both players, got %d events", len(events))
	}
}

func TestActionCardWildCard(t *testing.T) {
	ranger, _ := NewPlayer("Robin", Ranger, true)
	wildCard := &ActionCard{Id: 666, Name: "Wild Card", Action: WildCardAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Arrow}}

	ranger.Hand = []PlayerCard{wildCard}
	ranger.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	door := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}
	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{},
	}

	game := &Game{
		Players:   []*Player{ranger},
		HandSize:  1,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(door, game)
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: ranger.Id, CardID: wildCard.Id})
	if err != nil {
		t.Fatalf("failed to play Wild CardID: %v", err)
	}

	// ActionCard is removed from playfield and replaced by a ResourceCard{Resources: [WildCard]}
	if len(game.PlayField.Field) != 1 {
		t.Fatalf("expected 1 card on playfield, got %d", len(game.PlayField.Field))
	}
	rc, ok := game.PlayField.Field[0].(*ResourceCard)
	if !ok || len(rc.Resources) != 1 || rc.Resources[0] != WildCard {
		t.Errorf("expected ResourceCard with WildCard on playfield, got %v", game.PlayField.Field[0])
	}
	if !game.PlayField.IsPlayfieldBeaten() {
		t.Errorf("expected playfield to be beaten with WildCard satisfying Sword requirement")
	}

	var cardRemovedFound bool
	for _, e := range events {
		if re, ok := e.(CardRemovedEvent); ok && re.CardID == wildCard.Id {
			cardRemovedFound = true
		}
	}
	if !cardRemovedFound {
		t.Errorf("expected CardRemovedEvent for wildCard action card")
	}
}

func TestActionCardMagicBomb(t *testing.T) {
	wizard, _ := NewPlayer("Mage", Wizard, true)
	magicBombCard := &ActionCard{Id: 666, Name: "Magic Bomb", Action: MagicBombAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Scroll}}

	wizard.Hand = []PlayerCard{magicBombCard}
	wizard.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	door := &DoorCard{Type: DoorMonster, Name: "Big Monster", Resources: []ResourceType{Sword, Shield, Arrow, Scroll, Jump}}
	game := &Game{
		Players:   []*Player{wizard},
		HandSize:  1,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(door, game)
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: wizard.Id, CardID: magicBombCard.Id})
	if err != nil {
		t.Fatalf("failed to play Magic Bomb: %v", err)
	}

	// Should have replaced MagicBomb with 5-resource card
	if len(game.PlayField.Field) != 1 {
		t.Fatalf("expected 1 card on playfield, got %d", len(game.PlayField.Field))
	}
	rc, ok := game.PlayField.Field[0].(*ResourceCard)
	if !ok || len(rc.Resources) != 5 {
		t.Fatalf("expected 5-resource card on playfield, got %v", game.PlayField.Field[0])
	}
	if !game.PlayField.IsPlayfieldBeaten() {
		t.Errorf("expected playfield to be beaten with Magic Bomb matching all 5 resources")
	}
}

func TestActionCardThrowingKnives(t *testing.T) {
	ninja, _ := NewPlayer("Ninja", Ninja, true)
	knivesCard := &ActionCard{Id: 666, Name: "Throwing Knives", Action: ThrowingKnivesAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Jump}}

	ninja.Hand = []PlayerCard{knivesCard}
	ninja.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	door := &DoorCard{Type: DoorMonster, Name: "Tough Monster", Resources: []ResourceType{Sword, Shield, Arrow}}
	game := &Game{
		Players:   []*Player{ninja},
		HandSize:  1,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(door, game)
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: ninja.Id, CardID: knivesCard.Id})
	if err != nil {
		t.Fatalf("failed to play Throwing Knives: %v", err)
	}

	if len(game.PlayField.Field) != 1 {
		t.Fatalf("expected 1 card on playfield, got %d", len(game.PlayField.Field))
	}
	rc, ok := game.PlayField.Field[0].(*ResourceCard)
	if !ok || len(rc.Resources) != 3 || rc.Resources[0] != WildCard {
		t.Fatalf("expected 3 WildCards on playfield, got %v", game.PlayField.Field[0])
	}
	if !game.PlayField.IsPlayfieldBeaten() {
		t.Errorf("expected playfield to be beaten with 3 WildCards satisfying 3 resources")
	}
}

func TestActionCardMonsterDefeatActions(t *testing.T) {
	monsterActions := []struct {
		name   string
		action CardAction
	}{
		{"Critical Hit", CriticalHitAction{}},
		{"Smite", SmiteAction{}},
		{"Fireball", FireballAction{}},
		{"Tame Creature", TameCreatureAction{}},
	}

	for _, tt := range monsterActions {
		t.Run(tt.name, func(t *testing.T) {
			player, _ := NewPlayer("Hero", Ranger, true)
			card := &ActionCard{Id: 666, Name: tt.name, Action: tt.action}
			drawCard := &ResourceCard{Resources: []ResourceType{Arrow}}
			player.Hand = []PlayerCard{card}
			player.Deck = &Deck{Cards: []PlayerCard{drawCard}}

			monsterDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword, Shield}}
			nextDoor := &DoorCard{Type: DoorObstacle, Name: "Obstacle", Resources: []ResourceType{Jump}}

			game := &Game{
				Players:   []*Player{player},
				HandSize:  1,
				Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
				PlayField: NewPlayfield(),
				Status:    Playing,
			}
			_, _ = game.PlayField.AddDungeonCard(monsterDoor, game)
			registerCardsInTestGame(game)

			_, err := game.Apply(PlayCardCmd{PlayerID: player.Id, CardID: card.Id})
			if err != nil {
				t.Fatalf("failed to play %s: %v", tt.name, err)
			}
			if game.PlayField.HasActiveDoor(monsterDoor) {
				t.Errorf("%s failed to defeat monster door", tt.name)
			}
		})
	}
}

func TestActionCardObstacleDefeatActions(t *testing.T) {
	obstacleActions := []struct {
		name   string
		action CardAction
	}{
		{"Mighty Leap", MightyLeapAction{}},
		{"Sprint", SprintAction{}},
		{"True Sight", TrueSightAction{}},
	}

	for _, tt := range obstacleActions {
		t.Run(tt.name, func(t *testing.T) {
			player, _ := NewPlayer("Hero", Barbarian, true)
			card := &ActionCard{Id: 666, Name: tt.name, Action: tt.action}
			drawCard := &ResourceCard{Resources: []ResourceType{Sword}}
			player.Hand = []PlayerCard{card}
			player.Deck = &Deck{Cards: []PlayerCard{drawCard}}

			obstacleDoor := &DoorCard{Type: DoorObstacle, Name: "Wall", Resources: []ResourceType{Jump, Jump}}
			nextDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

			game := &Game{
				Players:   []*Player{player},
				HandSize:  1,
				Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
				PlayField: NewPlayfield(),
				Status:    Playing,
			}
			_, _ = game.PlayField.AddDungeonCard(obstacleDoor, game)
			registerCardsInTestGame(game)

			_, err := game.Apply(PlayCardCmd{PlayerID: player.Id, CardID: card.Id})
			if err != nil {
				t.Fatalf("failed to play %s: %v", tt.name, err)
			}
			if game.PlayField.HasActiveDoor(obstacleDoor) {
				t.Errorf("%s failed to defeat obstacle door", tt.name)
			}
		})
	}
}

func TestActionCardPersonDefeatActions(t *testing.T) {
	personActions := []struct {
		name   string
		action CardAction
	}{
		{"Backstab", BackstabAction{}},
		{"Living Vines", LivingVinesAction{}},
		{"Snipe", SnipeAction{}},
	}

	for _, tt := range personActions {
		t.Run(tt.name, func(t *testing.T) {
			player, _ := NewPlayer("Hero", Thief, true)
			card := &ActionCard{Id: 666, Name: tt.name, Action: tt.action}
			drawCard := &ResourceCard{Resources: []ResourceType{Jump}}
			player.Hand = []PlayerCard{card}
			player.Deck = &Deck{Cards: []PlayerCard{drawCard}}

			personDoor := &DoorCard{Type: DoorPerson, Name: "Guard", Resources: []ResourceType{Arrow}}
			nextDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

			game := &Game{
				Players:   []*Player{player},
				HandSize:  1,
				Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
				PlayField: NewPlayfield(),
				Status:    Playing,
			}
			_, _ = game.PlayField.AddDungeonCard(personDoor, game)
			registerCardsInTestGame(game)

			_, err := game.Apply(PlayCardCmd{PlayerID: player.Id, CardID: card.Id})
			if err != nil {
				t.Fatalf("failed to play %s: %v", tt.name, err)
			}
			if game.PlayField.HasActiveDoor(personDoor) {
				t.Errorf("%s failed to defeat person door", tt.name)
			}
		})
	}
}

func TestActionCardMiniBossDefeatActions(t *testing.T) {
	miniBossActions := []struct {
		name   string
		action CardAction
	}{
		{"Crush", CrushAction{}},
	}

	for _, tt := range miniBossActions {
		t.Run(tt.name, func(t *testing.T) {
			barbarian, _ := NewPlayer("Hero", Barbarian, true)
			card := &ActionCard{Id: 666, Name: tt.name, Action: tt.action}
			drawCard := &ResourceCard{Resources: []ResourceType{Sword}}
			barbarian.Hand = []PlayerCard{card}
			barbarian.Deck = &Deck{Cards: []PlayerCard{drawCard}}

			miniBoss := &MiniBossCard{Name: "Gargoyle", Resources: []ResourceType{Sword, Shield, Arrow}}
			nextDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

			game := &Game{
				Players:   []*Player{barbarian},
				HandSize:  1,
				Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
				PlayField: NewPlayfield(),
				Status:    Playing,
			}
			_, _ = game.PlayField.AddDungeonCard(miniBoss, game)
			registerCardsInTestGame(game)

			_, err := game.Apply(PlayCardCmd{PlayerID: barbarian.Id, CardID: card.Id})
			if err != nil {
				t.Fatalf("failed to play %s: %v", tt.name, err)
			}
			if game.PlayField.HasActiveDoor(miniBoss) {
				t.Errorf("expected mini boss to be defeated by %s", tt.name)
			}
		})
	}
}

func TestActionCardCancelEvent(t *testing.T) {
	wizard, _ := NewPlayer("Mage", Wizard, true)
	cancelCard := &ActionCard{Id: 666, Name: "Cancel", Action: CancelAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{cancelCard}
	wizard.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	eventDoor := &EventCard{Name: "Trap"}
	nextDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{wizard},
		HandSize:  1,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(eventDoor, game)
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: wizard.Id, CardID: cancelCard.Id})
	if err != nil {
		t.Fatalf("failed to play Cancel: %v", err)
	}
	if game.PlayField.HasActiveDoor(eventDoor) {
		t.Error("expected event door to be defeated by Cancel")
	}
}

func TestActionCardDivineShield(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Valkyrie, true)

	divineShieldCard := &ActionCard{Id: 666, Name: "Divine Shield", Action: DivineShieldAction{}}
	d1 := &ResourceCard{Resources: []ResourceType{Sword}}
	d2 := &ResourceCard{Resources: []ResourceType{Shield}}
	p1.Hand = []PlayerCard{divineShieldCard}
	p1.Deck = &Deck{Cards: []PlayerCard{d1, d1}}
	p2.Deck = &Deck{Cards: []PlayerCard{d2}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: divineShieldCard.Id})
	if err != nil {
		t.Fatalf("failed to play Divine Shield: %v", err)
	}

	// Time should be frozen
	if !game.IsTimeFrozen {
		t.Error("expected time to be frozen")
	}
	// P2 should have drawn 1 card from deck
	if len(p2.Hand) != 1 || p2.Hand[0] != d2 {
		t.Errorf("expected P2 to draw d2 into hand, got %v", p2.Hand)
	}

	var frozenEvtFound bool
	for _, e := range events {
		if _, ok := e.(TimeFrozenEvent); ok {
			frozenEvtFound = true
		}
	}
	if !frozenEvtFound {
		t.Error("expected TimeFrozenEvent in events")
	}
}

func TestActionCardExtraQuiver(t *testing.T) {
	p1, _ := NewPlayer("P1", Ranger, true)
	p2, _ := NewPlayer("P2", Huntress, true)

	quiverCard := &ActionCard{Id: 666, Name: "Extra Quiver", Action: ExtraQuiverAction{}}
	d1 := &ResourceCard{Resources: []ResourceType{Arrow}}
	d2 := &ResourceCard{Resources: []ResourceType{Arrow}}
	p1.Hand = []PlayerCard{quiverCard}
	p1.Deck = &Deck{Cards: []PlayerCard{d1, d1, d1}}
	p2.Deck = &Deck{Cards: []PlayerCard{d2, d2}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: quiverCard.Id})
	if err != nil {
		t.Fatalf("failed to play Extra Quiver: %v", err)
	}

	// P2 should have drawn 2 cards
	if len(p2.Hand) != 2 {
		t.Errorf("expected P2 to have 2 cards in hand, got %d", len(p2.Hand))
	}
}

func TestActionCardHealingHerbs(t *testing.T) {
	p1, _ := NewPlayer("P1", Ranger, true)
	p2, _ := NewPlayer("P2", Paladin, true)

	herbsCard := &ActionCard{Id: 666, Name: "Healing Herbs", Action: HealingHerbsAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Arrow}}
	p1.Hand = []PlayerCard{herbsCard}
	p1.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	p2.Discard.PutAtop(c1, c2)

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: herbsCard.Id})
	if err != nil {
		t.Fatalf("failed to play Healing Herbs: %v", err)
	}

	// P2 should draw 2 cards from discard into hand
	if len(p2.Hand) != 2 {
		t.Errorf("expected P2 to have drawn 2 cards from discard into hand, got %d", len(p2.Hand))
	}
	if p2.Discard.Length() != 0 {
		t.Errorf("expected P2 discard to be empty, got %d", p2.Discard.Length())
	}
}

func TestActionCardAncientHealing(t *testing.T) {
	p1, _ := NewPlayer("P1", Valkyrie, true)
	p2, _ := NewPlayer("P2", Paladin, true)

	ancientHealingCard := &ActionCard{Id: 666, Name: "Ancient Healing", Action: AncientHealingAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Shield}}
	p1.Hand = []PlayerCard{ancientHealingCard}
	p1.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Jump}}
	p1.Discard.PutAtop(c1, c2)
	p2.Discard.PutAtop(c3)

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: ancientHealingCard.Id})
	if err != nil {
		t.Fatalf("failed to play Ancient Healing: %v", err)
	}

	// Both players should draw from discard (up to 2 cards each)
	if p1.Discard.Length() != 0 {
		t.Errorf("expected P1 discard to be empty, got %d", p1.Discard.Length())
	}
	if p2.Discard.Length() != 0 {
		t.Errorf("expected P2 discard to be empty, got %d", p2.Discard.Length())
	}
	if len(p2.Hand) != 1 || p2.Hand[0] != c3 {
		t.Errorf("expected P2 hand to contain [c3], got %v", p2.Hand)
	}
}

func TestActionCardBattleRage(t *testing.T) {
	barbarian, _ := NewPlayer("Conan", Barbarian, true)
	battleRageCard := &ActionCard{Id: 666, Name: "Battle Rage", Action: BattleRageAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Sword}}
	barbarian.Hand = []PlayerCard{battleRageCard}
	barbarian.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Curse of Doom",
		Effect: AbilitiesCannotBePlayed,
	}
	nextDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{barbarian},
		HandSize:  1,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: barbarian.Id, CardID: battleRageCard.Id})
	if err != nil {
		t.Fatalf("failed to play Battle Rage: %v", err)
	}

	if slices.Contains(game.PlayField.ActiveCurses, curse) {
		t.Error("expected curse to be removed from active curses")
	}
}

func TestActionCardCleanse(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	cleanseCard := &ActionCard{Id: 666, Name: "Cleanse", Action: CleanseAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Sword}}
	paladin.Hand = []PlayerCard{cleanseCard}
	paladin.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	curse := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Curse of Weakness",
		Effect: AbilitiesCannotBePlayed,
	}
	nextDoor := &DoorCard{Type: DoorMonster, Name: "Monster", Resources: []ResourceType{Sword}}

	game := &Game{
		Players:   []*Player{paladin},
		HandSize:  1,
		Dungeon:   &Dungeon{Boss: &BossMat{Name: "Boss"}, Doors: []DungeonCard{nextDoor}},
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(curse, game)
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: paladin.Id, CardID: cleanseCard.Id})
	if err != nil {
		t.Fatalf("failed to play Cleanse: %v", err)
	}

	if slices.Contains(game.PlayField.ActiveCurses, curse) {
		t.Error("expected curse to be removed from active curses")
	}
}

func TestActionCardPortal(t *testing.T) {
	wizard, _ := NewPlayer("Gandalf", Wizard, true)
	portalCard := &ActionCard{Id: 666, Name: "Portal", Action: PortalAction{}}
	drawCard := &ResourceCard{Resources: []ResourceType{Scroll}}
	wizard.Hand = []PlayerCard{portalCard}
	wizard.Deck = &Deck{Cards: []PlayerCard{drawCard}}

	door1 := &DoorCard{Type: DoorMonster, Name: "Monster 1", Resources: []ResourceType{Sword}}
	door2 := &DoorCard{Type: DoorObstacle, Name: "Obstacle 2", Resources: []ResourceType{Jump}}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{door2},
	}

	game := &Game{
		Players:   []*Player{wizard},
		HandSize:  1,
		Dungeon:   dungeon,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	_, _ = game.PlayField.AddDungeonCard(door1, game)
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: wizard.Id, CardID: portalCard.Id})
	if err != nil {
		t.Fatalf("failed to play Portal: %v", err)
	}

	if game.PlayField.HasActiveDoor(door1) {
		t.Error("expected door1 to be removed from playfield")
	}
	if !game.PlayField.HasActiveDoor(door2) {
		t.Error("expected door2 to be opened")
	}
	// door1 should be at the bottom of dungeon doors
	if len(dungeon.Doors) == 0 || dungeon.Doors[0] != door1 {
		t.Errorf("expected door1 to be at the bottom of dungeon deck, got %v", dungeon.Doors)
	}
}

func TestActionCardDonate(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	donateCard := &ActionCard{Id: 666, Name: "Donate", Action: DonateAction{}}
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	p1.Hand = []PlayerCard{donateCard, c1, c2}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Jump}}}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: donateCard.Id})
	if err != nil {
		t.Fatalf("failed to play Donate: %v", err)
	}

	// P2 should have received c1 and c2
	if !p2.HasCardInHand(c1) || !p2.HasCardInHand(c2) {
		t.Errorf("expected P2 to receive c1 and c2, got %v", p2.Hand)
	}

	var donateEvtFound bool
	for _, e := range events {
		if de, ok := e.(HandDonatedEvent); ok && de.FromPlayerID == p1.Id && de.ToPlayerID == p2.Id {
			donateEvtFound = true
		}
	}
	if !donateEvtFound {
		t.Error("expected HandDonatedEvent in events")
	}
}

func TestActionCardDonateAutoTarget(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Ranger, true)

	donateCard := &ActionCard{Id: 666, Name: "Donate", Action: DonateAction{}}
	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	p1.Hand = []PlayerCard{donateCard, c1}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Jump}}}}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: donateCard.Id})
	if err != nil {
		t.Fatalf("failed to auto-target Donate in 2-player game: %v", err)
	}

	if !p2.HasCardInHand(c1) {
		t.Errorf("expected P2 to receive donated c1, got %v", p2.Hand)
	}
}

func TestActionCardSteal(t *testing.T) {
	p1, _ := NewPlayer("P1", Thief, true)
	p2, _ := NewPlayer("P2", Paladin, true)

	stealCard := &ActionCard{Id: 666, Name: "Steal", Action: StealAction{}}
	p1.Hand = []PlayerCard{stealCard}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Jump}}}}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	p2.Hand = []PlayerCard{c1, c2}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	events, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: stealCard.Id})
	if err != nil {
		t.Fatalf("failed to play Steal: %v", err)
	}

	// P1 should now have c1 and c2 in hand (+ refilled card)
	if !p1.HasCardInHand(c1) || !p1.HasCardInHand(c2) {
		t.Errorf("expected P1 to have stolen cards c1 and c2, got %v", p1.Hand)
	}
	if len(p2.Hand) != 0 {
		t.Errorf("expected P2 hand to be empty after theft, got %v", p2.Hand)
	}

	var stealEvtFound bool
	for _, e := range events {
		if se, ok := e.(HandStolenEvent); ok && se.FromPlayerID == p2.Id && se.ToPlayerID == p1.Id {
			stealEvtFound = true
		}
	}
	if !stealEvtFound {
		t.Error("expected HandStolenEvent in events")
	}
}

func TestActionCardStealAutoTarget(t *testing.T) {
	p1, _ := NewPlayer("P1", Thief, true)
	p2, _ := NewPlayer("P2", Paladin, true)

	stealCard := &ActionCard{Id: 666, Name: "Steal", Action: StealAction{}}
	p1.Hand = []PlayerCard{stealCard}
	p1.Deck = &Deck{Cards: []PlayerCard{&ResourceCard{Resources: []ResourceType{Jump}}}}

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	p2.Hand = []PlayerCard{c1}

	game := &Game{
		Players:   []*Player{p1, p2},
		HandSize:  1,
		PlayField: NewPlayfield(),
		Status:    Playing,
	}
	registerCardsInTestGame(game)

	_, err := game.Apply(PlayCardCmd{PlayerID: p1.Id, CardID: stealCard.Id})
	if err != nil {
		t.Fatalf("failed to auto-target Steal in 2-player game: %v", err)
	}

	if !p1.HasCardInHand(c1) {
		t.Errorf("expected P1 to have stolen c1, got %v", p1.Hand)
	}
	if len(p2.Hand) != 0 {
		t.Errorf("expected P2 hand to be empty, got %v", p2.Hand)
	}
}
