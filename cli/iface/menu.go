package iface

import (
	"strings"

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

func renderMenuBar(width int, items []MenuItem) string {
	parts := []string{}
	used := 0

	for _, item := range items {
		text := menuKeyStyle.Render(" ["+item.Key+"]") +
			menuItemStyle.Render(" "+item.Title+" ")
		w := lipgloss.Width(text)
		if used+w > width && len(parts) > 0 {
			break
		}
		parts = append(parts, text)
		used += w
	}

	bar := strings.Join(parts, "")
	if used < width {
		bar += menuBarStyle.Render(strings.Repeat(" ", width-used))
	}

	return bar
}
