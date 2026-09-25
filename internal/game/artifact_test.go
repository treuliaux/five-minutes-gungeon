package game

import (
	"testing"
)

func TestSetupArtifacts(t *testing.T) {
	tests := []struct {
		name              string
		playerDeckColors  []DeckColor
		expectedArtifacts []string
	}{
		{
			name:             "2 players (Red, Blue) gets Green, Yellow, Purple, Black artifacts",
			playerDeckColors: []DeckColor{Red, Blue},
			expectedArtifacts: []string{
				"Rainbow Herbs", // Green
				"M-jh'öilnør",   // Yellow
				"Sundial Watch", // Purple
				"Curse Zapper",  // Black
			},
		},
		{
			name:             "4 players (Red, Blue, Green, Yellow) gets Purple, Black artifacts",
			playerDeckColors: []DeckColor{Red, Blue, Green, Yellow},
			expectedArtifacts: []string{
				"Sundial Watch", // Purple
				"Curse Zapper",  // Black
			},
		},
		{
			name:              "6 players (all colors) gets 0 artifacts",
			playerDeckColors:  []DeckColor{Red, Blue, Green, Yellow, Purple, Black},
			expectedArtifacts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := NewPlayfield()
			pf.SetupArtifacts(tt.playerDeckColors)

			if len(pf.Artifacts) != len(tt.expectedArtifacts) {
				t.Fatalf("expected %d artifacts, got %d", len(tt.expectedArtifacts), len(pf.Artifacts))
			}

			for i, expectedName := range tt.expectedArtifacts {
				if pf.Artifacts[i].Name != expectedName {
					t.Errorf("expected artifact[%d] to be %q, got %q", i, expectedName, pf.Artifacts[i].Name)
				}
				if pf.Artifacts[i].Used {
					t.Errorf("expected artifact %q to start with Used=false", expectedName)
				}
				if pf.Artifacts[i].Action == nil {
					t.Errorf("expected artifact %q to have non-nil Action", expectedName)
				}
			}
		})
	}
}

func TestArtifactCanBePlayed(t *testing.T) {
	pf := NewPlayfield()
	axe := &ArtifactCard{Color: Red, Name: "Battle Axe", Used: false}
	scroll := &ArtifactCard{Color: Blue, Name: "The Infinity Scroll", Used: true}
	herbsNotInPlay := &ArtifactCard{Color: Green, Name: "Rainbow Herbs", Used: false}

	pf.Artifacts = []*ArtifactCard{axe, scroll}

	if !pf.ArtifactCanBePlayed(axe) {
		t.Error("expected unused artifact on playfield to be playable")
	}
	if pf.ArtifactCanBePlayed(scroll) {
		t.Error("expected used artifact on playfield to NOT be playable")
	}
	if pf.ArtifactCanBePlayed(herbsNotInPlay) {
		t.Error("expected artifact not in playfield to NOT be playable")
	}
}

func TestBattleAxeArtifact_DefeatMonster(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	monster := &DoorCard{Type: DoorMonster, Name: "Dragon"}
	obstacle := &DoorCard{Type: DoorObstacle, Name: "Wall"}
	nextDoor := &DoorCard{Type: DoorPerson, Name: "Guard"}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{monster, nextDoor},
	}
	game := &Game{
		Players:      []*Player{p1},
		Dungeon:      dungeon,
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	_, _ = game.PlayField.AddDungeonCard(monster, game)

	axe := &ArtifactCard{Color: Red, Name: "Battle Axe", Action: &BattleAxeArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{axe}

	// 1. Defeat Monster with explicit target
	events, err := game.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    axe,
		ActionIndex: FirstArtifactAction,
		Target:      monster,
	})
	if err != nil {
		t.Fatalf("failed to use Battle Axe on monster: %v", err)
	}
	if !axe.Used {
		t.Error("expected Battle Axe to be marked Used")
	}
	if game.PlayField.HasActiveDoor(monster) {
		t.Error("expected monster to be defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened")
	}

	foundArtifactUsed := false
	for _, e := range events {
		if ae, ok := e.(ArtifactUsedEvent); ok && ae.ByPlayer == p1 {
			foundArtifactUsed = true
		}
	}
	if !foundArtifactUsed {
		t.Error("expected ArtifactUsedEvent to be emitted")
	}

	// 2. Reject targeting non-monster
	axe2 := &ArtifactCard{Color: Red, Name: "Battle Axe 2", Action: &BattleAxeArtifact{}}
	game.PlayField.Artifacts = append(game.PlayField.Artifacts, axe2)
	game.PlayField.OpenedDoors = []DungeonCard{obstacle}

	_, err = game.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    axe2,
		ActionIndex: FirstArtifactAction,
		Target:      obstacle,
	})
	if err == nil {
		t.Error("expected error when targeting DoorObstacle with Battle Axe, got nil")
	}
}

