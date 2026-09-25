package game

import (
	"math/rand/v2"
	"slices"
)

type Dungeon struct {
	Boss  *BossMat
	Doors []DungeonCard
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

	return &Dungeon{
		Boss:  bossMat,
		Doors: buildDungeonDeck(bossMat, nbPlayers, false),
	}
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

	return &Dungeon{
		Boss:  bossMat,
		Doors: buildDungeonDeck(bossMat, nbPlayers, true),
	}
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

func buildDungeonDeck(boss *BossMat, nbPlayer int, includeExtension bool) []DungeonCard {
	dungeonDeck := make([]DungeonCard, 0, boss.DeckSize+boss.AdditionalChallenges+nbPlayer*2)
	if includeExtension {
		bossAbilities := slices.Clone(boss.SpecialAbilities)
		if len(bossAbilities) == 0 {
			bosses := BossList()
			for i := range 5 {
				shuffleCards(bosses[i].SpecialAbilities)
				bossAbilities = append(bossAbilities, bosses[i].SpecialAbilities[0])
			}
		}
		dungeonDeck = append(dungeonDeck, bossAbilities...)
	}

	doorsList := DungeonDoorCards()
	shuffleCards(doorsList)
	doorsToAdd := boss.DeckSize - len(dungeonDeck)
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
	if boss.AdditionalChallenges > 0 {
		if boss.AdditionalChallenges > len(challengeList) {
			boss.AdditionalChallenges = len(challengeList)
		}
		var additionalChallenges []DungeonCard
		challengeList, additionalChallenges = challengeList[boss.AdditionalChallenges:], challengeList[:boss.AdditionalChallenges]
		dungeonDeck = append(dungeonDeck, additionalChallenges...)
	}
	challengesToAdd := min(nbPlayer*2, len(challengeList))
	dungeonDeck = append(dungeonDeck, challengeList[:challengesToAdd]...)

	shuffleCards(dungeonDeck)

	return dungeonDeck
}

func BossList() [7]*BossMat {
	return [7]*BossMat{
		{
			Name:      "Baby Barbarian",
			Resources: []ResourceType{Sword, Sword, Arrow, Arrow, Jump, Jump, Jump},
			DeckSize:  20,
			SpecialAbilities: []DungeonCard{
				&EventCard{Type: ChallengeEvent, Name: "Poisoned Milk", Action: &PoisonedMilkEvent{}, Extension: true},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Blanket", Effect: PlayersCanOnlyUseOneHandToPlay},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Blocks", Effect: HandSizeLimitedToThree},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Blocks", Effect: HandSizeLimitedToThree},
				&CurseCard{Type: ChallengeCurse, Name: "Rattle of Time", Effect: TimeCannotBeStopped},
			},
		},
		{
			Name:      "The Grime Reaper",
			Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Scroll, Scroll, Scroll, Shield, Shield, Shield},
			DeckSize:  25,
			SpecialAbilities: []DungeonCard{
				&EventCard{Type: ChallengeEvent, Name: "Acid Polish", Action: &AcidPolishEvent{}},
				&EventCard{Type: ChallengeEvent, Name: "Waxed Floor", Action: &WaxedFloorEvent{}},
				&MiniBossCard{Type: ChallengeMiniBoss, Name: "Reaper Jr.", Resources: []ResourceType{Sword, Sword, Jump, Arrow, Arrow, Arrow}},
				&CurseCard{Type: ChallengeCurse, Name: "Waffles Waffles !", Effect: PlayersMustOnlySayWaffles},
				&CurseCard{Type: ChallengeCurse, Name: "Waffles Waffles !", Effect: PlayersMustOnlySayWaffles},
			},
		},
		{
			Name:      "Zola the Gorgon",
			Resources: []ResourceType{Sword, Sword, Sword, Sword, Shield, Shield, Shield, Jump, Jump, Jump},
			DeckSize:  30,
			SpecialAbilities: []DungeonCard{
				&EventCard{Type: ChallengeEvent, Name: "Corrosive Spit", Action: &CorrosiveSpitEvent{}},
				&EventCard{Type: ChallengeEvent, Name: "Ensnared!", Action: &EnsnaredEvent{}},
				&EventCard{Type: ChallengeEvent, Name: "My Swords!", Action: &MySwordsEvent{}},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Cosplayers", Effect: FlippedHeroMat,
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
				&CurseCard{Type: ChallengeCurse, Name: "Gorgon's Gaze", Effect: AbilitiesCannotBePlayed},
			},
		},
		{
			Name:      "A Freakin' Dragon!!!",
			Resources: []ResourceType{Sword, Arrow, Arrow, Arrow, Arrow, Jump, Jump, Jump, Jump, Shield},
			DeckSize:  35,
			SpecialAbilities: []DungeonCard{
				&EventCard{Type: ChallengeEvent, Name: "Fire Breath", Action: &FireBreathEvent{}},
				&EventCard{Type: ChallengeEvent, Name: "Tail Swipe", Action: &TailSwipeEvent{}},
				&EventCard{Type: ChallengeEvent, Name: "Tail Swipe", Action: &TailSwipeEvent{}},
				&CurseCard{Type: ChallengeCurse, Name: "Blinding Light", Effect: HandsHidden},
				&CurseCard{Type: ChallengeCurse, Name: "Endless Ambush", Effect: DoorsOpenInPairs},
			},
		},
		{
			Name:      "The Dungeon Master",
			Resources: []ResourceType{Sword, Sword, Sword, Arrow, Arrow, Arrow, Shield, Shield, Shield, Scroll, Scroll, Scroll},
			DeckSize:  40,
			SpecialAbilities: []DungeonCard{
				&EventCard{Type: ChallengeEvent, Name: "A 20-Sided Boulder", Action: &ATwentySidedBoulderEvent{}},
				&MiniBossCard{Type: ChallengeMiniBoss, Name: "\"Dungeon Master\"", Resources: []ResourceType{Jump, Jump, Shield, Shield, Shield}, Extension: true},
				&MiniBossCard{Type: ChallengeMiniBoss, Name: "The Necro-Nom-Icon", Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Scroll}, Extension: true},
				&CurseCard{Type: ChallengeCurse, Name: "Clock Blocked", Effect: TimeCannotBeStopped},
				&CurseCard{Type: ChallengeCurse, Name: "Sheepified!", Effect: ActionsCannotBePlayed},
			},
		},
		{
			Name:             "The Dungeon Master (Final Form)",
			Resources:        []ResourceType{Sword, Sword, Arrow, Arrow, Shield, Shield, Scroll, Scroll, Scroll, Scroll, Scroll, Jump, Jump, Jump, Jump},
			DeckSize:         50,
			SpecialAbilities: []DungeonCard{},
		},
		{
			Name:                 "The K.I.C.K. 9000",
			Resources:            []ResourceType{Sword, Sword, Sword, Sword, Arrow, Arrow, Arrow, Arrow, Shield, Shield, Jump},
			DeckSize:             20,
			AdditionalChallenges: 10,
			SpecialAbilities: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "A Boot-Alion of Kittens", Effect: PlayersMustOnlySayMeow},
				&CurseCard{Type: ChallengeCurse, Name: "Conga-Rats!", Effect: ThreeDiscardsWhenTimeStops,
					Cure: func(ctx Context) ([]Event, error) {
						ctx.Engine().ClearStopTimeCurse()

						return nil, nil
					},
				},
				&EventCard{Type: ChallengeEvent, Name: "Feeding the Trolls"},
				&CurseCard{Type: ChallengeCurse, Name: "House Rules!", Effect: PlayersHandFacingAway},
				&EventCard{Type: ChallengeEvent, Name: "The Early Bird"},
			},
		},
	}
}

