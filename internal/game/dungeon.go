package game

import (
	"slices"
)

type Dungeon struct {
	Boss     *BossMat
	Doors    []DungeonCard
	bossList []*BossMat
}

func NewBaseDungeon(lvl int, nbPlayers int) *Dungeon {
	bossList := BossList()
	if lvl > len(bossList) {
		lvl = len(bossList)
	}
	if lvl < 1 {
		lvl = 1
	}
	bossMat := bossList[lvl-1]

	d := &Dungeon{
		Boss:     bossMat,
		Doors:    nil,
		bossList: bossList,
	}
	d.Doors = d.buildDungeonDeck(nbPlayers, false)

	return d
}

func NewExtensionDungeon(lvl int, nbPlayers int) *Dungeon {
	bossList := BossList()
	if lvl > len(bossList) {
		lvl = len(bossList)
	}
	if lvl < 1 {
		lvl = 1
	}
	bossMat := bossList[lvl-1]

	d := &Dungeon{
		Boss:     bossMat,
		Doors:    nil,
		bossList: bossList,
	}
	d.Doors = d.buildDungeonDeck(nbPlayers, true)

	return d
}

func (d *Dungeon) OpenDoor() DungeonCard {
	if len(d.Doors) == 0 {
		return d.Boss
	}
	drawnCard := d.Doors[len(d.Doors)-1]
	d.Doors = d.Doors[:len(d.Doors)-1]

	return drawnCard
}

func (d *Dungeon) PutDoorBelowDeck(card DungeonCard) {
	switch card.(type) {
	case *BossMat, *EventCard:
	default:
		d.Doors = slices.Insert(d.Doors, 0, card)
	}
}

func (d *Dungeon) buildDungeonDeck(nbPlayer int, includeExtension bool) []DungeonCard {
	if len(d.bossList) == 0 {
		d.bossList = BossList()
	}
	dungeonDeck := make([]DungeonCard, 0, d.Boss.DeckSize+d.Boss.AdditionalChallenges+nbPlayer*2)
	if includeExtension {
		bossAbilities := slices.Clone(d.Boss.SpecialAbilities)
		if len(bossAbilities) == 0 {
			bosses := slices.Clone(d.bossList[:])
			for i := range 5 {
				shuffleCards(bosses[i].SpecialAbilities)
				bossAbilities = append(bossAbilities, bosses[i].SpecialAbilities[0])
			}
		}
		dungeonDeck = append(dungeonDeck, bossAbilities...)
	}

	doorsList := DungeonDoorCards()
	shuffleCards(doorsList)
	doorsToAdd := d.Boss.DeckSize - len(dungeonDeck)
	dungeonDeck = append(dungeonDeck, doorsList[:doorsToAdd]...)

	var challengeList []DungeonCard
	for _, c := range DungeonChallengeCards() {
		if !includeExtension {
			if mb, ok := c.(*MiniBossCard); ok && mb.Extension {
				continue
			}
			if ev, ok := c.(*EventCard); ok && ev.Extension {
				continue
			}
		}
		challengeList = append(challengeList, c)
	}
	shuffleCards(challengeList)
	if d.Boss.AdditionalChallenges > 0 {
		if d.Boss.AdditionalChallenges > len(challengeList) {
			d.Boss.AdditionalChallenges = len(challengeList)
		}
		var additionalChallenges []DungeonCard
		challengeList, additionalChallenges = challengeList[d.Boss.AdditionalChallenges:], challengeList[:d.Boss.AdditionalChallenges]
		dungeonDeck = append(dungeonDeck, additionalChallenges...)
	}
	challengesToAdd := min(nbPlayer*2, len(challengeList))
	dungeonDeck = append(dungeonDeck, challengeList[:challengesToAdd]...)

	shuffleCards(dungeonDeck)

	return dungeonDeck
}