func TestBattleAxeArtifact_AutoTargetMonster(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	monster := &DoorCard{Type: DoorMonster, Name: "Dragon"}
	nextDoor := &DoorCard{Type: DoorPerson, Name: "Guard"}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{monster, nextDoor},
	}
	game := &Game{
		Players:      []*Player{p1},
		Dungeon:      dungeon,
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	_, _ = game.PlayField.AddDungeonCard(monster, game)

	axe := &ArtifactCard{Color: Red, Name: "Battle Axe", Action: &BattleAxeArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{axe}

	// Target nil -> auto-targets active Monster door
	_, err := game.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    axe,
		ActionIndex: FirstArtifactAction,
		Target:      nil,
	})
	if err != nil {
		t.Fatalf("failed auto-targeting monster with Battle Axe: %v", err)
	}
	if game.PlayField.HasActiveDoor(monster) {
		t.Error("expected monster to be defeated")
	}
}

func TestBattleAxeArtifact_DrawTwoCards(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	p2, _ := NewPlayer("Robin", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c4 := &ResourceCard{Resources: []ResourceType{Jump}}

	p1.Hand = []PlayerCard{}
	p1.Deck = &Deck{Cards: []PlayerCard{c1, c2}}

	p2.Hand = []PlayerCard{}
	p2.Deck = &Deck{Cards: []PlayerCard{c3, c4}}

	game := &Game{
		Players:      []*Player{p1, p2},
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}

	axe := &ArtifactCard{Color: Red, Name: "Battle Axe", Action: &BattleAxeArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{axe}

	events, err := game.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    axe,
		ActionIndex: SecondArtifactAction,
	})
	if err != nil {
		t.Fatalf("failed to use Battle Axe second action: %v", err)
	}

	if len(p1.Hand) != 2 || len(p2.Hand) != 2 {
		t.Errorf("expected both players to draw 2 cards, got P1: %d, P2: %d", len(p1.Hand), len(p2.Hand))
	}

	var drawnEvents int
	for _, e := range events {
		if _, ok := e.(CardDrawnFromDeckEvent); ok {
			drawnEvents++
		}
	}
	if drawnEvents != 4 {
		t.Errorf("expected 4 CardDrawnFromDeckEvent, got %d", drawnEvents)
	}
}

func TestTheInfinityScrollArtifact_DefeatMiniBoss(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	miniBoss := &MiniBossCard{Name: "Tinkles, Destroyer of Soles", Resources: []ResourceType{Jump, Jump}}
	nextDoor := &DoorCard{Type: DoorPerson, Name: "Guard"}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{miniBoss, nextDoor},
	}
	game := &Game{
		Players:      []*Player{p1},
		Dungeon:      dungeon,
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	_, _ = game.PlayField.AddDungeonCard(miniBoss, game)

	scroll := &ArtifactCard{Color: Blue, Name: "The Infinity Scroll", Action: &TheInfinityScrollArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{scroll}

	// 1. Defeat MiniBoss
	_, err := game.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    scroll,
		ActionIndex: FirstArtifactAction,
		Target:      nil, // auto-targets MiniBoss
	})
	if err != nil {
		t.Fatalf("failed to defeat mini-boss with The Infinity Scroll: %v", err)
	}
	if game.PlayField.HasActiveDoor(miniBoss) {
		t.Error("expected mini-boss to be defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened")
	}
}

func TestTheInfinityScrollArtifact_CounterEvent(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	eventCard := &EventCard{Name: "Sudden Illness", Action: SuddenIllnessEvent{}}
	nextDoor := &DoorCard{Type: DoorPerson, Name: "Guard"}

	dungeon := &Dungeon{
		Boss:  &BossMat{Name: "Boss"},
		Doors: []DungeonCard{eventCard, nextDoor},
	}
	game := &Game{
		Players:      []*Player{p1},
		Dungeon:      dungeon,
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	_, _ = game.PlayField.AddDungeonCard(eventCard, game)

	scroll := &ArtifactCard{Color: Blue, Name: "The Infinity Scroll", Action: &TheInfinityScrollArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{scroll}

	// 2. Counter active Event
	events, err := game.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    scroll,
		ActionIndex: SecondArtifactAction,
		Target:      nil, // auto-targets active EventCard
	})
	if err != nil {
		t.Fatalf("failed to counter event with The Infinity Scroll: %v", err)
	}
	if game.PlayField.HasActiveDoor(eventCard) {
		t.Error("expected event card to be countered and defeated")
	}
	if !game.PlayField.HasActiveDoor(nextDoor) {
		t.Error("expected nextDoor to be opened")
	}

	foundCounteredEvent := false
	for _, e := range events {
		if ce, ok := e.(EventCounteredEvent); ok {
			if ce.EventCard != eventCard {
				t.Errorf("expected EventCounteredEvent to have eventCard %v, got %v", eventCard, ce.EventCard)
			}
			foundCounteredEvent = true
		}
	}
	if !foundCounteredEvent {
		t.Error("expected EventCounteredEvent to be emitted")
	}
}

