package connection

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/platform"
	"github.com/pritunl/pritunl-client/service/utils"
)

func GetBashPath() string {
	switch runtime.GOOS {
	case "darwin":
		if constants.Development {
			return filepath.Join(utils.GetRootDir(), "..",
				"wireguard_macos", "bash")
		}

		return filepath.Join(string(os.PathSeparator), "Applications",
			"Pritunl.app", "Contents", "Resources", "bash")
	case "linux":
		break
	case "windows":
		break
	default:
		panic("paths: Bash path not implemented")
	}

	return ""
}

func GetWgPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(utils.GetRootDir(), "wireguard", "wg.exe")
	case "darwin":
		if constants.Development {
			return filepath.Join(utils.GetRootDir(), "..",
				"wireguard_macos", "wg")
		}

		return filepath.Join(string(os.PathSeparator), "Applications",
			"Pritunl.app", "Contents", "Resources", "wg")
	case "linux":
		path, _ := exec.LookPath("wg")
		if path != "" {
			return path
		}

		break
	default:
		panic("paths: WG path not implemented")
	}

	return ""
}

func GetWgQuickPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(utils.GetRootDir(), "wireguard", "wg-quick.exe")
	case "darwin":
		if constants.Development {
			return filepath.Join(utils.GetRootDir(), "..",
				"wireguard_macos", "wg-quick")
		}

		return filepath.Join(string(os.PathSeparator), "Applications",
			"Pritunl.app", "Contents", "Resources", "wg-quick")
	case "linux":
		path := filepath.Join(string(os.PathSeparator),
			"usr", "bin", "wg-quick")
		exists, _ := utils.Exists(path)
		if exists {
			return path
		}

		path, _ = exec.LookPath("wg-quick")
		if path != "" {
			return path
		}

		break
	default:
		panic("paths: WG quick path not implemented")
	}

	return ""
}

func GetWgUtilPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(utils.GetRootDir(), "wireguard", "wireguard.exe")
	case "darwin":
		break
	case "linux":
		break
	default:
		panic("paths: WG util path not implemented")
	}

	return ""
}

func GetWgConfDir() (dir1 string, dir2 string, err error) {
	switch runtime.GOOS {
	case "windows":
		dir1, err = utils.GetTempDir()
		if err != nil {
			return
		}

		return
	case "darwin":
		dir1 = "/etc/wireguard"
		return
	case "linux":
		dir1 = "/etc/wireguard"
		return
	default:
		panic("paths: WG util path not implemented")
	}
}

func GetOvpnConfPath() (pth string, err error) {
	if runtime.GOOS == "windows" {
		pth = filepath.Join(utils.GetWinDrive(),
			"ProgramData", "Pritunl", "Temp")
		err = platform.MkdirSecure(pth)
		if err != nil {
			err = &utils.IoError{
				errors.Wrap(
					err, "utils: Failed to create temp directory"),
			}
			return
		}
	} else {
		pth = filepath.Join(string(filepath.Separator), "etc", "openvpn")
		exists, _ := utils.ExistsDir(pth)
		if exists {
			return
		}

		pth = filepath.Join(string(filepath.Separator), "tmp", "pritunl")
		err = platform.MkdirSecure(pth)
		if err != nil {
			err = &utils.IoError{
				errors.Wrap(
					err, "utils: Failed to create temp directory"),
			}
			return
		}
	}

	return
}

func GetOvpnDir() (pth string) {
	if constants.Development {
		switch runtime.GOOS {
		case "windows":
			if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
				pth = filepath.Join(utils.GetRootDir(),
					"..", "openvpn_win", "openvpn_arm64")
			} else {
				pth = filepath.Join(utils.GetRootDir(),
					"..", "openvpn_win", "openvpn_amd64")
			}
		case "darwin":
			pth = ""
		case "linux":
			pth = ""
		default:
			panic("paths: Ovpn dev dir not implemented")
		}

		return
	}

	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(utils.GetRootDir(), "openvpn")
	case "darwin":
		pth = ""
	case "linux":
		pth = ""
	default:
		panic("paths: Ovpn dir not implemented")
	}

	return
}

func GetOvpnPath() (pth string) {
	if constants.Development {
		switch runtime.GOOS {
		case "windows":
			if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
				pth = filepath.Join(utils.GetRootDir(), "..",
					"openvpn_win", "openvpn_arm64", "openvpn.exe")
			} else {
				pth = filepath.Join(utils.GetRootDir(), "..",
					"openvpn_win", "openvpn_amd64", "openvpn.exe")
			}
			break
		case "darwin":
			if constants.Macos10 {
				pth = filepath.Join(utils.GetRootDir(), "..",
					"openvpn_macos", "openvpn10")
			} else {
				pth = filepath.Join(utils.GetRootDir(), "..",
					"openvpn_macos", "openvpn_arm64")
			}
			break
		case "linux":
			pth = "openvpn"
			break
		default:
			panic("paths: Ovpn dev path not implemented")
		}

		return
	}

	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(utils.GetRootDir(), "openvpn", "openvpn.exe")
		break
	case "darwin":
		if constants.Macos10 {
			pth = filepath.Join(string(os.PathSeparator), "Applications",
				"Pritunl.app", "Contents", "Resources", "pritunl-openvpn10")
		} else {
			pth = filepath.Join(string(os.PathSeparator), "Applications",
				"Pritunl.app", "Contents", "Resources", "pritunl-openvpn")
		}
		break
	case "linux":
		pth = "openvpn"
		break
	default:
		panic("paths: Ovpn path not implemented")
	}

	return
}
