package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"
)

const debugEnvVar = "FMG_DEBUG"

func main() {
	iconsFlag := flag.String("icons", "auto", "icon set: auto, emoji, symbols or ascii (env: "+iconsEnvVar+")")
	debugFlag := flag.Bool("debug", false, "show the events, snapshots and errors log panels (env: "+debugEnvVar+")")
	flag.Parse()

	activeIconTier = resolveIconTier(*iconsFlag, os.Getenv, runtime.GOOS)

	p := tea.NewProgram(NewModel(*debugFlag || os.Getenv(debugEnvVar) != ""))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
