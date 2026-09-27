package game

import (
	"fmt"
	"slices"
)

func (g *Game) addPlayer(cmd AddPlayerCmd) ([]Event, error) {
	if g.LevelState.Status == Playing {
		return nil, ErrGameAlreadyStarted
	}
	if slices.ContainsFunc(g.Players, func(player *Player) bool {
		return player.Hero.Color == ColorFromHeroClass(cmd.Class)
	}) {
		return nil, ErrClassAlreadyPicked
	}
	player, err := NewPlayer(cmd.Name, cmd.Class, g.Config.UseExtension)
	if err != nil {
		return nil, err
	}
	g.Players = append(g.Players, player)

	return []Event{PlayerAddedEvent{PlayerID: player.Id}}, nil
}

func (g *Game) choosePlayerHero(cmd ChoosePlayerHeroCmd) ([]Event, error) {
	if g.LevelState.Status == Playing {
		return nil, ErrGameAlreadyStarted
	}
	if !g.Config.UseExtension && (cmd.Class == Druid || cmd.Class == Shaman) {
		return nil, ErrGameExtensionDisabled
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}

	if slices.ContainsFunc(otherPlayers(g.Players, p), func(player *Player) bool {
		return player.Hero.Color == ColorFromHeroClass(cmd.Class)
	}) {
		return nil, ErrClassAlreadyPicked
	}
	hero, err := NewHeroFromHeroClass(cmd.Class)
	if err != nil {
		return nil, err
	}
	deck := hero.NewBaseDeck()
	if g.Config.UseExtension {
		deck.IncludeExtension()
	}
	p.Hero = hero
	p.Deck = deck

	return []Event{PlayerChangedHeroEvent{Player: PlayerToDTO(p)}}, nil
}

