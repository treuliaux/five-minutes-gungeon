package main

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

func TestColorizeDescriptionKeepsText(t *testing.T) {
	for _, text := range []string{"Defeat a [Monster]", "Cure a [Curse] or a [Mini-Boss]", "Counts as any [Resource]", "No keyword"} {
		if got := ansi.Strip(colorizeDescription(text)); got != text {
			t.Errorf("colorizeDescription(%q) changed visible text to %q", text, got)
		}
	}
}

func TestEveryDungeonCardKindHasAStyle(t *testing.T) {
	for _, kind := range []game.DungeonCardKind{game.CardMonster, game.CardObstacle, game.CardPerson, game.CardMiniBoss, game.CardEvent, game.CardCurse, game.CardBoss} {
		style := styleOfCardKind(kind)
		if style.label == "UNKNOWN" || style.color == nil {
			t.Errorf("kind %s has no style", kind)
		}
		if lipgloss.Width(style.border.TopLeft) != 1 || lipgloss.Width(style.border.Top) != 1 || lipgloss.Width(style.border.TopRight) != 1 {
			t.Errorf("kind %s has a border whose top glyphs are not 1 cell wide", kind)
		}
	}
}
