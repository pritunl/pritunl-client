// Package sprofile manages the client profiles. System profiles are stored
// by the service and can autostart, user profiles are stored in the user
// profiles directory and are synced and connected by the client, matching
// the desktop client profile handling.
package sprofile

import (
	"fmt"
	"math"
	"strings"

	"github.com/pritunl/pritunl-client/cli/profile"
)

// Sprofile is a system or user profile, the System field selects the
// storage and connection handling.
type Sprofile struct {
	Id                    string                `json:"id"`
	Name                  string                `json:"name"`
	State                 bool                  `json:"state"`
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
	DisableReconnect      bool                  `json:"disable_reconnect"`
	DisableReconnectLocal bool                  `json:"disable_reconnect_local"`
	DisableGateway        bool                  `json:"disable_gateway"`
	DisableDns            bool                  `json:"disable_dns"`
	DisableIpv6           bool                  `json:"disable_ipv6"`
	Dco                   bool                  `json:"dco"`
	DebugOutput           bool                  `json:"debug_output"`
	RestrictClient        bool                  `json:"restrict_client"`
	ForceDns              bool                  `json:"force_dns"`
	SsoAuth               bool                  `json:"sso_auth"`
	PasswordMode          string                `json:"password_mode"`
	Token                 bool                  `json:"token"`
	TokenTtl              int                   `json:"token_ttl"`
	Disabled              bool                  `json:"disabled"`
	SyncTime              int64                 `json:"sync_time"`
	SyncHosts             []string              `json:"sync_hosts"`
	SyncHash              string                `json:"sync_hash"`
	SyncSecret            string                `json:"sync_secret"`
	SyncToken             string                `json:"sync_token"`
	ServerPublicKey       []string              `json:"server_public_key"`
	ServerBoxPublicKey    string                `json:"server_box_public_key"`
	RegistrationKey       string                `json:"registration_key"`
	OvpnData              string                `json:"ovpn_data"`
	Password              string                `json:"password"`
	KeyData               string                `json:"key_data"`

	// System profiles are stored by the service, user profiles are stored
	// in the user profiles directory.
	System bool `json:"-"`

	// Profile is the connection state from the service when the profile
	// has a connection.
	Profile *profile.Profile `json:"-"`
}

type RemoteData struct {
	Priority int `json:"priority"`
}

func (s *Sprofile) FormatedName() (name string) {
	name = s.Name

	if name == "" {
		if s.Server != "" {
			name = s.Server
			if s.User != "" {
				name += fmt.Sprintf(" (%s)", strings.SplitN(s.User, "@", 2)[0])
			}
		} else {
			name = "Unknown Profile"
		}
	}

	return
}

func (s *Sprofile) FormatedNameShort() (name string) {
	name = s.Name

	if name == "" {
		if s.Server != "" {
			name = s.Server
		} else {
			name = "Unknown Profile"
		}
	}

	return
}

// FormatedType returns the profile storage type.
func (s *Sprofile) FormatedType() string {
	if s.System {
		return "System"
	}
	return "User"
}

func (s *Sprofile) FormatedRunState() string {
	if s.State {
		return "Active"
	} else {
		return "Inactive"
	}
}

// FormatedState returns the autostart state, only system profiles can
// autostart.
func (s *Sprofile) FormatedState() string {
	if !s.System {
		return "-"
	}
	if s.Disabled {
		return "Disabled"
	} else {
		return "Enabled"
	}
}

// IsConnected returns true when the profile has an active connection,
// system profiles are active from the service state.
func (s *Sprofile) IsConnected() bool {
	if s.System {
		return s.State
	}
	return s.Profile != nil && s.Profile.Status != "" &&
		s.Profile.Status != "disconnected"
}

func (s *Sprofile) FormatedStatus() (label, status string) {
	if s.Profile == nil {
		if s.System && s.State {
			return "Status", "Connecting"
		}
		return "Status", "Disconnected"
	}

	if s.Profile.Status == "" {
		if s.System && s.State {
			return "Status", "Connecting"
		}
		return "Status", "Disconnected"
	}

	switch s.Profile.Status {
	case "connected":
		uptime := s.Profile.Uptime()
		unitItems := []string{}

		if uptime >= 86400 {
			units := int64(math.Floor(float64(uptime) / 86400))
			uptime -= units * 86400
			unitStr := fmt.Sprintf("%dd", units)
			unitItems = append(unitItems, unitStr)
		}

		if uptime >= 3600 || len(unitItems) > 0 {
			units := int64(math.Floor(float64(uptime) / 3600))
			uptime -= units * 3600
			unitStr := fmt.Sprintf("%dh", units)
			unitItems = append(unitItems, unitStr)
		}

		if uptime >= 60 || len(unitItems) > 0 {
			units := int64(math.Floor(float64(uptime) / 60))
			uptime -= units * 60
			unitStr := fmt.Sprintf("%dm", units)
			unitItems = append(unitItems, unitStr)
		}

		unitStr := fmt.Sprintf("%ds", uptime)
		unitItems = append(unitItems, unitStr)

		return "Online For", strings.Join(unitItems, " ")
	case "connecting":
		return "Status", "Connecting"
	case "authenticating":
		return "Status", "Authenticating"
	case "reconnecting":
		return "Status", "Reconnecting"
	case "disconnecting":
		if s.System && s.State {
			return "Status", "Reconnecting"
		}
		return "Status", "Disconnecting"
	default:
		return "Status", s.Profile.Status
	}
}

// UvName returns the device name set in the profile configuration.
func (s *Sprofile) UvName() string {
	for _, line := range strings.Split(s.OvpnData, "\n") {
		if strings.HasPrefix(line, "setenv UV_NAME ") {
			lineSpl := strings.Split(strings.TrimSpace(line), " ")
			return lineSpl[len(lineSpl)-1]
		}
	}
	return ""
}

// GetLogs reads the profile log from the service.
func (s *Sprofile) GetLogs() (data string, err error) {
	pth := "/log/" + s.Id
	if s.System {
		pth = "/sprofile/" + s.Id + "/log"
	}

	body, err := serviceGet(pth)
	if err != nil {
		return
	}

	data = strings.TrimSpace(string(body)) + "\n"

	return
}

// ClearLogs clears the profile log output.
func (s *Sprofile) ClearLogs() (err error) {
	pth := "/log/" + s.Id
	if s.System {
		pth = "/sprofile/" + s.Id + "/log"
	}

	return serviceCall("DELETE", pth, nil)
}
