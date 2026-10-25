package registry

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/chickeniq/commandlink/command"
)

func (r *Registry) Text(ctx context.Context, prefix, content string, identify IdentifyProvider) (string, error) {
	content, ok := strings.CutPrefix(content, prefix)
	words := strings.Fields(content)
	if !ok || len(words) == 0 {
		return "", nil
	}

	cmd, def := r.command(words[0])
	c, ok := cmd.(command.TextCommand)
	if !ok && words[0] != "help" {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	commandCtx, err := identify.newCtx(ctx)
	if err != nil {
		return "", err
	}

	if !ok {
		return r.help(prefix, commandCtx.User), nil
	}

	args, words := def.Args, words[1:]
	commandCtx.Values = make(command.Values, len(args))
	for i, arg := range args[:min(len(args), len(words))] {
		if words[i] != "_" || arg.Required {
			commandCtx.Values[arg.Name] = command.ValueOf(words[i])
		}
	}

	text, err := run(commandCtx, def, c.Text)
	if err == nil && text == "" {
		return "", fmt.Errorf("command %q returned nothing", def.Name)
	}

	return text, err
}

func (r *Registry) help(prefix string, user command.User) string {
	var usages []string
	skips := false
	for name, e := range r.commands {
		if _, ok := e.new().(command.TextCommand); !ok || e.def.Name != name || !user.HasPermission(e.def.Permission) {
			continue
		}

		usage := prefix + name
		if len(e.def.Aliases) > 0 {
			usage += " (" + strings.Join(e.def.Aliases, ", ") + ")"
		}

		optional := 0
		for _, arg := range e.def.Args {
			if arg.Required {
				usage += " <" + arg.Name + ">"
				continue
			}

			usage += " [" + arg.Name + "]"
			optional++
		}

		skips = skips || optional > 1
		usages = append(usages, usage)
	}

	slices.Sort(usages)
	help := "Cmds: " + strings.Join(usages, ", ")
	if skips {
		help += " | Skip optionals with `_`"
	}
	return help
}
