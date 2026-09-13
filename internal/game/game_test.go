package game

import (
	"testing"
	"time"
)

func TestNewGame(t *testing.T) {
	game := NewGame()

	if game == nil {
		t.Fatal("expected game to be created")
	}

	if len(game.players) != 0 {
		t.Errorf("expected 0 players, got %d", len(game.players))
	}
}

func TestGameAddPlayer(t *testing.T) {
	game := NewGame()

	game.addPlayer(AddPlayerCmd{
		Name:  "Tanguy",
		Class: Paladin,
		reply: nil,
	})

	if len(game.players) != 1 {
		t.Errorf("expected 1 player, got %d", len(game.players))
	}
	if game.players[0].name != "Tanguy" {
		t.Error("expected added player to be in the game")
	}
}

func TestGameAddPlayersAndStart(t *testing.T) {
	game := NewGame()

	game.addPlayer(AddPlayerCmd{
		Name:  "Player1",
		Class: Paladin,
		reply: nil,
	})
	game.addPlayer(AddPlayerCmd{
		Name:  "Player2",
		Class: Barbarian,
		reply: nil,
	})

	err := game.start()

	if err != nil {
		t.Error("expected game to start successfully")
	}

	if (len(game.players[0].hand) == 0) || (len(game.players[1].hand) == 0) {
		t.Error("player1 and player2 should have cards")
	}
}

func TestGameCannotStartWithLessThanTwoPlayers(t *testing.T) {
	game := NewGame()

	err := game.start()

	if err == nil {
		t.Error("expected game not to start with less than two players")
	}
}

func TestGameCannotStartWithMoreThanSixPlayers(t *testing.T) {
	game := NewGame()
	for range 7 {
		game.addPlayer(AddPlayerCmd{
			Name:  "Player",
			Class: Paladin,
			reply: nil,
		})
	}

	err := game.start()

	if err == nil {
		t.Error("expected game not to start with more than six players")
	}
}

func TestGameCannotStartTwice(t *testing.T) {
	game := NewGame()
	game.addPlayer(AddPlayerCmd{
		Name:  "Player1",
		Class: Paladin,
		reply: nil,
	})
	game.addPlayer(AddPlayerCmd{
		Name:  "Player2",
		Class: Barbarian,
		reply: nil,
	})

	err := game.start()
	if err != nil {
		t.Error("expected game to start successfully")
	}

	err = game.start()
	if err == nil {
		t.Error("expected game not to start twice")
	}
}

func TestSimpleTurn(t *testing.T) {
	paladin, _ := NewPlayer(
		"Paladin",
		Paladin,
	)
	barbarian, _ := NewPlayer(
		"Barbarian",
		Barbarian,
	)
	game := &Game{
		players: []*Player{paladin, barbarian},
		dungeon: &Dungeon{
			boss: nil,
			doors: []DungeonCard{
				&DoorCard{
					Type:      DoorMonster,
					name:      "Goblin",
					resources: []ResourceType{Sword, Shield},
				},
			},
		},
		Status: Waiting,
	}

	if err := game.start(); err != nil {
		t.Error("expected game to start successfully")
	}
	if game.isPlayfieldBeaten() {
		t.Error("expected playfield NOT to be beaten")
	}
	game.playCard(paladin, paladin.hand[0])
	if game.isPlayfieldBeaten() {
		t.Error("expected playfield NOT to be beaten")
	}
	game.playCard(barbarian, barbarian.hand[0])
	if !game.isPlayfieldBeaten() {
		t.Error("expected playfield to be beaten")
	}
}

func TestCompleteCycle(t *testing.T) {
	paladin, _ := NewPlayer(
		"Paladin",
		Paladin,
	)
	barbarian, _ := NewPlayer(
		"Barbarian",
		Barbarian,
	)
	game := &Game{
		players: []*Player{paladin, barbarian},
		dungeon: NewDungeon(),
		Status:  Waiting,
	}

	_ = game.start()
	_, ok := game.currentDungeonCard.(*DoorCard)
	if !ok {
		t.Error("expected game to start with a door card")
	}
	_ = game.Tick(time.Second)
	game.playCard(paladin, paladin.hand[0])
	_ = game.Tick(time.Second)
	if game.Status != Playing {
		t.Error("expected game state to be playing")
	}
	game.playCard(barbarian, barbarian.hand[0])
	_ = game.Tick(time.Second)
	_, ok = game.currentDungeonCard.(*BossMat)
	if !ok {
		t.Error("expected game to be at the boss stage")
	}
	game.playCard(paladin, paladin.hand[0])
	_ = game.Tick(time.Second)
	if game.Status != Playing {
		t.Error("expected game state to be playing")
	}
	game.playCard(barbarian, barbarian.hand[0])
	_ = game.Tick(1)
	if game.Status != Victory {
		t.Error("expected game state to be victory")
	}
}

func TestSimpleDefeat(t *testing.T) {
	paladin, _ := NewPlayer(
		"Paladin",
		Paladin,
	)
	barbarian, _ := NewPlayer(
		"Barbarian",
		Barbarian,
	)
	game := &Game{
		players: []*Player{paladin, barbarian},
		dungeon: NewDungeon(),
		Status:  Waiting,
	}

	_ = game.start()
	_ = game.Tick(1)
	if game.Status != Playing {
		t.Error("expected game status to be playing")
	}
	_ = game.Tick(4*time.Minute + 59*time.Second)
	if game.Status != Playing {
		t.Error("expected game status to be playing")
	}
	_ = game.Tick(1 * time.Second)
	if game.Status != Defeat {
		t.Error("expected game status to be defeated")
	}
}
