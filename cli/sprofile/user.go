package sprofile

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/utils"
	"github.com/pritunl/tools/logger"
)

// userConf mirrors the desktop client profile configuration file stored
// next to the profile data in the user profiles directory.
type userConf struct {
	Name                  string                `json:"name"`
	Wg                    bool                  `json:"wg"`
	LastMode              string                `json:"last_mode"`
	OrganizationId        string                `json:"organization_id"`
	Organization          string                `json:"organization"`
	ServerId              string                `json:"server_id"`
	Server                string                `json:"server"`
	UserId                string                `json:"user_id"`
	User                  string                `json:"user"`
	PreConnectMsg         string                `json:"pre_connect_msg"`
	RemotesData           map[string]RemoteData `json:"remotes_data"`
	HideOvpn              bool                  `json:"hide_ovpn"`
	DynamicFirewall       bool                  `json:"dynamic_firewall"`
	GeoSort               string                `json:"geo_sort"`
	ForceConnect          bool                  `json:"force_connect"`
	DeviceAuth            bool                  `json:"device_auth"`
	DisableReconnectLocal bool                  `json:"disable_reconnect_local"`
	DisableGateway        bool                  `json:"disable_gateway"`
	DisableDns            bool                  `json:"disable_dns"`
	DisableIpv6           bool                  `json:"disable_ipv6"`
	Dco                   bool                  `json:"dco"`
	DebugOutput           bool                  `json:"debug_output"`
	ForceDns              bool                  `json:"force_dns"`
	SsoAuth               bool                  `json:"sso_auth"`
	PasswordMode          string                `json:"password_mode"`
	Token                 bool                  `json:"token"`
	TokenTtl              int                   `json:"token_ttl"`
	DisableReconnect      bool                  `json:"disable_reconnect"`
	RestrictClient        bool                  `json:"restrict_client"`
	Disabled              bool                  `json:"disabled"`
	SyncTime              int64                 `json:"sync_time"`
	SyncHosts             []string              `json:"sync_hosts"`
	SyncHash              string                `json:"sync_hash"`
	SyncSecret            string                `json:"sync_secret"`
	SyncToken             string                `json:"sync_token"`
	ServerPublicKey       []string              `json:"server_public_key"`
	ServerBoxPublicKey    string                `json:"server_box_public_key"`
	RegistrationKey       string                `json:"registration_key"`
	KeyData               string                `json:"key_data"`
}

// GetUserPath returns the user profiles directory shared with the desktop
// client.
func GetUserPath() string {
	return filepath.Join(utils.GetDataPath(), "profiles")
}

// ConfPath returns the user profile configuration file path.
func (s *Sprofile) ConfPath() string {
	return filepath.Join(GetUserPath(), s.Id+".conf")
}

// DataPath returns the user profile OpenVPN data file path.
func (s *Sprofile) DataPath() string {
	return filepath.Join(GetUserPath(), s.Id+".ovpn")
}

func (s *Sprofile) exportConf() *userConf {
	return &userConf{
		Name:                  s.Name,
		Wg:                    s.Wg,
		LastMode:              s.LastMode,
		OrganizationId:        s.OrganizationId,
		Organization:          s.Organization,
		ServerId:              s.ServerId,
		Server:                s.Server,
		UserId:                s.UserId,
		User:                  s.User,
		PreConnectMsg:         s.PreConnectMsg,
		RemotesData:           s.RemotesData,
		HideOvpn:              s.HideOvpn,
		DynamicFirewall:       s.DynamicFirewall,
		GeoSort:               s.GeoSort,
		ForceConnect:          s.ForceConnect,
		DeviceAuth:            s.DeviceAuth,
		DisableReconnectLocal: s.DisableReconnectLocal,
		DisableGateway:        s.DisableGateway,
		DisableDns:            s.DisableDns,
		DisableIpv6:           s.DisableIpv6,
		Dco:                   s.Dco,
		DebugOutput:           s.DebugOutput,
		ForceDns:              s.ForceDns,
		SsoAuth:               s.SsoAuth,
		PasswordMode:          s.PasswordMode,
		Token:                 s.Token,
		TokenTtl:              s.TokenTtl,
		DisableReconnect:      s.DisableReconnect,
		RestrictClient:        s.RestrictClient,
		Disabled:              s.Disabled,
		SyncTime:              s.SyncTime,
		SyncHosts:             s.SyncHosts,
		SyncHash:              s.SyncHash,
		SyncSecret:            s.SyncSecret,
		SyncToken:             s.SyncToken,
		ServerPublicKey:       s.ServerPublicKey,
		ServerBoxPublicKey:    s.ServerBoxPublicKey,
		RegistrationKey:       s.RegistrationKey,
		KeyData:               s.KeyData,
	}
}

