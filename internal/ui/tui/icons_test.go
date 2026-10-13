package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func envFrom(env map[string]string) envLookup {
	return func(key string) string { return env[key] }
}

func TestIconsAreSafeInEveryTier(t *testing.T) {
	for _, tier := range []iconTier{iconTierEmoji, iconTierSymbols, iconTierASCII} {
		for id := iconSword; id <= iconBoss; id++ {
			glyph := iconIn(tier, id)
			if glyph == "" {
				t.Errorf("tier %s: icon %d is empty", tier, id)
			}
			if strings.ContainsRune(glyph, '\uFE0F') {
				t.Errorf("tier %s: icon %d (%q) uses a variation selector", tier, id, glyph)
			}
			if tier == iconTierASCII && strings.ContainsFunc(glyph, func(r rune) bool { return r > 127 }) {
				t.Errorf("ascii tier: icon %d (%q) is not ASCII", id, glyph)
			}
			if tier == iconTierSymbols && lipgloss.Width(glyph) != 1 {
				t.Errorf("symbols tier: icon %d (%q) is not 1 cell wide", id, glyph)
			}
			if tier == iconTierEmoji && id != iconInfinite && lipgloss.Width(glyph) != 2 {
				t.Errorf("emoji tier: icon %d (%q) is not 2 cells wide", id, glyph)
			}
		}
	}
}

func TestResolveIconTier(t *testing.T) {
	utf8 := map[string]string{"LANG": "en_US.UTF-8", "TERM": "xterm-256color"}

	tests := []struct {
		name  string
		flag  string
		env   map[string]string
		goos  string
		wants iconTier
	}{
		{"flag wins over env", "ascii", map[string]string{IconsEnvVar: "emoji"}, "linux", iconTierASCII},
		{"env wins over detection", "auto", map[string]string{IconsEnvVar: "symbols", "LANG": "en_US.UTF-8", "KITTY_WINDOW_ID": "1"}, "linux", iconTierSymbols},
		{"non utf8 locale", "", map[string]string{"LANG": "C"}, "linux", iconTierASCII},
		{"empty locale on linux", "", map[string]string{}, "linux", iconTierASCII},
		{"raw console", "", map[string]string{"LANG": "en_US.UTF-8", "TERM": "linux"}, "linux", iconTierASCII},
		{"tmux", "", map[string]string{"LANG": "en_US.UTF-8", "TMUX": "/tmp/tmux", "KITTY_WINDOW_ID": "1"}, "linux", iconTierSymbols},
		{"kitty", "", map[string]string{"LANG": "en_US.UTF-8", "TERM": "xterm-kitty"}, "linux", iconTierEmoji},
		{"recent VTE", "", map[string]string{"LANG": "en_US.UTF-8", "VTE_VERSION": "7600"}, "linux", iconTierEmoji},
		{"unknown terminal", "", utf8, "linux", iconTierSymbols},
		{"windows terminal", "", map[string]string{"WT_SESSION": "abc"}, "windows", iconTierEmoji},
		{"legacy windows console", "", map[string]string{}, "windows", iconTierSymbols},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResolveIconTier(tt.flag, envFrom(tt.env), tt.goos)
			if activeIconTier != tt.wants {
				t.Errorf("got %s, want %s", activeIconTier, tt.wants)
			}
		})
	}
}