func TestRainbowHerbsArtifact(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	p2, _ := NewPlayer("Robin", Ranger, true)

	c1 := &ResourceCard{Resources: []ResourceType{Sword}}
	c2 := &ResourceCard{Resources: []ResourceType{Shield}}
	c3 := &ResourceCard{Resources: []ResourceType{Arrow}}
	c4 := &ResourceCard{Resources: []ResourceType{Jump}}

	p1.Deck = &Deck{Cards: []PlayerCard{}}
	p1.Discard = &Discard{Cards: []PlayerCard{c1, c2}}

	p2.Deck = &Deck{Cards: []PlayerCard{}}
	p2.Discard = &Discard{Cards: []PlayerCard{c3, c4}}

	game := &Game{
		Players:      []*Player{p1, p2},
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}

	herbs := &ArtifactCard{Color: Green, Name: "Rainbow Herbs", Action: &RainbowHerbsArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{herbs}

	events, err := game.Apply(UseArtifactCmd{
		Player:   p1,
		Artifact: herbs,
	})
	if err != nil {
		t.Fatalf("failed to use Rainbow Herbs: %v", err)
	}

	// All discard cards should be moved to Deck
	if p1.Discard.Length() != 0 || p1.Deck.Length() != 2 {
		t.Errorf("expected P1 discard=0, deck=2; got discard=%d, deck=%d", p1.Discard.Length(), p1.Deck.Length())
	}
	if p2.Discard.Length() != 0 || p2.Deck.Length() != 2 {
		t.Errorf("expected P2 discard=0, deck=2; got discard=%d, deck=%d", p2.Discard.Length(), p2.Deck.Length())
	}

	var healEvents int
	for _, e := range events {
		if _, ok := e.(PlayerHealedEvent); ok {
			healEvents++
		}
	}
	if healEvents != 2 {
		t.Errorf("expected 2 PlayerHealedEvent (1 per player), got %d", healEvents)
	}
}

func TestMJhoilnorArtifact(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	d1 := &DoorCard{Type: DoorMonster, Name: "D1"}
	d2 := &DoorCard{Type: DoorObstacle, Name: "D2"}
	d3 := &DoorCard{Type: DoorPerson, Name: "D3"}
	boss := &BossMat{Name: "Final Boss"}

	dungeon := &Dungeon{
		Boss:  boss,
		Doors: []DungeonCard{d1, d2, d3},
	}
	game := &Game{
		Players:      []*Player{p1},
		Dungeon:      dungeon,
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}

	mj := &ArtifactCard{Color: Yellow, Name: "M-jh'öilnør", Action: &MJhoilnorArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{mj}

	events, err := game.Apply(UseArtifactCmd{
		Player:   p1,
		Artifact: mj,
	})
	if err != nil {
		t.Fatalf("failed to use M-jh'öilnør: %v", err)
	}

	// Should have discarded 2 doors from dungeon
	if len(dungeon.Doors) != 1 || dungeon.Doors[0] != d1 {
		t.Errorf("expected 1 door remaining (d1), got %d doors", len(dungeon.Doors))
	}

	var discardEvents int
	for _, e := range events {
		if _, ok := e.(DungeonCardDiscardedEvent); ok {
			discardEvents++
		}
	}
	if discardEvents != 2 {
		t.Errorf("expected 2 DungeonCardDiscardedEvent, got %d", discardEvents)
	}

	// When 0 cards remain in dungeon, DiscardTopCardFromDungeon must NOT discard BossMat
	dungeon.Doors = []DungeonCard{}
	mj2 := &ArtifactCard{Color: Yellow, Name: "M-jh'öilnør 2", Action: &MJhoilnorArtifact{}}
	game.PlayField.Artifacts = append(game.PlayField.Artifacts, mj2)

	events2, err2 := game.Apply(UseArtifactCmd{
		Player:   p1,
		Artifact: mj2,
	})
	if err2 != nil {
		t.Fatalf("expected no error when dungeon is empty, got: %v", err2)
	}
	for _, e := range events2 {
		if de, ok := e.(DungeonCardDiscardedEvent); ok {
			if _, isBoss := de.Card.(*BossMat); isBoss {
				t.Error("M-jh'öilnør must NEVER discard the BossMat")
			}
		}
	}
}

func TestSundialWatchArtifact(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	game := &Game{
		Players:      []*Player{p1},
		PlayField:    NewPlayfield(),
		IsTimeFrozen: false,
		Status:       Playing,
		UseExtension: true,
	}

	sundial := &ArtifactCard{Color: Purple, Name: "Sundial Watch", Action: &SundialWatchArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{sundial}

	events, err := game.Apply(UseArtifactCmd{
		Player:   p1,
		Artifact: sundial,
	})
	if err != nil {
		t.Fatalf("failed to use Sundial Watch: %v", err)
	}
	if !game.IsTimeFrozen {
		t.Error("expected IsTimeFrozen to be true after Sundial Watch")
	}

	foundTimeFrozenEvent := false
	for _, e := range events {
		if te, ok := e.(TimeFrozenEvent); ok && te.ByPlayer == p1 {
			foundTimeFrozenEvent = true
		}
	}
	if !foundTimeFrozenEvent {
		t.Error("expected TimeFrozenEvent to be emitted")
	}

	// Attempting to freeze time again when already frozen returns error
	sundial2 := &ArtifactCard{Color: Purple, Name: "Sundial Watch 2", Action: &SundialWatchArtifact{}}
	game.PlayField.Artifacts = append(game.PlayField.Artifacts, sundial2)

	_, err = game.Apply(UseArtifactCmd{
		Player:   p1,
		Artifact: sundial2,
	})
	if err == nil {
		t.Error("expected error when time is already frozen, got nil")
	}
}

func TestCurseZapperArtifact(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	curse1 := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Curse of Weakness",
		Effect: TimeCannotBeStopped,
	}
	curse2 := &CurseCard{
		Type:   ChallengeCurse,
		Name:   "Curse of Sloth",
		Effect: TimeCannotBeStopped,
	}

	game := &Game{
		Players:      []*Player{p1},
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	game.PlayField.ActiveCurses = []*CurseCard{curse1, curse2}

	zapper := &ArtifactCard{Color: Black, Name: "Curse Zapper", Action: &CurseZapperArtifact{}}
	game.PlayField.Artifacts = []*ArtifactCard{zapper}

	events, err := game.Apply(UseArtifactCmd{
		Player:   p1,
		Artifact: zapper,
	})
	if err != nil {
		t.Fatalf("failed to use Curse Zapper: %v", err)
	}

	if len(game.PlayField.ActiveCurses) != 0 {
		t.Errorf("expected 0 active curses, got %d", len(game.PlayField.ActiveCurses))
	}

	var curseRemovedEvents int
	for _, e := range events {
		if _, ok := e.(CurseRemovedEvent); ok {
			curseRemovedEvents++
		}
	}
	if curseRemovedEvents != 2 {
		t.Errorf("expected 2 CurseRemovedEvent, got %d", curseRemovedEvents)
	}
}

func TestGameUseArtifact_GuardsAndValidation(t *testing.T) {
	p1, _ := NewPlayer("Arthur", Paladin, true)
	axe := &ArtifactCard{Color: Red, Name: "Battle Axe", Action: &BattleAxeArtifact{}, Used: false}

	// 1. Error when Extension is disabled
	gameNoExt := &Game{
		Players:      []*Player{p1},
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: false,
	}
	gameNoExt.PlayField.Artifacts = []*ArtifactCard{axe}
	_, err := gameNoExt.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    axe,
		ActionIndex: FirstArtifactAction,
	})
	if err == nil {
		t.Error("expected error when UseExtension=false, got nil")
	}

	// 2. Error when PendingInteraction is active
	gamePending := &Game{
		Players:            []*Player{p1},
		PlayField:          NewPlayfield(),
		Status:             Playing,
		UseExtension:       true,
		PendingInteraction: &PlayerDiscardCardsInteraction{},
	}
	gamePending.PlayField.Artifacts = []*ArtifactCard{axe}
	_, err = gamePending.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    axe,
		ActionIndex: FirstArtifactAction,
	})
	if err == nil {
		t.Error("expected error when PendingInteraction != nil, got nil")
	}

	// 3. Error when artifact is already used
	usedAxe := &ArtifactCard{Color: Red, Name: "Battle Axe", Action: &BattleAxeArtifact{}, Used: true}
	gameUsed := &Game{
		Players:      []*Player{p1},
		PlayField:    NewPlayfield(),
		Status:       Playing,
		UseExtension: true,
	}
	gameUsed.PlayField.Artifacts = []*ArtifactCard{usedAxe}
	_, err = gameUsed.Apply(UseArtifactCmd{
		Player:      p1,
		Artifact:    usedAxe,
		ActionIndex: FirstArtifactAction,
	})
	if err == nil {
		t.Error("expected error when artifact is already used, got nil")
	}
}
