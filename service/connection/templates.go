package connection

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
)

const (
	wgConfTempl = `[Interface]
Address = {{.Address}}
PrivateKey = {{.PrivateKey}}{{if .HasMtu}}
MTU = {{.Mtu}}{{end}}{{if .HasDns}}
DNS = {{.DnsServers}}{{end}}

[Peer]
PublicKey = {{.PublicKey}}
AllowedIPs = {{.AllowedIps}}
Endpoint = {{.Endpoint}}
`
)

var (
	wgIfaceMacReg = regexp.MustCompile("\\((utun[0-9]+)\\)")
	WgConfTempl   = template.Must(template.New("wg_conf").Parse(wgConfTempl))
)

type WgConfData struct {
	Address    string
	PrivateKey string
	HasMtu     bool
	Mtu        int
	HasDns     bool
	DnsServers string
	PublicKey  string
	AllowedIps string
	Endpoint   string
}

func (w *WgConfData) Validate() (err error) {
	addrs := []string{}
	for _, addr := range strings.Split(w.Address, ",") {
		prefix, ok := parseWgNetwork(addr)
		if !ok {
			err = &errortypes.ParseError{
				errors.New("connection: Invalid wg conf address"),
			}
			return
		}
		addrs = append(addrs, prefix.String())
	}
	w.Address = strings.Join(addrs, ",")

	w.PrivateKey = utils.FilterBase64(w.PrivateKey)
	if !validWgKey(w.PrivateKey) {
		err = &errortypes.ParseError{
			errors.New("connection: Invalid wg conf private key"),
		}
		return
	}

	if w.HasMtu {
		if w.Mtu < 1 || w.Mtu > 65535 {
			err = &errortypes.ParseError{
				errors.New("connection: Invalid wg conf mtu"),
			}
			return
		}
	} else {
		w.Mtu = 0
	}

	if w.HasDns {
		dnsServers := []string{}
		for _, server := range strings.Split(w.DnsServers, ",") {
			ip, ok := parseWgIp(server)
			if ok {
				dnsServers = append(dnsServers, ip.String())
				continue
			}

			domain := filterSearchDomain(server)
			if domain != "" {
				dnsServers = append(dnsServers, domain)
			}
		}
		w.DnsServers = strings.Join(dnsServers, ",")
		if w.DnsServers == "" {
			w.HasDns = false
		}
	} else {
		w.DnsServers = ""
	}

	w.PublicKey = utils.FilterBase64(w.PublicKey)
	if !validWgKey(w.PublicKey) {
		err = &errortypes.ParseError{
			errors.New("connection: Invalid wg conf public key"),
		}
		return
	}

	if w.AllowedIps != "" {
		allowedIps := []string{}
		for _, network := range strings.Split(w.AllowedIps, ",") {
			prefix, ok := parseWgNetwork(network)
			if !ok {
				err = &errortypes.ParseError{
					errors.New("connection: Invalid wg conf allowed ips"),
				}
				return
			}
			allowedIps = append(allowedIps, prefix.Masked().String())
		}
		w.AllowedIps = strings.Join(allowedIps, ",")
	}

	host, port, e := net.SplitHostPort(w.Endpoint)
	if e != nil {
		err = &errortypes.ParseError{
			errors.New("connection: Invalid wg conf endpoint"),
		}
		return
	}

	portNum, e := strconv.Atoi(port)
	if e != nil || portNum < 1 || portNum > 65535 {
		err = &errortypes.ParseError{
			errors.New("connection: Invalid wg conf endpoint port"),
		}
		return
	}

	if ip, ok := parseWgIp(host); ok {
		host = ip.String()
	} else if strings.Contains(host, ":") {
		err = &errortypes.ParseError{
			errors.New("connection: Invalid wg conf endpoint host"),
		}
		return
	} else {
		host = utils.FilterDomain(host)
		if host == "" {
			err = &errortypes.ParseError{
				errors.New("connection: Invalid wg conf endpoint host"),
			}
			return
		}
	}
	w.Endpoint = net.JoinHostPort(host, strconv.Itoa(portNum))

	return
}