func (g *Game) start() ([]Event, error) {
	if g.LevelState.Status == Playing {
		return nil, ErrGameAlreadyStarted
	}
	if len(g.Players) < 2 || len(g.Players) > 6 {
		return nil, ErrGameIncorrectPlayerNumber
	}

	g.generateDungeon()
	var deckColors []DeckColor
	for _, p := range g.Players {
		deckColors = append(deckColors, p.Hero.Color)
		if p.Deck != nil {
			continue
		}
		if g.Config.UseExtension {
			p.Deck = p.Hero.NewExtensionDeck()
			continue
		}
		p.Deck = p.Hero.NewBaseDeck()
	}
	if g.Config.UseExtension {
		g.LevelState.Playfield.SetupArtifacts(deckColors)
	}
	g.registerCards()

	var events []Event
	events = append(events, GameStartedEvent{})

	for _, player := range g.Players {
		drawEvents, err := g.RefillPlayerHand(player)
		events = append(events, drawEvents...)
		if err != nil {
			return events, err
		}
	}

	g.LevelState.Status = Playing
	g.LevelState.CurseExpectingDiscards = make(map[*Player]int, len(g.Players))

	openDoorEvents, err := g.OpenDoor()
	events = append(events, openDoorEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) playCard(cmd PlayCardCmd) ([]Event, error) {
	if g.LevelState.Status != Playing {
		return nil, ErrGameNotPlaying
	}
	if g.LevelState.PendingInteraction != nil {
		return nil, ErrPendingInteraction
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	c, err := g.PlayerCardByID(cmd.CardID)
	if err != nil {
		return nil, err
	}
	if !p.HasCardInHand(c) {
		return nil, ErrCardNotInHand
	}
	if g.HasActiveCurseEffect(HandSizeLimitedToThree) && len(p.Hand) > 3 {
		return p.VoidHand(HandSizeLimitedToThree, g)
	}
	if g.HasActiveCurseEffect(ThreeDiscardsWhenTimeStops) {
		if _, ok := g.LevelState.CurseExpectingDiscards[p]; ok {
			delete(g.LevelState.CurseExpectingDiscards, p)
			return p.VoidHand(ThreeDiscardsWhenTimeStops, g)
		}
	}
	if _, ok := c.(*ActionCard); ok && g.HasActiveCurseEffect(ActionsCannotBePlayed) {
		return p.VoidHand(ActionsCannotBePlayed, g)
	}

	p.RemoveCardFromHand(c)
	fieldEvents, err := g.LevelState.Playfield.AddPlayerCard(p, c)
	if err != nil {
		p.Hand = append(p.Hand, c)
		return nil, err
	}

	var events []Event
	prevLastPlayedCardTimer := g.LevelState.LastPlayedCardTimer
	g.LevelState.LastPlayedCardTimer = g.LevelState.RealElapsedTime
	events = append(events, fieldEvents...)
	wasTimeFrozen := g.LevelState.IsTimeFrozen
	if g.LevelState.IsTimeFrozen {
		g.LevelState.IsTimeFrozen = false
		events = append(events, TimeUnfrozenEvent{ByPlayerID: p.Id})
	}

	var actionEvents []Event
	if ac, ok := c.(*ActionCard); ok {
		var targetCard DungeonCard
		if cmd.TargetCardID != 0 {
			targetCard, err = g.DungeonCardByID(cmd.TargetCardID)
			if err != nil {
				p.Hand = append(p.Hand, c)

				return events, err
			}
		}
		var targetPlayers []*Player
		if len(cmd.TargetPlayerIDs) > 0 {
			targetPlayers, err = g.PlayersByIDs(cmd.TargetPlayerIDs)
			if err != nil {
				p.Hand = append(p.Hand, c)

				return events, err
			}
		}
		ctx := &CardActionContext{
			engine:        g,
			Player:        p,
			Card:          c,
			TargetCard:    targetCard,
			TargetPlayers: targetPlayers,
		}
		actionEvents, err = ac.Action.Execute(ctx)
		if err != nil {
			p.Hand = append(p.Hand, c)
			g.LevelState.Playfield.Field = slices.DeleteFunc(g.LevelState.Playfield.Field, func(item PlayerCard) bool { return item == c })
			g.LevelState.IsTimeFrozen = wasTimeFrozen
			g.LevelState.LastPlayedCardTimer = prevLastPlayedCardTimer

			return nil, err
		}
	}
	events = append(events, actionEvents...)

	// Action cards may defeat the boss (e.g., Holy Hand Grenade)
	if g.LevelState.Status != Playing {
		return events, nil
	}

	cardDrawnEvents, err := g.RefillPlayerHand(p)
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) discardCard(cmd DiscardCardsCmd) ([]Event, error) {
	if g.LevelState.Status != Playing {
		return nil, ErrGameNotPlaying
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	cards, err := g.PlayerCardsByIDs(cmd.CardIDs)
	if err != nil {
		return nil, err
	}

	if _, ok := g.LevelState.CurseExpectingDiscards[p]; ok {
		g.LevelState.CurseExpectingDiscards[p] -= len(cards)
		if g.LevelState.CurseExpectingDiscards[p] <= 0 {
			delete(g.LevelState.CurseExpectingDiscards, p)
		}
	}

	events, err := p.DiscardCards(cards)
	if err != nil {
		return events, err
	}
	refillEvents, err := g.RefillPlayerHand(p)
	events = append(events, refillEvents...)
	if err != nil {
		return events, err
	}

	return events, err
}

func (g *Game) useHeroAbility(cmd UseHeroAbilityCmd) ([]Event, error) {
	if g.LevelState.Status != Playing {
		return nil, ErrGameNotPlaying
	}
	if g.LevelState.PendingInteraction != nil {
		return nil, ErrPendingInteraction
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	discardCards, err := g.PlayerCardsByIDs(cmd.DiscardCardIDs)
	if err != nil {
		return nil, err
	}
	var targetCard DungeonCard
	if cmd.TargetCardID != 0 {
		targetCard, err = g.DungeonCardByID(cmd.TargetCardID)
		if err != nil {
			return nil, err
		}
	}
	var targetPlayer *Player
	if cmd.TargetPlayerID != "" {
		targetPlayer, err = g.PlayerByID(cmd.TargetPlayerID)
		if err != nil {
			return nil, err
		}
	}
	if g.HasActiveCurseEffect(AbilitiesCannotBePlayed) {
		return p.VoidHand(AbilitiesCannotBePlayed, g)
	}
	if g.HasActiveCurseEffect(HandSizeLimitedToThree) && len(p.Hand) > 3 {
		return p.VoidHand(HandSizeLimitedToThree, g)
	}
	if g.HasActiveCurseEffect(ThreeDiscardsWhenTimeStops) {
		if _, ok := g.LevelState.CurseExpectingDiscards[p]; ok {
			delete(g.LevelState.CurseExpectingDiscards, p)
			return p.VoidHand(ThreeDiscardsWhenTimeStops, g)
		}
	}

	discardCards = uniqueCards(discardCards)
	if len(discardCards) != 3 {
		return nil, fmt.Errorf("player must discard 3 different cards")
	}
	for _, card := range discardCards {
		if !p.HasCardInHand(card) {
			return nil, ErrCardNotInHand
		}
	}

	var events []Event
	for _, card := range discardCards {
		discardEvents, err := p.DiscardCard(card)
		events = append(events, discardEvents...)
		if err != nil {
			return nil, err
		}
	}

	ctx := &AbilityContext{
		engine:       g,
		Player:       p,
		TargetCard:   targetCard,
		TargetPlayer: targetPlayer,
	}
	abilityEvents, err := p.Hero.Ability.Execute(ctx)
	if err != nil {
		for _, c := range discardCards {
			if !slices.Contains(p.Hand, c) {
				p.Hand = append(p.Hand, c)
			}
		}
		return nil, err
	}

	events = append(events, HeroAbilityUsedEvent{ByPlayerID: p.Id})
	events = append(events, abilityEvents...)

	cardDrawnEvents, err := g.RefillPlayerHand(p)
	events = append(events, cardDrawnEvents...)
	if err != nil {
		return events, err
	}

	return events, nil
}

func (g *Game) submitPromptChoice(cmd SubmitPromptChoiceCmd) ([]Event, error) {
	if g.LevelState.Status != Playing {
		return nil, ErrGameNotPlaying
	}
	if g.LevelState.PendingInteraction == nil {
		return nil, ErrNoPendingInteraction
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	var targetPlayer *Player
	if cmd.TargetPlayerID != "" {
		targetPlayer, err = g.PlayerByID(cmd.TargetPlayerID)
		if err != nil {
			return nil, err
		}
	}
	cards, err := g.PlayerCardsByIDs(cmd.CardIDs)
	if err != nil {
		return nil, err
	}
	var targetArtifact *ArtifactCard
	if cmd.TargetArtifactID != 0 {
		targetArtifact, err = g.ArtifactByID(cmd.TargetArtifactID)
		if err != nil {
			return nil, err
		}
	}

	var events []Event
	var resolveEventFunc func(Context) ([]Event, error)
	var eventCard DungeonCard
	switch i := g.LevelState.PendingInteraction.(type) {
	case *TeamChoicePlayerInteraction:
		if targetPlayer == nil {
			return nil, ErrExpectedTarget
		}
		i.CollectedChoices[p] = targetPlayer
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		eventCard = i.Card
		resolveEventFunc = i.OnComplete

	case *PlayerDonatesHandInteraction:
		if targetPlayer == nil {
			return nil, ErrExpectedTarget
		}
		i.CollectedChoices[p] = targetPlayer
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		eventCard = i.Card
		resolveEventFunc = i.OnComplete

	case *PlayerDiscardCardsInteraction:
		discardCards := uniqueCards(cards)
		expectedCount := min(i.requiredCounts[p], len(p.Hand))
		if len(discardCards) != expectedCount {
			return nil, fmt.Errorf("expected %d cards to discard, got %d", expectedCount, len(discardCards))
		}
		for _, c := range discardCards {
			if !p.HasCardInHand(c) {
				return nil, ErrCardNotInHand
			}
		}
		i.CollectedChoices[p] = discardCards
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		eventCard = i.Card
		resolveEventFunc = i.OnComplete

	case *TeamChoiceResourceInteraction:
		if cmd.Resource == NoResource {
			return nil, ErrExpectedTarget
		}
		i.CollectedChoices[p] = cmd.Resource
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		eventCard = i.Card
		resolveEventFunc = i.OnComplete

	case *TeamChoiceArtifactInteraction:
		if targetArtifact == nil {
			return nil, ErrExpectedTarget
		}
		i.CollectedChoices[p] = targetArtifact
		delete(i.PendingPlayers, p)
		events = append(events, PlayerEventChoiceSubmittedEvent{PlayerID: p.Id})
		if len(i.PendingPlayers) != 0 {
			return events, nil
		}
		eventCard = i.Card
		resolveEventFunc = i.OnComplete

	default:
		return nil, fmt.Errorf("unexpected interaction type: %v", i)
	}

	ctx := &CardEventContext{
		engine: g,
		Card:   eventCard,
		Input:  g.LevelState.PendingInteraction,
	}

	return g.finalizeEventInteraction(ctx, resolveEventFunc)
}

func (g *Game) finalizeEventInteraction(ctx *CardEventContext, onComplete func(Context) ([]Event, error)) ([]Event, error) {
	actionEvents, err := onComplete(ctx)
	if err != nil {
		return actionEvents, err
	}

	for _, p := range g.Players {
		drawEvents, err := g.RefillPlayerHand(p)
		actionEvents = append(actionEvents, drawEvents...)
		if err != nil {
			return actionEvents, err
		}
	}

	defeatEvents, err := g.DefeatDoor(ctx.Card)
	if err != nil {
		return append(actionEvents, defeatEvents...), err
	}
	g.LevelState.PendingInteraction = nil

	return append(actionEvents, defeatEvents...), nil
}

func (g *Game) useArtifact(cmd UseArtifactCmd) ([]Event, error) {
	if g.LevelState.Status != Playing {
		return nil, ErrGameNotPlaying
	}
	if !g.Config.UseExtension {
		return nil, ErrGameExtensionDisabled
	}
	if g.LevelState.PendingInteraction != nil {
		return nil, ErrPendingInteraction
	}
	p, err := g.PlayerByID(cmd.PlayerID)
	if err != nil {
		return nil, err
	}
	var target DungeonCard
	if cmd.TargetID != 0 {
		target, err = g.DungeonCardByID(cmd.TargetID)
		if err != nil {
			return nil, err
		}
	}
	artifact, err := g.ArtifactByID(cmd.ArtifactID)
	if err != nil {
		return nil, err
	}
	if !g.LevelState.Playfield.ArtifactCanBePlayed(artifact) {
		return nil, ErrArtifactAlreadyUsed
	}

	ctx := ArtifactActionContext{
		engine:       g,
		Player:       p,
		Artifact:     artifact,
		ChosenAction: cmd.ActionIndex,
		Target:       target,
	}
	artifactEvents, err := artifact.Action.Execute(ctx)
	if err != nil {
		return nil, err
	}
	artifact.Used = true

	events := []Event{ArtifactUsedEvent{ByPlayerID: p.Id}}

	return append(events, artifactEvents...), nil
}
