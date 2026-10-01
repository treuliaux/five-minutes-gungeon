package game

import (
	"fmt"
	"slices"
	"time"
)

//go:generate go run github.com/fairjungle/enumer -type=Status -json
type Status uint8

const (
	Waiting Status = iota
	Victory
	Defeat
	Playing
)

const (
	cardPlayDebounceDuration = time.Second / 2
	gameDuration             = 5 * time.Minute
)

type Game struct {
	// Current level state
	LevelState *LevelState

	// Campaign & Session Configuration (Persistent across levels)
	Config   Config
	Players  []*Player
	HandSize int
	level    int
}

type GameEngine interface {
	DefeatDoor(target DungeonCard) ([]Event, error)
	OpenDoor() ([]Event, error)
	SendDungeonCardBottomDungeon(card DungeonCard) ([]Event, error)
	StopTime(player *Player) ([]Event, error)
	ClearStopTimeCurse()

	ListPlayers() []*Player
	ActiveDoors(f *DoorsFilter) []DungeonCard
	ListArtifacts() []*ArtifactCard
	DungeonCardByID(id CardID) (DungeonCard, error)
	PlayerCardsByIDs(id []CardID) ([]PlayerCard, error)
	PlayerByID(id PlayerID) (*Player, error)
	ArtifactByID(id ArtifactID) (*ArtifactCard, error)

	PlayArbitraryCard(player *Player, card PlayerCard) ([]Event, error)
	RemoveDungeonCard(card DungeonCard) ([]Event, error)
	RemovePlayerCard(card PlayerCard) ([]Event, error)
	DiscardTopCardFromDungeon() ([]Event, error)
	RefillPlayerHand(player *Player) ([]Event, error)
}

func NewGameWithExtension() *Game {
	return &Game{
		LevelState: newRoundState(Waiting),
		level:      1,
		Config:     Config{UseExtension: true, ResetLevelOnDefeat: true, EventReactionTime: 2},
	}
}

func NewGame(cfg Config) *Game {
	return &Game{
		LevelState: newRoundState(Waiting),
		level:      1,
		Config:     cfg,
	}
}

func (g *Game) Tick(delta time.Duration) ([]Event, error) {
	g.LevelState.RealElapsedTime += delta
	events := []Event{TimerTickEvent{TimeLeftDuration: gameDuration - g.LevelState.InGameTimer}}
	if g.LevelState.Status == Victory || g.LevelState.Status == Defeat {
		return nil, nil
	}
	if g.LevelState.Status != Playing {
		return nil, ErrGameNotPlaying
	}
	if g.LevelState.IsTimeFrozen {
		delta = 0
	}
	g.LevelState.InGameTimer += delta
	if g.LevelState.InGameTimer >= gameDuration {
		g.defeat()

		return append(events, GameLostEvent{}), nil
	}
	if g.LevelState.PendingInteraction != nil {
		return events, nil
	}
	resolveActiveEventsEvents, err := g.resolveActiveEvents()
	events = append(events, resolveActiveEventsEvents...)
	if err != nil {
		return events, err
	}
	if (g.actionDebounce() && !g.LevelState.IsFightingBoss) || !g.LevelState.Playfield.IsPlayfieldBeaten() {
		return events, nil
	}

	defeatAllDoorsEvents, err := g.LevelState.Playfield.DefeatAllDoors()
	events = append(events, defeatAllDoorsEvents...)
	if err != nil {
		return events, err
	}
	if g.LevelState.IsFightingBoss {
		events = append(events, GameWonEvent{})
		events = append(events, g.victory()...)

		return events, nil
	}
	openDoorEvents, err := g.OpenDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) Apply(cmd Command) ([]Event, error) {
	var err error
	var events []Event

	switch cmd := cmd.(type) {
	case AddPlayerCmd:
		events, err = g.addPlayer(cmd)
	case ChangeHeroCmd:
		events, err = g.changePlayerHero(cmd)
	case StartCmd:
		events, err = g.start()
	case PlayCardCmd:
		events, err = g.playCard(cmd)
	case DiscardCardsCmd:
		events, err = g.discardCard(cmd)
	case UseHeroAbilityCmd:
		events, err = g.useHeroAbility(cmd)
	case SubmitPromptChoiceCmd:
		events, err = g.submitPromptChoice(cmd)
	case UseArtifactCmd:
		events, err = g.useArtifact(cmd)
	case GetSnapshotCmd:
		cmd.reply <- NewGameSnapshot(g)

		return nil, nil
	}
	if cmd.Reply() != nil {
		cmd.Reply() <- err
	}

	return events, err
}