func BossList() []*BossMat {
	return []*BossMat{
		{
			Id:        nextCardId(),
			Name:      "Baby Barbarian",
			Resources: []ResourceType{Sword, Sword, Arrow, Arrow, Jump, Jump, Jump},
			DeckSize:  20,
			SpecialAbilities: []DungeonCard{
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Poisoned Milk",
					Description: "The player(s) with the most cards in your [Hand]: [Discard] your [Hand]",
					Action:      &PoisonedMilkEvent{},
					Extension:   true,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Cursed Blanket",
					Description: "You can only use one [Hand]",
					Effect:      PlayersCanOnlyUseOneHandToPlay,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Cursed Blocks",
					Description: "You cannot have more than 3 cards in your [Hand]",
					Effect:      HandSizeLimitedToThree,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Cursed Blocks",
					Description: "You cannot have more than 3 cards in your [Hand]",
					Effect:      HandSizeLimitedToThree,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Rattle of Time",
					Description: "The [Timer] cannot be paused",
					Effect:      TimeCannotBeStopped,
				},
			},
		},
		{
			Id:        nextCardId(),
			Name:      "The Grime Reaper",
			Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Scroll, Scroll, Scroll, Shield, Shield, Shield},
			DeckSize:  25,
			SpecialAbilities: []DungeonCard{
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Acid Polish",
					Description: "All players: [Discard] all cards with 🛡️",
					Action:      &AcidPolishEvent{},
				},
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Waxed Floor",
					Description: "All players with more than 5 cards in their [Hand] must discard their entire [Hand]",
					Action:      &WaxedFloorEvent{},
				},
				&MiniBossCard{
					Id:        nextCardId(),
					Type:      ChallengeMiniBoss,
					Name:      "Reaper Jr.",
					Resources: []ResourceType{Sword, Sword, Jump, Arrow, Arrow, Arrow},
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Waffles Waffles !",
					Description: "All players must only say \"Waffles\"",
					Effect:      PlayersMustOnlySayWaffles,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Waffles Waffles !",
					Description: "All players must only say \"Waffles\"",
					Effect:      PlayersMustOnlySayWaffles,
				},
			},
		},
		{
			Id:        nextCardId(),
			Name:      "Zola the Gorgon",
			Resources: []ResourceType{Sword, Sword, Sword, Sword, Shield, Shield, Shield, Jump, Jump, Jump},
			DeckSize:  30,
			SpecialAbilities: []DungeonCard{
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Corrosive Spit",
					Description: "All players: [Discard] all cards in your [Hand] with a 🛡️ on them",
					Action:      &CorrosiveSpitEvent{},
				},
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Ensnared!",
					Description: "All players: [Discard] all cards in your [Hand] with a 🦵 on them",
					Action:      &EnsnaredEvent{},
				},
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "My Swords!",
					Description: "All players: [Discard] all cards in your [Hand] with a 🗡️ on them",
					Action:      &MySwordsEvent{},
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Cursed Cosplayers",
					Description: "All players: Flip your [Hero mat] until this [Curse] is removed",
					Effect:      FlippedHeroMat,
					Apply: func(ctx Context) ([]Event, error) {
						var events []Event
						for _, p := range ctx.Engine().ListPlayers() {
							events = append(events, p.FlipHeroMat()...)
						}
						return events, nil
					},
					Cure: func(ctx Context) ([]Event, error) {
						var events []Event
						for _, p := range ctx.Engine().ListPlayers() {
							events = append(events, p.FlipHeroMat()...)
						}
						return events, nil
					},
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Gorgon's Gaze",
					Description: "All players: You cannot use your special [Ability]",
					Effect:      AbilitiesCannotBePlayed,
				},
			},
		},
		{
			Id:        nextCardId(),
			Name:      "A Freakin' Dragon!!!",
			Resources: []ResourceType{Sword, Arrow, Arrow, Arrow, Arrow, Jump, Jump, Jump, Jump, Shield},
			DeckSize:  35,
			SpecialAbilities: []DungeonCard{
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Fire Breath",
					Description: "Choose a player: All other players must discard their [Hand]",
					Action:      &FireBreathEvent{},
				},
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Tail Swipe",
					Description: "All players: Pass your [Hand] to another player",
					Action:      &TailSwipeEvent{},
				},
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "Tail Swipe",
					Description: "All players: Pass your [Hand] to another player",
					Action:      &TailSwipeEvent{},
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Blinding Light",
					Description: "All players: You must play with your cards face down",
					Effect:      HandsHidden,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Endless Ambush",
					Description: "You must open 2 [Dungeon Cards] at a time",
					Effect:      DoorsOpenInPairs,
				},
			},
		},
		{
			Id:        nextCardId(),
			Name:      "The Dungeon Master",
			Resources: []ResourceType{Sword, Sword, Sword, Arrow, Arrow, Arrow, Shield, Shield, Shield, Scroll, Scroll, Scroll},
			DeckSize:  40,
			SpecialAbilities: []DungeonCard{
				&EventCard{
					Id:          nextCardId(),
					Type:        ChallengeEvent,
					Name:        "A 20-Sided Boulder",
					Description: "All players: [Discard] your [Hand]",
					Action:      &ATwentySidedBoulderEvent{},
				},
				&MiniBossCard{
					Id:        nextCardId(),
					Type:      ChallengeMiniBoss,
					Name:      "\"Dungeon Master\"",
					Resources: []ResourceType{Jump, Jump, Shield, Shield, Shield},
					Extension: true,
				},
				&MiniBossCard{
					Id:        nextCardId(),
					Type:      ChallengeMiniBoss,
					Name:      "The Necro-Nom-Icon",
					Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Scroll},
					Extension: true,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Clock Blocked",
					Description: "The [Timer] cannot be paused",
					Effect:      TimeCannotBeStopped,
				},
				&CurseCard{
					Id:          nextCardId(),
					Type:        ChallengeCurse,
					Name:        "Sheepified!",
					Description: "All players: You cannot play [Action] cards",
					Effect:      ActionsCannotBePlayed,
				},
			},
		},
		{
			Id:               nextCardId(),
			Name:             "The Dungeon Master (Final Form)",
			Resources:        []ResourceType{Sword, Sword, Arrow, Arrow, Shield, Shield, Scroll, Scroll, Scroll, Scroll, Scroll, Jump, Jump, Jump, Jump},
			DeckSize:         50,
			SpecialAbilities: []DungeonCard{},
		},
		{
			Id:                   nextCardId(),
			Name:                 "The K.I.C.K. 9000",
			Resources:            []ResourceType{Sword, Sword, Sword, Sword, Arrow, Arrow, Arrow, Arrow, Shield, Shield, Jump},
			DeckSize:             20,
			AdditionalChallenges: 10,
			SpecialAbilities: []DungeonCard{
				&CurseCard{Id: nextCardId(), Type: ChallengeCurse, Name: "A Boot-Alion of Kittens", Description: "", Effect: PlayersMustOnlySayMeow},
				&CurseCard{Id: nextCardId(), Type: ChallengeCurse, Name: "Conga-Rats!", Description: "", Effect: ThreeDiscardsWhenTimeStops,
					Cure: func(ctx Context) ([]Event, error) {
						ctx.Engine().ClearStopTimeCurse()

						return nil, nil
					},
				},
				&EventCard{
					Id: nextCardId(), Type: ChallengeEvent, Name: "Feeding the Trolls", Description: ""},
				&CurseCard{Id: nextCardId(), Type: ChallengeCurse, Name: "House Rules!", Description: "", Effect: PlayersHandFacingAway},
				&EventCard{
					Id: nextCardId(), Type: ChallengeEvent, Name: "The Early Bird", Description: ""},
			},
		},
	}
}

