package game

type Command interface {
	isCommand()
	Reply() chan error
}

type StartCmd struct {
	reply chan error
}

func (StartCmd) isCommand() {}
func (c StartCmd) Reply() chan error {
	return c.reply
}

type AddPlayerCmd struct {
	Name  string
	Class HeroClass
	reply chan error
}

func (AddPlayerCmd) isCommand() {}
func (c AddPlayerCmd) Reply() chan error {
	return c.reply
}

type PlayCardCmd struct {
	PlayerID        PlayerID
	CardID          CardID
	TargetCardID    CardID
	TargetPlayerIDs []PlayerID
	reply           chan error
}

func (PlayCardCmd) isCommand() {}
func (c PlayCardCmd) Reply() chan error {
	return c.reply
}

type DiscardCardsCmd struct {
	PlayerID PlayerID
	CardIDs  []CardID
	reply    chan error
}

func (DiscardCardsCmd) isCommand() {}
func (c DiscardCardsCmd) Reply() chan error {
	return c.reply
}

type UseHeroAbilityCmd struct {
	PlayerID       PlayerID
	DiscardCardIDs []CardID
	TargetCardID   CardID
	TargetPlayerID PlayerID
	reply          chan error
}

func (UseHeroAbilityCmd) isCommand() {}
func (c UseHeroAbilityCmd) Reply() chan error {
	return c.reply
}

type SubmitPromptChoiceCmd struct {
	PlayerID       PlayerID
	TargetPlayerID PlayerID
	CardIDs        []CardID
	Resource       *ResourceType
	reply          chan error
}

func (SubmitPromptChoiceCmd) isCommand() {}
func (c SubmitPromptChoiceCmd) Reply() chan error {
	return c.reply
}

type UseArtifactCmd struct {
	PlayerID    PlayerID
	ArtifactID  ArtifactID
	ActionIndex ArtifactActionIndex
	TargetID    CardID
	reply       chan error
}

func (UseArtifactCmd) isCommand() {}
func (c UseArtifactCmd) Reply() chan error {
	return c.reply
}
