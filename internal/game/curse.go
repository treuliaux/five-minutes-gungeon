package game

type CurseHook func(ctx Context) ([]Event, error)

//go:generate go run github.com/fairjungle/enumer -type=GameCurseEffect -json
type GameCurseEffect uint8

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
