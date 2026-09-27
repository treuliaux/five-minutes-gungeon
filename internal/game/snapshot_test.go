package game

import (
	"context"
	"testing"
	"time"
)

func TestNewGameSnapshot_FullGameState(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	barbarian, _ := NewPlayer("Conan", Barbarian, true)

	cHand1 := &ResourceCard{Id: 101, Resources: []ResourceType{Sword, Shield}}
	cHand2 := &ActionCard{Id: 102, Name: "Holy Hand Grenade", Description: "Defeats any door"}
	paladin.Hand = []PlayerCard{cHand1, cHand2}

	cDisc1 := &ResourceCard{Id: 103, Resources: []ResourceType{Arrow}}
	paladin.Discard.PutAtop(cDisc1)

	cDeck1 := &ResourceCard{Id: 104, Resources: []ResourceType{Jump}}
	paladin.Deck = &Deck{Cards: []PlayerCard{cDeck1}}

	doorMonster := &DoorCard{Id: 201, Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	miniBoss := &MiniBossCard{Id: 202, Type: ChallengeMiniBoss, Name: "Dragon", Resources: []ResourceType{Shield, Arrow}}
	curse := &CurseCard{Id: 203, Type: ChallengeCurse, Name: "Curse of Silence", Description: "Cannot talk"}
	boss := &BossMat{Id: 204, Name: "Gungeon Boss", Resources: []ResourceType{Scroll, WildCard}}
	nextDoor := &DoorCard{Id: 205, Type: DoorObstacle, Name: "Stone Wall", Resources: []ResourceType{Jump}}

	fieldResource := &ResourceCard{Id: 301, Resources: []ResourceType{Sword}}
	fieldAction := &ActionCard{Id: 302, Name: "Sprint", Description: "Bypasses obstacle"}

	artAxe := &ArtifactCard{
		Id:          1,
		Color:       Red,
		Name:        "Battle Axe",
		Description: "Defeat a monster or draw cards",
		MultiAction: true,
		Used:        false,
	}
	artScroll := &ArtifactCard{
		Id:          2,
		Color:       Blue,
		Name:        "Infinity Scroll",
		Description: "Defeat mini-boss or counter event",
		MultiAction: true,
		Used:        true,
	}

	round := newRoundState(Playing)
	round.InGameTimer = 45 * time.Second
	round.RealElapsedTime = 50 * time.Second
	round.IsTimeFrozen = true
	round.Dungeon = &Dungeon{Boss: boss, Doors: []DungeonCard{nextDoor}}
	round.LastPlayedCardTimer = 40 * time.Second

	game := &Game{
		Players:    []*Player{paladin, barbarian},
		LevelState: round,
	}
	game.LevelState.Playfield.OpenedDoors = []DungeonCard{doorMonster, miniBoss}
	game.LevelState.Playfield.ActiveCurses = []*CurseCard{curse}
	game.LevelState.Playfield.Field = []PlayerCard{fieldResource, fieldAction}
	game.LevelState.Playfield.Artifacts = []*ArtifactCard{artAxe, artScroll}

	snapshot := NewGameSnapshot(game)

	// Game level checks
	if snapshot.Status != Playing {
		t.Errorf("expected Status Playing, got %v", snapshot.Status)
	}
	if snapshot.InGameTimer != 45*time.Second {
		t.Errorf("expected InGameTimer 45s, got %v", snapshot.InGameTimer)
	}
	if snapshot.TotalDuration != 50*time.Second {
		t.Errorf("expected TotalDuration 50s, got %v", snapshot.TotalDuration)
	}
	if !snapshot.IsTimeFrozen {
		t.Error("expected IsTimeFrozen true")
	}
	if snapshot.IsFightingBoss {
		t.Error("expected IsFightingBoss false")
	}
	if snapshot.RemainingDoorsCount != 1 {
		t.Errorf("expected RemainingDoorsCount 1, got %d", snapshot.RemainingDoorsCount)
	}
	if snapshot.Boss == nil || snapshot.Boss.Id != boss.ID() || snapshot.Boss.Kind != CardBoss || snapshot.Boss.Name != "Gungeon Boss" {
		t.Errorf("unexpected Boss DTO: %+v", snapshot.Boss)
	}

	// Players checks
	if len(snapshot.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(snapshot.Players))
	}
	p1DTO := snapshot.Players[0]
	if p1DTO.Id != paladin.Id || p1DTO.Name != "Arthur" || p1DTO.HeroClass != Paladin || p1DTO.HeroName != "Paladin" {
		t.Errorf("unexpected Player1 DTO: %+v", p1DTO)
	}
	if len(p1DTO.Hand) != 2 || p1DTO.Hand[0].Id != 101 || p1DTO.Hand[0].Kind != PlayerCardResource {
		t.Errorf("unexpected Player1 Hand: %+v", p1DTO.Hand)
	}
	if p1DTO.Hand[1].Id != 102 || p1DTO.Hand[1].Kind != PlayerCardAction || p1DTO.Hand[1].Name != "Holy Hand Grenade" {
		t.Errorf("unexpected Player1 Action card in Hand: %+v", p1DTO.Hand[1])
	}
	if len(p1DTO.Discard) != 1 || p1DTO.Discard[0].Id != 103 {
		t.Errorf("unexpected Player1 Discard: %+v", p1DTO.Discard)
	}
	if p1DTO.DeckCount != 1 {
		t.Errorf("expected DeckCount 1, got %d", p1DTO.DeckCount)
	}

	// Opened Doors checks
	if len(snapshot.OpenedDoors) != 2 {
		t.Fatalf("expected 2 opened doors, got %d", len(snapshot.OpenedDoors))
	}
	if snapshot.OpenedDoors[0].Id != doorMonster.ID() || snapshot.OpenedDoors[0].Kind != CardMonster || snapshot.OpenedDoors[0].Name != "Goblin" {
		t.Errorf("unexpected door 0: %+v", snapshot.OpenedDoors[0])
	}
	if snapshot.OpenedDoors[1].Id != miniBoss.ID() || snapshot.OpenedDoors[1].Kind != CardMiniBoss || snapshot.OpenedDoors[1].Name != "Dragon" {
		t.Errorf("unexpected door 1: %+v", snapshot.OpenedDoors[1])
	}

	// Active Curses checks
	if len(snapshot.ActiveCurses) != 1 {
		t.Fatalf("expected 1 active curse, got %d", len(snapshot.ActiveCurses))
	}
	if snapshot.ActiveCurses[0].Id != curse.ID() || snapshot.ActiveCurses[0].Kind != CardCurse || snapshot.ActiveCurses[0].Name != "Curse of Silence" {
		t.Errorf("unexpected active curse: %+v", snapshot.ActiveCurses[0])
	}

	// Played Field checks
	if len(snapshot.PlayedField) != 2 {
		t.Fatalf("expected 2 played field cards, got %d", len(snapshot.PlayedField))
	}
	if snapshot.PlayedField[0].Id != 301 || snapshot.PlayedField[0].Kind != PlayerCardResource {
		t.Errorf("unexpected played field resource: %+v", snapshot.PlayedField[0])
	}
	if snapshot.PlayedField[1].Id != 302 || snapshot.PlayedField[1].Kind != PlayerCardAction || snapshot.PlayedField[1].Description != "Bypasses obstacle" {
		t.Errorf("unexpected played field action: %+v", snapshot.PlayedField[1])
	}

	// Artifacts checks
	if len(snapshot.Artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(snapshot.Artifacts))
	}
	if snapshot.Artifacts[0].Id != 1 || snapshot.Artifacts[0].Color != Red || snapshot.Artifacts[0].Used != false || !snapshot.Artifacts[0].SelectAction {
		t.Errorf("unexpected artifact 0: %+v", snapshot.Artifacts[0])
	}
	if snapshot.Artifacts[1].Id != 2 || snapshot.Artifacts[1].Color != Blue || snapshot.Artifacts[1].Used != true || !snapshot.Artifacts[1].SelectAction {
		t.Errorf("unexpected artifact 1: %+v", snapshot.Artifacts[1])
	}

	// Pending Interaction
	if snapshot.PendingInteraction != nil {
		t.Errorf("expected nil PendingInteraction, got %+v", snapshot.PendingInteraction)
	}
}

