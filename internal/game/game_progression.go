package game

func (g *Game) defeat() []Event {
	level := g.level
	if g.Config.ResetLevelOnDefeat {
		level = 1
	}
	currentBossDTO := DungeonCardToDTO(g.LevelState.Dungeon.Boss)
	nextBossDTO := DungeonCardToDTO(BossList()[level-1])
	g.prepareNextLevel(Defeat, level)

	return []Event{DungeonPrevailed{
		Boss:     currentBossDTO,
		NextBoss: nextBossDTO,
	}}
}

func (g *Game) victory() []Event {
	currentBossDTO := DungeonCardToDTO(g.LevelState.Dungeon.Boss)
	if g.level == len(g.LevelState.Dungeon.bossList)-1 { // Bonus boss is currently disabled
		g.prepareNextLevel(Victory, 1)

		return []Event{CampaignEnded{Boss: currentBossDTO}}
	}
	nextBossDTO := DungeonCardToDTO(BossList()[g.level])
	g.prepareNextLevel(Victory, g.level+1)

	return []Event{DungeonDefeated{
		Boss:     currentBossDTO,
		NextBoss: nextBossDTO,
	}}
}

func (g *Game) prepareNextLevel(status Status, nextLevel int) {
	g.LevelState = newRoundState(status)
	g.level = nextLevel
	for _, p := range g.Players {
		p.PrepareForNextLevel()
	}
}
