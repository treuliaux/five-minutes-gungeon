package game

type Context interface {
	isContext()
	Engine() GameEngine
}

type AbilityContext struct {
	engine       GameEngine
	Player       *Player
	TargetCard   DungeonCard
	TargetPlayer *Player
}

func (AbilityContext) isContext() {}
func (a AbilityContext) Engine() GameEngine {
	return a.engine
}

type CardActionContext struct {
	engine        GameEngine
	Player        *Player
	Card          PlayerCard
	TargetCard    DungeonCard
	TargetPlayers []*Player
}

func (CardActionContext) isContext() {}
func (a CardActionContext) Engine() GameEngine {
	return a.engine
}

type CardEventContext struct {
	engine GameEngine
	Card   DungeonCard
	Input  PendingInteraction
}

func (CardEventContext) isContext() {}
func (a CardEventContext) Engine() GameEngine {
	return a.engine
}

type CardCurseContext struct {
	engine GameEngine
	Card   DungeonCard
	Input  PendingInteraction
}

func (CardCurseContext) isContext() {}
func (c CardCurseContext) Engine() GameEngine {
	return c.engine
}

type ArtifactActionContext struct {
	engine       GameEngine
	Player       *Player
	Artifact     *ArtifactCard
	ChosenAction ArtifactActionIndex
	Target       DungeonCard
}

func (ArtifactActionContext) isContext() {}
func (a ArtifactActionContext) Engine() GameEngine {
	return a.engine
}
