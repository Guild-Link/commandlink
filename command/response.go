package command

import (
	"slices"

	d "github.com/bwmarrin/discordgo"
)

func MkOptions[T any](items []T, option func(T) d.SelectMenuOption) []d.SelectMenuOption {
	options := make([]d.SelectMenuOption, len(items))
	for i, item := range items {
		options[i] = option(item)
	}
	return options
}

func OptionIndex(options []d.SelectMenuOption, value string) int {
	return slices.IndexFunc(options, func(o d.SelectMenuOption) bool { return o.Value == value })
}
