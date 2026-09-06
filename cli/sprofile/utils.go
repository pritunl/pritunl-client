package sprofile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/profile"
	"github.com/pritunl/pritunl-client/cli/service"
	"github.com/pritunl/pritunl-client/cli/terminal"
	"github.com/pritunl/tools/logger"
)

// profileRequest is the request body for system profile connections and
// disconnect and remove requests.
type profileRequest struct {
	Id       string `json:"id"`
	Mode     string `json:"mode,omitempty"`
	Password string `json:"password,omitempty"`
}

// connectData is the request body for user profile connections, the
// profile configuration is sent to the service with the data.
type connectData struct {
	Id                 string                `json:"id"`
	Mode               string                `json:"mode"`
	OrgId              string                `json:"org_id"`
	UserId             string                `json:"user_id"`
	ServerId           string                `json:"server_id"`
	SyncHosts          []string              `json:"sync_hosts"`
	SyncToken          string                `json:"sync_token"`
	SyncSecret         string                `json:"sync_secret"`
	SyncHash           string                `json:"sync_hash"`
	Username           string                `json:"username"`
	Password           string                `json:"password"`
	RemotesData        map[string]RemoteData `json:"remotes_data"`
	HideOvpn           bool                  `json:"hide_ovpn"`
	DynamicFirewall    bool                  `json:"dynamic_firewall"`
	GeoSort            string                `json:"geo_sort"`
	ForceConnect       bool                  `json:"force_connect"`
	DeviceAuth         bool                  `json:"device_auth"`
	DisableGateway     bool                  `json:"disable_gateway"`
	DisableDns         bool                  `json:"disable_dns"`
	DisableIpv6        bool                  `json:"disable_ipv6"`
	Dco                bool                  `json:"dco"`
	DebugOutput        bool                  `json:"debug_output"`
	ForceDns           bool                  `json:"force_dns"`
	RestrictClient     bool                  `json:"restrict_client"`
	SsoAuth            bool                  `json:"sso_auth"`
	ServerPublicKey    string                `json:"server_public_key"`
	ServerBoxPublicKey string                `json:"server_box_public_key"`
	TokenTtl           int                   `json:"token_ttl"`
	Timeout            bool                  `json:"timeout"`
	Reconnect          bool                  `json:"reconnect"`
	Data               string                `json:"data"`
}

type tokenRequest struct {
	Profile            string `json:"profile"`
	ServerPublicKey    string `json:"server_public_key"`
	ServerBoxPublicKey string `json:"server_box_public_key"`
	Ttl                int    `json:"ttl"`
}

type tokenResponse struct {
	Valid bool `json:"valid"`
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

	return sprfl.Disconnect()
}

func Delete(sprflId string) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	return sprfl.Remove()
}

// GetAll loads the system profiles from the service and the user
// profiles from the user profiles directory with the connection state
// of each, matching the desktop client profile loading.
func GetAll() (sprfls Sprofiles, err error) {
	sprfls = []*Sprofile{}
	err = serviceJson(service.GetPollClient(), "GET", "/sprofile",
		nil, &sprfls)
	if err != nil {
		return
	}

	for _, sprfl := range sprfls {
		sprfl.System = true
	}

	userPrfls, err := getAllUser()
	if err != nil {
		return
	}
	sprfls = append(sprfls, userPrfls...)

	prfls, err := profile.GetAll()
	if err != nil {
		return
	}

	for _, sprfl := range sprfls {
		prfl := prfls[sprfl.Id]
		if prfl == nil {
			continue
		}

		sprfl.Profile = prfl
		if !sprfl.System {
			sprfl.State = sprfl.IsConnected()

			// Store the registration key from the service as the
			// desktop client does on the registration event
			if sprfl.RegistrationKey == "" && prfl.RegistrationKey != "" {
				sprfl.RegistrationKey = prfl.RegistrationKey
				e := sprfl.writeConf()
				if e != nil {
					logger.WithFields(logger.Fields{
						"profile_id": sprfl.Id,
						"error":      e,
					}).Error("sprofile: Failed to save registration key")
				}
			}
		}
	}

	sprfls.Sort()

	return
}

