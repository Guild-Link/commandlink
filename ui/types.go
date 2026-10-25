package ui

import (
	"sync"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
)

const (
	maxButtons   = 5
	maxOptions   = 25
	pageOptions  = 23
	previousPage = "commandlink:previous"
	nextPage     = "commandlink:next"
)

type InteractiveCommand interface {
	Interactive(*command.Ctx, *UI) error
	command.Command
}

type Session struct {
	handlers map[string]func(Event) error
	pending  map[string]func(Event) error
	mu       sync.Mutex

	pages   map[int]int
	selects int

	open    *d.InteractionResponseData
	command InteractiveCommand
	ctx     command.Ctx
	prefix  string

	next int
}

type node interface {
	build() []d.MessageComponent
}

type UI struct {
	session *Session
	target  *[]node
	row     *row
}

type Event struct {
	ID       string
	Selected []string
	Inputs   command.Values
}
