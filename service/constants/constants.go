package constants

import (
	"os"
)

const (
	Version        = "1.3.4729.52"
	MacosHelperDir = "/Library/PrivilegedHelperTools/pritunl-client"
)

var (
	Development = false
	Macos10     = false
	Interrupt   = false
	Flatpak     = os.Getenv("FLATPAK_MODE") == "true"
)
