package auth

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/platform"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/utils"
)

var Key = ""

func Init() (err error) {
	pth := utils.GetAuthPath()

	if runtime.GOOS == "windows" {
		tempPth := filepath.Join(
			utils.GetWinDrive(), "ProgramData", "Pritunl", "Temp")

		exists, _ := utils.ExistsDir(tempPth)
		if exists {
			_ = utils.RemoveAll(tempPth)
		}

		err = platform.MkdirSecure(tempPth)
		if err != nil {
			err = &WriteError{
				errors.Wrap(
					err, "utils: Failed to create temp directory"),
			}
			return
		}

		dataPth := filepath.Join(utils.GetWinDrive(), "ProgramData", "Pritunl")
		err = platform.MkdirReadSecure(dataPth)
		if err != nil {
			err = &WriteError{
				errors.Wrap(
					err, "utils: Failed to create data directory"),
			}
			return
		}

		profilesPth := filepath.Join(utils.GetWinDrive(), "ProgramData",
			"Pritunl", "Profiles")
		err = platform.MkdirSecure(profilesPth)
		if err != nil {
			err = &WriteError{
				errors.Wrap(
					err, "utils: Failed to create profiles directory"),
			}
			return
		}
	} else {
		prflsPath := sprofile.GetPath()

		err = platform.MkdirLinkedSecure(prflsPath)
		if err != nil {
			err = &WriteError{
				errors.Wrap(
					err, "utils: Failed to create profiles directory"),
			}
			return
		}
	}

	if _, e := os.Stat(pth); os.IsNotExist(e) {
		Key, err = utils.RandStr(64)
		if err != nil {
			return
		}

		err = ioutil.WriteFile(pth, []byte(Key), os.FileMode(0644))
		if err != nil {
			err = &WriteError{
				errors.Wrap(err, "auth: Failed to auth key"),
			}
			return
		}
	} else {
		data, e := ioutil.ReadFile(pth)
		if e != nil {
			err = &WriteError{
				errors.Wrap(e, "auth: Failed to auth key"),
			}
			return
		}

		Key = strings.TrimSpace(string(data))

		if Key == "" {
			err = os.Remove(pth)
			if err != nil {
				err = &WriteError{
					errors.Wrap(err, "auth: Failed to reset auth key"),
				}
				return
			}
			Init()
		}
	}

	return
}
