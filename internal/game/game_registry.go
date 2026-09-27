package game

func (g *Game) registerCards() {
	if g.LevelState.DungeonCardsMap == nil {
		g.LevelState.DungeonCardsMap = make(map[CardID]DungeonCard, 64)
	}
	if g.LevelState.PlayerCardsMap == nil {
		g.LevelState.PlayerCardsMap = make(map[CardID]PlayerCard, 256)
	}
	if g.LevelState.Status == Playing {
		return
	}

	for _, c := range append(g.LevelState.Dungeon.Doors, g.LevelState.Dungeon.Boss) {
		g.LevelState.DungeonCardsMap[c.ID()] = c
	}

	for _, p := range g.Players {
		for _, c := range p.Deck.Cards {
			g.LevelState.PlayerCardsMap[c.ID()] = c
		}
	}
}

func (g *Game) PlayerByID(id PlayerID) (*Player, error) {
	for _, player := range g.Players {
		if player.Id != id {
			continue
		}

		return player, nil
	}

	return nil, ErrPlayerNotFound
}

func (g *Game) PlayersByIDs(ids []PlayerID) ([]*Player, error) {
	players := make([]*Player, 0, len(ids))
	for _, id := range ids {
		p, err := g.PlayerByID(id)
		if err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, nil
}
func (g *Game) PlayerCardByID(id CardID) (PlayerCard, error) {
	if card, ok := g.LevelState.PlayerCardsMap[id]; ok {
		return card, nil
	}

	return nil, ErrCardNotFound
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
	if card, ok := g.LevelState.DungeonCardsMap[id]; ok {
		return card, nil
	}

	return nil, ErrCardNotFound
}

func (g *Game) ArtifactByID(id ArtifactID) (*ArtifactCard, error) {
	for _, card := range g.LevelState.Playfield.Artifacts {
		if card.ID() != id {
			continue
		}

		return card, nil
	}

	return nil, ErrCardNotFound
}

func (g *Game) ClearStopTimeCurse() {
	g.LevelState.CurseExpectingDiscards = make(map[*Player]int)
}