func (g *Game) StopTime(player *Player) ([]Event, error) {
	if g.LevelState.IsTimeFrozen {
		return nil, fmt.Errorf("time is already frozen")
	}
	if g.HasActiveCurseEffect(TimeCannotBeStopped) {
		return player.VoidHand(TimeCannotBeStopped, g)
	}
	if g.HasActiveCurseEffect(ThreeDiscardsWhenTimeStops) {
		for _, p := range g.Players {
			g.LevelState.CurseExpectingDiscards[p] = 3
		}
	}
	g.LevelState.IsTimeFrozen = true

	return []Event{TimeFrozenEvent{ByPlayerID: player.Id}}, nil
}

func (g *Game) ListPlayers() []*Player {
	return g.Players
}

func (g *Game) ListArtifacts() []*ArtifactCard {
	return g.LevelState.Playfield.Artifacts
}

func (g *Game) PlayArbitraryCard(p *Player, card PlayerCard) ([]Event, error) {
	g.LevelState.PlayerCardsMap[card.ID()] = card

	return g.LevelState.Playfield.AddPlayerCard(p, card)
}

func (g *Game) RemovePlayerCard(card PlayerCard) ([]Event, error) {
	return g.LevelState.Playfield.RemovePlayerCard(card)
}

func (g *Game) RemoveDungeonCard(card DungeonCard) ([]Event, error) {
	return g.LevelState.Playfield.RemoveDungeonCard(g, card)
}

func (g *Game) HasActiveCurseEffect(effect GameCurseEffect) bool {
	if g.LevelState.Playfield == nil {
		return false
	}
	for _, curse := range g.LevelState.Playfield.ActiveCurses {
		if curse.Effect == effect {
			return true
		}
	}

	return false
}

func (g *Game) TargetHandSize() int {
	if g.HasActiveCurseEffect(HandSizeLimitedToThree) {
		return min(g.HandSize, 3)
	}

	return g.HandSize
}

func (g *Game) RefillPlayerHand(player *Player) ([]Event, error) {
	if g.HandSize == 0 {
		g.determineHandSize()
	}

	var events []Event
	var err error
	nbToDraw := g.TargetHandSize() - len(player.Hand)
	if nbToDraw > 0 {
		events, err = player.DrawCardsFromDeck(nbToDraw)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}

func (g *Game) clearField() []Event {
	fieldEvents, err := g.LevelState.Playfield.ClearField()
	if err != nil {
		return nil
	}
	g.LevelState.LastPlayedCardTimer = g.LevelState.RealElapsedTime

	return fieldEvents
}

func (g *Game) determineHandSize() {
	switch len(g.Players) {
	case 2:
		g.HandSize = 5
	case 3:
		g.HandSize = 4
	case 4, 5, 6:
		g.HandSize = 3
	}
}

func (g *Game) generateDungeon() {
	if g.LevelState.Dungeon == nil {
		if g.Config.UseExtension {
			g.LevelState.Dungeon = NewExtensionDungeon(g.level, len(g.Players))

			return
		}
		g.LevelState.Dungeon = NewBaseDungeon(g.level, len(g.Players))
	}
}

func (g *Game) actionDebounce() bool {
	return g.LevelState.RealElapsedTime-g.LevelState.LastPlayedCardTimer <= cardPlayDebounceDuration
}

func (g *Game) resolveActiveEvents() ([]Event, error) {
	var events []Event
	for _, card := range slices.Clone(g.LevelState.Playfield.OpenedDoors) {
		eventCard, ok := card.(*EventCard)
		if !ok || eventCard.Action == nil {
			continue
		}
		if g.LevelState.InGameTimer-eventCard.OpenedTime < time.Duration(g.Config.EventReactionTime)*time.Second {
			return nil, nil
		}

		ctx := &CardEventContext{
			engine: g,
			Card:   eventCard,
		}
		if interaction := eventCard.Action.Interaction(ctx); interaction != nil {
			g.LevelState.PendingInteraction = interaction

			initEvents, err := interaction.Init(ctx)
			events = append(events, initEvents...)
			if err != nil {
				return events, err
			}

			dto := PendingInteractionToDTO(interaction)
			promptEvent := EventPromptOpenedEvent{
				Kind:           dto.Kind,
				RequiredCounts: dto.RequiredCounts,
			}
			return append(events, promptEvent), nil
		}
		resolveEvents, err := g.LevelState.Playfield.ResolveEvent(ctx)
		events = append(events, resolveEvents...)
		if err != nil {
			return events, err
		}

		defeatDoorEvents, err := g.DefeatDoor(eventCard)
		events = append(events, defeatDoorEvents...)
		if err != nil {
			return events, err
		}
	}

	return events, nil
}
