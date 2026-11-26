package constants

import (
	"os"
)

const (
	Version = "1.4.4823.50"
)

var (
	Development = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
