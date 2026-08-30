package profile

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Route struct {
	NextHop    string `json:"next_hop"`
	Network    string `json:"network"`
	Metric     int    `json:"metric"`
	NetGateway bool   `json:"net_gateway"`
}

type Profile struct {
	Id            string   `json:"id"`
	Mode          string   `json:"mode"`
	Iface         string   `json:"iface"`
	Tuniface      string   `json:"tun_iface"`
	Routes        []*Route `json:"routes"`
	Routes6       []*Route `json:"routes6"`
	Reconnect     bool     `json:"reconnect"`
	AuthReconnect bool     `json:"auth_reconnect"`
	Status        string   `json:"status"`
	Timestamp     int64    `json:"timestamp"`
	GatewayAddr   string   `json:"gateway_addr"`
	GatewayAddr6  string   `json:"gateway_addr6"`
	ServerAddr    string   `json:"server_addr"`
	ClientAddr    string   `json:"client_addr"`
	MacAddr       string   `json:"mac_addr"`
	MacAddrs      []string `json:"mac_addrs"`
	SsoUrl        string   `json:"sso_url"`
}

func (p *Profile) Uptime() int64 {
	return time.Now().Unix() - p.Timestamp
}

func (p *Profile) FormatedTime() string {
	if p.Timestamp == 0 {
		return "Connecting"
	}

	uptime := p.Uptime()
	unitItems := []string{}

	if uptime > 86400 {
		units := int64(math.Floor(float64(uptime) / 86400))
		uptime -= units * 86400
		unitStr := fmt.Sprintf("%d day", units)
		if units > 1 {
			unitStr += "s"
		}
		unitItems = append(unitItems, unitStr)
	}

	if uptime > 3600 {
		units := int64(math.Floor(float64(uptime) / 3600))
		uptime -= units * 3600
		unitStr := fmt.Sprintf("%d hour", units)
		if units > 1 {
			unitStr += "s"
		}
		unitItems = append(unitItems, unitStr)
	}

	if uptime > 60 {
		units := int64(math.Floor(float64(uptime) / 60))
		uptime -= units * 60
		unitStr := fmt.Sprintf("%d min", units)
		if units > 1 {
			unitStr += "s"
		}
		unitItems = append(unitItems, unitStr)
	}

	if uptime > 0 {
		unitStr := fmt.Sprintf("%d sec", uptime)
		if uptime > 1 {
			unitStr += "s"
		}
		unitItems = append(unitItems, unitStr)
	}

	return strings.Join(unitItems, " ")
}

func (p *Profile) FormatedTimeShort() string {
	if p.Timestamp == 0 {
		return "Connecting"
	}

	uptime := p.Uptime()
	if uptime < 0 {
		uptime = 0
	}

	days := uptime / 86400
	uptime -= days * 86400
	hours := uptime / 3600
	uptime -= hours * 3600
	mins := uptime / 60
	secs := uptime - mins*60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, mins, secs)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