func (s *Sprofile) loadConf(conf *userConf) {
	s.Name = conf.Name
	s.Wg = conf.Wg
	s.LastMode = conf.LastMode
	s.OrganizationId = conf.OrganizationId
	s.Organization = conf.Organization
	s.ServerId = conf.ServerId
	s.Server = conf.Server
	s.UserId = conf.UserId
	s.User = conf.User
	s.PreConnectMsg = conf.PreConnectMsg
	s.RemotesData = conf.RemotesData
	s.HideOvpn = conf.HideOvpn
	s.DynamicFirewall = conf.DynamicFirewall
	s.GeoSort = conf.GeoSort
	s.ForceConnect = conf.ForceConnect
	s.DeviceAuth = conf.DeviceAuth
	s.DisableReconnectLocal = conf.DisableReconnectLocal
	s.DisableGateway = conf.DisableGateway
	s.DisableDns = conf.DisableDns
	s.DisableIpv6 = conf.DisableIpv6
	s.Dco = conf.Dco
	s.DebugOutput = conf.DebugOutput
	s.ForceDns = conf.ForceDns
	s.SsoAuth = conf.SsoAuth
	s.PasswordMode = conf.PasswordMode
	s.Token = conf.Token
	s.TokenTtl = conf.TokenTtl
	s.DisableReconnect = conf.DisableReconnect
	s.RestrictClient = conf.RestrictClient
	s.Disabled = conf.Disabled
	s.SyncTime = conf.SyncTime
	s.SyncHosts = conf.SyncHosts
	s.SyncHash = conf.SyncHash
	s.SyncSecret = conf.SyncSecret
	s.SyncToken = conf.SyncToken
	s.ServerPublicKey = conf.ServerPublicKey
	s.ServerBoxPublicKey = conf.ServerBoxPublicKey
	s.RegistrationKey = conf.RegistrationKey
	s.KeyData = conf.KeyData
}

// importConf copies the configuration of an imported profile onto an
// existing profile for the same user, matching the desktop client.
func (s *Sprofile) importConf(data *Sprofile) {
	s.Name = data.Name
	s.Wg = data.Wg
	s.OrganizationId = data.OrganizationId
	s.Organization = data.Organization
	s.ServerId = data.ServerId
	s.Server = data.Server
	s.UserId = data.UserId
	s.User = data.User
	s.PreConnectMsg = data.PreConnectMsg
	s.RemotesData = data.RemotesData
	s.HideOvpn = data.HideOvpn
	s.DynamicFirewall = data.DynamicFirewall
	s.GeoSort = data.GeoSort
	s.ForceConnect = data.ForceConnect
	s.DeviceAuth = data.DeviceAuth
	s.DisableReconnectLocal = data.DisableReconnectLocal
	s.DisableGateway = data.DisableGateway
	s.DisableDns = data.DisableDns
	s.DisableIpv6 = data.DisableIpv6
	s.Dco = data.Dco
	s.DebugOutput = data.DebugOutput
	s.ForceDns = data.ForceDns
	s.SsoAuth = data.SsoAuth
	s.PasswordMode = data.PasswordMode
	s.Token = data.Token
	s.TokenTtl = data.TokenTtl
	s.DisableReconnect = data.DisableReconnect
	s.RestrictClient = data.RestrictClient
	s.SyncTime = data.SyncTime
	s.SyncHosts = data.SyncHosts
	if s.SyncHosts == nil {
		s.SyncHosts = []string{}
	}
	s.SyncHash = data.SyncHash
	s.SyncSecret = data.SyncSecret
	s.SyncToken = data.SyncToken
	s.ServerPublicKey = data.ServerPublicKey
	s.ServerBoxPublicKey = data.ServerBoxPublicKey
	s.KeyData = data.KeyData
}