// PasswordPrompt reads the system profile authentication from the
// terminal in the order expected by the service.
func PasswordPrompt(sprfl *Sprofile) (pass string, err error) {
	passModes := set.NewSet()

	passModesStr := strings.Split(sprfl.PasswordMode, "_")
	for _, passMode := range passModesStr {
		passModes.Add(passMode)
	}

	if passModes.Contains("pin") {
		part := terminal.ReadPassword("Pin")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: Pin is empty"),
			}
			return
		}
		pass += part
	}

	if passModes.Contains("duo") {
		part := terminal.ReadPassword("Duo Passcode")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: Duo Passcode is empty"),
			}
			return
		}
		pass += part
	}

	if passModes.Contains("onelogin") {
		part := terminal.ReadPassword("OneLogin Passcode")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: OneLogin Passcode is empty"),
			}
			return
		}
		pass += part
	}

	if passModes.Contains("okta") {
		part := terminal.ReadPassword("Okta Passcode")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: Okta Passcode is empty"),
			}
			return
		}
		pass += part
	}

	if passModes.Contains("otp") {
		part := terminal.ReadPassword("Authenticator Passcode")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: Authenticator Passcode is empty"),
			}
			return
		}
		pass += part
	}

	if passModes.Contains("yubikey") {
		part := terminal.ReadPassword("YubiKey")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: YubiKey is empty"),
			}
			return
		}
		pass += part
	}

	if pass == "" {
		part := terminal.ReadPassword("Password")
		if part == "" {
			err = errortypes.ParseError{
				errors.New("sprofile: Password is empty"),
			}
			return
		}
		pass += part
	}

	return
}

// PasswordPrompts returns the authentication prompts of a system
// profile from the password mode.
func PasswordPrompts(sprfl *Sprofile) (prompts []Prompt) {
	passModes := set.NewSet()

	passModesStr := strings.Split(sprfl.PasswordMode, "_")
	for _, passMode := range passModesStr {
		passModes.Add(passMode)
	}

	if passModes.Contains("pin") {
		prompts = append(prompts, promptPin)
	}
	if passModes.Contains("duo") {
		prompts = append(prompts, promptDuo)
	}
	if passModes.Contains("onelogin") {
		prompts = append(prompts, promptOnelogin)
	}
	if passModes.Contains("okta") {
		prompts = append(prompts, promptOkta)
	}
	if passModes.Contains("otp") {
		prompts = append(prompts, promptOtp)
	}
	if passModes.Contains("yubikey") {
		prompts = append(prompts, promptYubikey)
	}
	if passModes.Contains("password") {
		prompts = append(prompts, promptPassword)
	}

	return
}

// authTypes returns the authentication types of a user profile from the
// password mode or the OpenVPN data, matching the desktop client.
func (s *Sprofile) authTypes(data string) []string {
	passwordMode := s.PasswordMode
	if passwordMode == "" && strings.Contains(data, "auth-user-pass") {
		if s.User != "" {
			passwordMode = "otp"
		} else {
			passwordMode = "username_password"
		}
	}

	if passwordMode == "" {
		return []string{}
	}

	return strings.Split(passwordMode, "_")
}

