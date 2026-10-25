package ui

import (
	"slices"

	d "github.com/bwmarrin/discordgo"
	"github.com/chickeniq/commandlink/command"
)

type selectMenu struct {
	session  *Session
	index    int
	menu     d.SelectMenu
	value    *string
	onChange func() error
}

func (u *UI) Select(value *string) *selectMenu {
	s := &selectMenu{session: u.session, index: u.session.selects, value: value}
	u.session.selects++
	s.menu.CustomID = u.handler(func(e Event) error {
		if s.menu.Disabled {
			return nil
		}

		switch value := command.ValueOf(e.Selected).String(); value {
		case previousPage:
			s.session.pages[s.index]--
		case nextPage:
			s.session.pages[s.index]++
		default:
			*s.value = value
			return call(s.onChange)
		}
		return nil
	})
	u.add(s)
	return s
}

func (s *selectMenu) Options(options []d.SelectMenuOption) *selectMenu {
	s.menu.Options = options
	return s
}

func (s *selectMenu) Roles() *selectMenu {
	s.menu.MenuType = d.RoleSelectMenu
	return s
}

func (s *selectMenu) Channels(types ...d.ChannelType) *selectMenu {
	s.menu.MenuType, s.menu.ChannelTypes = d.ChannelSelectMenu, types
	return s
}

func (s *selectMenu) Placeholder(placeholder string) *selectMenu {
	s.menu.Placeholder = placeholder
	return s
}

func (s *selectMenu) Optional() *selectMenu {
	s.menu.MinValues = new(0)
	return s
}

func (s *selectMenu) Disabled(disabled bool) *selectMenu {
	s.menu.Disabled = disabled
	return s
}

func (s *selectMenu) OnChange(onChange func() error) *selectMenu {
	s.onChange = onChange
	return s
}

func (s *selectMenu) Value() command.Value {
	return command.ValueOf(s.selected())
}

func (s *selectMenu) Index() int {
	selected := s.selected()
	if len(selected) == 0 {
		return -1
	}
	return command.OptionIndex(s.menu.Options, selected[0])
}

func (s *selectMenu) build() []d.MessageComponent {
	if !s.typed() && len(s.menu.Options) == 0 {
		return nil
	}
	return []d.MessageComponent{d.ActionsRow{Components: []d.MessageComponent{s.preselected()}}}
}

func (s *selectMenu) selected() []string {
	if s.typed() {
		if *s.value == "" {
			return nil
		}
		return []string{*s.value}
	}

	options := s.menu.Options
	if command.OptionIndex(options, *s.value) < 0 {
		if s.optional() || len(options) == 0 {
			return nil
		}
		*s.value = options[s.page()*pageOptions].Value
	}

	return []string{*s.value}
}

func (s *selectMenu) page() int {
	options := s.menu.Options
	if len(options) <= maxOptions {
		return 0
	}

	page, ok := s.session.pages[s.index]
	if !ok {
		page = max(command.OptionIndex(options, *s.value), 0) / pageOptions
	}

	page = min(max(page, 0), (len(options)-1)/pageOptions)
	s.session.pages[s.index] = page
	return page
}

func (s *selectMenu) shown() []d.SelectMenuOption {
	options := s.menu.Options
	if len(options) <= maxOptions {
		return slices.Clone(options)
	}

	page := s.page()
	shown := slices.Clone(options[page*pageOptions : min((page+1)*pageOptions, len(options))])
	if page > 0 {
		shown = append(shown, d.SelectMenuOption{Label: "◀ Previous page", Value: previousPage})
	}

	if (page+1)*pageOptions < len(options) {
		shown = append(shown, d.SelectMenuOption{Label: "Next page ▶", Value: nextPage})
	}
	return shown
}

func (s *selectMenu) optional() bool {
	return s.menu.MinValues != nil && *s.menu.MinValues == 0
}

func (s *selectMenu) typed() bool {
	return s.menu.Type() != d.SelectMenuComponent
}

func (s *selectMenu) preselected() d.SelectMenu {
	menu, selected := s.menu, s.selected()
	menu.Options = s.shown()
	for i, option := range menu.Options {
		menu.Options[i].Default = slices.Contains(selected, option.Value)
	}

	if !s.typed() {
		if len(selected) > 0 && command.OptionIndex(menu.Options, selected[0]) < 0 {
			menu.Placeholder = s.menu.Options[command.OptionIndex(s.menu.Options, selected[0])].Label
		}
		return menu
	}

	kind := d.SelectMenuDefaultValueRole
	if menu.MenuType == d.ChannelSelectMenu {
		kind = d.SelectMenuDefaultValueChannel
	}

	menu.DefaultValues = make([]d.SelectMenuDefaultValue, len(selected))
	for i, id := range selected {
		menu.DefaultValues[i] = d.SelectMenuDefaultValue{ID: id, Type: kind}
	}
	return menu
}
