package webhook

import (
	"sync"

	d "github.com/bwmarrin/discordgo"
)

type Client struct {
	session  *d.Session
	name     string
	webhooks sync.Map
}
