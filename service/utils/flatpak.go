package utils

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/platform"
)

const (
	FlatpakDefaultId = "com.pritunl.Client"
)

func IsFlatpak() bool {
	return runtime.GOOS == "linux" && constants.Flatpak
}

func GetFlatpakId() string {
	id := os.Getenv("FLATPAK_ID")
	if id == "" {
		id = FlatpakDefaultId
	}
	return id
}

func GetFlatpakRunDir() string {
	return filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "app", GetFlatpakId())
}

func GetFlatpakConfigDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "pritunl", "service")
}

func FlatpakInit() (err error) {
	if !IsFlatpak() {
		return
	}

	if os.Getenv("XDG_RUNTIME_DIR") == "" {
		err = &errortypes.ReadError{
			errors.New("utils: XDG_RUNTIME_DIR not defined in flatpak mode"),
		}
		return
	}

	err = platform.MkdirSecure(GetFlatpakRunDir())
	if err != nil {
		return
	}

	err = platform.MkdirSecure(GetFlatpakConfigDir())
	if err != nil {
		return
	}

	return
}

func GetSockPath() string {
	if IsFlatpak() {
		return filepath.Join(GetFlatpakRunDir(), "pritunl.sock")
	}
	return filepath.Join(string(filepath.Separator),
		"var", "run", "pritunl.sock")
}