func TestNewGameSnapshot_PendingInteractions(t *testing.T) {
	p1, _ := NewPlayer("P1", Paladin, true)
	p2, _ := NewPlayer("P2", Barbarian, true)
	sourceCard := &EventCard{Id: 999, Name: "Interaction Card"}

	t.Run("TeamChoiceArtifactInteraction", func(t *testing.T) {
		pi := &TeamChoiceArtifactInteraction{
			Kind:           InteractionTeamChoiceArtifact,
			Card:           sourceCard,
			PendingPlayers: map[*Player]bool{p1: true},
		}
		dto := PendingInteractionToDTO(pi)
		if dto == nil {
			t.Fatal("expected non-nil DTO")
		}
		if dto.Kind != InteractionTeamChoiceArtifact || dto.SourceCardID != 999 {
			t.Errorf("unexpected Kind or SourceCardID: %+v", dto)
		}
		if len(dto.PendingPlayers) != 1 || dto.PendingPlayers[0] != p1.Id {
			t.Errorf("unexpected PendingPlayers: %v", dto.PendingPlayers)
		}
	})

	t.Run("TeamChoiceResourceInteraction", func(t *testing.T) {
		pi := &TeamChoiceResourceInteraction{
			Kind:           InteractionTeamChoiceResource,
			Card:           sourceCard,
			PendingPlayers: map[*Player]bool{p1: true, p2: true},
		}
		dto := PendingInteractionToDTO(pi)
		if dto == nil {
			t.Fatal("expected non-nil DTO")
		}
		if dto.Kind != InteractionTeamChoiceResource || dto.SourceCardID != 999 {
			t.Errorf("unexpected Kind or SourceCardID: %+v", dto)
		}
		if len(dto.PendingPlayers) != 2 {
			t.Errorf("expected 2 PendingPlayers, got %d", len(dto.PendingPlayers))
		}
	})

	t.Run("TeamChoicePlayerInteraction", func(t *testing.T) {
		pi := &TeamChoicePlayerInteraction{
			Kind:           InteractionTeamChoicePlayer,
			Card:           sourceCard,
			PendingPlayers: map[*Player]bool{p2: true},
		}
		dto := PendingInteractionToDTO(pi)
		if dto == nil {
			t.Fatal("expected non-nil DTO")
		}
		if dto.Kind != InteractionTeamChoicePlayer || dto.SourceCardID != 999 {
			t.Errorf("unexpected Kind or SourceCardID: %+v", dto)
		}
		if len(dto.PendingPlayers) != 1 || dto.PendingPlayers[0] != p2.Id {
			t.Errorf("unexpected PendingPlayers: %v", dto.PendingPlayers)
		}
	})

	t.Run("PlayerDiscardCardsInteraction", func(t *testing.T) {
		pi := &PlayerDiscardCardsInteraction{
			Kind:           InteractionPlayerDiscardCards,
			Card:           sourceCard,
			requiredCounts: map[*Player]int{p1: 2, p2: 1},
			PendingPlayers: map[*Player]bool{p1: true, p2: true},
		}
		dto := PendingInteractionToDTO(pi)
		if dto == nil {
			t.Fatal("expected non-nil DTO")
		}
		if dto.Kind != InteractionPlayerDiscardCards || dto.SourceCardID != 999 {
			t.Errorf("unexpected Kind or SourceCardID: %+v", dto)
		}
		if len(dto.RequiredCounts) != 2 || dto.RequiredCounts[p1.Id] != 2 || dto.RequiredCounts[p2.Id] != 1 {
			t.Errorf("unexpected RequiredCounts: %v", dto.RequiredCounts)
		}
		if len(dto.PendingPlayers) != 2 {
			t.Errorf("expected 2 PendingPlayers, got %d", len(dto.PendingPlayers))
		}
	})

	t.Run("PlayerDonatesHandInteraction", func(t *testing.T) {
		pi := &PlayerDonatesHandInteraction{
			Kind:           InteractionPlayerDonatesHand,
			Card:           sourceCard,
			PendingPlayers: map[*Player]bool{p1: true},
		}
		dto := PendingInteractionToDTO(pi)
		if dto == nil {
			t.Fatal("expected non-nil DTO")
		}
		if dto.Kind != InteractionPlayerDonatesHand || dto.SourceCardID != 999 {
			t.Errorf("unexpected Kind or SourceCardID: %+v", dto)
		}
		if len(dto.PendingPlayers) != 1 || dto.PendingPlayers[0] != p1.Id {
			t.Errorf("unexpected PendingPlayers: %v", dto.PendingPlayers)
		}
	})
}

