package game

import "slices"

type DoorsFilter struct {
	DoorKinds      []DoorKind
	ChallengeKinds []ChallengeKind
	IncludeBossMat bool
}

func NewDoorsFilter() *DoorsFilter {
	return &DoorsFilter{}
}
func (d *DoorsFilter) AddEverything() *DoorsFilter {
	d.IncludeBossMat = true
	d.ChallengeKinds = []ChallengeKind{ChallengeEvent, ChallengeCurse, ChallengeMiniBoss}
	d.DoorKinds = []DoorKind{DoorObstacle, DoorPerson, DoorMonster}

	return d
}
func (d *DoorsFilter) AddAllDoors() *DoorsFilter {
	d.DoorKinds = []DoorKind{DoorObstacle, DoorPerson, DoorMonster}

	return d
}
func (d *DoorsFilter) AddDoors(doorKinds ...DoorKind) *DoorsFilter {
	for _, dk := range doorKinds {
		if !slices.Contains(d.DoorKinds, dk) {
			d.DoorKinds = append(d.DoorKinds, dk)
		}
	}

	return d
}
func (d *DoorsFilter) AddCurses() *DoorsFilter {
	if !slices.Contains(d.ChallengeKinds, ChallengeCurse) {
		d.ChallengeKinds = append(d.ChallengeKinds, ChallengeCurse)
	}

	return d
}
func (d *DoorsFilter) AddEvents() *DoorsFilter {
	if !slices.Contains(d.ChallengeKinds, ChallengeEvent) {
		d.ChallengeKinds = append(d.ChallengeKinds, ChallengeEvent)
	}

	return d
}
func (d *DoorsFilter) AddMiniBoss() *DoorsFilter {
	if !slices.Contains(d.ChallengeKinds, ChallengeMiniBoss) {
		d.ChallengeKinds = append(d.ChallengeKinds, ChallengeMiniBoss)
	}

	return d
}
