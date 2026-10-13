package tui

import (
	"strconv"
	"strings"
)

const IconsEnvVar = "FMG_ICONS"

// minEmojiVTEVersion is the first VTE release (0.54) from which wide emoji are laid out reliably.
const minEmojiVTEVersion = 5400

type envLookup func(key string) string

func parseIconTier(value string) (iconTier, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "emoji":
		return iconTierEmoji, true
	case "symbols":
		return iconTierSymbols, true
	case "ascii":
		return iconTierASCII, true
	}

	return iconTierEmoji, false
}

// ResolveIconTier picks the tier from, in order: the --icons flag, the FMG_ICONS env var, auto-detection.
// "auto" (or an unknown value) falls through to the next source.
func ResolveIconTier(flagValue string, getenv envLookup, goos string) {
	if tier, ok := parseIconTier(flagValue); ok {
		activeIconTier = tier
	}

	if tier, ok := parseIconTier(getenv(IconsEnvVar)); ok {
		activeIconTier = tier
	}

	activeIconTier = detectIconTier(getenv, goos)
}

// detectIconTier guesses the richest tier the terminal can draw without breaking the layout. When unsure it picks
// "symbols", which is safe everywhere UTF-8 works.
func detectIconTier(getenv envLookup, goos string) iconTier {
	if !localeIsUTF8(getenv, goos) {
		return iconTierASCII
	}

	term := getenv("TERM")
	switch {
	case term == "linux", term == "dumb":
		return iconTierASCII
	case getenv("TMUX") != "", getenv("STY") != "", strings.HasPrefix(term, "screen"), strings.HasPrefix(term, "tmux"):
		// Multiplexers use their own width tables, which often differ from the outer terminal's.
		return iconTierSymbols
	}

	if goos == "windows" {
		if getenv("WT_SESSION") != "" {
			return iconTierEmoji
		}

		return iconTierSymbols
	}

	if terminalHandlesEmoji(getenv, term) {
		return iconTierEmoji
	}

	return iconTierSymbols
}

func localeIsUTF8(getenv envLookup, goos string) bool {
	locale := ""
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if locale = getenv(key); locale != "" {
			break
		}
	}

	if locale == "" {
		// Windows doesn't use these variables; elsewhere an empty locale means "C".
		return goos == "windows"
	}

	locale = strings.ToLower(locale)

	return strings.Contains(locale, "utf-8") || strings.Contains(locale, "utf8")
}

func terminalHandlesEmoji(getenv envLookup, term string) bool {
	switch getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm", "ghostty":
		return true
	}

	switch {
	case term == "xterm-kitty", term == "xterm-ghostty", strings.Contains(term, "alacritty"):
		return true
	case getenv("KITTY_WINDOW_ID") != "", getenv("WEZTERM_PANE") != "", getenv("GHOSTTY_RESOURCES_DIR") != "", getenv("KONSOLE_VERSION") != "":
		return true
	}

	vte, err := strconv.Atoi(getenv("VTE_VERSION"))
	if err != nil {
		return false
	}

	return vte >= minEmojiVTEVersion
}