func TestNewGameSnapshot_NilAndEmptyEdgeCases(t *testing.T) {
	// Game with nil sub-components
	emptyGame := &Game{}
	snap := NewGameSnapshot(emptyGame)
	if snap.Boss != nil {
		t.Error("expected Boss to be nil for empty game")
	}
	if snap.RemainingDoorsCount != 0 {
		t.Errorf("expected RemainingDoorsCount 0, got %d", snap.RemainingDoorsCount)
	}
	if len(snap.Players) != 0 || len(snap.OpenedDoors) != 0 || len(snap.ActiveCurses) != 0 || len(snap.PlayedField) != 0 || len(snap.Artifacts) != 0 {
		t.Error("expected empty slices for empty game")
	}
	if snap.PendingInteraction != nil {
		t.Error("expected nil PendingInteraction")
	}

	// Individual converter nil checks
	pDTO := PlayerToDTO(nil)
	if pDTO.Id != "" {
		t.Errorf("expected empty PlayerDTO for nil player, got %+v", pDTO)
	}

	dDTO := DungeonCardToDTO(nil)
	if dDTO.Id != 0 {
		t.Errorf("expected empty DungeonCardDTO for nil card, got %+v", dDTO)
	}

	pcDTO := PlayerCardToDTO(nil)
	if pcDTO.Id != 0 {
		t.Errorf("expected empty PlayerCardDTO for nil card, got %+v", pcDTO)
	}

	aDTO := ArtifactToDTO(nil)
	if aDTO.Id != 0 {
		t.Errorf("expected empty ArtifactDTO for nil artifact, got %+v", aDTO)
	}

	piDTO := PendingInteractionToDTO(nil)
	if piDTO != nil {
		t.Errorf("expected nil PendingInteractionDTO for nil interaction, got %+v", piDTO)
	}
}

