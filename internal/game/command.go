package game

type Command interface {
	isCommand()
	Reply() chan error
}

type StartCmd struct {
	reply chan error
}

func (cmd StartCmd) isCommand() {}
func (cmd StartCmd) Reply() chan error {
	return cmd.reply
}

type AddPlayerCmd struct {
	Name  string
	Class HeroClass
	reply chan error
}

func (cmd AddPlayerCmd) isCommand() {}
func (cmd AddPlayerCmd) Reply() chan error {
	return cmd.reply
}

type PlayCardCmd struct {
	Player *Player
	Card   PlayerCard
	reply  chan error
}

func (cmd PlayCardCmd) isCommand() {}
func (cmd PlayCardCmd) Reply() chan error {
	return cmd.reply
}

type DiscardCardCmd struct {
	Player *Player
	Card   PlayerCard
	reply  chan error
}

func (cmd DiscardCardCmd) isCommand() {}
func (cmd DiscardCardCmd) Reply() chan error {
	return cmd.reply
}

type UseHeroAbilityCmd struct {
	Player       *Player
	DiscardCards []PlayerCard
	Ability      Ability
	reply        chan error
}

func (cmd UseHeroAbilityCmd) isCommand() {}
func (cmd UseHeroAbilityCmd) Reply() chan error {
	return cmd.reply
}
