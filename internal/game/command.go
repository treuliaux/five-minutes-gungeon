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
	Player *Player
	Card   PlayerCard
	reply  chan error
}

func (PlayCardCmd) isCommand() {}
func (c PlayCardCmd) Reply() chan error {
	return c.reply
}

type DiscardCardCmd struct {
	Player *Player
	Card   PlayerCard
	reply  chan error
}

func (DiscardCardCmd) isCommand() {}
func (c DiscardCardCmd) Reply() chan error {
	return c.reply
}

type UseHeroAbilityCmd struct {
	Player       *Player
	DiscardCards []PlayerCard
	Ability      Ability
	reply        chan error
}

func (UseHeroAbilityCmd) isCommand() {}
func (c UseHeroAbilityCmd) Reply() chan error {
	return c.reply
}
