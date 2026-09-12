package game

import (
	"testing"

	"github.com/gookit/goutil/dump"
)

func TestNewGame(t *testing.T) {
	game := NewGame()

	if game == nil {
		t.Fatal("expected game to be created")
	}

	if len(game.Players) != 0 {
		t.Errorf("expected 0 players, got %d", len(game.Players))
	}
}

func TestGameAddPlayer(t *testing.T) {
	game := NewGame()
	player := NewPlayer("Tanguy", Paladin())

	game.AddPlayer(player)

	if len(game.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(game.Players))
	}
	if game.Players[0] != player {
		t.Error("expected added player to be in the game")
	}
}

func TestGameAddPlayersAndStart(t *testing.T) {
	game := NewGame()
	player1 := NewPlayer(
		"Player1",
		Paladin(),
	)
	player2 := NewPlayer(
		"Player2",
		Barbarian(),
	)

	game.AddPlayer(player1)
	game.AddPlayer(player2)

	err := game.Start()

	if err != nil {
		t.Error("expected game to start successfully")
	}

	if (len(player1.Hand) == 0) || (len(player2.Hand) == 0) {
		t.Error("player1 and player2 should have cards")
	}
}

func TestGameCannotStartWithLessThanTwoPlayers(t *testing.T) {
	game := NewGame()

	err := game.Start()

	if err == nil {
		t.Error("expected game not to start with less than two players")
	}
}

func TestGameCannotStartWithMoreThanSixPlayers(t *testing.T) {
	game := NewGame()
	for range 7 {
		game.AddPlayer(NewPlayer(
			"Player",
			Paladin(),
		))
	}

	err := game.Start()

	if err == nil {
		t.Error("expected game not to start with more than six players")
	}
}

func TestGameCannotStartTwice(t *testing.T) {
	game := NewGame()
	game.AddPlayer(NewPlayer(
		"Player1",
		Paladin(),
	))
	game.AddPlayer(NewPlayer(
		"Player2",
		Barbarian(),
	))

	err := game.Start()
	if err != nil {
		t.Error("expected game to start successfully")
	}

	err = game.Start()
	if err == nil {
		t.Error("expected game not to start twice")
	}
}

func TestSimpleTurn(t *testing.T) {
	paladin := NewPlayer(
		"Paladin",
		Paladin(),
	)
	barbarian := NewPlayer(
		"Barbarian",
		Barbarian(),
	)
	game := &Game{
		Players: []*Player{paladin, barbarian},
		Dungeon: &Dungeon{
			Boss: nil,
			doors: []DungeonCard{
				DoorCard{
					Type:      Monster,
					Name:      "Goblin",
					Resources: []ResourceType{Sword, Shield},
				},
			},
		},
		Status: Waiting,
	}

	if err := game.Start(); err != nil {
		t.Error("expected game to start successfully")
	}
	dump.P(game.PlayedCards)
	game.PlayCard(paladin, paladin.Hand[0])
	game.PlayCard(barbarian, barbarian.Hand[0])
	dump.P(game.PlayedCards)
}
