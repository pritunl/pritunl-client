package connection

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/parser"
	"github.com/sirupsen/logrus"
)

type Remote struct {
	Host      string
	Addr4     string
	Addr6     string
	OvpnPort  int
	OvpnProto string
	Type      string
}

type Remotes []*Remote

func (r Remotes) GetHosts() (hosts []string) {
	hosts = []string{}

	for _, remote := range r {
		hosts = append(hosts, remote.Host)
	}

	return
}

func (r Remotes) GetAddrs() (addrs []string) {
	addrs = []string{}

	for _, remote := range r {
		if remote.Addr4 != "" {
			addrs = append(addrs, remote.Addr4)
		}
		if remote.Addr6 != "" {
			addrs = append(addrs, remote.Addr6)
		}
	}

	return
}

func (r Remotes) GetFormatted() (hosts []string) {
	hosts = []string{}

	for _, remote := range r {
		hosts = append(hosts, remote.GetFormatted())
	}

	return
}

func (r Remotes) GetAddrMap() (addrMap map[string]*Remote, other []*Remote) {
	addrMap = map[string]*Remote{}
	other = []*Remote{}

	for _, remote := range r {
		if remote.Addr4 != "" {
			addrMap[remote.Addr4] = remote
		}
		if remote.Addr6 != "" {
			addrMap[remote.Addr6] = remote
		}

		if remote.Addr4 == "" && remote.Addr6 == "" {
			other = append(other, remote)
		}
	}

	return
}

func (r Remotes) GetParser() (remotes parser.Remotes) {
	remotes = parser.Remotes{}

	for _, remote := range r {
		remotes = append(remotes, remote.GetParser()...)
	}

	return
}

func (r *Remote) Lookup() {
	ip := net.ParseIP(r.Host)
	if ip != nil {
		ipStr := ip.String()
		if ip.To4() == nil {
			r.Addr6 = ipStr
		} else {
			r.Addr4 = ipStr
		}
	} else {
		remoteIps, err := net.LookupIP(r.Host)
		if err != nil {
			err = &errortypes.RequestError{
				errors.Wrap(err, "remotes: Failed to resolve remote"),
			}

			logrus.WithFields(logrus.Fields{
				"host":  r.Host,
				"error": err,
			}).Info("profile: Failed to resolve remote")
			return
		}

		for _, remoteIp := range remoteIps {
			remoteIpStr := remoteIp.String()
			if remoteIp.To4() == nil {
				if r.Addr6 == "" {
					r.Addr6 = remoteIpStr
				} else {
					// TODO Handle multiple
				}
			} else {
				if r.Addr4 == "" {
					r.Addr4 = remoteIpStr
				} else {
					// TODO Handle multiple
				}
			}
		}
	}
}

func (r *Remote) Equal(addr string) bool {
	if strings.Contains(addr, ":") {
		var hostIp6 net.IP
		if strings.Contains(r.Host, ":") {
			hostIp6 = net.ParseIP(r.Host)
		}

		var addrIp6 net.IP
		if strings.Contains(r.Addr6, ":") {
			addrIp6 = net.ParseIP(r.Addr6)
		}

		ip6 := net.ParseIP(addr)
		if ip6 != nil {
			if ip6.Equal(hostIp6) || ip6.Equal(addrIp6) {
				return true
			}
		}
	}

	if addr == r.Host || addr == r.Addr4 || addr == r.Addr6 {
		return true
	}

	return false
}

func (r *Remote) GetUrl(path string) *url.URL {
	remote := r.Host

	if strings.Count(remote, ":") > 1 && !strings.Contains(remote, "[") {
		remote = "[" + remote + "]"
	}

	return &url.URL{
		Scheme: "https",
		Host:   remote,
		Path:   path,
	}
}

func (r *Remote) GetFormatted() (host string) {
	host = r.Host

	if r.Type == SyncRemote {
		host += "*"
	}
	if r.Addr4 != "" {
		host += fmt.Sprintf("[%s]", r.Addr4)
	}
	if r.Addr6 != "" {
		host += fmt.Sprintf("[%s]", r.Addr6)
	}

	return
}

func (r *Remote) GetParser() (remotes parser.Remotes) {
	remotes = parser.Remotes{}

	if r.Addr4 != "" {
		remotes = append(remotes, parser.Remote{
			Host:  r.Addr4,
			Port:  r.OvpnPort,
			Proto: r.OvpnProto,
		})
	}
	if r.Addr6 != "" {
		remotes = append(remotes, parser.Remote{
			Host:  r.Addr6,
			Port:  r.OvpnPort,
			Proto: r.OvpnProto,
		})
	}

	if r.Addr4 == "" && r.Addr6 == "" && r.Host != "" {
		remotes = append(remotes, parser.Remote{
			Host:  r.Host,
			Port:  r.OvpnPort,
			Proto: r.OvpnProto,
		})
	}

	return
}
