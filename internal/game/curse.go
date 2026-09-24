package game

type CurseEffect interface {
	isCurseCard()
	Execute(ctx CardCurseContext) ([]Event, error)
}

type CardCurseContext struct {
	Engine GameEngine
	Card   DungeonCard
}
