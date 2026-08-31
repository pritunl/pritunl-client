package logs

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/service"
	"github.com/pritunl/pritunl-client/cli/utils"
)

// ClientLogPath returns the path of the local client log written by the
// command line interface.
func ClientLogPath() string {
	return filepath.Join(utils.GetDataPath(), "pritunl.log")
}

// GetServiceLog reads the service log from the service.
func GetServiceLog() (data string, err error) {
	reqUrl := service.GetAddress() + "/log/service"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "logs: Get request failed"),
		}
		return
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		req.Host = "unix"
	}
	req.Header.Set("Auth-Key", authKey)
	req.Header.Set("User-Agent", "pritunl")

	resp, err := service.GetClient().Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "logs: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("logs: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrap(err, "logs: Failed to read response"),
		}
		return
	}

	data = strings.TrimSpace(string(body)) + "\n"

	return
}

// ClearServiceLog clears the service log.
func ClearServiceLog() (err error) {
	reqUrl := service.GetAddress() + "/log/service"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	req, err := http.NewRequest("DELETE", reqUrl, nil)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "logs: Delete request failed"),
		}
		return
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		req.Host = "unix"
	}
	req.Header.Set("Auth-Key", authKey)
	req.Header.Set("User-Agent", "pritunl")

	resp, err := service.GetClient().Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "logs: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("logs: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}

// GetClientLog reads the local client log file.
func GetClientLog() (data string, err error) {
	body, err := ioutil.ReadFile(ClientLogPath())
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		err = errortypes.ReadError{
			errors.Wrap(err, "logs: Failed to read client log"),
		}
		return
	}

	data = strings.TrimSpace(string(body)) + "\n"

	return
}

// ClearClientLog truncates the local client log file.
func ClearClientLog() (err error) {
	err = os.WriteFile(ClientLogPath(), []byte{}, 0644)
	if err != nil {
		err = errortypes.WriteError{
			errors.Wrap(err, "logs: Failed to clear client log"),
		}
		return
	}

	return
}
