package iface

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	DialogCancel = -1
	DialogOk     = 1
)

var (
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B82F6")).
			Padding(1, 2)

	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Bold(true).
				PaddingBottom(1)

	dialogHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			PaddingTop(1)
)

type DialogKeyMap struct {
	Left     key.Binding
	Right    key.Binding
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Esc      key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Space    key.Binding
	Quit     key.Binding
	Close    key.Binding
}

var dialogKeys = DialogKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "right"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Close: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "close"),
	),
}

type DialogCloseMsg struct {
	Return int
}

type Dialog struct {
	title   string
	message string
	width   int
	options []Option
}

func NewDialog(title, message string, opts ...Option) Dialog {
	d := Dialog{
		title:   title,
		message: message,
		width:   60,
		options: opts,
	}

	d.init()

	return d
}

func (d *Dialog) init() {
	contentWidth := d.contentWidth()

	for _, opt := range d.options {
		opt.Init(contentWidth)
	}

	if d.activeIndex() == -1 {
		d.focusIndex(0)
	}
}

func (d *Dialog) contentWidth() int {
	frameX, _ := dialogBoxStyle.GetFrameSize()
	return max(d.width-frameX, 10)
}

func (d *Dialog) SetSize(width, height int) {
	if d.width == width {
		return
	}
	d.width = max(width, 20)

	contentWidth := d.contentWidth()
	for _, opt := range d.options {
		if txt, ok := opt.(*OptionText); ok {
			txt.model.Width = max(contentWidth-4, 10)
		}
	}
}

func (d *Dialog) activeIndex() int {
	for i, opt := range d.options {
		if opt.Focused() {
			return i
		}
	}
	return -1
}

func (d *Dialog) GetActiveOption() Option {
	i := d.activeIndex()
	if i == -1 {
		return nil
	}
	return d.options[i]
}

func (d *Dialog) focusIndex(index int) tea.Cmd {
	if len(d.options) == 0 {
		return nil
	}

	index = ((index % len(d.options)) + len(d.options)) % len(d.options)

	var cmd tea.Cmd
	for i, opt := range d.options {
		if i == index {
			cmd = opt.Focus()
		} else if opt.Focused() {
			opt.Unfocus()
		}
	}

	return cmd
}

func (d *Dialog) focusNext() tea.Cmd {
	return d.focusIndex(d.activeIndex() + 1)
}

func (d *Dialog) focusPrev() tea.Cmd {
	cur := d.activeIndex()
	if cur == -1 {
		cur = len(d.options)
	}
	return d.focusIndex(cur - 1)
}

func (d *Dialog) focusFooter(dir int) tea.Cmd {
	cur := d.activeIndex()
	if cur == -1 || !d.options[cur].Footer() {
		return nil
	}

	for i := cur + dir; i >= 0 && i < len(d.options); i += dir {
		if d.options[i].Footer() {
			return d.focusIndex(i)
		}
	}

	return nil
}

func (d *Dialog) defaultReturn() (int, bool) {
	for i := len(d.options) - 1; i >= 0; i-- {
		if btn, ok := d.options[i].(*OptionButton); ok {
			return btn.Return, true
		}
	}
	return 0, false
}

func closeDialog(ret int) tea.Cmd {
	return func() tea.Msg {
		return DialogCloseMsg{
			Return: ret,
		}
	}
}

func (d Dialog) View() string {
	contentWidth := d.contentWidth()

	fields := []string{
		dialogTitleStyle.Render(d.title),
	}

	if d.message != "" {
		fields = append(fields,
			lipgloss.NewStyle().Width(contentWidth).Render(d.message),
			"",
		)
	}

	footerFields := []string{}
	hasToggle := false

	prevToggle := false
	for _, opt := range d.options {
		if opt.Footer() {
			footerFields = append(footerFields, opt.View())
			continue
		}

		_, isToggle := opt.(*OptionToggle)
		if isToggle {
			hasToggle = true
			if prevToggle {
				fields = fields[:len(fields)-1]
			}
		}
		fields = append(fields, opt.View(), "")
		prevToggle = isToggle
	}

	if len(footerFields) > 0 {
		fields = append(fields, lipgloss.JoinHorizontal(
			lipgloss.Top, footerFields...))
	}

	helpText := "tab/↑↓: move  enter: select  esc: close"
	if hasToggle {
		helpText = "tab/↑↓: move  space: toggle  enter: select  esc: close"
	}
	fields = append(fields, dialogHelpStyle.Width(contentWidth).Render(
		helpText))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		fields...,
	)

	return dialogBoxStyle.Width(d.width).Render(content)
}

func (d Dialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		active := d.GetActiveOption()
		if active != nil {
			return d, active.Update(msg)
		}
		return d, nil
	}

	active := d.GetActiveOption()

	switch {
	case key.Matches(keyMsg, dialogKeys.Quit):
		return d, tea.Quit
	case key.Matches(keyMsg, dialogKeys.Esc):
		return d, closeDialog(DialogCancel)
	case key.Matches(keyMsg, dialogKeys.Close):
		if _, isText := active.(*OptionText); !isText {
			return d, closeDialog(DialogCancel)
		}
	case key.Matches(keyMsg, dialogKeys.Tab),
		key.Matches(keyMsg, dialogKeys.Down):
		return d, d.focusNext()
	case key.Matches(keyMsg, dialogKeys.ShiftTab),
		key.Matches(keyMsg, dialogKeys.Up):
		return d, d.focusPrev()
	case key.Matches(keyMsg, dialogKeys.Left):
		if active != nil && active.Footer() {
			return d, d.focusFooter(-1)
		}
	case key.Matches(keyMsg, dialogKeys.Right):
		if active != nil && active.Footer() {
			return d, d.focusFooter(1)
		}
	case key.Matches(keyMsg, dialogKeys.Space):
		if active != nil && active.OnSpace() {
			return d, nil
		}
	case key.Matches(keyMsg, dialogKeys.Enter):
		if active != nil {
			ret, close, handled := active.OnEnter()
			if handled {
				if close {
					return d, closeDialog(ret)
				}
				return d, nil
			}
		}

		ret, ok := d.defaultReturn()
		if ok {
			return d, closeDialog(ret)
		}
		return d, closeDialog(DialogOk)
	}

	if active != nil {
		return d, active.Update(msg)
	}

	return d, nil
}
