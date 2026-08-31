package config

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/service"
)

// Config holds the global service advanced settings.
type Config struct {
	EnableDnsRefresh bool `json:"enable_dns_refresh"`
	DisableDnsWatch  bool `json:"disable_dns_watch"`
	DisableWgDns     bool `json:"disable_wg_dns"`
	DisableWakeWatch bool `json:"disable_wake_watch"`
	DisableNetClean  bool `json:"disable_net_clean"`
	DisableBrowser   bool `json:"disable_browser"`
	InterfaceMetric  int  `json:"interface_metric"`
}

// Get loads the global settings from the service.
func Get() (conf *Config, err error) {
	reqUrl := service.GetAddress() + "/config"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "config: Get request failed"),
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
			errors.Wrap(err, "config: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("config: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	conf = &Config{}
	err = json.NewDecoder(resp.Body).Decode(conf)
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "config: Failed to parse response"),
		}
		return
	}

	return
}

// Commit saves the global settings to the service.
func (c *Config) Commit() (err error) {
	reqUrl := service.GetAddress() + "/config"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	data, err := json.Marshal(c)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "config: Json marshal error"),
		}
		return
	}

	req, err := http.NewRequest("PUT", reqUrl, bytes.NewBuffer(data))
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "config: Put request failed"),
		}
		return
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		req.Host = "unix"
	}
	req.Header.Set("Auth-Key", authKey)
	req.Header.Set("User-Agent", "pritunl")
	req.Header.Set("Content-Type", "application/json")

	resp, err := service.GetClient().Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "config: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("config: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}
