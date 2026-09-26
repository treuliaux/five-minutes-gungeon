package game

func registerCardsInTestGame(g *Game, extraCards ...any) {
	if g.PlayerCardsMap == nil {
		g.PlayerCardsMap = make(map[CardID]PlayerCard, 64)
	}
	if g.DungeonCardsMap == nil {
		g.DungeonCardsMap = make(map[CardID]DungeonCard, 32)
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
		g.PlayerCardsMap[c.ID()] = c
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
		g.DungeonCardsMap[c.ID()] = c
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
	if g.PlayField != nil {
		for _, c := range g.PlayField.OpenedDoors {
			assignDungeonCard(c)
		}
		for _, c := range g.PlayField.ActiveCurses {
			assignDungeonCard(c)
		}
		for _, c := range g.PlayField.Field {
			assignPlayerCard(c)
		}
	}
	if g.Dungeon != nil {
		if g.Dungeon.Boss != nil {
			assignDungeonCard(g.Dungeon.Boss)
		}
		for _, c := range g.Dungeon.Doors {
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
