package iface

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	optionButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#3B82F6")).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(2)
	optionButtonActiveStyle = optionButtonStyle.
				Foreground(lipgloss.Color("#3B82F6")).
				Background(lipgloss.Color("#FFFFFF")).
				Underline(true)

	optionLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))
	optionLabelActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Bold(true)

	toggleOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Background(lipgloss.Color("#E5E7EB")).
			Padding(0, 1).
			MarginRight(1)
	toggleOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3B82F6")).
			Padding(0, 1).
			MarginRight(1)
)

// Option is a single field or button in a Dialog.
type Option interface {
	// Init prepares the option for display in a dialog of the given
	// content width.
	Init(width int)

	// Footer options are rendered horizontally at the bottom of the dialog.
	Footer() bool
	Update(tea.Msg) tea.Cmd
	Focused() bool
	Focus() tea.Cmd
	Unfocus()

	// OnEnter returns the dialog return value, whether the dialog should
	// close with that value and whether the key was handled. Unhandled
	// enter presses activate the default button.
	OnEnter() (ret int, close bool, handled bool)

	// OnSpace returns true when the key was consumed by the option.
	OnSpace() bool
	View() string
}

type OptionText struct {
	Label       string
	Placeholder string
	Value       string
	Password    bool
	model       textinput.Model
}

func (o *OptionText) Init(width int) {
	o.model = textinput.New()
	o.model.Placeholder = o.Placeholder
	o.model.CharLimit = 2048
	o.model.Width = max(width-4, 10)
	o.model.Prompt = "> "
	if o.Password {
		o.model.EchoMode = textinput.EchoPassword
		o.model.EchoCharacter = '•'
	}
	if o.Value != "" {
		o.model.SetValue(o.Value)
	}
}

func (o *OptionText) Footer() bool {
	return false
}

func (o *OptionText) Update(msg tea.Msg) (cmd tea.Cmd) {
	o.model, cmd = o.model.Update(msg)
	return
}

func (o *OptionText) Focused() bool {
	return o.model.Focused()
}

func (o *OptionText) Focus() (cmd tea.Cmd) {
	cmd = o.model.Focus()
	return
}

func (o *OptionText) Unfocus() {
	o.model.Blur()
}

func (o *OptionText) OnEnter() (int, bool, bool) {
	return 0, false, false
}

func (o *OptionText) OnSpace() bool {
	return false
}

func (o *OptionText) View() string {
	var label string
	if o.Focused() {
		label = optionLabelActiveStyle.Render(o.Label)
	} else {
		label = optionLabelStyle.Render(o.Label)
	}
	return lipgloss.JoinVertical(lipgloss.Left, label, o.model.View())
}

func (o *OptionText) GetValue() string {
	return o.model.Value()
}

type OptionToggle struct {
	Label   string
	Value   bool
	focused bool
}

func (o *OptionToggle) Init(width int) {
}

func (o *OptionToggle) Footer() bool {
	return false
}

func (o *OptionToggle) Update(msg tea.Msg) (cmd tea.Cmd) {
	return
}

func (o *OptionToggle) Focused() bool {
	return o.focused
}

func (o *OptionToggle) Focus() (cmd tea.Cmd) {
	o.focused = true
	return
}

func (o *OptionToggle) Unfocus() {
	o.focused = false
}

func (o *OptionToggle) Toggle() {
	o.Value = !o.Value
}

func (o *OptionToggle) OnEnter() (int, bool, bool) {
	o.Toggle()
	return 0, false, true
}

func (o *OptionToggle) OnSpace() bool {
	o.Toggle()
	return true
}

func (o *OptionToggle) GetValue() bool {
	return o.Value
}

func (o *OptionToggle) View() string {
	var state string
	if o.Value {
		state = toggleOnStyle.Render("ON ")
	} else {
		state = toggleOffStyle.Render("OFF")
	}

	var label string
	if o.focused {
		label = optionLabelActiveStyle.Render(o.Label)
	} else {
		label = optionLabelStyle.Render(o.Label)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, state, label)
}

type OptionButton struct {
	Label   string
	Return  int
	focused bool
}

func (o *OptionButton) Init(width int) {
}

func (o *OptionButton) Footer() bool {
	return true
}

func (o *OptionButton) Update(msg tea.Msg) (cmd tea.Cmd) {
	return
}

func (o *OptionButton) Focused() bool {
	return o.focused
}

func (o *OptionButton) Focus() (cmd tea.Cmd) {
	o.focused = true
	return
}

func (o *OptionButton) Unfocus() {
	o.focused = false
}

func (o *OptionButton) OnEnter() (int, bool, bool) {
	return o.Return, true, true
}

func (o *OptionButton) OnSpace() bool {
	return false
}

func (o *OptionButton) View() string {
	if o.focused {
		return optionButtonActiveStyle.Render(o.Label)
	}
	return optionButtonStyle.Render(o.Label)
}
