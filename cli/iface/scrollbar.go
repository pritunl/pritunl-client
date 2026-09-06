package iface

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Columns used by the scrollbar and its gap beside a viewport.
	scrollbarWidth = 2
)

var (
	scrollbarTrackStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#374151"))
	scrollbarThumbStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6"))
)

// renderScrollbar renders a vertical scrollbar matching the viewport
// height, the thumb size and position reflect the visible portion. The
// column is blank when the content fits in the viewport.
func renderScrollbar(view viewport.Model) string {
	height := view.Height
	total := view.TotalLineCount()

	if height <= 0 {
		return ""
	}

	if total <= height {
		return strings.TrimRight(strings.Repeat(" \n", height), "\n")
	}

	thumb := max(height*height/total, 1)
	track := height - thumb
	pos := 0
	if track > 0 {
		pos = view.YOffset * track / (total - height)
		if view.AtBottom() {
			pos = track
		}
		pos = min(pos, track)
	}

	lines := make([]string, 0, height)
	for i := 0; i < height; i++ {
		if i >= pos && i < pos+thumb {
			lines = append(lines, scrollbarThumbStyle.Render("┃"))
		} else {
			lines = append(lines, scrollbarTrackStyle.Render("│"))
		}
	}

	return strings.Join(lines, "\n")
}

// renderScrollView renders the viewport with a scrollbar on the right, the
// viewport width must already leave scrollbarWidth columns for the bar.
func renderScrollView(view viewport.Model) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		view.View(),
		lipgloss.NewStyle().MarginLeft(1).Render(renderScrollbar(view)),
	)
}
