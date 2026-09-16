package constants

import (
	"os"
)

const (
	Version        = "1.4.4752.50"
	MacosHelperDir = "/Library/PrivilegedHelperTools/pritunl-client"
)

var (
	Development = false
	Macos10     = false
	Interrupt   = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
