package constants

import (
	"os"
)

const (
	Version = "1.4.4752.50"
)

var (
	Development = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
