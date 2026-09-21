package game

import (
	"math/rand/v2"
	"slices"
	"time"
)

type Dungeon struct {
	Boss  *BossMat
	Doors []DungeonCard
}

func NewDungeon(lvl int, nbPlayers int, includingExtension bool) *Dungeon {
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
		Doors: buildDungeonDeck(bossMat, nbPlayers, includingExtension),
	}
}

func (d *Dungeon) OpenDoor() DungeonCard {
	if len(d.Doors) == 0 {
		return d.RevealBoss()
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

func (d *Dungeon) RevealBoss() DungeonCard {
	return d.Boss
}

func BossList() [7]*BossMat {
	return [7]*BossMat{
		{
			Name:      "Baby Barbarian",
			Resources: []ResourceType{Sword, Sword, Arrow, Arrow, Jump, Jump, Jump},
			DeckSize:  20,
			Curses: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Blanket"},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Blocks"},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Blocks"},
				&CurseCard{Type: ChallengeCurse, Name: "Poisoned Milk"},
				&CurseCard{Type: ChallengeCurse, Name: "Rattle of Time"},
			},
		},
		{
			Name:      "The Grime Reaper",
			Resources: []ResourceType{Scroll, Scroll, Scroll, Scroll, Scroll, Scroll, Scroll, Shield, Shield, Shield},
			DeckSize:  25,
			Curses: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "Acid Polish"},
				&CurseCard{Type: ChallengeCurse, Name: "Reaper Jr."},
				&CurseCard{Type: ChallengeCurse, Name: "Waffles Waffles !"},
				&CurseCard{Type: ChallengeCurse, Name: "Waffles Waffles !"},
				&CurseCard{Type: ChallengeCurse, Name: "Waxed Floor"},
			},
		},
		{
			Name:      "Zola the Gorgon",
			Resources: []ResourceType{Sword, Sword, Sword, Sword, Shield, Shield, Shield, Jump, Jump, Jump},
			DeckSize:  30,
			Curses: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "Corrosive Spit"},
				&CurseCard{Type: ChallengeCurse, Name: "Cursed Cosplayers"},
				&CurseCard{Type: ChallengeCurse, Name: "Ensnared!"},
				&CurseCard{Type: ChallengeCurse, Name: "Gorgon's Gaze"},
				&CurseCard{Type: ChallengeCurse, Name: "My Swords!"},
			},
		},
		{
			Name:      "A Freakin' Dragon!!!",
			Resources: []ResourceType{Sword, Arrow, Arrow, Arrow, Arrow, Jump, Jump, Jump, Jump, Shield},
			DeckSize:  35,
			Curses: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "Blinding Light"},
				&CurseCard{Type: ChallengeCurse, Name: "Endless Ambush"},
				&CurseCard{Type: ChallengeCurse, Name: "Fire Breath"},
				&CurseCard{Type: ChallengeCurse, Name: "Tail Swipe"},
				&CurseCard{Type: ChallengeCurse, Name: "Tail Swipe"},
			},
		},
		{
			Name:      "The Dungeon Master",
			Resources: []ResourceType{Sword, Sword, Sword, Arrow, Arrow, Arrow, Shield, Shield, Shield, Scroll, Scroll, Scroll},
			DeckSize:  40,
			Curses: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "\"Dungeon Master\""},
				&CurseCard{Type: ChallengeCurse, Name: "A 20-Sided Boulder"},
				&CurseCard{Type: ChallengeCurse, Name: "Clock Blocked"},
				&CurseCard{Type: ChallengeCurse, Name: "Sheepified!"},
				&CurseCard{Type: ChallengeCurse, Name: "The Necro-Nom-Icon"},
			},
		},
		{
			Name:                 "The K.I.C.K. 9000",
			Resources:            []ResourceType{Sword, Sword, Sword, Sword, Arrow, Arrow, Arrow, Arrow, Shield, Shield, Jump},
			DeckSize:             20,
			AdditionalChallenges: 10,
			Curses: []DungeonCard{
				&CurseCard{Type: ChallengeCurse, Name: "A Boot-Alion of Kittens"},
				&CurseCard{Type: ChallengeCurse, Name: "Conga-Rats!"},
				&CurseCard{Type: ChallengeCurse, Name: "Feeding the Trolls"},
				&CurseCard{Type: ChallengeCurse, Name: "House Rules!"},
				&CurseCard{Type: ChallengeCurse, Name: "The Early Bird"},
			},
		},
		{
			Name:      "The Dungeon Master (Final Form)",
			Resources: []ResourceType{Sword, Sword, Arrow, Arrow, Shield, Shield, Scroll, Scroll, Scroll, Scroll, Scroll, Jump, Jump, Jump, Jump},
			DeckSize:  50,
			Curses:    []DungeonCard{},
		},
	}
}

func buildDungeonDeck(boss *BossMat, nbPlayer int, includingExtension bool) []DungeonCard {
	var dungeonDeck []DungeonCard
	if includingExtension {
		dungeonDeck = append(dungeonDeck, boss.Curses...)
	}

	doorsList := DungeonDoorCards()
	shuffleCards(doorsList)
	doorsToAdd := boss.DeckSize - len(dungeonDeck)
	dungeonDeck = append(dungeonDeck, doorsList[:doorsToAdd]...)

	challengeList := DungeonChallengeCards()
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
			Type:       ChallengeEvent,
			Name:       "Yet More Spikes!",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Gimme a Hand!",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Sudden Illness",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "A Boo-Boo",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Dungeon Error in Your Favor",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "An Ungodly Amount of Porcupines",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Locked Door!",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Confusion",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Ambush!",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Trap Door",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  false,
		},
		&EventCard{
			Type:       ChallengeEvent,
			Name:       "Crowd Funding",
			Action:     nil,
			OpenedTime: 2 * 5 * time.Minute,
			Extension:  true,
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
