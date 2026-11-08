package webhook

import (
	"regexp"
	"sync"

	d "github.com/bwmarrin/discordgo"
)

const maxUsername = 80

var discordExpr = regexp.MustCompile(`(?i)(disc)o(rd)`)
var clydeExpr = regexp.MustCompile(`(?i)(clyd)e`)

type Client struct {
	session  *d.Session
	name     string
	webhooks sync.Map
}
