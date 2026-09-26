package game

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"sync/atomic"
)

var globalCardIdIdx atomic.Uint32

func nextCardId() CardID {
	return CardID(globalCardIdIdx.Add(1))
}

func resetGlobalCardIndex() {
	globalCardIdIdx.Store(0)
}

func (g *Game) PlayerByID(id PlayerID) (*Player, error) {
	for _, player := range g.Players {
		if player.Id != id {
			continue
		}

		return player, nil
	}

	return nil, fmt.Errorf("player '%s' not found", id)
}

func (g *Game) PlayersByIDs(ids []PlayerID) ([]*Player, error) {
	players := make([]*Player, len(ids))
	idx := 0
	for _, player := range g.Players {
		if !slices.Contains(ids, player.Id) {
			continue
		}
		players[idx] = player
		idx++
	}
	if len(players) != len(ids) {
		return nil, fmt.Errorf("not all players were found")
	}

	return players, nil
}
func (g *Game) PlayerCardByID(id CardID) (PlayerCard, error) {
	if card, ok := g.PlayerCardsMap[id]; ok {
		return card, nil
	}

	return nil, fmt.Errorf("card with id '%v' not found", id)
}
func (g *Game) PlayerCardsByIDs(ids []CardID) ([]PlayerCard, error) {
	cards := make([]PlayerCard, len(ids))
	for i, id := range ids {
		c, err := g.PlayerCardByID(id)
		if err != nil {
			return nil, err
		}
		cards[i] = c
	}

	return cards, nil
}
func (g *Game) DungeonCardByID(id CardID) (DungeonCard, error) {
	if card, ok := g.DungeonCardsMap[id]; ok {
		return card, nil
	}

	return nil, fmt.Errorf("card with id '%v' not found", id)
}
func (g *Game) ArtifactByID(id ArtifactID) (*ArtifactCard, error) {
	for _, card := range g.PlayField.Artifacts {
		if card.ID() != id {
			continue
		}

		return card, nil
	}

	return nil, fmt.Errorf("artifact with id '%v' not found", id)
}

func pluckCardIDs[T IdentifiableCard](cards []T) []CardID {
	cardIDs := make([]CardID, len(cards))
	for i, c := range cards {
		cardIDs[i] = c.ID()
	}

	return cardIDs
}

func defeatDoorByFilter(engine GameEngine, source any, target DungeonCard, filter *DoorsFilter) ([]Event, error) {
	if target != nil {
		if !slices.Contains(engine.ActiveDoors(filter), target) {
			return nil, fmt.Errorf("invalid target")
		}
	}
	if target == nil {
		var err error
		target, err = smartTargeting(source, engine.ActiveDoors(filter))
		if err != nil {
			return nil, err
		}
	}
	if _, ok := target.(*BossMat); ok && !filter.IncludeBossMat {
		return nil, fmt.Errorf("boss mat cannot be targeted")
	}

	return engine.DefeatDoor(target)
}

func smartTargeting[T any](source any, candidates []T) (T, error) {
	var zero T
	switch len(candidates) {
	case 0:
		return zero, fmt.Errorf("no candidate found")
	case 1:
		return candidates[0], nil
	default:
		return zero, &AmbiguousTargetError[T]{
			Source:       source,
			ValidTargets: candidates,
		}
	}
}

func smartTwoPlayersTargeting(ctx *CardActionContext, targets []*Player) ([]*Player, error) {
	if len(targets) > 2 {
		return nil, fmt.Errorf("cannot target more than 2 players")
	}
	if len(targets) == 0 {
		candidates := ctx.Engine().ListPlayers()
		switch len(candidates) {
		case 0:
			return nil, fmt.Errorf("no valid target player")
		case 1, 2:
			targets = candidates
		default:
			return nil, &AmbiguousTargetError[*Player]{
				Source:       ctx.Card,
				ValidTargets: candidates,
			}
		}
	}

	return targets, nil
}

func otherPlayers(allPlayers []*Player, omit *Player) []*Player {
	result := make([]*Player, 0, len(allPlayers))
	for _, p := range allPlayers {
		if p == omit {
			continue
		}
		result = append(result, p)
	}

	return result
}

func shuffleCards[T any](cards []T) {
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}
