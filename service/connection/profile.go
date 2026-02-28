package connection

import (
	"strings"

	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/types"
	"github.com/sirupsen/logrus"
)

type Profile struct {
	conn               *Connection                 `json:"-"`
	Id                 string                      `json:"id"`
	Mode               string                      `json:"mode"`
	OrgId              string                      `json:"org_id"`
	UserId             string                      `json:"user_id"`
	ServerId           string                      `json:"server_id"`
	SyncHosts          []string                    `json:"sync_hosts"`
	SyncToken          string                      `json:"sync_token"`
	SyncSecret         string                      `json:"sync_secret"`
	Data               string                      `json:"data"`
	Username           string                      `json:"username"`
	Password           string                      `json:"password"`
	RemotesData        map[string]types.RemoteData `json:"remotes_data"`
	HideOvpn           bool                        `json:"hide_ovpn"`
	DynamicFirewall    bool                        `json:"dynamic_firewall"`
	GeoSort            string                      `json:"geo_sort"`
	ForceConnect       bool                        `json:"force_connect"`
	DeviceAuth         bool                        `json:"device_auth"`
	DisableGateway     bool                        `json:"disable_gateway"`
	DisableDns         bool                        `json:"disable_dns"`
	Dco                bool                        `json:"dco"`
	RestrictClient     bool                        `json:"restrict_client"`
	ForceDns           bool                        `json:"force_dns"`
	SsoAuth            bool                        `json:"sso_auth"`
	ServerPublicKey    string                      `json:"server_public_key"`
	ServerBoxPublicKey string                      `json:"server_box_public_key"`
	RegistrationKey    string                      `json:"registration_key"`
	TokenTtl           int                         `json:"token_ttl"`
	Reconnect          bool                        `json:"reconnect"`
	Timeout            bool                        `json:"timeout"`
	SystemProfile      bool                        `json:"-"`
}

func (p *Profile) Fields() logrus.Fields {
	return logrus.Fields{
		"profile_id":               p.Id,
		"profile_mode":             p.Mode,
		"profile_dynamic_firewall": p.DynamicFirewall,
		"profile_device_auth":      p.DeviceAuth,
		"profile_disable_gateway":  p.DisableGateway,
		"profile_disable_dns":      p.DisableDns,
		"profile_dco":              p.Dco,
		"profile_geo_sort":         p.IsGeoSort(),
		"profile_force_connect":    p.ForceConnect,
		"profile_force_dns":        p.ForceDns,
		"profile_sso_auth":         p.SsoAuth,
		"profile_reconnect":        p.Reconnect,
		"profile_timeout":          p.Timeout,
		"profile_system_profile":   p.SystemProfile,
	}
}

func (p *Profile) IsGeoSort() bool {
	return p.GeoSort != ""
}

func (p *Profile) Sync() {
	if p.SystemProfile {
		sprfl := sprofile.Get(p.Id)
		if sprfl == nil {
			logrus.WithFields(p.conn.Fields(nil)).Error(
				"profile: Missing system profile in sync")
			return
		}

		updated, err := sprfl.Sync()
		if err != nil {
			logrus.WithFields(p.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("profile: Failed to sync system profile")
			sprfl.SyncTime = -1
			sprfl.Commit()
		}

		if updated {
			p.ImportSystemProfile(sprfl)
		}
	} else {
		// TODO Sync non system profiles
	}

	return
}

func (p *Profile) ImportSystemProfile(sprfl *sprofile.Sprofile) {
	serverPublicKey := ""
	if sprfl.ServerPublicKey != nil && len(sprfl.ServerPublicKey) > 0 {
		serverPublicKey = strings.Join(sprfl.ServerPublicKey, "\n")
	}

	lastMode := sprfl.LastMode
	if lastMode == "" {
		lastMode = OvpnMode
	}

	p.Id = sprfl.Id
	p.Mode = lastMode
	p.OrgId = sprfl.OrganizationId
	p.UserId = sprfl.UserId
	p.ServerId = sprfl.ServerId
	p.SyncHosts = sprfl.SyncHosts
	p.SyncToken = sprfl.SyncToken
	p.SyncSecret = sprfl.SyncSecret
	p.Data = sprfl.OvpnData
	p.Username = "pritunl"
	p.Password = sprfl.Password
	p.RemotesData = sprfl.RemotesData
	p.HideOvpn = sprfl.HideOvpn
	p.DynamicFirewall = sprfl.DynamicFirewall
	p.GeoSort = sprfl.GeoSort
	p.ForceConnect = sprfl.ForceConnect
	p.DeviceAuth = sprfl.DeviceAuth
	p.DisableGateway = sprfl.DisableGateway
	p.DisableDns = sprfl.DisableDns
	p.Dco = sprfl.Dco
	p.RestrictClient = sprfl.RestrictClient
	p.ForceDns = sprfl.ForceDns
	p.SsoAuth = sprfl.SsoAuth
	p.ServerPublicKey = serverPublicKey
	p.ServerBoxPublicKey = sprfl.ServerBoxPublicKey
	p.RegistrationKey = sprfl.RegistrationKey
	p.TokenTtl = sprfl.TokenTtl
	p.Reconnect = true
	p.SystemProfile = true
}