// userPrompts returns the authentication prompts of a user profile in
// the desktop client order, a valid token removes the second factor
// prompts.
func userPrompts(authTypes []string, tokenValid bool) (prompts []Prompt) {
	types := set.NewSet()
	for _, authType := range authTypes {
		types.Add(authType)
	}

	if tokenValid {
		types.Remove("pin")
		types.Remove("duo")
		types.Remove("onelogin")
		types.Remove("okta")
		types.Remove("yubikey")
		types.Remove("otp")
	}

	if types.Contains("username") {
		prompts = append(prompts, promptUsername)
	}
	if types.Contains("password") {
		prompts = append(prompts, promptPassword)
	}
	if types.Contains("pin") {
		prompts = append(prompts, promptPin)
	}
	if types.Contains("duo") {
		prompts = append(prompts, promptDuo)
	}
	if types.Contains("onelogin") {
		prompts = append(prompts, promptOnelogin)
	}
	if types.Contains("okta") {
		prompts = append(prompts, promptOkta)
	}
	if types.Contains("otp") && !types.Contains("duo") &&
		!types.Contains("onelogin") && !types.Contains("okta") {

		prompts = append(prompts, promptOtp)
	}
	if types.Contains("yubikey") {
		prompts = append(prompts, promptYubikey)
	}

	return
}

// Prompts is the authentication input required to connect a profile.
type Prompts struct {
	Fields []Prompt

	// TokenValid is set when the service holds a valid authentication
	// token for the profile and second factors are not required.
	TokenValid bool

	// SyncErr is a non fatal failure syncing the profile configuration
	// before connecting.
	SyncErr error
}

// Empty returns true when no input is required.
func (p *Prompts) Empty() bool {
	return len(p.Fields) == 0
}

// PreConnect prepares the profile for connecting and returns the
// authentication prompts. User profiles are synced from the server and
// the service token is updated first, matching the desktop client.
func (s *Sprofile) PreConnect() (prompts *Prompts, err error) {
	prompts = &Prompts{}

	if s.System {
		prompts.Fields = PasswordPrompts(s)
		return
	}

	prompts.SyncErr = s.Sync()

	if s.Token {
		valid, e := s.tokenUpdate()
		if e != nil {
			logger.WithFields(logger.Fields{
				"profile_id": s.Id,
				"error":      e,
			}).Error("sprofile: Failed to update token")
		} else {
			prompts.TokenValid = valid
		}
	} else {
		e := s.tokenDelete()
		if e != nil {
			logger.WithFields(logger.Fields{
				"profile_id": s.Id,
				"error":      e,
			}).Error("sprofile: Failed to clear token")
		}
	}

	data, err := s.ReadData()
	if err != nil {
		return
	}

	prompts.Fields = userPrompts(s.authTypes(data), prompts.TokenValid)

	return
}

func (s *Sprofile) tokenUpdate() (valid bool, err error) {
	resp := &tokenResponse{}
	err = serviceJson(service.GetClient(), "PUT", "/token", &tokenRequest{
		Profile:            s.Id,
		ServerPublicKey:    strings.Join(s.ServerPublicKey, "\n"),
		ServerBoxPublicKey: s.ServerBoxPublicKey,
		Ttl:                s.TokenTtl,
	}, resp)
	if err != nil {
		return
	}

	valid = resp.Valid

	return
}

func (s *Sprofile) tokenDelete() (err error) {
	return serviceCall("DELETE", "/token/"+s.Id, nil)
}

