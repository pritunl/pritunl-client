package iface

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	menuBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3B82F6"))
	menuItemStyle = menuBarStyle
	menuKeyStyle  = menuBarStyle.Bold(true)
)

type MenuItem struct {
	Title string
	Key   string
}

// KeyMsg returns the key press the menu item represents so clicking the
// item runs the same action as the key. Items that only describe
// navigation keys have no click action.
func (i MenuItem) KeyMsg() (tea.KeyMsg, bool) {
	switch i.Key {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}, true
	case "←/→":
		return tea.KeyMsg{Type: tea.KeyRight}, true
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}, true
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}, true
	}

	runes := []rune(i.Key)
	if len(runes) == 1 {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}, true
	}

	return tea.KeyMsg{}, false
}

// menuPart is a rendered menu item and its column position in the bar.
type menuPart struct {
	item  MenuItem
	text  string
	x     int
	width int
}

// menuBarParts renders the items that fit in the bar width.
func menuBarParts(width int, items []MenuItem) []menuPart {
	parts := []menuPart{}
	used := 0

	for _, item := range items {
		text := menuKeyStyle.Render(" ["+item.Key+"]") +
			menuItemStyle.Render(" "+item.Title+" ")
		w := lipgloss.Width(text)
		if used+w > width && len(parts) > 0 {
			break
		}
		parts = append(parts, menuPart{
			item:  item,
			text:  text,
			x:     used,
			width: w,
		})
		used += w
	}

	return parts
}

func renderMenuBar(width int, items []MenuItem) string {
	parts := menuBarParts(width, items)

	used := 0
	texts := []string{}
	for _, part := range parts {
		texts = append(texts, part.text)
		used += part.width
	}

	bar := strings.Join(texts, "")
	if used < width {
		bar += menuBarStyle.Render(strings.Repeat(" ", width-used))
	}

	return bar
}

// menuBarClick returns the key press for the menu item at the column.
func menuBarClick(width int, items []MenuItem, x int) (tea.KeyMsg, bool) {
	for _, part := range menuBarParts(width, items) {
		if x >= part.x && x < part.x+part.width {
			return part.item.KeyMsg()
		}
	}
	return tea.KeyMsg{}, false
}

// isLeftClick returns true for a left mouse button press.
func isLeftClick(msg tea.MouseMsg) bool {
	return msg.Action == tea.MouseActionPress &&
		msg.Button == tea.MouseButtonLeft
}
