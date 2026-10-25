package registry

import (
	"context"
	"fmt"
	"strconv"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
	"github.com/chickeniq/commandlink/ui"
)

func (f IdentifyProvider) newCtx(ctx context.Context) (*command.Ctx, error) {
	user, serverID, err := f(ctx)
	if err != nil {
		return nil, err
	}

	return &command.Ctx{Context: ctx, User: user, ServerID: serverID}, nil
}

func check(c command.Command, def command.Definition) error {
	_, interactive := c.(ui.InteractiveCommand)
	_, text := c.(command.TextCommand)

	if !text && !interactive {
		return fmt.Errorf("command %q: must implement Text or Interactive", def.Name)
	}

	optional := false
	for _, arg := range def.Args {
		if optional && arg.Required {
			return fmt.Errorf("command %q: required arg %q follows optional arg", def.Name, arg.Name)
		}
		optional = !arg.Required
	}

	return nil
}

func valid(arg command.Arg, value string) bool {
	var err error
	switch arg.Type {
	case d.ApplicationCommandOptionInteger:
		_, err = strconv.Atoi(value)
	case d.ApplicationCommandOptionNumber:
		_, err = strconv.ParseFloat(value, 64)
	case d.ApplicationCommandOptionBoolean:
		_, err = strconv.ParseBool(value)
	}
	return err == nil
}

func values(options []*d.ApplicationCommandInteractionDataOption) command.Values {
	values := make(command.Values, len(options))
	for _, option := range options {
		values[option.Name] = command.ValueOf(option.Value)
	}
	return values
}

func inputs(components []d.MessageComponent) command.Values {
	inputs := command.Values{}
	for _, component := range components {
		if label, ok := component.(*d.Label); ok {
			if input, ok := label.Component.(*d.TextInput); ok {
				inputs[input.CustomID] = command.ValueOf(input.Value)
			}
		}
	}
	return inputs
}