func DungeonDoorCards() []DungeonCard {
	return []DungeonCard{
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Rosetta Stone Golem",
			Resources: []ResourceType{Jump, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "An Overpriced Merchant",
			Resources: []ResourceType{Scroll, Scroll, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Wall of Spikes",
			Resources: []ResourceType{Scroll, Scroll, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Exactly 26 Ninjas",
			Resources: []ResourceType{Scroll, Jump, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "The Duck of Canterbury",
			Resources: []ResourceType{Scroll, Jump, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Just a Bunch of Stairs",
			Resources: []ResourceType{Scroll, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Fantôme Arthritique",
			Resources: []ResourceType{Scroll, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Timber-Wolf",
			Resources: []ResourceType{Sword, Sword},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Gorblin",
			Resources: []ResourceType{Sword, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Two Guys, One Bow",
			Resources: []ResourceType{Arrow, Shield, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Squire Nedward",
			Resources: []ResourceType{Shield, Shield, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Suspicious Looking Crate",
			Resources: []ResourceType{Shield, Sword, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Disappearing Blocks",
			Resources: []ResourceType{Jump, Arrow, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "One Guy, Two-Bows",
			Resources: []ResourceType{Shield, Sword, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Bottomless Pit",
			Resources: []ResourceType{Jump, Jump, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Rather Unpleasant Pheasant",
			Resources: []ResourceType{Sword, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "A Warrior Princess",
			Resources: []ResourceType{Shield, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "An Arm Dealer",
			Resources: []ResourceType{Scroll, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "A Sleeping Giant",
			Resources: []ResourceType{Sword, Sword, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "7 Unhelpful Dwarfs",
			Resources: []ResourceType{Sword, Scroll, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Chair that is Very Uncomfortable",
			Resources: []ResourceType{Sword, Jump, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Very Long Loading Screen",
			Resources: []ResourceType{Sword, Jump, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Creature of Unfathomable Evil",
			Resources: []ResourceType{Sword, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Jack the Ripper in a Box",
			Resources: []ResourceType{Shield, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Grozznak the Tall",
			Resources: []ResourceType{Jump, Shield, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Definitely Not Booby-Trapped Chest",
			Resources: []ResourceType{Jump, Shield, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Steve",
			Resources: []ResourceType{Scroll, Jump, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Sir Fuzzylumps",
			Resources: []ResourceType{Jump, Arrow, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Jacked O'Lanterns",
			Resources: []ResourceType{Arrow, Arrow, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Adorable Slime",
			Resources: []ResourceType{Jump, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Collapsed Ceiling",
			Resources: []ResourceType{Sword, Scroll, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Quicksand",
			Resources: []ResourceType{Jump, Jump, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Surprise Dodgeball Tournament",
			Resources: []ResourceType{Jump, Jump, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Gab-erwocky ...Get It? You Got It.",
			Resources: []ResourceType{Scroll, Sword, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Griffin-Door",
			Resources: []ResourceType{Arrow, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Ludicrously Tall Wall of Ice",
			Resources: []ResourceType{Jump, Jump, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Shark with Legs!!",
			Resources: []ResourceType{Sword, Arrow, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Lots and Lots of Zombies",
			Resources: []ResourceType{Sword, Sword, Sword},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "The Necrobouncer",
			Resources: []ResourceType{Sword, Scroll, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "An Overly-Dramatic Monologue",
			Resources: []ResourceType{Scroll, Arrow, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "The Chromicorn",
			Resources: []ResourceType{Jump, Sword, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Literal Strawman",
			Resources: []ResourceType{Sword, Scroll, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Massive Pauldrons",
			Resources: []ResourceType{Sword, Scroll, Sword},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Uugghh...Boots",
			Resources: []ResourceType{Sword, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Living Vines",
			Resources: []ResourceType{Scroll, Scroll, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "The Carpal Tunnel",
			Resources: []ResourceType{Scroll, Arrow, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "Invisible Wall",
			Resources: []ResourceType{Scroll, Scroll},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "Eeeewwwwwww...",
			Resources: []ResourceType{Shield, Scroll, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "A Terrible, No-Good, Awful Puppet Show",
			Resources: []ResourceType{Shield, Scroll, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "Barber-Arian",
			Resources: []ResourceType{Sword, Sword, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "A \"Ghost\"",
			Resources: []ResourceType{Sword, Sword, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A Deadly Game of Hopscotch",
			Resources: []ResourceType{Jump, Arrow, Jump},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorMonster,
			Name:      "A Cactus that Wants a Hug",
			Resources: []ResourceType{Shield, Shield, Shield},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorPerson,
			Name:      "A Gaggle of Screaming Children",
			Resources: []ResourceType{Sword, Shield, Arrow},
		},
		&DoorCard{
			Id:        nextCardId(),
			Type:      DoorObstacle,
			Name:      "A \"Shortcut\"",
			Resources: []ResourceType{Sword, Shield, Shield},
		},
	}
}

func DungeonChallengeCards() []DungeonCard {
	return []DungeonCard{
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "The Dreaded Tri-Bread",
			Resources: []ResourceType{Jump, Jump, Scroll, Arrow, Arrow},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "Das Boot!",
			Resources: []ResourceType{Sword, Sword, Jump, Jump, Jump},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "Giant Enemy Crab",
			Resources: []ResourceType{Arrow, Arrow, Arrow, Shield, Shield, Shield},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "The Rate King",
			Resources: []ResourceType{Jump, Jump, Jump, Sword, Sword, Sword},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "A Miniature T-Rex",
			Resources: []ResourceType{Shield, Shield, Arrow, Arrow, Sword, Sword},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "A Very Mini Mini-Boss",
			Resources: []ResourceType{Jump, Shield, Scroll},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "The Goblin King",
			Resources: []ResourceType{Scroll, Scroll, Sword, Shield, Shield},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "A Low-Tech Mech",
			Resources: []ResourceType{Sword, Sword, Arrow, Arrow, Arrow},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "A Wizard of Ill Repute",
			Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Jump, Jump},
			Extension: false,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "The Collector",
			Resources: []ResourceType{Jump, Shield, Scroll, Arrow, Sword},
			Extension: false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Yet More Spikes!",
			Description: "Choose a player to discard their [Hand]",
			Action:      &YetMoreSpikesEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Gimme a Hand!",
			Description: "All players: Pass your [Hand] to the same player",
			Action:      &GimmeAHandEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Sudden Illness",
			Description: "All players: [Discard] your [Hand]",
			Action:      &SuddenIllnessEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "A Boo-Boo",
			Description: "All players: [Discard] a card",
			Action:      &ABooBooEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Dungeon Error in Your Favor",
			Description: "All players: Draw 5 cards",
			Action:      &DungeonErrorInYourFavorEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "An Ungodly Amount of Porcupines",
			Description: "All players: Draw 3 cards, then discard 3 cards",
			Action:      &AnUngodlyAmountOfPorcupinesEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Locked Door!",
			Description: "Pick a [Resource] - All players must discard all cards with that [Resource]",
			Action:      &LockedDoorEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Confusion",
			Description: "All players: Pass your [Hand] to another player",
			Action:      &ConfusionEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Ambush!",
			Description: "Flip over the next 2 [Dungeon Cards] - You must defeat both before moving on",
			Action:      &AmbushEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Trap Door",
			Description: "All players: [Discard] 3 cards",
			Action:      &TrapDoorEvent{},
			Extension:   false,
		},
		&EventCard{
			Id:          nextCardId(),
			Type:        ChallengeEvent,
			Name:        "Crowd Funding",
			Description: "Re-activate an [Artifact] - All players: [Discard] your [Hand]",
			Action:      &CrowdFundingEvent{},
			Extension:   true,
		},
		&MiniBossCard{
			Id:        nextCardId(),
			Type:      ChallengeMiniBoss,
			Name:      "Tinkles, Destroyer of Soles",
			Resources: []ResourceType{Jump, Jump, Jump, Jump, Jump},
			Extension: true,
		},
	}
}
