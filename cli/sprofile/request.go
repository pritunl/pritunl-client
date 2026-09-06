package sprofile

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/service"
)

// serviceRequest sends an authenticated request to the service, data is
// sent as a json body when not nil.
func serviceRequest(client *http.Client, method, pth string,
	data interface{}) (resp *http.Response, err error) {

	reqUrl := service.GetAddress() + pth

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	var body io.Reader
	if data != nil {
		buf, e := json.Marshal(data)
		if e != nil {
			err = errortypes.RequestError{
				errors.Wrap(e, "sprofile: Json marshal error"),
			}
			return
		}
		body = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: %s request failed", method),
		}
		return
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		req.Host = "unix"
	}
	req.Header.Set("Auth-Key", authKey)
	req.Header.Set("User-Agent", "pritunl")
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err = client.Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}

	return
}

// serviceCall sends a request that must return status 200, the response
// body is discarded.
func serviceCall(method, pth string, data interface{}) (err error) {
	resp, err := serviceRequest(service.GetClient(), method, pth, data)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}

// serviceGet reads the response body of a request that must return
// status 200.
func serviceGet(pth string) (body []byte, err error) {
	resp, err := serviceRequest(service.GetClient(), "GET", pth, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrap(err, "sprofile: Failed to read response"),
		}
		return
	}

	return
}

// serviceJson decodes the json response of a request that must return
// status 200.
func serviceJson(client *http.Client, method, pth string,
	data interface{}, out interface{}) (err error) {

	resp, err := serviceRequest(client, method, pth, data)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf("sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to parse response"),
		}
		return
	}

	return
}
