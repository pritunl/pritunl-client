package service

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/platform"
	"github.com/pritunl/pritunl-client/cli/utils"
)

const (
	serviceBin          = "pritunl-client-service"
	serviceStartTimeout = 15 * time.Second
	serviceStartPoll    = 200 * time.Millisecond
)

func Ping() bool {
	authKey, err := GetAuthKey()
	if err != nil {
		return false
	}

	req, err := http.NewRequest("GET", GetAddress()+"/ping", nil)
	if err != nil {
		return false
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		req.Host = "unix"
	}
	req.Header.Set("Auth-Key", authKey)
	req.Header.Set("User-Agent", "pritunl")

	resp, err := GetPollClient().Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func findServiceBin() (pth string, err error) {
	exe, e := os.Executable()
	if e == nil {
		pth = filepath.Join(filepath.Dir(exe), serviceBin)
		_, e = os.Stat(pth)
		if e == nil {
			return
		}
	}

	pth, e = exec.LookPath(serviceBin)
	if e != nil {
		err = errortypes.NotFoundError{
			errors.Wrap(e, "service: Background service binary not found"),
		}
		return
	}

	return
}

// EnsureRunning starts the background service in Flatpak mode when it is
// not reachable. The Flatpak launcher starts the service with the desktop
// client, this covers running the command line client directly. The
// service exits on its own when another instance holds the instance lock.
func EnsureRunning() (err error) {
	if !utils.IsFlatpak() {
		return
	}

	if Ping() {
		return
	}

	pth, err := findServiceBin()
	if err != nil {
		return
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrap(err, "service: Failed to open null device"),
		}
		return
	}
	defer devNull.Close()

	cmd := exec.Command(pth)
	cmd.Env = append(os.Environ(), "FLATPAK_MODE=true")
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.SysProcAttr = platform.DetachAttr()

	err = cmd.Start()
	if err != nil {
		err = errortypes.ExecError{
			errors.Wrap(err, "service: Failed to start background service"),
		}
		return
	}

	go func() {
		_ = cmd.Wait()
	}()

	deadline := time.Now().Add(serviceStartTimeout)
	for time.Now().Before(deadline) {
		if Ping() {
			return
		}
		time.Sleep(serviceStartPoll)
	}

	err = errortypes.ExecError{
		errors.New("service: Background service failed to start, " +
			"check the service log"),
	}
	return
}
