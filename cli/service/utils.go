package service

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/utils"
)

var httpClient = &http.Client{
	Timeout: 1 * time.Minute,
}

var unixClient = &http.Client{
	Timeout: 1 * time.Minute,
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/pritunl.sock")
		},
	},
}

func GetAddress() string {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return "http://unix"
	} else {
		return "http://127.0.0.1:9770"
	}
}

func GetAuthKey() (key string, err error) {
	pth := utils.GetAuthPath()

	data, err := ioutil.ReadFile(pth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "auth: Failed to auth key"),
		}
		return
	}

	key = strings.TrimSpace(string(data))

	return
}

func GetClient() *http.Client {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return unixClient
	} else {
		return httpClient
	}
}
