package text

import (
	"slices"
	"strings"
	"unicode"

	d "github.com/bwmarrin/discordgo"
)

func Mentions(state *d.State, m *d.Message) string {
	return mentions(state, m, func(name string) string { return name })
}

func PlainMentions(state *d.State, m *d.Message) string {
	return mentions(state, m, func(name string) string {
		return strings.TrimSpace(strings.Map(func(r rune) rune {
			if unicode.Is(unicode.So, r) || r == '\u200d' || r == '\ufe0f' {
				return -1
			}

			return r
		}, name))
	})
}

func Nickname(state *d.State, guildID string, user *d.User) string {
	if member, err := state.Member(guildID, user.ID); err == nil && member.Nick != "" {
		return member.Nick
	}

	return user.DisplayName()
}

func mentions(state *d.State, m *d.Message, name func(string) string) string {
	var pairs []string
	for _, user := range m.Mentions {
		nick := "@" + Nickname(state, m.GuildID, user)
		pairs = append(pairs, "<@"+user.ID+">", nick, "<@!"+user.ID+">", nick)
	}

	if guild, err := state.Guild(m.GuildID); err == nil {
		for _, channel := range slices.Concat(guild.Channels, guild.Threads) {
			pairs = append(pairs, "<#"+channel.ID+">", "#"+name(channel.Name))
		}

		for _, role := range guild.Roles {
			pairs = append(pairs, "<@&"+role.ID+">", "@"+name(role.Name))
		}
	}

	return strings.NewReplacer(pairs...).Replace(m.Content)
}
