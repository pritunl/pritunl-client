package iface

import (
	"fmt"
)

func renderCol(width int, format string, args ...interface{}) string {
	data := []rune(fmt.Sprintf(format, args...))
	if width <= 0 {
		return ""
	}
	if len(data) <= width {
		return string(data)
	}
	if width < 4 {
		return string(data[:width])
	}
	return string(data[:width-3]) + "..."
}
