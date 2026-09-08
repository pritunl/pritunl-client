package constants

import (
	"os"
)

const (
	Version = "1.4.4744.47"
)

var (
	Development = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
