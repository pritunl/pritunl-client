package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
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

func GetAuthPath() (pth string) {
	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(GetWinDrive(), "ProgramData", "Pritunl", "auth")
		break
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "run", "pritunl.auth")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}
