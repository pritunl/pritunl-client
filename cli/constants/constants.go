package constants

import (
	"os"
)

const (
	Version = "1.3.4729.52"
)

var (
	Development = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
