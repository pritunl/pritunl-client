package iface

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	// Destructive buttons are red like the card disconnect button
	optionButtonDangerStyle = optionButtonStyle.
				Background(lipgloss.Color("#EF4444"))
	optionButtonDangerActiveStyle = optionButtonActiveStyle.
					Foreground(lipgloss.Color("#EF4444"))

	optionLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))
	optionLabelActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Bold(true)
	optionErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))

	optionLinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Underline(true)
	optionLinkActiveStyle = optionLinkStyle.
				Foreground(lipgloss.Color("#93C5FD")).
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

// optionInputStyles returns the text input styles matching the dialog
// colors, the prompt follows the label focus color.
func optionInputStyles() textinput.Styles {
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = optionLabelActiveStyle
	styles.Blurred.Prompt = optionLabelStyle
	styles.Focused.Placeholder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))
	styles.Blurred.Placeholder = styles.Focused.Placeholder
	return styles
}

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

	// Validate checks the value as it is typed, the error is shown under
	// the input and blocks the dialog from being accepted.
	Validate func(string) error

	model textinput.Model
}

func (o *OptionText) Init(width int) {
	o.model = textinput.New()
	o.model.Placeholder = o.Placeholder
	o.model.CharLimit = 2048
	o.model.Prompt = "> "
	o.model.Validate = o.Validate
	o.model.SetStyles(optionInputStyles())
	o.SetWidth(width)
	if o.Password {
		o.model.EchoMode = textinput.EchoPassword
		o.model.EchoCharacter = '•'
	}
	if o.Value != "" {
		o.model.SetValue(o.Value)
	}
}

// SetWidth sets the input width for the dialog content width.
func (o *OptionText) SetWidth(width int) {
	o.model.SetWidth(max(width-4, 10))
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

// Err returns the validation error of the current value.
func (o *OptionText) Err() error {
	return o.model.Err
}

func (o *OptionText) View() string {
	var label string
	if o.Focused() {
		label = optionLabelActiveStyle.Render(o.Label)
	} else {
		label = optionLabelStyle.Render(o.Label)
	}

	parts := []string{label, o.model.View()}
	if err := o.model.Err; err != nil {
		parts = append(parts, optionErrorStyle.Render(err.Error()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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
	Label  string
	Return int

	// Copy is copied to the clipboard instead of closing the dialog when
	// the button is activated.
	Copy string

	focused bool
	copied  bool
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
	o.copied = false
}

func (o *OptionButton) OnEnter() (int, bool, bool) {
	if o.Copy != "" {
		return 0, false, true
	}
	return o.Return, true, true
}

// copyCmd copies the button text to the clipboard and marks the button as
// copied until it loses focus.
func (o *OptionButton) copyCmd() tea.Cmd {
	if o.Copy == "" {
		return nil
	}
	o.copied = true
	return copyText(o.Copy)
}

func (o *OptionButton) label() string {
	if o.copied {
		return "Copied"
	}
	return o.Label
}

func (o *OptionButton) OnSpace() bool {
	return false
}

// Danger returns true for buttons that remove or clear something.
func (o *OptionButton) Danger() bool {
	label := strings.ToLower(o.Label)
	return strings.Contains(label, "remove") ||
		strings.Contains(label, "delete") ||
		strings.Contains(label, "clear")
}

func (o *OptionButton) View() string {
	label := o.label()

	if o.Danger() {
		if o.focused {
			return optionButtonDangerActiveStyle.Render(label)
		}
		return optionButtonDangerStyle.Render(label)
	}

	if o.focused {
		return optionButtonActiveStyle.Render(label)
	}
	return optionButtonStyle.Render(label)
}

// OptionLink is a URL shown as a terminal hyperlink, activating or
// clicking the link opens it in the default browser. The hyperlink is
// declared with OSC 8 so terminals treat the full URL as one link even
// when it wraps onto multiple lines.
type OptionLink struct {
	Url string

	width   int
	focused bool
	err     error
}

func (o *OptionLink) Init(width int) {
	o.SetWidth(width)
}

// SetWidth sets the wrap width for the dialog content width.
func (o *OptionLink) SetWidth(width int) {
	o.width = max(width, 10)
}

// MinWidth returns the content width needed to show the link on one line.
func (o *OptionLink) MinWidth() int {
	return lipgloss.Width(o.Url)
}

func (o *OptionLink) Footer() bool {
	return false
}

func (o *OptionLink) Update(msg tea.Msg) (cmd tea.Cmd) {
	return
}

func (o *OptionLink) Focused() bool {
	return o.focused
}

func (o *OptionLink) Focus() (cmd tea.Cmd) {
	o.focused = true
	return
}

func (o *OptionLink) Unfocus() {
	o.focused = false
}

func (o *OptionLink) OnEnter() (int, bool, bool) {
	o.err = openUrl(o.Url)
	return 0, false, true
}

func (o *OptionLink) OnSpace() bool {
	return false
}

func (o *OptionLink) View() string {
	style := optionLinkStyle
	if o.focused {
		style = optionLinkActiveStyle
	}

	// The id joins the wrapped segments into one hover target
	style = style.Hyperlink(o.Url, "id=link")

	// Wrap the URL by hand so the lines are not padded, padding would
	// underline and link the blank space after short lines.
	lines := []string{}
	for _, chunk := range chunkRunes(o.Url, o.width) {
		lines = append(lines, style.Render(chunk))
	}
	if o.err != nil {
		lines = append(lines,
			optionErrorStyle.Width(o.width).Render(errorMessage(o.err)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// chunkRunes splits the text into pieces of at most width runes.
func chunkRunes(text string, width int) []string {
	runes := []rune(text)
	chunks := []string{}
	for len(runes) > width {
		chunks = append(chunks, string(runes[:width]))
		runes = runes[width:]
	}
	return append(chunks, string(runes))
}
