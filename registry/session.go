package registry

import (
	"crypto/rand"
	"errors"

	"github.com/chickeniq/commandlink/command"
	"github.com/chickeniq/commandlink/ui"
)

func (r *Registry) startSession(ctx *command.Ctx, c ui.InteractiveCommand, def command.Definition) (*command.Response, error) {
	id := rand.Text()
	s := &session{componentID: id, userID: ctx.User.ID, command: def, values: ctx.Values, ui: ui.New(c, id)}

	return r.handleSession(ctx, s, ui.Event{})
}

func (r *Registry) loadSession(id string) (*session, error) {
	s, ok := r.sessions.Get(id)
	if !ok {
		return nil, errors.New("this interaction has expired")
	}
	return s, nil
}

func (r *Registry) handleSession(ctx *command.Ctx, s *session, event ui.Event) (*command.Response, error) {
	if s.userID != ctx.User.ID {
		return nil, errors.New("this interaction belongs to someone else")
	}

	ctx.Values = s.values
	resp, err := run(ctx, s.command, func(ctx *command.Ctx) (*command.Response, error) {
		return s.ui.Handle(ctx, event)
	})
	if err == nil {
		r.sessions.Add(s.componentID, s)
	}
	return resp, err
}
