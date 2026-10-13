package tui

import (
	"context"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

type mockGameController struct {
	lastDispatched game.Command
}

func (m *mockGameController) Dispatch(_ context.Context, cmd game.Command) error {
	m.lastDispatched = cmd
	return nil
}

func (m *mockGameController) GetSnapshot() (game.GameSnapshotDTO, error) {
	return game.GameSnapshotDTO{}, nil
}

func (m *mockGameController) Events() <-chan game.Event {
	return nil
}

func (m *mockGameController) Snapshots() <-chan game.GameSnapshotDTO {
	return nil
}

func (m *mockGameController) Errors() <-chan error {
	return nil
}

func (m *mockGameController) Close() error {
	return nil
}

func TestHeroSelectionExtensionGating(t *testing.T) {
	// When extension is disabled, Druid and Shaman should not be available
	baseModel := newSelectHeroModel(false)
	for _, c := range baseModel.choices {
		if c == game.Druid || c == game.Shaman {
			t.Fatalf("expected Druid and Shaman to be excluded when extension is false, found %v", c)
		}
	}

	// When extension is enabled, Druid and Shaman should be available
	extModel := newSelectHeroModel(true)
	hasDruid := slices.Contains(extModel.choices, game.Druid)
	hasShaman := slices.Contains(extModel.choices, game.Shaman)
	if !hasDruid || !hasShaman {
		t.Fatalf("expected Druid and Shaman to be included when extension is true")
	}
}

func TestPlayModelDoorTargeting(t *testing.T) {
	mockCtrl := &mockGameController{}
	m := newPlayModel()
	m.controller = mockCtrl
	m.ctx = context.Background()
	m.snapshot = game.GameSnapshotDTO{
		Status: game.Playing,
		OpenedDoors: []game.DungeonCardDTO{
			{Id: 101, Name: "Gungeon Bat", Kind: game.CardMonster},
			{Id: 102, Name: "Spike Trap", Kind: game.CardObstacle},
		},
		Players: []game.PlayerDTO{
			{
				Id:        "Player1",
				Name:      "Player1",
				HeroClass: game.Paladin,
				Hand: []game.PlayerCardDTO{
					{Id: 1, Name: "Smite Card", Kind: game.PlayerCardAction, TargetType: game.TargetCard},
				},
			},
		},
	}

	// Select the action card
	m = m.toggleSelectCard(1)
	if !m.selectedCards[1] {
		t.Fatalf("expected card 1 to be selected")
	}

	// Trigger play selected cards (should enter targeting mode because there are 2 opened doors)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.targeting.kind != TargetKindDoor {
		t.Fatalf("expected targeting mode to be TargetKindDoor, got %v", m.targeting.kind)
	}
	if m.targeting.targetCardID != 101 {
		t.Fatalf("expected default target door to be 101, got %v", m.targeting.targetCardID)
	}

	// Press '2' to switch target to door 102
	m, _ = m.Update(tea.KeyPressMsg{Code: '2'})
	if m.targeting.targetCardID != 102 {
		t.Fatalf("expected target door to be 102, got %v", m.targeting.targetCardID)
	}

	// Confirm targeting with enter
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.targeting.kind != TargetKindNone {
		t.Fatalf("expected targeting mode to reset to TargetKindNone after confirmation")
	}
	playCmd, ok := mockCtrl.lastDispatched.(game.PlayCardCmd)
	if !ok {
		t.Fatalf("expected PlayCardCmd to be dispatched, got %T", mockCtrl.lastDispatched)
	}
	if playCmd.TargetCardID != 102 || playCmd.CardID != 1 {
		t.Fatalf("expected PlayCardCmd with CardID 1 and TargetCardID 102, got %+v", playCmd)
	}

	// Test Esc cancels targeting
	m.selectedCards[1] = true
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.targeting.kind != TargetKindDoor {
		t.Fatalf("expected targeting mode to be active")
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.targeting.kind != TargetKindNone {
		t.Fatalf("expected targeting mode to be canceled on Esc")
	}
}

func TestPlayModelPendingPromptModalAndDismissal(t *testing.T) {
	mockCtrl := &mockGameController{}
	m := newPlayModel()
	m.controller = mockCtrl
	m.ctx = context.Background()
	m.snapshot = game.GameSnapshotDTO{
		Status: game.Playing,
		PendingInteraction: &game.PendingInteractionDTO{
			Kind:           game.InteractionTeamChoiceResource,
			PendingPlayers: []game.PlayerID{"Player1", "Player2"},
		},
		Players: []game.PlayerDTO{
			{Id: "Player1", Name: "Player1", HeroClass: game.Paladin},
			{Id: "Player2", Name: "Player2", HeroClass: game.Sorceress},
		},
	}

	// View should render modal overlay
	view := m.View()
	if !slices.Contains([]rune(view), '⚡') {
		t.Fatalf("expected view to contain modal content")
	}

	// Press Esc to dismiss modal
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !m.dismissedPrompt {
		t.Fatalf("expected dismissedPrompt to be true after pressing esc")
	}

	// Press Space to re-open modal
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.dismissedPrompt {
		t.Fatalf("expected dismissedPrompt to be false after pressing space")
	}

	// Select option 3 (Shield) and confirm
	m, _ = m.Update(tea.KeyPressMsg{Code: '3'})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	promptCmd, ok := mockCtrl.lastDispatched.(game.SubmitPromptChoiceCmd)
	if !ok {
		t.Fatalf("expected SubmitPromptChoiceCmd, got %T", mockCtrl.lastDispatched)
	}
	if promptCmd.PlayerID != "Player1" || promptCmd.Resource != game.Shield {
		t.Fatalf("expected SubmitPromptChoiceCmd with PlayerID Player1 and Resource Shield, got %+v", promptCmd)
	}
}

func TestPlayModelPendingPromptDiscardSelection(t *testing.T) {
	mockCtrl := &mockGameController{}
	m := newPlayModel()
	m.controller = mockCtrl
	m.ctx = context.Background()
	m.snapshot = game.GameSnapshotDTO{
		Status: game.Playing,
		PendingInteraction: &game.PendingInteractionDTO{
			Kind:           game.InteractionPlayerDiscardCards,
			PendingPlayers: []game.PlayerID{"Player1"},
			RequiredCounts: map[game.PlayerID]int{"Player1": 1},
		},
		Players: []game.PlayerDTO{
			{
				Id:        "Player1",
				Name:      "Player1",
				HeroClass: game.Paladin,
				Hand: []game.PlayerCardDTO{
					{Id: 10, Name: "Card 1"},
					{Id: 11, Name: "Card 2"},
				},
			},
		},
	}

	// Toggle card 1
	m, _ = m.Update(tea.KeyPressMsg{Code: '1'})
	if !m.selectedCards[10] {
		t.Fatalf("expected card 10 to be selected")
	}

	// Submit with Enter
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	promptCmd, ok := mockCtrl.lastDispatched.(game.SubmitPromptChoiceCmd)
	if !ok {
		t.Fatalf("expected SubmitPromptChoiceCmd, got %T", mockCtrl.lastDispatched)
	}
	if len(promptCmd.CardIDs) != 1 || promptCmd.CardIDs[0] != 10 {
		t.Fatalf("expected SubmitPromptChoiceCmd with CardIDs [10], got %+v", promptCmd.CardIDs)
	}
}

func TestRenderOpenedDoors(t *testing.T) {
	m := newPlayModel()
	m.snapshot = game.GameSnapshotDTO{
		Status: game.Playing,
		OpenedDoors: []game.DungeonCardDTO{
			{
				Id:        1,
				Name:      "A Warrior Princess",
				Kind:      game.CardPerson,
				Resources: []game.ResourceType{game.Shield, game.Arrow},
			},
		},
	}
	rendered := m.renderOpenedDoors()
	t.Logf("Rendered opened doors:\n%s", rendered)
}
