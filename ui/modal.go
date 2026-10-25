package ui

import (
	"slices"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
)

type modal struct {
	u        *UI
	fields   []*textInput
	modal    d.InteractionResponseData
	onSubmit func() error
}

type textInput struct {
	label       string
	description string
	input       d.TextInput
	value       *string
	onChange    func() error
}

func (u *UI) Modal(title string) *modal {
	return &modal{u: u, modal: d.InteractionResponseData{Title: title}}
}

func (m *modal) Text(text string) *modal {
	m.modal.Components = append(m.modal.Components, d.TextDisplay{Content: text})
	return m
}

func (m *modal) Input(t *textInput) *modal {
	m.fields = append(m.fields, t)
	return m
}

func (m *modal) OnSubmit(onSubmit func() error) *modal {
	m.onSubmit = onSubmit
	return m
}

func (m *modal) Open() error {
	s := m.u.session
	modal := m.modal
	modal.CustomID, modal.Components = s.id(), slices.Clone(modal.Components)

	var submits []func(command.Values) error
	for _, f := range m.fields {
		component, submit := f.field(s.id())
		modal.Components = append(modal.Components, component)
		submits = append(submits, submit)
	}

	s.handlers[modal.CustomID] = func(e Event) error {
		for _, submit := range submits {
			if err := submit(e.Inputs); err != nil {
				return err
			}
		}

		if m.onSubmit == nil {
			return nil
		}
		return m.onSubmit()
	}
	s.open = &modal
	return nil
}

func TextInput(label string, value *string) *textInput {
	input := d.TextInput{Style: d.TextInputShort, Value: *value}
	return &textInput{label: label, input: input, value: value}
}

func (t *textInput) OnChange(onChange func() error) *textInput {
	t.onChange = onChange
	return t
}

func (t *textInput) Description(description string) *textInput {
	t.description = description
	return t
}

func (t *textInput) Placeholder(placeholder string) *textInput {
	t.input.Placeholder = placeholder
	return t
}

func (t *textInput) Optional() *textInput {
	t.input.Required = new(false)
	return t
}

func (t *textInput) field(id string) (d.MessageComponent, func(command.Values) error) {
	input := t.input
	input.CustomID = id
	return d.Label{Label: t.label, Description: t.description, Component: input}, func(v command.Values) error {
		*t.value = v.Get(id).String()
		return call(t.onChange)
	}
}
