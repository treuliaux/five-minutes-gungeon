package game

type CurseHook func(ctx Context) ([]Event, error)

type GameCurseEffect int

const (
	NoEffect GameCurseEffect = iota

	TimeCannotBeStopped
	ActionsCannotBePlayed
	AbilitiesCannotBePlayed
	DoorsOpenInPairs
	HandSizeLimitedToThree
	FlippedHeroMat
	ThreeDiscardsWhenTimeStops

	HandsHidden
	PlayersHandFacingAway

	PlayersCanOnlyUseOneHandToPlay
	PlayersMustOnlySayWaffles
	PlayersMustOnlySayMeow
)
