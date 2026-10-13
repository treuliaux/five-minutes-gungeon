package main

import (
	"image/color"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/treuliaux/five-minutes-gungeon/internal/game"
)

// cardKindStyle is the visual identity of a dungeon card kind. Three redundant cues (icon, colour, border shape)
// make a kind recognisable even without colour support.
type cardKindStyle struct {
	icon   iconID
	label  string
	color  color.Color
	border lipgloss.Border
}

func (s cardKindStyle) text() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.color).Bold(true)
}

func styleOfCardKind(kind game.DungeonCardKind) cardKindStyle {
	switch kind {
	case game.CardMonster:
		return cardKindStyle{icon: iconMonster, label: "MONSTER", color: lipgloss.Color("88"), border: lipgloss.NormalBorder()}
	case game.CardObstacle:
		return cardKindStyle{icon: iconObstacle, label: "OBSTACLE", color: lipgloss.Color("213"), border: lipgloss.NormalBorder()}
	case game.CardPerson:
		return cardKindStyle{icon: iconPerson, label: "PERSON", color: lipgloss.Color("214"), border: lipgloss.NormalBorder()}
	case game.CardMiniBoss:
		return cardKindStyle{icon: iconMiniBoss, label: "MINI-BOSS", color: lipgloss.Color("166"), border: lipgloss.ThickBorder()}
	case game.CardEvent:
		return cardKindStyle{icon: iconEvent, label: "EVENT", color: lipgloss.Color("39"), border: lipgloss.NormalBorder()}
	case game.CardCurse:
		return cardKindStyle{icon: iconCurse, label: "CURSE", color: lipgloss.Color("93"), border: lipgloss.RoundedBorder()}
	case game.CardBoss:
		return cardKindStyle{icon: iconBoss, label: "BOSS", color: lipgloss.Color("9"), border: lipgloss.DoubleBorder()}
	}

	return cardKindStyle{icon: iconWarning, label: "UNKNOWN", color: lipgloss.Color("241"), border: lipgloss.NormalBorder()}
}

var kindKeywordRegexp = regexp.MustCompile(`\[(Monster|Obstacle|Person|Mini-Boss|Event|Curse|Boss)\]`)

var kindByKeyword = map[string]game.DungeonCardKind{
	"Monster":   game.CardMonster,
	"Obstacle":  game.CardObstacle,
	"Person":    game.CardPerson,
	"Mini-Boss": game.CardMiniBoss,
	"Event":     game.CardEvent,
	"Curse":     game.CardCurse,
	"Boss":      game.CardBoss,
}

// colorizeDescription paints the card kind keywords of a description (e.g. "[Monster]") with the colour of the kind.
// Other bracketed keywords ([Resource], [Hand]...) are left untouched.
func colorizeDescription(text string) string {
	return kindKeywordRegexp.ReplaceAllStringFunc(text, func(match string) string {
		kind, ok := kindByKeyword[strings.Trim(match, "[]")]
		if !ok {
			return match
		}

		return styleOfCardKind(kind).text().Render(match)
	})
}
