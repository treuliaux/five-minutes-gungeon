package game

import (
	"fmt"
	"slices"
)

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
	var result []*Player
	for _, p := range allPlayers {
		if p == omit {
			continue
		}
		result = append(result, p)
	}

	return result
}
