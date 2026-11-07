package ui

import (
	"slices"

	d "github.com/bwmarrin/discordgo"
	"github.com/guild-link/commandlink/command"
)

type modal struct {
	u        *UI
	fields   []field
	modal    d.InteractionResponseData
	onSubmit func() error
}

type field interface {
	field(id string) (d.MessageComponent, func(command.Values) error)
}

type fileInput struct {
	label string
	input d.FileUpload
	value *[]string
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

func (m *modal) Input(t field) *modal {
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

func (t *textInput) Paragraph() *textInput {
	t.input.Style = d.TextInputParagraph
	return t
}

func (t *textInput) MaxLength(maxLength int) *textInput {
	t.input.MaxLength = maxLength
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

func FileInput(label string, value *[]string) *fileInput {
	return &fileInput{label: label, input: d.FileUpload{MaxValues: 10}, value: value}
}

func (f *fileInput) Max(maxValues int) *fileInput {
	f.input.MaxValues = maxValues
	return f
}

func (f *fileInput) Optional() *fileInput {
	f.input.Required = new(false)
	return f
}

func (f *fileInput) field(id string) (d.MessageComponent, func(command.Values) error) {
	input := f.input
	input.CustomID = id
	return d.Label{Label: f.label, Component: input}, func(v command.Values) error {
		*f.value = v.Get(id).Strings()
		return nil
	}
}
