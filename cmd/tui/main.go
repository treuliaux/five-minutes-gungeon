package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func main() {
	p := tea.NewProgram(NewModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	return
}

func formatDebug(input string, length int) string {
	asRunes := []rune(input)

	asString := string(asRunes[:min(length, len(asRunes))])
	asString += strings.Repeat(".", length-len(asString))

	return asString
}