func TestDungeonCardToDTO_AllCardKinds(t *testing.T) {
	tests := []struct {
		name     string
		card     DungeonCard
		wantKind DungeonCardKind
		wantName string
		wantRes  []ResourceType
		wantDesc string
	}{
		{
			name:     "DoorMonster",
			card:     &DoorCard{Id: 1, Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}},
			wantKind: CardMonster,
			wantName: "Goblin",
			wantRes:  []ResourceType{Sword},
		},
		{
			name:     "DoorObstacle",
			card:     &DoorCard{Id: 2, Type: DoorObstacle, Name: "Trap", Resources: []ResourceType{Jump, Shield}},
			wantKind: CardObstacle,
			wantName: "Trap",
			wantRes:  []ResourceType{Jump, Shield},
		},
		{
			name:     "DoorPerson",
			card:     &DoorCard{Id: 3, Type: DoorPerson, Name: "Merchant", Resources: []ResourceType{Scroll}},
			wantKind: CardPerson,
			wantName: "Merchant",
			wantRes:  []ResourceType{Scroll},
		},
		{
			name:     "MiniBoss",
			card:     &MiniBossCard{Id: 4, Type: ChallengeMiniBoss, Name: "Dragon", Resources: []ResourceType{Sword, Arrow}},
			wantKind: CardMiniBoss,
			wantName: "Dragon",
			wantRes:  []ResourceType{Sword, Arrow},
		},
		{
			name:     "EventCard",
			card:     &EventCard{Id: 5, Type: ChallengeEvent, Name: "Cave In", Description: "All players discard"},
			wantKind: CardEvent,
			wantName: "Cave In",
			wantDesc: "All players discard",
		},
		{
			name:     "CurseCard",
			card:     &CurseCard{Id: 6, Type: ChallengeCurse, Name: "Curse of Blindness", Description: "Hands hidden"},
			wantKind: CardCurse,
			wantName: "Curse of Blindness",
			wantDesc: "Hands hidden",
		},
		{
			name:     "BossMat",
			card:     &BossMat{Id: 7, Name: "The Final Boss", Resources: []ResourceType{Sword, Shield, Arrow, Jump, Scroll}},
			wantKind: CardBoss,
			wantName: "The Final Boss",
			wantRes:  []ResourceType{Sword, Shield, Arrow, Jump, Scroll},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := DungeonCardToDTO(tt.card)
			if dto.Id != tt.card.ID() {
				t.Errorf("Id mismatch: got %d, want %d", dto.Id, tt.card.ID())
			}
			if dto.Kind != tt.wantKind {
				t.Errorf("Kind mismatch: got %v, want %v", dto.Kind, tt.wantKind)
			}
			if dto.Name != tt.wantName {
				t.Errorf("Name mismatch: got %q, want %q", dto.Name, tt.wantName)
			}
			if dto.Description != tt.wantDesc {
				t.Errorf("Description mismatch: got %q, want %q", dto.Description, tt.wantDesc)
			}
			if len(dto.Resources) != len(tt.wantRes) {
				t.Errorf("Resources length mismatch: got %d, want %d", len(dto.Resources), len(tt.wantRes))
			}
		})
	}
}