// upsertConf applies a synced configuration keeping the current values
// for empty identity fields, matching the desktop client.
func (s *Sprofile) upsertConf(conf *userConf) {
	if conf.Name != "" {
		s.Name = conf.Name
	}
	s.Wg = conf.Wg
	if conf.OrganizationId != "" {
		s.OrganizationId = conf.OrganizationId
	}
	if conf.Organization != "" {
		s.Organization = conf.Organization
	}
	if conf.ServerId != "" {
		s.ServerId = conf.ServerId
	}
	if conf.Server != "" {
		s.Server = conf.Server
	}
	if conf.UserId != "" {
		s.UserId = conf.UserId
	}
	if conf.User != "" {
		s.User = conf.User
	}
	s.PreConnectMsg = conf.PreConnectMsg
	s.RemotesData = conf.RemotesData
	s.HideOvpn = conf.HideOvpn
	s.DynamicFirewall = conf.DynamicFirewall
	s.GeoSort = conf.GeoSort
	s.ForceConnect = conf.ForceConnect
	s.DeviceAuth = conf.DeviceAuth
	s.SsoAuth = conf.SsoAuth
	s.PasswordMode = conf.PasswordMode
	s.Token = conf.Token
	s.TokenTtl = conf.TokenTtl
	s.DisableReconnect = conf.DisableReconnect
	s.RestrictClient = conf.RestrictClient
	s.SyncHosts = conf.SyncHosts
	s.SyncHash = conf.SyncHash
	s.ServerPublicKey = conf.ServerPublicKey
	s.ServerBoxPublicKey = conf.ServerBoxPublicKey
}

// ensureFileMode restricts profile files to the owner as the desktop
// client does when loading profiles.
func ensureFileMode(pth string) {
	if runtime.GOOS == "windows" {
		return
	}

	info, err := os.Stat(pth)
	if err != nil {
		return
	}

	if info.Mode().Perm() != 0600 {
		err = os.Chmod(pth, 0600)
		if err != nil {
			logger.WithFields(logger.Fields{
				"path":  pth,
				"error": err,
			}).Warn("sprofile: Failed to set profile file mode")
		}
	}
}

func loadUser(prflId string) (sprfl *Sprofile, err error) {
	sprfl = &Sprofile{
		Id:     prflId,
		System: false,
	}

	confPath := sprfl.ConfPath()
	dataPath := sprfl.DataPath()
	logPath := filepath.Join(GetUserPath(), prflId+".log")

	ensureFileMode(confPath)
	ensureFileMode(dataPath)
	ensureFileMode(logPath)

	confData, err := os.ReadFile(confPath)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to read profile '%s'",
				confPath),
		}
		return
	}

	conf := &userConf{}
	err = json.Unmarshal(confData, conf)
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrapf(err, "sprofile: Failed to parse profile '%s'",
				confPath),
		}
		return
	}
	sprfl.loadConf(conf)

	data, err := os.ReadFile(dataPath)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to read profile data '%s'",
				dataPath),
		}
		return
	}
	sprfl.OvpnData = string(data)

	return
}

// getAllUser loads the profiles in the user profiles directory, profiles
// that fail to load are logged and skipped.
func getAllUser() (sprfls []*Sprofile, err error) {
	sprfls = []*Sprofile{}

	entries, err := os.ReadDir(GetUserPath())
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		err = errortypes.ReadError{
			errors.Wrap(err, "sprofile: Failed to read profiles directory"),
		}
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".conf") {
			continue
		}

		prflId := strings.TrimSuffix(name, ".conf")
		if prflId == "" {
			continue
		}

		sprfl, e := loadUser(prflId)
		if e != nil {
			logger.WithFields(logger.Fields{
				"profile_id": prflId,
				"error":      e,
			}).Error("sprofile: Failed to load user profile")
			continue
		}

		sprfls = append(sprfls, sprfl)
	}

	return
}

