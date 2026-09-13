package game

type Dungeon struct {
	boss  *BossMat
	doors []DungeonCard
}

func NewDungeon() *Dungeon {
	return &Dungeon{
		boss: &BossMat{
			name: "Piti Amenou",
			resources: []ResourceType{
				Sword,
				Shield,
				Scroll,
				Scroll,
			},
		},
		doors: []DungeonCard{
			&DoorCard{
				Type:      DoorMonster,
				name:      "Goblin",
				resources: []ResourceType{Sword, Shield},
			},
		},
	}
}

func (d *Dungeon) OpenDoor() DungeonCard {
	if len(d.doors) == 0 {
		return d.revealBoss()
	}
	drawnCard, doors := d.doors[0], d.doors[1:]
	d.doors = doors

	return drawnCard
}

func (d *Dungeon) revealBoss() DungeonCard {
	return d.boss
}
