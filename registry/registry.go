package registry

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	d "github.com/bwmarrin/discordgo"
	"github.com/guild-link/commandlink/command"
	"github.com/guild-link/commandlink/ui"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

func New(commands []func() command.Command, timeout, ttl time.Duration) (*Registry, error) {
	r := &Registry{
		sessions: expirable.NewLRU[string, *session](0, nil, ttl),
		commands: make(map[string]entry, len(commands)),
		aliases:  map[string]string{},
		timeout:  timeout,
	}

	for _, f := range commands {
		c := f()
		def := c.Definition()
		if err := validate(c, def); err != nil {
			return nil, err
		}

		for _, name := range append([]string{def.Name}, def.Aliases...) {
			if c, _ := r.command(name); c != nil {
				return nil, fmt.Errorf("command %q: name %q is already used", def.Name, name)
			}
		}

		r.commands[def.Name] = entry{def: def, new: f}
		for _, alias := range def.Aliases {
			r.aliases[alias] = def.Name
		}
	}

	return r, nil
}

func Cmd[T any, P ptr[T]](c T) func() command.Command {
	return func() command.Command { return P(new(c)) }
}

func (r *Registry) Register(ctx context.Context, session *d.Session, appID, guildID string) error {
	var commands []*d.ApplicationCommand
	for _, e := range r.commands {
		def := e.def
		if _, ok := e.new().(ui.InteractiveCommand); !ok {
			continue
		}

		slash := &d.ApplicationCommand{Name: def.Name, Description: def.Desc}
		for _, arg := range def.Args {
			slash.Options = append(slash.Options, &d.ApplicationCommandOption{
				Type:         cmp.Or(arg.Type, d.ApplicationCommandOptionString),
				Autocomplete: arg.Completion != nil,
				Required:     arg.Required,
				Description:  arg.Desc,
				Name:         arg.Name,
			})
		}

		commands = append(commands, slash)
	}

	if _, err := session.ApplicationCommandBulkOverwrite(appID, guildID, commands, d.WithContext(ctx)); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}
	return nil
}

func (r *Registry) command(name string) (command.Command, command.Definition) {
	if real, ok := r.aliases[name]; ok {
		name = real
	}

	e, ok := r.commands[name]
	if !ok {
		return nil, command.Definition{}
	}

	return e.new(), e.def
}

func run[T any](ctx *command.Ctx, def command.Definition, fn func(*command.Ctx) (T, error)) (T, error) {
	var zero T
	if !ctx.User.HasPermission(def.Permission) {
		return zero, fmt.Errorf("command requires %s permission", def.Permission)
	}

	for _, arg := range def.Args {
		value := ctx.Get(arg.Name).String()
		switch {
		case value == "" && arg.Required:
			return zero, fmt.Errorf("missing required argument %q", arg.Name)
		case value != "" && !valid(arg, value):
			return zero, fmt.Errorf("invalid value %q for argument %q", value, arg.Name)
		}
	}

	data, err := fn(ctx)
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		log.Printf("command %q timed out: %v", def.Name, err)
		return data, errors.New("command timed out")
	}

	return data, err
}
