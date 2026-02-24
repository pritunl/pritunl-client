package update

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
)

var (
	Upgrade  bool
	Security bool

	lastCheck time.Time
	client    = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 30 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			},
		},
		Timeout: 30 * time.Second,
	}
)

type updateRespData struct {
	Upgrade  bool `json:"upgrade"`
	Security bool `json:"security"`
}

func Check() (err error) {
	if time.Since(lastCheck) < 3*time.Hour {
		return
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://app.pritunl.com/update/%s",
			constants.Version,
		),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "update: Update error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-client")

	res, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "update: Update request error"),
		}
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = utils.LogRequestError(res, "")
		return
	}

	data := &updateRespData{}
	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "update: Failed to parse response body"),
		}
		return
	}

	Upgrade = data.Upgrade
	Security = data.Security
	lastCheck = time.Now()

	return
}