func (s *Sprofile) writeConf() (err error) {
	err = os.MkdirAll(GetUserPath(), 0700)
	if err != nil {
		err = errortypes.WriteError{
			errors.Wrap(err, "sprofile: Failed to create profiles directory"),
		}
		return
	}

	data, err := json.Marshal(s.exportConf())
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to marshal profile"),
		}
		return
	}

	err = os.WriteFile(s.ConfPath(), data, 0600)
	if err != nil {
		err = errortypes.WriteError{
			errors.Wrap(err, "sprofile: Failed to write profile"),
		}
		return
	}

	return
}

func (s *Sprofile) writeData(data string) (err error) {
	err = os.MkdirAll(GetUserPath(), 0700)
	if err != nil {
		err = errortypes.WriteError{
			errors.Wrap(err, "sprofile: Failed to create profiles directory"),
		}
		return
	}

	err = os.WriteFile(s.DataPath(), []byte(data), 0600)
	if err != nil {
		err = errortypes.WriteError{
			errors.Wrap(err, "sprofile: Failed to write profile data"),
		}
		return
	}

	s.OvpnData = data

	return
}

// hasKeyData returns true when the OpenVPN data contains inline keys.
func hasKeyData(data string) bool {
	return extractBlock(data, "<tls-auth>", "</tls-auth>\n") != "" ||
		extractBlock(data, "<tls-crypt>", "</tls-crypt>\n") != "" ||
		extractBlock(data, "<key>", "</key>\n") != ""
}

// extractBlock returns the section of data from the start tag through
// the end tag or an empty string when not found.
func extractBlock(data, start, end string) string {
	sIndex := strings.Index(data, start)
	if sIndex < 0 {
		return ""
	}
	eIndex := strings.Index(data[sIndex:], end)
	if eIndex < 0 {
		return ""
	}
	return data[sIndex : sIndex+eIndex+len(end)]
}

// ReadData returns the profile OpenVPN data including the keys. System
// profiles return the data from the service.
func (s *Sprofile) ReadData() (data string, err error) {
	if s.System {
		data = s.OvpnData
		return
	}

	dataByt, err := os.ReadFile(s.DataPath())
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrap(err, "sprofile: Failed to read profile data"),
		}
		return
	}
	data = string(dataByt)
	s.OvpnData = data

	if s.KeyData != "" {
		err = errortypes.ReadError{
			errors.New("sprofile: Profile keys are encrypted by the " +
				"desktop client, disable safe storage in the desktop " +
				"client settings to use the profile"),
		}
		return
	}

	if runtime.GOOS == "darwin" && !hasKeyData(data) {
		keyData, e := keychainKey(s.Id)
		if e != nil {
			logger.WithFields(logger.Fields{
				"profile_id": s.Id,
				"error":      e,
			}).Error("sprofile: Failed to get profile key from keychain")
		} else {
			data += keyData
		}
	}

	return
}

// removeUser deletes the user profile files.
func (s *Sprofile) removeUser() (err error) {
	if runtime.GOOS == "darwin" {
		keychainDelete(s.Id)
	}

	for _, pth := range []string{s.ConfPath(), s.DataPath()} {
		e := os.Remove(pth)
		if e != nil && !os.IsNotExist(e) {
			err = errortypes.WriteError{
				errors.Wrapf(e, "sprofile: Failed to remove '%s'", pth),
			}
			return
		}
	}

	return
}

// keychainKey reads profile keys stored in the macOS keychain by older
// desktop clients.
func keychainKey(prflId string) (keyData string, err error) {
	output, err := exec.Command(
		"/usr/bin/security",
		"find-generic-password",
		"-w",
		"-s", "pritunl",
		"-a", prflId,
	).Output()
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrap(err, "sprofile: Keychain read failed"),
		}
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(
		strings.TrimSpace(string(output)))
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to decode keychain key"),
		}
		return
	}

	keyData = string(decoded)

	return
}

func keychainDelete(prflId string) {
	_ = exec.Command(
		"/usr/bin/security",
		"delete-generic-password",
		"-s", "pritunl",
		"-a", prflId,
	).Run()
}
