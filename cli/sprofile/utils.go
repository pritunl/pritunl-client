package sprofile

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/profile"
	"github.com/pritunl/pritunl-client/cli/service"
	"github.com/pritunl/pritunl-client/cli/terminal"
	"github.com/pritunl/pritunl-client/cli/utils"
	"github.com/spf13/cobra"
)

var (
	clientSecure = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 12 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			},
		},
		Timeout: 12 * time.Second,
	}
	clientInsecure = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 12 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			},
		},
		Timeout: 12 * time.Second,
	}
	ip4reg = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	ip6reg = regexp.MustCompile("/\\[[a-fA-F0-9:]*\\]/")
)

type SprofileData struct {
	Id                 string `json:"id"`
	Mode               string `json:"mode"`
	Disabled           bool   `json:"disabled"`
	OrgId              string `json:"org_id"`
	UserId             string `json:"user_id"`
	ServerId           string `json:"server_id"`
	SyncToken          string `json:"sync_token"`
	SyncSecret         string `json:"sync_secret"`
	Data               string `json:"data"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	ServerPublicKey    string `json:"server_public_key"`
	ServerBoxPublicKey string `json:"server_box_public_key"`
	TokenTtl           int    `json:"token_ttl"`
	Reconnect          bool   `json:"reconnect"`
	Timeout            bool   `json:"timeout"`
}

func Match(sprflId string) (sprfl *Sprofile, err error) {
	sprfls, err := GetAll()
	if err != nil {
		return
	}

	for _, spfl := range sprfls {
		if sprflId == spfl.Id {
			sprfl = spfl
		} else if len(sprflId) <= len(spfl.Id) &&
			spfl.Id[:len(sprflId)] == sprflId {

			if sprfl != nil {
				err = errortypes.NotFoundError{
					errors.New("sprofile: Profile duplicate match"),
				}
				return
			}
			sprfl = spfl
		}
	}

	if sprfl == nil {
		err = errortypes.NotFoundError{
			errors.New("sprofile: Profile not found"),
		}
		return
	}

	return
}

func Stop(sprflId string) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	reqUrl := service.GetAddress() + "/profile"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	data, err := json.Marshal(&SprofileData{
		Id: sprfl.Id,
	})
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Json marshal error"),
		}
		return
	}

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("DELETE", reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Delete request failed"),
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}

func Delete(sprflId string) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	reqUrl := service.GetAddress() + "/sprofile"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	data, err := json.Marshal(&SprofileData{
		Id: sprfl.Id,
	})
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Json marshal error"),
		}
		return
	}

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("DELETE", reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Delete request failed"),
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}

func GetAll() (sprfls Sprofiles, err error) {
	reqUrl := service.GetAddress() + "/sprofile"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Get request failed"),
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	sprfls = []*Sprofile{}
	err = json.NewDecoder(resp.Body).Decode(&sprfls)
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to parse response"),
		}
		return
	}

	sprflsMap := map[string]*Sprofile{}
	for _, sprfl := range sprfls {
		sprflsMap[sprfl.Id] = sprfl
	}

	prfls, err := profile.GetAll()
	if err != nil {
		return
	}

	for _, prfl := range prfls {
		sprfl := sprflsMap[prfl.Id]
		if sprfl != nil {
			sprfl.Profile = prfl
		}
	}

	sprfls.Sort()

	return
}

func PasswordPrompt(sprfl *Sprofile) (pass string, err error) {
	passModes := set.NewSet()

	passModesStr := strings.Split(sprfl.PasswordMode, "_")
	for _, passMode := range passModesStr {
		passModes.Add(passMode)
	}

	if passModes.Contains("pin") {
		part := terminal.ReadPassword("Pin")
		if part == "" {
			cobra.CheckErr("sprofile: Pin is empty")
		}
		pass += part
	}

	if passModes.Contains("duo") {
		part := terminal.ReadPassword("Duo Passcode")
		if part == "" {
			cobra.CheckErr("sprofile: Duo Passcode is empty")
		}
		pass += part
	}

	if passModes.Contains("onelogin") {
		part := terminal.ReadPassword("OneLogin Passcode")
		if part == "" {
			cobra.CheckErr("sprofile: OneLogin Passcode is empty")
		}
		pass += part
	}

	if passModes.Contains("okta") {
		part := terminal.ReadPassword("Okta Passcode")
		if part == "" {
			cobra.CheckErr("sprofile: Okta Passcode is empty")
		}
		pass += part
	}

	if passModes.Contains("otp") {
		part := terminal.ReadPassword("Authenticator Passcode")
		if part == "" {
			cobra.CheckErr("sprofile: Authenticator Passcode is empty")
		}
		pass += part
	}

	if passModes.Contains("yubikey") {
		part := terminal.ReadPassword("YubiKey")
		if part == "" {
			cobra.CheckErr("sprofile: YubiKey is empty")
		}
		pass += part
	}

	if pass == "" {
		part := terminal.ReadPassword("Password")
		if part == "" {
			cobra.CheckErr("sprofile: Password is empty")
		}
		pass += part
	}

	return
}

func PasswordPrompts(sprfl *Sprofile) (prompts []Prompt) {
	passModes := set.NewSet()

	passModesStr := strings.Split(sprfl.PasswordMode, "_")
	for _, passMode := range passModesStr {
		passModes.Add(passMode)
	}

	if passModes.Contains("pin") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "pin",
			Label:       "Pin",
			Placeholder: "Enter pin...",
		})
	}

	if passModes.Contains("duo") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "duo",
			Label:       "Duo Passcode",
			Placeholder: "Enter passcode...",
		})
	}

	if passModes.Contains("onelogin") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "onelogin",
			Label:       "OneLogin Passcode",
			Placeholder: "Enter passcode...",
		})
	}

	if passModes.Contains("okta") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "okta",
			Label:       "Okta Passcode",
			Placeholder: "Enter passcode...",
		})
	}

	if passModes.Contains("otp") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "otp",
			Label:       "Authenticator Passcode",
			Placeholder: "Enter passcode...",
		})
	}

	if passModes.Contains("yubikey") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "yubikey",
			Label:       "YubiKey OTP",
			Placeholder: "Enter YubiKey...",
		})
	}

	if passModes.Contains("password") {
		prompts = append(prompts, Prompt{
			Type:        PromptInput,
			Key:         "password",
			Label:       "Password",
			Placeholder: "Enter password...",
		})
	}

	return
}

func Start(sprflId, mode, password string, passwordPrompt bool) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	if mode == "" {
		if sprfl.HideOvpn {
			mode = "wg"
		} else {
			mode = sprfl.LastMode
			if mode == "" {
				mode = "ovpn"
			}
		}
	}

	switch mode {
	case "ovpn", "wg":
		break
	default:
		err = errortypes.NotFoundError{
			errors.New("sprofile: Invalid profile mode"),
		}
		return
	}

	reqUrl := service.GetAddress() + "/profile"

	if passwordPrompt {
		password, err = PasswordPrompt(sprfl)
		if err != nil {
			return
		}
	}

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	data, err := json.Marshal(&SprofileData{
		Id:       sprfl.Id,
		Mode:     mode,
		Password: password,
	})
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Json marshal error"),
		}
		return
	}

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Post request failed"),
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	if sprfl.SsoAuth {
		for i := 0; i < 50; i++ {
			prfl, e := profile.Get(sprfl.Id)
			if e != nil {
				break
			}

			if prfl != nil && prfl.SsoUrl != "" {
				fmt.Println("Single sign-on authentication required, " +
					"open link to complete authentication:")
				fmt.Println(prfl.SsoUrl)
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return
}

type Callback func(prompts []Prompt, callback func(PromptValues), err error)

type Prompt struct {
	Key         string
	Type        int
	Label       string
	Placeholder string
	Value       string
}

var PromptInput = 1
var PromptLink = 2

type PromptValues map[string]string

func StartCallback(sprflId, mode string, callback Callback) {
	sprfl, err := Match(sprflId)
	if err != nil {
		callback(nil, nil, err)
		return
	}

	if mode == "" {
		mode = sprfl.LastMode
		if mode == "" {
			mode = "ovpn"
		}
	}

	switch mode {
	case "ovpn", "wg":
		break
	default:
		err = errortypes.NotFoundError{
			errors.New("sprofile: Invalid profile mode"),
		}
		callback(nil, nil, err)
		return
	}

	prompts := PasswordPrompts(sprfl)
	if len(prompts) > 0 {
		callback(prompts, func(values PromptValues) {
			password := ""

			if values["pin"] != "" {
				password += values["pin"]
			}
			if values["duo"] != "" {
				password += values["duo"]
			}
			if values["onelogin"] != "" {
				password += values["onelogin"]
			}
			if values["okta"] != "" {
				password += values["okta"]
			}
			if values["otp"] != "" {
				password += values["otp"]
			}
			if values["yubikey"] != "" {
				password += values["yubikey"]
			}
			if values["password"] != "" {
				password += values["password"]
			}

			startCallback2(sprflId, mode, password, callback)
		}, nil)
		return
	}

	startCallback2(sprflId, mode, "", callback)
	return
}

func startCallback2(sprflId, mode, password string, callback Callback) {
	sprfl, err := Match(sprflId)
	if err != nil {
		callback(nil, nil, err)
		return
	}

	reqUrl := service.GetAddress() + "/profile"

	authKey, err := service.GetAuthKey()
	if err != nil {
		callback(nil, nil, err)
		return
	}

	data, err := json.Marshal(&SprofileData{
		Id:       sprfl.Id,
		Mode:     mode,
		Password: password,
	})
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Json marshal error"),
		}
		callback(nil, nil, err)
		return
	}

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Post request failed"),
		}
		callback(nil, nil, err)
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		callback(nil, nil, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		callback(nil, nil, err)
		return
	}

	if sprfl.SsoAuth {
		for i := 0; i < 50; i++ {
			prfl, e := profile.Get(sprfl.Id)
			if e != nil {
				break
			}

			if prfl != nil && prfl.SsoUrl != "" {
				prompts := []Prompt{{
					Type:  PromptLink,
					Label: "Single sign-on authentication required:",
					Value: prfl.SsoUrl,
				}}

				callback(prompts, nil, nil)
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func SetState(sprflId string, state bool) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	if sprfl.ForceConnect && !state {
		err = errortypes.ParseError{
			errors.New("sprofile: Autostart enforced by server"),
		}
		return
	}

	sprfl.Disabled = !state

	reqUrl := service.GetAddress() + "/sprofile"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	data, err := json.Marshal(sprfl)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Json marshal error"),
		}
		return
	}

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("PUT", reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Post request failed"),
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}

func Import(data string) (err error) {
	proflId, err := utils.RandStr(16)
	if err != nil {
		return
	}

	profl := &Sprofile{
		Id: strings.ToLower(proflId),
	}

	jsonData := ""
	jsonFound := false
	jsonLoaded := false

	dataLines := strings.Split(data, "\n")
	data = ""
	for _, line := range dataLines {
		if !jsonLoaded && !jsonFound && line == "#{" {
			jsonFound = true
			jsonLoaded = true
		}

		if jsonFound && strings.HasPrefix(line, "#") {
			if line == "#}" {
				jsonFound = false
			}
			jsonData += strings.Replace(line, "#", "", 1)
		} else {
			data += line + "\n"
		}
	}

	if jsonLoaded {
		err = json.Unmarshal([]byte(jsonData), profl)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "profile: Failed to parse sync conf data"),
			}
			return
		}
	} else {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Conf data missing"),
		}
		return
	}

	profl.OvpnData = data

	reqUrl := service.GetAddress() + "/sprofile"

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	reqData, err := json.Marshal(profl)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Json marshal error"),
		}
		return
	}

	body := bytes.NewBuffer(reqData)

	req, err := http.NewRequest("PUT", reqUrl, body)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Post request failed"),
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
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(err, "sprofile: Unknown request error %d",
				resp.StatusCode),
		}
		return
	}

	return
}

func ImportTar(filename string) (err error) {
	tarFile, err := os.Open(filename)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to open tar '%s'", tarFile),
		}
		return
	}
	defer tarFile.Close()

	tr := tar.NewReader(tarFile)
	for {
		_, err = tr.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}

			err = errortypes.ReadError{
				errors.Wrap(err, "sprofile: Failed to read tar header"),
			}
			return
		}

		data := bytes.NewBuffer(nil)
		_, err = io.Copy(data, tr)
		if err != nil {
			err = errortypes.ReadError{
				errors.Wrap(err, "sprofile: Failed to read tar data"),
			}
			return
		}

		err = Import(data.String())
		if err != nil {
			return
		}
	}

	return
}

func ImportUri(uri string) (err error) {
	uri = strings.Replace(uri, "pritunl://", "https://", 1)
	uri = strings.Replace(uri, "/k/", "/ku/", 1)

	req, err := http.NewRequest(
		"GET",
		uri,
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "profile: Sync profile request error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl")

	var client *http.Client
	if len(ip4reg.FindAllString(uri, -1)) > 0 ||
		len(ip6reg.FindAllString(uri, -1)) > 0 {

		client = clientInsecure
	} else {
		client = clientSecure
	}

	resp, err := client.Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Invalid profile uri"),
		}
		return
	}

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Wrapf(
				err,
				"sprofile: Unknown profile uri error %d",
				resp.StatusCode,
			),
		}
		return
	}

	data := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to parse uri response body"),
		}
		return
	}

	for _, proflData := range data {
		err = Import(proflData)
		if err != nil {
			return
		}
	}

	return
}
