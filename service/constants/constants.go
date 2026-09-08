package constants

import (
	"os"
)

const (
	Version        = "1.4.4744.47"
	MacosHelperDir = "/Library/PrivilegedHelperTools/pritunl-client"
)

var (
	Development = false
	Macos10     = false
	Interrupt   = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