func DungeonDoorCards() []DungeonCard {
	return []DungeonCard{
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Rosetta Stone Golem",
			Resources: []ResourceType{Jump, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "An Overpriced Merchant",
			Resources: []ResourceType{Scroll, Scroll, Jump},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Wall of Spikes",
			Resources: []ResourceType{Scroll, Scroll, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Exactly 26 Ninjas",
			Resources: []ResourceType{Scroll, Jump, Jump},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "The Duck of Canterbury",
			Resources: []ResourceType{Scroll, Jump, Shield},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Just a Bunch of Stairs",
			Resources: []ResourceType{Scroll, Jump},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Fantôme Arthritique",
			Resources: []ResourceType{Scroll, Shield},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Timber-Wolf",
			Resources: []ResourceType{Sword, Sword},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Gorblin",
			Resources: []ResourceType{Sword, Jump},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Two Guys, One Bow",
			Resources: []ResourceType{Arrow, Shield, Arrow},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Squire Nedward",
			Resources: []ResourceType{Shield, Shield, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Suspicious Looking Crate",
			Resources: []ResourceType{Shield, Sword, Scroll},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Disappearing Blocks",
			Resources: []ResourceType{Jump, Arrow, Scroll},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "One Guy, Two-Bows",
			Resources: []ResourceType{Shield, Sword, Shield},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Bottomless Pit",
			Resources: []ResourceType{Jump, Jump, Jump},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Rather Unpleasant Pheasant",
			Resources: []ResourceType{Sword, Scroll},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "A Warrior Princess",
			Resources: []ResourceType{Shield, Arrow},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "An Arm Dealer",
			Resources: []ResourceType{Scroll, Arrow},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "A Sleeping Giant",
			Resources: []ResourceType{Sword, Sword, Jump},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "7 Unhelpful Dwarfs",
			Resources: []ResourceType{Sword, Scroll, Shield},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Chair that is Very Uncomfortable",
			Resources: []ResourceType{Sword, Jump, Shield},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Very Long Loading Screen",
			Resources: []ResourceType{Sword, Jump, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Creature of Unfathomable Evil",
			Resources: []ResourceType{Sword, Shield},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Jack the Ripper in a Box",
			Resources: []ResourceType{Shield, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Grozznak the Tall",
			Resources: []ResourceType{Jump, Shield, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Definitely Not Booby-Trapped Chest",
			Resources: []ResourceType{Jump, Shield, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Steve",
			Resources: []ResourceType{Scroll, Jump, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Sir Fuzzylumps",
			Resources: []ResourceType{Jump, Arrow, Arrow},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Jacked O'Lanterns",
			Resources: []ResourceType{Arrow, Arrow, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Adorable Slime",
			Resources: []ResourceType{Jump, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Collapsed Ceiling",
			Resources: []ResourceType{Sword, Scroll, Scroll},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Quicksand",
			Resources: []ResourceType{Jump, Jump, Shield},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Surprise Dodgeball Tournament",
			Resources: []ResourceType{Jump, Jump, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Gab-erwocky ...Get It? You Got It.",
			Resources: []ResourceType{Scroll, Sword, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Griffin-Door",
			Resources: []ResourceType{Arrow, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Ludicrously Tall Wall of Ice",
			Resources: []ResourceType{Jump, Jump, Jump},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Shark with Legs!!",
			Resources: []ResourceType{Sword, Arrow, Arrow},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Lots and Lots of Zombies",
			Resources: []ResourceType{Sword, Sword, Sword},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "The Necrobouncer",
			Resources: []ResourceType{Sword, Scroll, Arrow},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "An Overly-Dramatic Monologue",
			Resources: []ResourceType{Scroll, Arrow, Scroll},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "The Chromicorn",
			Resources: []ResourceType{Jump, Sword, Jump},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Literal Strawman",
			Resources: []ResourceType{Sword, Scroll, Jump},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Massive Pauldrons",
			Resources: []ResourceType{Sword, Scroll, Sword},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Uugghh...Boots",
			Resources: []ResourceType{Sword, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Living Vines",
			Resources: []ResourceType{Scroll, Scroll, Scroll},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "The Carpal Tunnel",
			Resources: []ResourceType{Scroll, Arrow, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "Invisible Wall",
			Resources: []ResourceType{Scroll, Scroll},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "Eeeewwwwwww...",
			Resources: []ResourceType{Shield, Scroll, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "A Terrible, No-Good, Awful Puppet Show",
			Resources: []ResourceType{Shield, Scroll, Arrow},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "Barber-Arian",
			Resources: []ResourceType{Sword, Sword, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "A \"Ghost\"",
			Resources: []ResourceType{Sword, Sword, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A Deadly Game of Hopscotch",
			Resources: []ResourceType{Jump, Arrow, Jump},
		},
		&DoorCard{
			Type:      DoorMonster,
			Name:      "A Cactus that Wants a Hug",
			Resources: []ResourceType{Shield, Shield, Shield},
		},
		&DoorCard{
			Type:      DoorPerson,
			Name:      "A Gaggle of Screaming Children",
			Resources: []ResourceType{Sword, Shield, Arrow},
		},
		&DoorCard{
			Type:      DoorObstacle,
			Name:      "A \"Shortcut\"",
			Resources: []ResourceType{Sword, Shield, Shield},
		},
	}
}

func DungeonChallengeCards() []DungeonCard {
	return []DungeonCard{
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "The Dreaded Tri-Bread",
			Resources: []ResourceType{Jump, Jump, Scroll, Arrow, Arrow},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "Das Boot!",
			Resources: []ResourceType{Sword, Sword, Jump, Jump, Jump},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "Giant Enemy Crab",
			Resources: []ResourceType{Arrow, Arrow, Arrow, Shield, Shield, Shield},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "The Rate King",
			Resources: []ResourceType{Jump, Jump, Jump, Sword, Sword, Sword},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "A Miniature T-Rex",
			Resources: []ResourceType{Shield, Shield, Arrow, Arrow, Sword, Sword},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "A Very Mini Mini-Boss",
			Resources: []ResourceType{Jump, Shield, Scroll},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "The Goblin King",
			Resources: []ResourceType{Scroll, Scroll, Sword, Shield, Shield},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "A Low-Tech Mech",
			Resources: []ResourceType{Sword, Sword, Arrow, Arrow, Arrow},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "A Wizard of Ill Repute",
			Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Jump, Jump},
			Extension: false,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "The Collector",
			Resources: []ResourceType{Jump, Shield, Scroll, Arrow, Sword},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Yet More Spikes!",
			Action:    &YetMoreSpikesEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Gimme a Hand!",
			Action:    &GimmeAHandEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Sudden Illness",
			Action:    &SuddenIllnessEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "A Boo-Boo",
			Action:    &ABooBooEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Dungeon Error in Your Favor",
			Action:    &DungeonErrorInYourFavorEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "An Ungodly Amount of Porcupines",
			Action:    &AnUngodlyAmountOfPorcupinesEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Locked Door!",
			Action:    &LockedDoorEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Confusion",
			Action:    &ConfusionEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Ambush!",
			Action:    &AmbushEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Trap Door",
			Action:    &TrapDoorEvent{},
			Extension: false,
		},
		&EventCard{
			Type:      ChallengeEvent,
			Name:      "Crowd Funding",
			Action:    &CrowdFundingEvent{},
			Extension: true,
		},
		&MiniBossCard{
			Type:      ChallengeMiniBoss,
			Name:      "Tinkles, Destroyer of Soles",
			Resources: []ResourceType{Jump, Jump, Jump, Jump, Jump},
			Extension: true,
		},
	}
}

func shuffleCards(cards []DungeonCard) {
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}
