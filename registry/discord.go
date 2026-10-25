package registry

import (
	"context"
	"fmt"
	"strings"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
	"github.com/chickeniq/commandlink/ui"
)

func (r *Registry) Handle(ctx context.Context, session *d.Session, event *d.InteractionCreate, identify IdentifyProvider) {
	i := &interaction{registry: r, ctx: ctx, session: session, event: event, identify: identify}
	switch event.Type {
	case d.InteractionApplicationCommand:
		i.command()
	case d.InteractionApplicationCommandAutocomplete:
		i.autocomplete()
	case d.InteractionMessageComponent:
		data := event.MessageComponentData()
		i.resumeSession(data.CustomID, ui.Event{Selected: data.Values})
	case d.InteractionModalSubmit:
		modal := event.ModalSubmitData()
		i.resumeSession(modal.CustomID, ui.Event{Inputs: inputs(modal.Components)})
	}
}

func (i *interaction) command() {
	data := i.event.ApplicationCommandData()
	cmd, def := i.registry.command(data.Name)
	c, _ := cmd.(ui.InteractiveCommand)

	i.reply(func(ctx *command.Ctx) (*command.Response, error) {
		if c == nil {
			return nil, fmt.Errorf("unknown command %q", data.Name)
		}

		ctx.Values = values(data.Options)
		return i.registry.startSession(ctx, c, def)
	})
}

func (i *interaction) autocomplete() {
	data := &d.InteractionResponseData{}
	defer i.respond(d.InteractionApplicationCommandAutocompleteResult, data)

	created, _ := d.SnowflakeTimestamp(i.event.ID)
	timed, cancel := context.WithDeadline(i.ctx, created.Add(deferAfter))
	defer cancel()

	ctx, err := i.identify.newCtx(timed)
	if logError("create Discord autocomplete context", err) != nil {
		return
	}

	cmd := i.event.ApplicationCommandData()
	ctx.Values = values(cmd.Options)

	for _, option := range cmd.Options {
		if option.Focused {
			completions := i.registry.complete(ctx, cmd.Name, option.Name)
			data.Choices = completions[:min(len(completions), maxChoices)]
			return
		}
	}
}

func (i *interaction) resumeSession(customID string, event ui.Event) {
	sessionID, _, _ := strings.Cut(customID, ":")
	event.ID = customID

	s, err := i.registry.loadSession(sessionID)
	if err != nil {
		i.fail(false, err)
		return
	}

	i.reply(func(ctx *command.Ctx) (*command.Response, error) {
		return i.registry.handleSession(ctx, s, event)
	})
}

func (r *Registry) complete(ctx *command.Ctx, name, arg string) []*d.ApplicationCommandOptionChoice {
	def := r.commands[name].def
	if !ctx.User.HasPermission(def.Permission) {
		return nil
	}

	for _, a := range def.Args {
		if a.Name == arg && a.Completion != nil {
			return a.Completion(ctx, ctx.Get(arg).String())
		}
	}
	return nil
}
