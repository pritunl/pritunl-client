package utils

import (
	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
)

func GetServiceLog() (data string, err error) {
	logPth := GetLogPath()

	exists, err := Exists(logPth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "log: Failed to check service log file"),
		}
		return
	}

	if exists {
		dataByt, e := ioutil.ReadFile(logPth)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "log: Failed to read service log file"),
			}
			return
		}

		data = string(dataByt)
	}

	return
}

func ClearServiceLog() (err error) {
	logPth := GetLogPath()

	err = CreateWriteLock(logPth, "", 0600)
	if err != nil {
		return
	}

	return
}
