package ui

import (
	"slices"

	d "github.com/bwmarrin/discordgo"
)

type button struct {
	onClick func() error
	button  d.Button
}

type row struct {
	buttons []*button
}

func (u *UI) NewRow() {
	u.row = nil
}

func (u *UI) Button(label string) *button {
	if u.row == nil {
		r := &row{}
		u.add(r)
		u.row = r
	}

	b := &button{button: d.Button{Label: label, Style: d.PrimaryButton}}
	b.button.CustomID = u.handler(func(Event) error {
		if b.button.Disabled || b.onClick == nil {
			return nil
		}
		return b.onClick()
	})
	u.row.buttons = append(u.row.buttons, b)
	return b
}

func (b *button) Style(style d.ButtonStyle) *button {
	b.button.Style = style
	return b
}

func (b *button) Link(url string) *button {
	b.button.URL = url
	return b
}

func (b *button) Disabled(disabled bool) *button {
	b.button.Disabled = disabled
	return b
}

func (b *button) OnClick(onClick func() error) *button {
	b.onClick = onClick
	return b
}

func (b *button) build() d.MessageComponent {
	button := b.button
	if button.URL != "" {
		button.CustomID, button.Style = "", d.LinkButton
	}
	return button
}

func (r *row) build() []d.MessageComponent {
	var buttons, rows []d.MessageComponent
	for _, b := range r.buttons {
		buttons = append(buttons, b.build())
	}

	for chunk := range slices.Chunk(buttons, maxButtons) {
		rows = append(rows, d.ActionsRow{Components: chunk})
	}
	return rows
}
