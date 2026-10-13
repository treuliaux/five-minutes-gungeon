package main

import "github.com/treuliaux/five-minutes-gungeon/internal/game"

type iconTier int

const (
	// iconTierEmoji only uses glyphs that are 2 cells wide by default (no U+FE0F sequences, no narrow symbols that
	// fonts like to render as wide emoji).
	iconTierEmoji iconTier = iota
	// iconTierSymbols only uses plain 1-cell Unicode symbols.
	iconTierSymbols
	// iconTierASCII only uses printable ASCII.
	iconTierASCII
)

func (t iconTier) String() string {
	switch t {
	case iconTierEmoji:
		return "emoji"
	case iconTierSymbols:
		return "symbols"
	case iconTierASCII:
		return "ascii"
	default:
		return "unknown"
	}
}

type iconID int

const (
	iconSword iconID = iota
	iconArrow
	iconShield
	iconJump
	iconScroll
	iconWild
	iconInfinite
	iconTarget
	iconFrozen
	iconHand
	iconDeck
	iconDiscard
	iconCurse
	iconWarning
	iconVictory
	iconDefeat
	iconPin
	iconTip
	iconEnabled
	iconDisabled
	iconRoster
	iconMonster
	iconObstacle
	iconPerson
	iconMiniBoss
	iconEvent
	iconBoss
)

type iconVariants struct {
	emoji   string
	symbols string
	ascii   string
}

var iconTable = map[iconID]iconVariants{
	iconSword:    {emoji: "🔪", symbols: "†", ascii: "Sw"},
	iconArrow:    {emoji: "🏹", symbols: "→", ascii: "Ar"},
	iconShield:   {emoji: "🔰", symbols: "■", ascii: "Sh"},
	iconJump:     {emoji: "🦵", symbols: "↑", ascii: "Ju"},
	iconScroll:   {emoji: "📜", symbols: "§", ascii: "Sc"},
	iconWild:     {emoji: "⭐", symbols: "★", ascii: "**"},
	iconInfinite: {emoji: "∞", symbols: "∞", ascii: "~"},
	iconTarget:   {emoji: "🎯", symbols: "◎", ascii: "(@)"},
	iconFrozen:   {emoji: "🧊", symbols: "*", ascii: "*"},
	iconHand:     {emoji: "✋", symbols: "H", ascii: "H"},
	iconDeck:     {emoji: "🎴", symbols: "D", ascii: "D"},
	iconDiscard:  {emoji: "🚮", symbols: "X", ascii: "X"},
	iconCurse:    {emoji: "💀", symbols: "x", ascii: "x"},
	iconWarning:  {emoji: "🚨", symbols: "!", ascii: "!"},
	iconVictory:  {emoji: "🏆", symbols: "★", ascii: "*"},
	iconDefeat:   {emoji: "💀", symbols: "x", ascii: "x"},
	iconPin:      {emoji: "📌", symbols: "»", ascii: ">"},
	iconTip:      {emoji: "💡", symbols: "»", ascii: ">"},
	iconEnabled:  {emoji: "✅", symbols: "✓", ascii: "+"},
	iconDisabled: {emoji: "❌", symbols: "✗", ascii: "-"},
	iconRoster:   {emoji: "🔰", symbols: "■", ascii: "#"},
	iconMonster:  {emoji: "👹", symbols: "Ø", ascii: "M"},
	iconObstacle: {emoji: "🧱", symbols: "▓", ascii: "O"},
	iconPerson:   {emoji: "🧑", symbols: "☻", ascii: "P"},
	iconMiniBoss: {emoji: "😈", symbols: "♛", ascii: "m"},
	iconEvent:    {emoji: "⚡", symbols: "↯", ascii: "E"},
	iconBoss:     {emoji: "👑", symbols: "♚", ascii: "B"},
}

// activeIconTier is chosen once at startup (see main) and then only read.
var activeIconTier = iconTierEmoji

func icon(id iconID) string {
	return iconIn(activeIconTier, id)
}

func iconIn(tier iconTier, id iconID) string {
	variants := iconTable[id]

	switch tier {
	case iconTierEmoji:
		return variants.emoji
	case iconTierSymbols:
		return variants.symbols
	case iconTierASCII:
		return variants.ascii
	}

	return variants.ascii
}

func resourceTypeToIcon(rt game.ResourceType) string {
	switch rt {
	case game.Sword:
		return icon(iconSword)
	case game.Arrow:
		return icon(iconArrow)
	case game.Shield:
		return icon(iconShield)
	case game.Jump:
		return icon(iconJump)
	case game.Scroll:
		return icon(iconScroll)
	case game.WildCard:
		return icon(iconWild)
	case game.InfiniteSword:
		return icon(iconSword) + icon(iconInfinite)
	case game.InfiniteArrow:
		return icon(iconArrow) + icon(iconInfinite)
	case game.InfiniteShield:
		return icon(iconShield) + icon(iconInfinite)
	case game.InfiniteJump:
		return icon(iconJump) + icon(iconInfinite)
	case game.InfiniteScroll:
		return icon(iconScroll) + icon(iconInfinite)
	default:
		return ""
	}
}