func TestPlayerCardToDTO_AllCardKinds(t *testing.T) {
	rc := &ResourceCard{Id: 10, Resources: []ResourceType{Sword, WildCard}}
	rcDTO := PlayerCardToDTO(rc)
	if rcDTO.Id != 10 || rcDTO.Kind != PlayerCardResource || len(rcDTO.Resources) != 2 {
		t.Errorf("unexpected ResourceCard DTO: %+v", rcDTO)
	}

	ac := &ActionCard{Id: 20, Name: "Heal", Description: "Recovers discard pile"}
	acDTO := PlayerCardToDTO(ac)
	if acDTO.Id != 20 || acDTO.Kind != PlayerCardAction || acDTO.Name != "Heal" || acDTO.Description != "Recovers discard pile" {
		t.Errorf("unexpected ActionCard DTO: %+v", acDTO)
	}
}

func TestRunner_Snapshot(t *testing.T) {
	paladin, _ := NewPlayer("Arthur", Paladin, true)
	door1 := &DoorCard{Id: nextCardId(), Type: DoorMonster, Name: "Goblin", Resources: []ResourceType{Sword}}
	boss := &BossMat{Id: nextCardId(), Name: "Boss", Resources: []ResourceType{Shield}}
	dungeon := &Dungeon{
		Boss:  boss,
		Doors: []DungeonCard{door1},
	}

	round := newRoundState(Playing)
	round.Dungeon = dungeon
	game := &Game{
		Players:    []*Player{paladin},
		HandSize:   3,
		LevelState: round,
	}
	_, _ = game.LevelState.Playfield.AddDungeonCard(door1, game)
	registerCardsInTestGame(game)

	runner := NewRunner(game)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = runner.Run(ctx)
	}()

	// Query snapshot through runner
	snap, err := runner.Snapshot(ctx)
	if err != nil {
		t.Fatalf("failed to take snapshot: %v", err)
	}

	if snap.Status != Playing {
		t.Errorf("expected snapshot Status Playing, got %v", snap.Status)
	}
	if len(snap.Players) != 1 || snap.Players[0].Name != "Arthur" {
		t.Errorf("unexpected players in snapshot: %+v", snap.Players)
	}
	if len(snap.OpenedDoors) != 1 || snap.OpenedDoors[0].Id != door1.ID() {
		t.Errorf("unexpected opened doors in snapshot: %+v", snap.OpenedDoors)
	}

	// Verify snapshot on canceled context returns error
	canceledCtx, cancelImmediate := context.WithCancel(context.Background())
	cancelImmediate()
	_, err = runner.Snapshot(canceledCtx)
	if err == nil {
		t.Error("expected error with canceled context, got nil")
	}

	// Stop runner and verify snapshot returns GameTerminatedError
	cancel()
	time.Sleep(50 * time.Millisecond) // Allow runner to finish shutdown
	_, err = runner.Snapshot(context.Background())
	if err == nil {
		t.Error("expected GameTerminatedError when taking snapshot on terminated runner, got nil")
	}
}
