package game

func registerCardsInTestGame(g *Game, extraCards ...any) {
	if g.LevelState.PlayerCardsMap == nil {
		g.LevelState.PlayerCardsMap = make(map[CardID]PlayerCard, 64)
	}
	if g.LevelState.DungeonCardsMap == nil {
		g.LevelState.DungeonCardsMap = make(map[CardID]DungeonCard, 32)
	}
	var testCardID uint32 = 1000

	assignPlayerCard := func(c PlayerCard) {
		if c == nil {
			return
		}
		if c.ID() == 0 {
			testCardID++
			switch rc := c.(type) {
			case *ResourceCard:
				rc.Id = CardID(testCardID)
			case *ActionCard:
				rc.Id = CardID(testCardID)
			}
		}
		g.LevelState.PlayerCardsMap[c.ID()] = c
	}

	assignDungeonCard := func(c DungeonCard) {
		if c == nil {
			return
		}
		if c.ID() == 0 {
			testCardID++
			switch dc := c.(type) {
			case *DoorCard:
				dc.Id = CardID(testCardID)
			case *EventCard:
				dc.Id = CardID(testCardID)
			case *MiniBossCard:
				dc.Id = CardID(testCardID)
			case *CurseCard:
				dc.Id = CardID(testCardID)
			case *BossMat:
				dc.Id = CardID(testCardID)
			}
		}
		g.LevelState.DungeonCardsMap[c.ID()] = c
	}

	for _, p := range g.Players {
		if p == nil {
			continue
		}
		for _, c := range p.Hand {
			assignPlayerCard(c)
		}
		if p.Deck != nil {
			for _, c := range p.Deck.Cards {
				assignPlayerCard(c)
			}
		}
		if p.Discard != nil {
			for _, c := range p.Discard.Cards {
				assignPlayerCard(c)
			}
		}
	}
	if g.LevelState.Playfield != nil {
		for _, c := range g.LevelState.Playfield.OpenedDoors {
			assignDungeonCard(c)
		}
		for _, c := range g.LevelState.Playfield.ActiveCurses {
			assignDungeonCard(c)
		}
		for _, c := range g.LevelState.Playfield.Field {
			assignPlayerCard(c)
		}
	}
	if g.LevelState.Dungeon != nil {
		if g.LevelState.Dungeon.Boss != nil {
			assignDungeonCard(g.LevelState.Dungeon.Boss)
		}
		for _, c := range g.LevelState.Dungeon.Doors {
			assignDungeonCard(c)
		}
	}
	for _, item := range extraCards {
		switch c := item.(type) {
		case PlayerCard:
			assignPlayerCard(c)
		case DungeonCard:
			assignDungeonCard(c)
		}
	}
}
