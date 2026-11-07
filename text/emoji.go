package text

import (
	"maps"
	"slices"
	"strings"

	d "github.com/bwmarrin/discordgo"
	"github.com/kyokomi/emoji/v2"
)

var emojiReplacer = newEmojiReplacer()

func EmojiNames(s string) string {
	s = d.EmojiRegex.ReplaceAllStringFunc(s, func(custom string) string {
		return ":" + strings.Split(custom, ":")[1] + ":"
	})

	return emojiReplacer.Replace(s)
}

func CustomEmojis(state *d.State, channelID, s string) string {
	channel, err := state.Channel(channelID)
	if err != nil {
		return s
	}

	guild, err := state.Guild(channel.GuildID)
	if err != nil {
		return s
	}

	var pairs []string
	for _, e := range guild.Emojis {
		pairs = append(pairs, ":"+e.Name+":", e.MessageFormat())
	}

	return strings.NewReplacer(pairs...).Replace(s)
}

func newEmojiReplacer() *strings.Replacer {
	names := emoji.RevCodeMap()
	codes := slices.Collect(maps.Keys(names))
	slices.SortFunc(codes, func(a, b string) int { return len(b) - len(a) })

	var pairs []string
	for _, code := range codes {
		pairs = append(pairs, code, names[code][0])
	}

	return strings.NewReplacer(pairs...)
}
