package ui

import (
	"strconv"

	d "github.com/bwmarrin/discordgo"
	"github.com/guild-link/commandlink/command"
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

	s.modal = nil
	s.ctx = *ctx

	var failed error
	if handler := s.handlers[event.ID]; handler != nil {
		failed = handler(event)
	}

	if s.modal != nil && failed == nil {
		return &command.Response{Modal: s.modal}, nil
	}

	s.modal = nil
	return s.render(failed)
}

func (s *Session) render(failed error) (*command.Response, error) {
	u := newUI(s)
	s.pending = map[string]func(Event) error{}
	s.selectCount = 0
	if err := s.command.Interactive(&s.ctx, u); err != nil {
		s.handlers = map[string]func(Event) error{}
		return nil, err
	}

	if s.modal != nil {
		return &command.Response{Modal: s.modal}, nil
	}

	s.handlers = s.pending
	return &command.Response{Components: build(*u.target), Error: failed}, nil
}

func (s *Session) id() string {
	s.lastID++
	return s.prefix + ":" + strconv.FormatInt(int64(s.lastID), 36)
}

func newUI(s *Session) *UI {
	return &UI{session: s, target: new([]node)}
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
