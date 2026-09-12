package game

type Dungeon struct {
	Boss  *BossMat
	doors []DungeonCard
}

func NewDungeon() *Dungeon {
	return &Dungeon{
		Boss: nil,
		doors: []DungeonCard{
			&DoorCard{
				Type:      Monster,
				Name:      "Goblin",
				Resources: []ResourceType{Sword},
			},
		},
	}
}

func (d *Dungeon) OpenDoor() *DungeonCard {
	drawnCard, doors := d.doors[0], d.doors[1:]
	if drawnCard == nil {
		return nil
	}
	d.doors = doors

	return &drawnCard
}
