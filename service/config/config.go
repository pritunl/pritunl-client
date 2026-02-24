package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

var (
	Config            = &ConfigData{}
	StaticRoot        = ""
	StaticTestingRoot = ""
)

type ConfigData struct {
	path              string `json:"-"`
	loaded            bool   `json:"-"`
	DisableDnsWatch   bool   `json:"disable_dns_watch"`
	EnableDnsRefresh  bool   `json:"enable_dns_refresh"`
	DisableWakeWatch  bool   `json:"disable_wake_watch"`
	DisableNetClean   bool   `json:"disable_net_clean"`
	DisableWgDns      bool   `json:"disable_wg_dns"`
	ForceLocalTpm     bool   `json:"force_local_tpm"`
	InterfaceMetric   int    `json:"interface_metric"`
	EnclavePrivateKey string `json:"enclave_private_key"`
}

func (c *ConfigData) Save() (err error) {
	if !c.loaded {
		err = &errortypes.WriteError{
			errors.New("config: Config file has not been loaded"),
		}
		return
	}

	pth := GetPath()

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File marshal error"),
		}
		return
	}

	err = utils.ExistsMkdir(filepath.Dir(pth), 0755)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(pth, data, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "config: File write error"),
		}
		return
	}

	return
}

func Load() (err error) {
	data := &ConfigData{}

	pth, exists, move, err := FindPath()
	if err != nil {
		return
	}

	if !exists {
		err = nil
		data.loaded = true
		Config = data
		return
	}

	file, err := ioutil.ReadFile(pth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File read error"),
		}
		return
	}

	err = json.Unmarshal(file, data)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "config: File unmarshal error"),
		}
		return
	}

	data.loaded = true

	Config = data

	if move {
		newPath := GetPath()

		logrus.WithFields(logrus.Fields{
			"old_path": pth,
			"new_path": newPath,
		}).Info("config: Moving config path")

		err = Save()
		if err != nil {
			return
		}
	}

	return
}

func Save() (err error) {
	err = Config.Save()
	if err != nil {
		return
	}

	return
}
