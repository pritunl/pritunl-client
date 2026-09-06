package iface

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pritunl/pritunl-client/cli/sprofile"
)

// infoSyncTime mirrors the desktop client last configuration sync text.
func infoSyncTime(syncTime int64) string {
	if syncTime == -1 {
		return "Failed to sync"
	}
	if syncTime == 0 {
		return "Never"
	}
	return time.Unix(syncTime, 0).Format("Jan 2 2006, 3:04PM")
}

// infoSyncHosts mirrors the desktop client formatted sync hosts, only the
// host and port of the first eight hosts are shown.
func infoSyncHosts(syncHosts []string) string {
	hosts := []string{}
	for i, hostAddr := range syncHosts {
		if i >= 8 {
			hosts = append(hosts, "...")
			break
		}

		hostUrl, err := url.Parse(hostAddr)
		if err != nil || hostUrl.Host == "" {
			hosts = append(hosts, hostAddr)
			continue
		}
		hosts = append(hosts, hostUrl.Host)
	}

	return strings.Join(hosts, "\n  ")
}

func infoRemotesData(remotes map[string]sprofile.RemoteData) string {
	if len(remotes) == 0 {
		return ""
	}

	data, err := json.Marshal(remotes)
	if err != nil {
		return ""
	}
	return string(data)
}

func infoBool(val bool) string {
	return strconv.FormatBool(val)
}

// profileInfoFields returns the profile fields shown in the desktop client
// settings dialog including the debugging fields.
func profileInfoFields(sprfl *sprofile.Sprofile) []InfoField {
	return []InfoField{
		{"ID", sprfl.Id},
		{"Last Configuration Sync", infoSyncTime(sprfl.SyncTime)},
		{"Configuration Sync Hosts", infoSyncHosts(sprfl.SyncHosts)},
		{"System", infoBool(sprfl.System)},
		{"UV Name", sprfl.UvName()},
		{"State", infoBool(sprfl.State)},
		{"WireGuard", infoBool(sprfl.Wg)},
		{"Last Mode", sprfl.LastMode},
		{"Organization ID", sprfl.OrganizationId},
		{"Organization", sprfl.Organization},
		{"Server ID", sprfl.ServerId},
		{"Server", sprfl.Server},
		{"User ID", sprfl.UserId},
		{"User", sprfl.User},
		{"Pre Connect Message", sprfl.PreConnectMsg},
		{"Disable Reconnect", infoBool(sprfl.DisableReconnect)},
		{"Disable Reconnect Local", infoBool(sprfl.DisableReconnectLocal)},
		{"Restrict Client", infoBool(sprfl.RestrictClient)},
		{"Remotes Data", infoRemotesData(sprfl.RemotesData)},
		{"Hide OpenVPN", infoBool(sprfl.HideOvpn)},
		{"Dynamic Firewall", infoBool(sprfl.DynamicFirewall)},
		{"Geo Sort", sprfl.GeoSort},
		{"Force Connect", infoBool(sprfl.ForceConnect)},
		{"Device Auth", infoBool(sprfl.DeviceAuth)},
		{"Disable Gateway", infoBool(sprfl.DisableGateway)},
		{"Disable DNS", infoBool(sprfl.DisableDns)},
		{"Disable IPv6", infoBool(sprfl.DisableIpv6)},
		{"Data Channel Offload", infoBool(sprfl.Dco)},
		{"Debug Output", infoBool(sprfl.DebugOutput)},
		{"Force DNS", infoBool(sprfl.ForceDns)},
		{"SSO Auth", infoBool(sprfl.SsoAuth)},
		{"Password Mode", sprfl.PasswordMode},
		{"Token", infoBool(sprfl.Token)},
		{"Token TTL", strconv.Itoa(sprfl.TokenTtl)},
		{"Sync Hash", sprfl.SyncHash},
	}
}
