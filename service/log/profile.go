package log

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
)

func getPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(utils.GetWinDrive(), "ProgramData",
			"Pritunl", "Profiles")
	case "darwin":
		return filepath.Join("/", "var",
			"lib", "pritunl-client", "profiles")
	case "linux":
		return filepath.Join("/", "var",
			"lib", "pritunl-client", "profiles")
	default:
		panic("profile: Not implemented")
	}
}

func ProfilePushLog(prflId string, output string) (err error) {
	prflsPath := getPath()
	logPth1 := filepath.Join(prflsPath, prflId+".log")
	logPth2 := logPth1 + ".1"

	file, err := os.OpenFile(logPth1,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "log: Failed to open profile log file"),
		}
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "log: Failed to stat profile log file"),
		}
		return
	}

	if stat.Size() >= 200000 {
		file.Close()

		os.Remove(logPth2)
		err = os.Rename(logPth1, logPth2)
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "log: Failed to rotate profile log file"),
			}
			return
		}

		file, err = os.OpenFile(logPth1,
			os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "log: Failed to open profile log file"),
			}
			return
		}
	}

	_, err = file.Write([]byte(output + "\n"))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "log: Failed to write to profile log file"),
		}
		return
	}

	return
}

func GetProfileLog(prflId string) (data string, err error) {
	prflsPath := getPath()
	logPth := filepath.Join(prflsPath, prflId+".log")

	exists, err := utils.Exists(logPth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "log: Failed to check profile log file"),
		}
		return
	}

	if exists {
		dataByt, e := ioutil.ReadFile(logPth)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "log: Failed to read profile log file"),
			}
			return
		}

		data = string(dataByt)
	}

	return
}

func ClearProfileLog(prflId string) (err error) {
	prflsPath := getPath()
	logPth := filepath.Join(prflsPath, prflId+".log")

	os.Remove(logPth)

	return
}
