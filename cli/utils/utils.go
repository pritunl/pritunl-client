package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/constants"
	"github.com/pritunl/pritunl-client/cli/errortypes"
)

const (
	FlatpakDefaultId = "com.pritunl.Client"
)

func Uuid() (id string) {
	idByte := make([]byte, 16)

	_, err := rand.Read(idByte)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to get random data"),
		}
		panic(err)
	}

	id = hex.EncodeToString(idByte[:])

	return
}

func GetWinDrive() string {
	systemDrv := os.Getenv("SYSTEMDRIVE")
	if systemDrv == "" {
		return "C:\\"
	}
	return systemDrv + "\\"
}

func GetRootDir() (pth string) {
	pth, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	return
}

func GetDataPath() (pth string) {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, "AppData", "Roaming")
		}
		pth = filepath.Join(base, "pritunl")
	case "darwin":
		home, _ := os.UserHomeDir()
		pth = filepath.Join(home, "Library", "Application Support", "pritunl")
	default:
		base := os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".config")
		}
		pth = filepath.Join(base, "pritunl")
	}

	return
}

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

	return
}

func GetSockPath() string {
	if IsFlatpak() {
		return filepath.Join(GetFlatpakRunDir(), "pritunl.sock")
	}
	return filepath.Join(string(filepath.Separator),
		"var", "run", "pritunl.sock")
}

func GetAuthPath() (pth string) {
	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(GetWinDrive(), "ProgramData", "Pritunl", "auth")
		break
	case "linux", "darwin":
		if IsFlatpak() {
			pth = filepath.Join(GetFlatpakRunDir(), "pritunl.auth")
			break
		}

		pth = filepath.Join(string(filepath.Separator),
			"var", "run", "pritunl.auth")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}
