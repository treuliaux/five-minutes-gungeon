package game

type Dungeon struct {
	Boss  *BossMat
	Doors []DungeonCard
}

func NewDungeon() *Dungeon {
	return &Dungeon{
		Boss: &BossMat{
			Name: "Piti Amenou",
			Resources: []ResourceType{
				Sword,
				Shield,
				Scroll,
				Scroll,
			},
		},
		Doors: []DungeonCard{
			&DoorCard{
				Type:      DoorMonster,
				Name:      "Goblin",
				Resources: []ResourceType{Sword, Shield},
			},
		},
	}
}

func (d *Dungeon) OpenDoor() DungeonCard {
	if len(d.Doors) == 0 {
		return d.RevealBoss()
	}
	drawnCard, doors := d.Doors[0], d.Doors[1:]
	d.Doors = doors

	return drawnCard
}

func (d *Dungeon) RevealBoss() DungeonCard {
	return d.Boss
}
