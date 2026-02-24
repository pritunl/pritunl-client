package config

import (
	"path/filepath"
	"runtime"

	"github.com/pritunl/pritunl-client/service/utils"
)

func FindPath() (pth string, exists, move bool, err error) {
	pth = GetPath()

	pthExists, err := utils.Exists(pth)
	if err != nil {
		return
	}

	if pthExists {
		exists = true
		return
	}

	oldPth := GetPathOld()
	if oldPth != "" {
		pthExists, err = utils.Exists(oldPth)
		if err != nil {
			return
		}

		if pthExists {
			exists = true
			move = true
			pth = oldPth
			return
		}
	}

	return
}

func GetPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(utils.GetWinDrive(), "ProgramData",
			"Pritunl", "pritunl-client.json")
	case "darwin":
		return filepath.Join("/", "Library",
			"Application Support", "Pritunl", "pritunl-client.json")
	case "linux":
		return filepath.Join("/", "var",
			"lib", "pritunl-client", "pritunl-client.json")
	default:
		panic("profile: Not implemented")
	}
}

func GetPathOld() string {
	switch runtime.GOOS {
	case "windows":
		return ""
	case "darwin":
		return filepath.Join("/", "var",
			"lib", "pritunl-client", "pritunl-client.json")
	case "linux":
		return ""
	default:
		panic("profile: Not implemented")
	}
}
