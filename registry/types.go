package registry

import (
	"context"
	"time"

	d "github.com/bwmarrin/discordgo"
	"github.com/guild-link/commandlink/command"
	"github.com/guild-link/commandlink/ui"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	deferAfter = d.InteractionDeadline - 500*time.Millisecond
	errorColor = 0xED4245
	maxChoices = 25
)

type Registry struct {
	sessions *expirable.LRU[string, *session]
	aliases  map[string]string
	commands map[string]entry
	timeout  time.Duration
}

type IdentifyProvider func(context.Context) (Identity, error)

type Identity struct {
	User     command.User
	ServerID string
	Source   string
}

type entry struct {
	new func() command.Command
	def command.Definition
}

type ptr[T any] interface {
	command.Command
	*T
}

type session struct {
	command command.Definition
	values  command.Values

	componentID string
	userID      string

	ui *ui.Session
}

type interaction struct {
	event    *d.InteractionCreate
	session  *d.Session
	registry *Registry

	ctx      context.Context
	identify IdentifyProvider
}

type result struct {
	resp *command.Response
	err  error
}