// Start connects a profile from the command line, the prompts are read
// from the terminal when passwordPrompt is set.
func Start(sprflId, mode, username, password string,
	passwordPrompt bool) (err error) {

	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	prompts, err := sprfl.PreConnect()
	if err != nil {
		return
	}

	if prompts.SyncErr != nil {
		fmt.Fprintln(os.Stderr, "Profile sync failed: "+
			strings.SplitN(prompts.SyncErr.Error(), "\n", 2)[0])
	}

	if passwordPrompt {
		if sprfl.System {
			password, err = PasswordPrompt(sprfl)
			if err != nil {
				return
			}
		} else {
			values := readPrompts(prompts.Fields)
			if values["username"] != "" {
				username = values["username"]
			}
			password = BuildUserPassword(values)
		}
	}

	err = sprfl.Connect(mode, &ConnectAuth{
		Username: username,
		Password: password,
		Token:    prompts.TokenValid,
	})
	if err != nil {
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

// readPrompts reads the prompt values from the terminal.
func readPrompts(prompts []Prompt) (values PromptValues) {
	values = PromptValues{}

	for _, prompt := range prompts {
		if prompt.Secret {
			values[prompt.Key] = terminal.ReadPassword(prompt.Label)
			continue
		}

		fmt.Print(prompt.Label + ": ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		values[prompt.Key] = strings.TrimSpace(line)
	}

	return
}

type Prompt struct {
	Key         string
	Type        int
	Label       string
	Placeholder string
	Value       string

	// Secret prompts are masked when entered.
	Secret bool
}

var PromptInput = 1

var (
	promptUsername = Prompt{
		Type:        PromptInput,
		Key:         "username",
		Label:       "Username",
		Placeholder: "Enter username...",
	}
	promptPassword = Prompt{
		Type:        PromptInput,
		Key:         "password",
		Label:       "Password",
		Placeholder: "Enter password...",
		Secret:      true,
	}
	promptPin = Prompt{
		Type:        PromptInput,
		Key:         "pin",
		Label:       "Pin",
		Placeholder: "Enter pin...",
		Secret:      true,
	}
	promptDuo = Prompt{
		Type:        PromptInput,
		Key:         "duo",
		Label:       "Duo Passcode",
		Placeholder: "Enter passcode...",
	}
	promptOnelogin = Prompt{
		Type:        PromptInput,
		Key:         "onelogin",
		Label:       "OneLogin Passcode",
		Placeholder: "Enter passcode...",
	}
	promptOkta = Prompt{
		Type:        PromptInput,
		Key:         "okta",
		Label:       "Okta Passcode",
		Placeholder: "Enter passcode...",
	}
	promptOtp = Prompt{
		Type:        PromptInput,
		Key:         "otp",
		Label:       "Authenticator Passcode",
		Placeholder: "Enter passcode...",
	}
	promptYubikey = Prompt{
		Type:        PromptInput,
		Key:         "yubikey",
		Label:       "YubiKey OTP",
		Placeholder: "Enter YubiKey...",
	}
)

type PromptValues map[string]string

// BuildPassword joins system profile prompt values in the order expected
// by the service.
func BuildPassword(values PromptValues) (password string) {
	for _, key := range []string{
		"pin", "duo", "onelogin", "okta", "otp", "yubikey", "password",
	} {
		password += values[key]
	}
	return
}

// BuildUserPassword joins user profile prompt values in the order used
// by the desktop client.
func BuildUserPassword(values PromptValues) (password string) {
	for _, key := range []string{
		"password", "pin", "duo", "onelogin", "okta", "otp", "yubikey",
	} {
		password += values[key]
	}
	return
}

// BuildAuth returns the connection authentication from prompt values,
// the password is joined in the order for the profile type.
func (s *Sprofile) BuildAuth(values PromptValues,
	tokenValid bool) *ConnectAuth {

	auth := &ConnectAuth{
		Username: values["username"],
		Token:    tokenValid,
	}

	if s.System {
		auth.Password = BuildPassword(values)
	} else {
		auth.Password = BuildUserPassword(values)
	}

	return auth
}

// ResolveMode returns the mode to connect with when none is specified.
func (s *Sprofile) ResolveMode(mode string) string {
	if mode == "" {
		if s.HideOvpn {
			mode = "wg"
		} else {
			mode = s.LastMode
			if mode == "" {
				mode = "ovpn"
			}
		}
	}
	return mode
}

// ConnectAuth is the authentication for a connection, Token is set when
// the service holds a valid token from PreConnect.
type ConnectAuth struct {
	Username string
	Password string
	Token    bool
}

// Connect sends the connect request for the profile, prompts must
// already be resolved into auth. User profiles send the profile data to
// the service matching the desktop client.
func (s *Sprofile) Connect(mode string, auth *ConnectAuth) (err error) {
	mode = s.ResolveMode(mode)

	switch mode {
	case "ovpn", "wg":
		break
	default:
		err = errortypes.NotFoundError{
			errors.New("sprofile: Invalid profile mode"),
		}
		return
	}

	if auth == nil {
		auth = &ConnectAuth{}
	}

	if s.System {
		return serviceCall("POST", "/profile", &profileRequest{
			Id:       s.Id,
			Mode:     mode,
			Password: auth.Password,
		})
	}

	data, err := s.ReadData()
	if err != nil {
		return
	}
	if data == "" {
		err = errortypes.ReadError{
			errors.New("sprofile: Profile data is empty"),
		}
		return
	}

	username := auth.Username
	if username == "" {
		username = "pritunl"
	}
	if !auth.Token && auth.Password == "" {
		username = ""
	}

	return serviceCall("POST", "/profile", &connectData{
		Id:                 s.Id,
		Mode:               mode,
		OrgId:              s.OrganizationId,
		UserId:             s.UserId,
		ServerId:           s.ServerId,
		SyncHosts:          s.SyncHosts,
		SyncToken:          s.SyncToken,
		SyncSecret:         s.SyncSecret,
		SyncHash:           s.SyncHash,
		Username:           username,
		Password:           auth.Password,
		RemotesData:        s.RemotesData,
		HideOvpn:           s.HideOvpn,
		DynamicFirewall:    s.DynamicFirewall,
		GeoSort:            s.GeoSort,
		ForceConnect:       s.ForceConnect,
		DeviceAuth:         s.DeviceAuth,
		DisableGateway:     s.DisableGateway,
		DisableDns:         s.DisableDns,
		DisableIpv6:        s.DisableIpv6,
		Dco:                s.Dco,
		DebugOutput:        s.DebugOutput,
		ForceDns:           s.ForceDns,
		RestrictClient:     s.RestrictClient,
		SsoAuth:            s.SsoAuth,
		ServerPublicKey:    strings.Join(s.ServerPublicKey, "\n"),
		ServerBoxPublicKey: s.ServerBoxPublicKey,
		TokenTtl:           s.TokenTtl,
		Timeout:            true,
		Reconnect:          !(s.DisableReconnect || s.DisableReconnectLocal),
		Data:               data,
	})
}

// Disconnect stops the running profile.
func (s *Sprofile) Disconnect() (err error) {
	return serviceCall("DELETE", "/profile", &profileRequest{
		Id: s.Id,
	})
}

// Remove deletes the profile, system profiles are removed from the
// service and user profiles from the user profiles directory.
func (s *Sprofile) Remove() (err error) {
	if s.System {
		return serviceCall("DELETE", "/sprofile", &profileRequest{
			Id: s.Id,
		})
	}

	_ = s.Disconnect()
	_ = serviceCall("DELETE", "/log/"+s.Id, nil)

	return s.removeUser()
}

// Commit saves the profile settings, system profiles to the service and
// user profiles to the configuration file.
func (s *Sprofile) Commit() (err error) {
	if !s.System {
		return s.writeConf()
	}

	if s.ForceConnect && s.Disabled {
		err = errortypes.ParseError{
			errors.New("sprofile: Autostart enforced by server"),
		}
		return
	}

	return serviceCall("PUT", "/sprofile", s)
}

// SetState enables or disables autostart, only system profiles autostart.
func SetState(sprflId string, state bool) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	if !sprfl.System {
		err = errortypes.ParseError{
			errors.New("sprofile: Autostart requires a system profile, " +
				"convert the profile with the convert command"),
		}
		return
	}

	sprfl.Disabled = !state

	return sprfl.Commit()
}

// Convert changes the profile storage type between system and user.
func Convert(sprflId string, system bool) (err error) {
	sprfl, err := Match(sprflId)
	if err != nil {
		return
	}

	if system {
		if sprfl.System {
			err = errortypes.ParseError{
				errors.New("sprofile: Profile is already a system profile"),
			}
			return
		}
		return sprfl.ConvertSystem()
	}

	if !sprfl.System {
		err = errortypes.ParseError{
			errors.New("sprofile: Profile is already a user profile"),
		}
		return
	}
	return sprfl.ConvertUser()
}
