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

const (
	UnixSocket = "/var/run/pritunl.sock"
	TcpAddress = "127.0.0.1:9770"
)

const pollTimeout = 5 * time.Second

var unixTransport = &http.Transport{
	DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", UnixSocket)
	},
}

var httpClient = &http.Client{
	Timeout: 1 * time.Minute,
}

var unixClient = &http.Client{
	Timeout:   1 * time.Minute,
	Transport: unixTransport,
}

var httpPollClient = &http.Client{
	Timeout: pollTimeout,
}

var unixPollClient = &http.Client{
	Timeout:   pollTimeout,
	Transport: unixTransport,
}

func GetAddress() string {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return "http://unix"
	} else {
		return "http://" + TcpAddress
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

func GetPollClient() *http.Client {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return unixPollClient
	} else {
		return httpPollClient
	}
}
