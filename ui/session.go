package ui

import (
	"strconv"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
)

func New(c InteractiveCommand, prefix string) *Session {
	return &Session{
		prefix:   prefix,
		command:  c,
		handlers: map[string]func(Event) error{},
		pages:    map[int]int{},
	}
}

func (s *Session) Handle(ctx *command.Ctx, event Event) (*command.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.open = nil
	s.ctx = *ctx

	var failed error
	if handler := s.handlers[event.ID]; handler != nil {
		failed = handler(event)
	}

	if s.open != nil && failed == nil {
		return &command.Response{Modal: s.open}, nil
	}

	s.open = nil
	return s.render(failed)
}

func (s *Session) render(failed error) (*command.Response, error) {
	u := &UI{session: s, target: new([]node)}
	s.pending, s.selects = map[string]func(Event) error{}, 0
	if err := s.command.Interactive(&s.ctx, u); err != nil {
		s.handlers = map[string]func(Event) error{}
		return nil, err
	}

	if s.open != nil {
		return &command.Response{Modal: s.open}, nil
	}

	s.handlers = s.pending
	return &command.Response{Components: build(*u.target), Error: failed}, nil
}

func (s *Session) id() string {
	s.next++
	return s.prefix + ":" + strconv.FormatInt(int64(s.next), 36)
}

func (u *UI) handler(handler func(Event) error) string {
	id := u.session.id()
	u.session.pending[id] = handler
	return id
}

func (u *UI) add(n node) {
	u.row = nil
	*u.target = append(*u.target, n)
}

func call(f func() error) error {
	if f == nil {
		return nil
	}
	return f()
}

func build(nodes []node) []d.MessageComponent {
	var components []d.MessageComponent
	for _, n := range nodes {
		components = append(components, n.build()...)
	}
	return components
}
