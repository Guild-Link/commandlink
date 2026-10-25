package registry

import (
	"context"
	"errors"
	"log"
	"time"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
)

func (i *interaction) reply(run func(*command.Ctx) (*command.Response, error)) {
	created, _ := d.SnowflakeTimestamp(i.event.ID)
	done := make(chan result, 1)
	go func() {
		timed, cancel := context.WithTimeout(i.ctx, i.registry.timeout)
		defer cancel()

		var resp *command.Response
		ctx, err := i.identify.newCtx(timed)
		if err == nil {
			resp, err = run(ctx)
		}
		done <- result{resp, err}
	}()

	select {
	case r := <-done:
		i.send(false, r)

	case <-time.After(time.Until(created.Add(deferAfter))):
		ackType := d.InteractionResponseDeferredMessageUpdate
		if i.event.Message == nil {
			ackType = d.InteractionResponseDeferredChannelMessageWithSource
		}

		i.respond(ackType, nil)
		i.send(true, <-done)
	}
}

func (i *interaction) send(deferred bool, r result) {
	if r.err == nil && r.resp.Modal != nil && deferred {
		r.err = errors.New("this took too long to open a modal, try again")
	}

	if r.err != nil {
		i.fail(deferred, r.err)
		return
	}

	responseType := d.InteractionResponseUpdateMessage
	if i.event.Message == nil {
		responseType = d.InteractionResponseChannelMessageWithSource
	}

	var err error
	switch {
	case r.resp.Modal != nil:
		err = i.respond(d.InteractionResponseModal, r.resp.Modal)
	case deferred:
		_, err = i.session.InteractionResponseEdit(i.event.Interaction, &d.WebhookEdit{
			Flags:      d.MessageFlagsIsComponentsV2,
			Components: &r.resp.Components,
		}, d.WithContext(i.ctx))
		logError("edit Discord interaction response", err)
	default:
		err = i.respond(responseType, &d.InteractionResponseData{
			Flags:      d.MessageFlagsIsComponentsV2,
			Components: r.resp.Components,
		})
	}

	if err != nil {
		i.fail(deferred, err)
		return
	}

	if r.resp.Error != nil {
		i.followup(mkError(r.resp.Error))
	}
}

func (i *interaction) fail(deferred bool, err error) {
	if !deferred {
		i.respond(d.InteractionResponseChannelMessageWithSource, &d.InteractionResponseData{
			Flags:      d.MessageFlagsEphemeral | d.MessageFlagsIsComponentsV2,
			Components: mkError(err),
		})
		return
	}

	if i.event.Message == nil {
		err := i.session.InteractionResponseDelete(i.event.Interaction, d.WithContext(i.ctx))
		logError("delete Discord interaction response", err)
	}
	i.followup(mkError(err))
}

func (i *interaction) followup(components []d.MessageComponent) {
	params := &d.WebhookParams{Flags: d.MessageFlagsEphemeral | d.MessageFlagsIsComponentsV2, Components: components}
	_, err := i.session.FollowupMessageCreate(i.event.Interaction, false, params, d.WithContext(i.ctx))
	logError("send Discord followup", err)
}

func (i *interaction) respond(responseType d.InteractionResponseType, data *d.InteractionResponseData) error {
	response := &d.InteractionResponse{Type: responseType, Data: data}
	err := i.session.InteractionRespond(i.event.Interaction, response, d.WithContext(i.ctx))
	return logError("respond to Discord interaction", err)
}

func logError(what string, err error) error {
	if err != nil {
		log.Printf("%s: %v", what, err)
	}
	return err
}

func mkError(err error) []d.MessageComponent {
	return []d.MessageComponent{d.Container{
		Components:  []d.MessageComponent{d.TextDisplay{Content: "**Error**\n" + err.Error()}},
		AccentColor: new(errorColor),
	}}
}
