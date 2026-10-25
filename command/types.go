package command

import (
	"context"

	d "github.com/bwmarrin/discordgo"
)

type Command interface {
	Definition() Definition
}

type TextCommand interface {
	Command
	Text(*Ctx) (string, error)
}

type Definition struct {
	Name       string
	Desc       string
	Permission Permission

	Args    []Arg
	Aliases []string
}

type Arg struct {
	Name     string
	Desc     string
	Required bool

	Type       d.ApplicationCommandOptionType
	Completion func(ctx *Ctx, current string) []*d.ApplicationCommandOptionChoice
}

type Ctx struct {
	context.Context
	Values

	ServerID string
	User     User
}

type User struct {
	Permission Permission
	Name       string
	ID         string
}

type Permission interface {
	String() string
	Level() int
}

type Response struct {
	Modal      *d.InteractionResponseData
	Components []d.MessageComponent
	Error      error
}

type Values map[string]Value
type Value struct{ value any }
