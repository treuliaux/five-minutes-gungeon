package game

import (
	"slices"
	"time"
)

type GameSnapshotDTO struct {
	Status              Status                 `json:"status"`
	InGameTimer         time.Duration          `json:"inGameTimer"`
	TotalDuration       time.Duration          `json:"totalDuration"`
	IsTimeFrozen        bool                   `json:"isTimeFrozen"`
	IsFightingBoss      bool                   `json:"isFightingBoss"`
	RemainingDoorsCount int                    `json:"remainingDoorsCount"`
	Boss                *DungeonCardDTO        `json:"boss,omitempty"`
	Players             []PlayerDTO            `json:"players"`
	OpenedDoors         []DungeonCardDTO       `json:"openedDoors"`
	ActiveCurses        []DungeonCardDTO       `json:"activeCurses"`
	PlayedField         []PlayerCardDTO        `json:"playedField"`
	Artifacts           []ArtifactDTO          `json:"artifacts"`
	PendingInteraction  *PendingInteractionDTO `json:"pendingInteraction,omitempty"`
}

func NewGameSnapshot(g *Game) GameSnapshotDTO {
	var boss *DungeonCardDTO
	var remainingDoors int
	if g.Dungeon != nil {
		if g.Dungeon.Boss != nil {
			b := DungeonCardToDTO(g.Dungeon.Boss)
			boss = &b
		}
		remainingDoors = len(g.Dungeon.Doors)
	}

	var openedDoors []DungeonCardDTO
	var activeCurses []DungeonCardDTO
	var playedField []PlayerCardDTO
	var artifacts []ArtifactDTO
	if g.PlayField != nil {
		openedDoors = DungeonCardsToDTO(g.PlayField.OpenedDoors)
		activeCurses = DungeonCardsToDTO(g.PlayField.ActiveCurses)
		playedField = PlayerCardsToDTO(g.PlayField.Field)
		artifacts = ArtifactsToDTO(g.PlayField.Artifacts)
	}

	return GameSnapshotDTO{
		Status:              g.Status,
		InGameTimer:         g.InGameTimer,
		TotalDuration:       g.RealElapsedTime,
		IsTimeFrozen:        g.IsTimeFrozen,
		IsFightingBoss:      g.IsFightingBoss,
		RemainingDoorsCount: remainingDoors,
		Boss:                boss,
		Players:             PlayersToDTO(g.Players),
		OpenedDoors:         openedDoors,
		ActiveCurses:        activeCurses,
		PlayedField:         playedField,
		Artifacts:           artifacts,
		PendingInteraction:  PendingInteractionToDTO(g.PendingInteraction),
	}
}

type PlayerDTO struct {
	Id        PlayerID        `json:"id"`
	Name      string          `json:"name"`
	HeroClass HeroClass       `json:"heroClass"`
	HeroName  string          `json:"heroName"`
	DeckCount int             `json:"deckCount"`
	Hand      []PlayerCardDTO `json:"hand"`
	Discard   []PlayerCardDTO `json:"discard"`
}

type DungeonCardDTO struct {
	Id          CardID          `json:"id"`
	Kind        DungeonCardKind `json:"kind"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Resources   []ResourceType  `json:"resources"`
}

type PlayerCardDTO struct {
	Id          CardID         `json:"id"`
	Kind        PlayerCardKind `json:"kind"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Resources   []ResourceType `json:"resources"`
}

