package iface

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Connect    key.Binding
	ConnectWg  key.Binding
	Disconnect key.Binding
	Import     key.Binding
	Logs       key.Binding
	Settings   key.Binding
	Remove     key.Binding
	Refresh    key.Binding
	Quit       key.Binding
}

var bindings = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Import: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "import"),
	),
	Connect: key.NewBinding(
		key.WithKeys("c", "enter"),
		key.WithHelp("c", "connect"),
	),
	ConnectWg: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "connect wireguard"),
	),
	Disconnect: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "disconnect"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
	Settings: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "settings"),
	),
	Remove: key.NewBinding(
		key.WithKeys("r", "delete"),
		key.WithHelp("r", "remove"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R", "ctrl+r"),
		key.WithHelp("R", "refresh"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type LogsKeyMap struct {
	Back  key.Binding
	Clear key.Binding
	Top   key.Binding
	End   key.Binding
	Quit  key.Binding
}

var logsKeys = LogsKeyMap{
	Back: key.NewBinding(
		key.WithKeys("esc", "q", "l", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear"),
	),
	Top: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home", "top"),
	),
	End: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end", "end"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}
