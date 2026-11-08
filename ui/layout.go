package ui

import (
	"fmt"
	"strings"

	d "github.com/bwmarrin/discordgo"
)

type container struct {
	*UI
	color *int
}

func (u *UI) Container() *container {
	c := &container{UI: newUI(u.session)}
	u.add(c)
	return c
}

func (c *container) Color(color int) *container {
	c.color = &color
	return c
}

func (c *container) build() []d.MessageComponent {
	return []d.MessageComponent{d.Container{AccentColor: c.color, Components: build(*c.target)}}
}

type text struct {
	session   *Session
	button    *button
	thumbnail string
	content   string
	lines     []string
	empty     string
}

func (u *UI) Text(content string) *text {
	t := &text{session: u.session, content: content}
	u.add(t)
	return t
}

func (t *text) Line(line string) *text {
	t.lines = append(t.lines, line)
	return t
}

func (t *text) Linef(format string, args ...any) *text {
	return t.Line(fmt.Sprintf(format, args...))
}

func (t *text) Empty(empty string) *text {
	t.empty = empty
	return t
}

func (t *text) Button(label string) *button {
	t.button = newUI(t.session).Button(label)
	return t.button
}

func (t *text) Thumbnail(url string) *text {
	t.thumbnail = url
	return t
}

func (t *text) build() []d.MessageComponent {
	lines := t.lines
	if len(lines) == 0 && t.empty != "" {
		lines = []string{t.empty}
	}

	if t.content != "" {
		lines = append([]string{t.content}, lines...)
	}
	display := d.TextDisplay{Content: strings.Join(lines, "\n")}
	var accessory d.MessageComponent
	switch {
	case t.button != nil:
		accessory = t.button.build()
	case t.thumbnail != "":
		accessory = d.Thumbnail{Media: d.UnfurledMediaItem{URL: t.thumbnail}}
	default:
		return []d.MessageComponent{display}
	}
	return []d.MessageComponent{d.Section{Components: []d.MessageComponent{display}, Accessory: accessory}}
}

type separator struct {
	separator d.Separator
}

func (u *UI) Separator() *separator {
	s := &separator{separator: d.Separator{Divider: new(true), Spacing: new(d.SeparatorSpacingSizeSmall)}}
	u.add(s)
	return s
}

func (s *separator) Large() *separator {
	s.separator.Spacing = new(d.SeparatorSpacingSizeLarge)
	return s
}

func (s *separator) Hidden() *separator {
	s.separator.Divider = new(false)
	return s
}

func (s *separator) build() []d.MessageComponent {
	return []d.MessageComponent{s.separator}
}

type gallery struct {
	urls []string
}

func (u *UI) Gallery(urls ...string) *gallery {
	g := &gallery{urls: urls}
	u.add(g)
	return g
}

func (g *gallery) build() []d.MessageComponent {
	var gallery d.MediaGallery
	for _, url := range g.urls {
		gallery.Items = append(gallery.Items, d.MediaGalleryItem{Media: d.UnfurledMediaItem{URL: url}})
	}
	if len(gallery.Items) == 0 {
		return nil
	}
	return []d.MessageComponent{gallery}
}