type ArtifactDTO struct {
	Id           ArtifactID `json:"id"`
	Color        DeckColor  `json:"color"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	SelectAction bool       `json:"selectAction"`
	Used         bool       `json:"used"`
}

type PendingInteractionDTO struct {
	Kind           InteractionKind  `json:"kind"`
	SourceCardID   CardID           `json:"sourceCardId,omitempty"`
	RequiredCounts map[PlayerID]int `json:"requiredCounts,omitempty"`
	PendingPlayers []PlayerID       `json:"pendingPlayers,omitempty"`
}

func PlayersToDTO(players []*Player) []PlayerDTO {
	playersDTO := make([]PlayerDTO, len(players))
	for i, p := range players {
		playersDTO[i] = PlayerToDTO(p)
	}

	return playersDTO
}

func PlayerToDTO(p *Player) PlayerDTO {
	if p == nil {
		return PlayerDTO{}
	}
	hand := make([]PlayerCardDTO, len(p.Hand))
	for i, card := range p.Hand {
		hand[i] = PlayerCardToDTO(card)
	}
	var discard []PlayerCardDTO
	if p.Discard != nil {
		discard = make([]PlayerCardDTO, p.Discard.Length())
		for i, card := range p.Discard.Cards {
			discard[i] = PlayerCardToDTO(card)
		}
	}
	var deckCount int
	if p.Deck != nil {
		deckCount = p.Deck.Length()
	}
	var heroClass HeroClass
	var heroName string
	if p.Hero != nil {
		heroClass = p.Hero.Class
		heroName = p.Hero.Name
	}

	return PlayerDTO{
		Id:        p.Id,
		Name:      p.Name,
		HeroClass: heroClass,
		HeroName:  heroName,
		Hand:      hand,
		Discard:   discard,
		DeckCount: deckCount,
	}
}

func DungeonCardsToDTO[T DungeonCard](cards []T) []DungeonCardDTO {
	cardsDTO := make([]DungeonCardDTO, len(cards))
	for i, c := range cards {
		cardsDTO[i] = DungeonCardToDTO(c)
	}

	return cardsDTO
}

func DungeonCardToDTO(card DungeonCard) DungeonCardDTO {
	if card == nil {
		return DungeonCardDTO{}
	}
	var kind DungeonCardKind
	var name string
	var description string
	var resources []ResourceType

	switch c := card.(type) {
	case *DoorCard:
		kind = DungeonCardKind(c.Type)
		name = c.Name
		resources = c.Resources
	case *MiniBossCard:
		kind = DungeonCardKind(c.Type)
		name = c.Name
		resources = c.Resources
	case *EventCard:
		kind = DungeonCardKind(c.Type)
		name = c.Name
		description = c.Description
	case *CurseCard:
		kind = DungeonCardKind(c.Type)
		name = c.Name
		description = c.Description
	case *BossMat:
		kind = CardBoss
		name = c.Name
		resources = c.Resources
	}

	return DungeonCardDTO{
		Id:          card.ID(),
		Kind:        kind,
		Name:        name,
		Description: description,
		Resources:   resources,
	}
}

func PlayerCardsToDTO(cards []PlayerCard) []PlayerCardDTO {
	cardsDTO := make([]PlayerCardDTO, len(cards))
	for i, c := range cards {
		cardsDTO[i] = PlayerCardToDTO(c)
	}

	return cardsDTO
}

func PlayerCardToDTO(card PlayerCard) PlayerCardDTO {
	if card == nil {
		return PlayerCardDTO{}
	}
	var kind PlayerCardKind
	var name string
	var description string
	var resources []ResourceType

	switch c := card.(type) {
	case *ResourceCard:
		kind = PlayerCardResource
		resources = c.Resources
	case *ActionCard:
		kind = PlayerCardAction
		name = c.Name
		description = c.Description
	}

	return PlayerCardDTO{
		Id:          card.ID(),
		Kind:        kind,
		Name:        name,
		Description: description,
		Resources:   resources,
	}
}

func ArtifactsToDTO(cards []*ArtifactCard) []ArtifactDTO {
	cardsDTO := make([]ArtifactDTO, len(cards))
	for i, c := range cards {
		cardsDTO[i] = ArtifactToDTO(c)
	}

	return cardsDTO
}

func ArtifactToDTO(artifact *ArtifactCard) ArtifactDTO {
	if artifact == nil {
		return ArtifactDTO{}
	}
	return ArtifactDTO{
		Id:           artifact.Id,
		Color:        artifact.Color,
		Name:         artifact.Name,
		Description:  artifact.Description,
		SelectAction: artifact.MultiAction,
		Used:         artifact.Used,
	}
}

func PendingInteractionToDTO(pendingInteraction PendingInteraction) *PendingInteractionDTO {
	if pendingInteraction == nil {
		return nil
	}
	var kind InteractionKind
	var sourceCardID CardID
	var requiredCounts map[PlayerID]int
	var pendingPlayers []PlayerID

	switch pi := pendingInteraction.(type) {
	case *TeamChoiceArtifactInteraction:
		kind = pi.Kind
		if pi.Card != nil {
			sourceCardID = pi.Card.ID()
		}
		requiredCounts = nil
		for p := range pi.PendingPlayers {
			pendingPlayers = append(pendingPlayers, p.Id)
		}
	case *TeamChoiceResourceInteraction:
		kind = pi.Kind
		if pi.Card != nil {
			sourceCardID = pi.Card.ID()
		}
		requiredCounts = nil
		for p := range pi.PendingPlayers {
			pendingPlayers = append(pendingPlayers, p.Id)
		}
	case *TeamChoicePlayerInteraction:
		kind = pi.Kind
		if pi.Card != nil {
			sourceCardID = pi.Card.ID()
		}
		requiredCounts = nil
		for p := range pi.PendingPlayers {
			pendingPlayers = append(pendingPlayers, p.Id)
		}
	case *PlayerDiscardCardsInteraction:
		kind = pi.Kind
		if pi.Card != nil {
			sourceCardID = pi.Card.ID()
		}
		requiredCounts = make(map[PlayerID]int, len(pi.requiredCounts))
		for p, count := range pi.requiredCounts {
			requiredCounts[p.Id] = count
		}
		for p := range pi.PendingPlayers {
			pendingPlayers = append(pendingPlayers, p.Id)
		}
	case *PlayerDonatesHandInteraction:
		kind = pi.Kind
		if pi.Card != nil {
			sourceCardID = pi.Card.ID()
		}
		requiredCounts = nil
		for p := range pi.PendingPlayers {
			pendingPlayers = append(pendingPlayers, p.Id)
		}
	}
	slices.Sort(pendingPlayers)

	return &PendingInteractionDTO{
		Kind:           kind,
		SourceCardID:   sourceCardID,
		RequiredCounts: requiredCounts,
		PendingPlayers: pendingPlayers,
	}
}
