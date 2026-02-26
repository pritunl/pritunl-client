package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/sirupsen/logrus"
)

type Remote struct {
	Host  string
	Port  int
	Proto string
}

func (r *Remote) GetFormatted() string {
	return r.Host + fmt.Sprintf("(%d/%s)", r.Port, r.Proto)
}

type Remotes []Remote

func (r Remotes) GetFormatted() (remotes []string) {
	remotes = []string{}

	for _, remote := range r {
		remotes = append(remotes, remote.GetFormatted())
	}

	return
}

type Ovpn struct {
	EnvId             string
	EnvName           string
	Dev               string
	DevType           string
	DataCiphers       string
	Remotes           []Remote
	NoBind            bool
	PersistTun        bool
	Cipher            string
	Auth              string
	Verb              int
	Mute              int
	PushPeerInfo      bool
	Ping              int
	PingExit          int
	HandWindow        int
	ServerPollTimeout int
	RenegSec          int
	RedirectGateway   string
	SndBuf            int
	RcvBuf            int
	MssFix            int
	TunMtu            int
	Fragment          int
	RemoteCertTls     string
	Compress          string
	CompLzo           string
	BlockOutsideDns   bool
	AuthUserPass      bool
	KeyDirection      int
	CaCert            string
	TlsAuth           string
	TlsCrypt          string
	Cert              string
	Key               string

	DisableGateway bool
	DisableDns     bool
}

func (o *Ovpn) Export(chown string) string {
	output := ""

	if o.EnvId != "" {
		output += fmt.Sprintf("setenv UV_ID %s\n", o.EnvId)
	}
	if o.EnvName != "" {
		output += fmt.Sprintf("setenv UV_NAME %s\n", o.EnvName)
	}
	output += fmt.Sprintf("setenv UV_PRITUNL_VER %s\n", constants.Version)

	if chown != "" {
		output += fmt.Sprintf("user %s\n", chown)
		output += fmt.Sprintf("group %s\n", chown)
	}

	output += "client\n"
	output += fmt.Sprintf("dev %s\n", o.Dev)
	output += fmt.Sprintf("dev-type %s\n", o.DevType)
	output += "connect-retry-max 1\n"
	for _, remote := range o.Remotes {
		output += fmt.Sprintf(
			"remote %s %d %s\n",
			remote.Host,
			remote.Port,
			remote.Proto,
		)
	}
	if o.NoBind {
		output += "nobind\n"
	}
	if o.PersistTun {
		output += "persist-tun\n"
	}
	if o.Cipher != "" && o.DataCiphers == "" {
		output += fmt.Sprintf("cipher %s\n", o.Cipher)
	}
	if o.Auth != "" {
		output += fmt.Sprintf("auth %s\n", o.Auth)
	}
	if o.Verb > 0 {
		output += fmt.Sprintf("verb %d\n", o.Verb)
	}
	if o.Mute > 0 {
		output += fmt.Sprintf("mute %d\n", o.Mute)
	}
	if o.PushPeerInfo {
		output += "push-peer-info\n"
	}
	if o.Ping > 0 {
		output += fmt.Sprintf("ping %d\n", o.Ping)
	}
	if o.PingExit > 0 {
		output += fmt.Sprintf("ping-exit %d\n", o.PingExit)
	}
	if o.HandWindow > 0 {
		output += fmt.Sprintf("hand-window %d\n", o.HandWindow)
	}
	if o.ServerPollTimeout > 0 {
		output += fmt.Sprintf("server-poll-timeout %d\n", o.ServerPollTimeout)
	}
	if o.RenegSec > 0 {
		output += fmt.Sprintf("reneg-sec %d\n", o.RenegSec)
	}
	if o.RedirectGateway != "" {
		output += fmt.Sprintf("redirect-gateway %s\n", o.RedirectGateway)
	}
	if o.SndBuf > 0 {
		output += fmt.Sprintf("sndbuf %d\n", o.SndBuf)
	}
	if o.RcvBuf > 0 {
		output += fmt.Sprintf("rcvbuf %d\n", o.RcvBuf)
	}
	if o.MssFix > 0 {
		output += fmt.Sprintf("mssfix %d\n", o.MssFix)
	}
	if o.Fragment > 0 {
		output += fmt.Sprintf("fragment %d\n", o.Fragment)
	}
	if o.TunMtu > 0 {
		output += fmt.Sprintf("tun-mtu %d\n", o.TunMtu)
	}
	if o.RemoteCertTls != "" {
		output += fmt.Sprintf("remote-cert-tls %s\n", o.RemoteCertTls)
	}
	if o.Compress != "" {
		output += fmt.Sprintf("compress %s\n", o.Compress)
	}
	if o.CompLzo != "" {
		output += fmt.Sprintf("comp-lzo %s\n", o.CompLzo)
	}
	if o.BlockOutsideDns {
		output += "ignore-unknown-option block-outside-dns\n"
		output += "block-outside-dns\n"
	}
	if o.AuthUserPass {
		output += "auth-user-pass\n"
	}
	if o.KeyDirection > 0 {
		output += fmt.Sprintf("key-direction %d\n", o.KeyDirection)
	}

	if o.DisableGateway {
		output += "pull-filter ignore \"redirect-gateway\"\n"
		output += "pull-filter ignore \"route-ipv6 2000::/3\"\n"
	}

	if o.DisableDns {
		output += "pull-filter ignore \"dhcp-option\"\n"
	}

	output += "pull-filter ignore \"ping-restart\"\n"
	output += "pull-filter ignore \"dhcp-option DOMAIN-ROUTE\"\n"

	if o.DataCiphers != "" {
		output += fmt.Sprintf("data-ciphers \"%s\"\n", o.DataCiphers)
	} else {
		output += "ignore-unknown-option data-ciphers\n"
		output += "data-ciphers \"AES-256-GCM:AES-128-GCM:" +
			"CHACHA20-POLY1305:AES-256-CBC:AES-128-CBC\"\n"
	}

	if o.CaCert != "" {
		output += fmt.Sprintf("<ca>\n%s</ca>\n", o.CaCert)
	}
	if o.TlsAuth != "" {
		output += fmt.Sprintf("<tls-auth>\n%s</tls-auth>\n", o.TlsAuth)
	}

	if o.TlsCrypt != "" {
		output += fmt.Sprintf("<tls-crypt>\n%s</tls-crypt>\n", o.TlsCrypt)
	}

	if o.Cert != "" {
		output += fmt.Sprintf("<cert>\n%s</cert>\n", o.Cert)
	}
	if o.Key != "" {
		output += fmt.Sprintf("<key>\n%s</key>\n", o.Key)
	}

	return output
}

func Import(data string, remotes []Remote,
	disableGateway, disableDns bool) (o *Ovpn) {

	o = &Ovpn{
		DisableGateway: disableGateway,
		DisableDns:     disableDns,
	}

	inCa := false
	inTlsAuth := false
	inTlsCrypt := false
	inCert := false
	inKey := false

	data = strings.ReplaceAll(data, "\r", "")

	for _, origLine := range strings.Split(data, "\n") {
		line := FilterStr(origLine, 256)

		if line != origLine {
			logrus.WithFields(logrus.Fields{
				"orig_line": origLine,
				"new_line":  line,
			}).Warn("parser: Configuration line filtered")
		}

		if inCa {
			if line == "</ca>" {
				inCa = false
				continue
			}
			o.CaCert += line + "\n"
		} else if inTlsAuth {
			if line == "</tls-auth>" {
				inTlsAuth = false
				continue
			}
			o.TlsAuth += line + "\n"
		} else if inTlsCrypt {
			if line == "</tls-crypt>" {
				inTlsCrypt = false
				continue
			}
			o.TlsCrypt += line + "\n"
		} else if inCert {
			if line == "</cert>" {
				inCert = false
				continue
			}
			o.Cert += line + "\n"
		} else if inKey {
			if line == "</key>" {
				inKey = false
				continue
			}
			o.Key += line + "\n"
		}

		lines := strings.Split(line, " ")

		key := strings.ToLower(lines[0])

		switch key {
		case "<ca>":
			inCa = true
			break
		case "<tls-auth>":
			inTlsAuth = true
			break
		case "<tls-crypt>":
			inTlsCrypt = true
			break
		case "<cert>":
			inCert = true
			break
		case "<key>":
			inKey = true
			break
		case "setenv":
			if len(lines) != 3 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [1]")
				continue
			}
			switch strings.ToLower(lines[1]) {
			case "uv_id":
				o.EnvId = lines[2]
				break
			case "uv_name":
				o.EnvName = lines[2]
				break
			}
			break
		case "dev":
			switch strings.ToLower(lines[1]) {
			case "tun":
				o.Dev = "tun"
				break
			case "tap":
				o.Dev = "tap"
				break
			}
			break
		case "dev-type":
			switch strings.ToLower(lines[1]) {
			case "tun":
				o.DevType = "tun"
				break
			case "tap":
				o.DevType = "tap"
				break
			}
			break
		case "data-ciphers":
			ciphersSet := set.NewSet()
			ciphers := []string{}
			for _, cipher := range strings.Split(lines[1], ":") {
				switch strings.ToLower(cipher) {
				case "aes-128-cbc":
					cipher = "AES-128-CBC"
				case "aes-256-cbc":
					cipher = "AES-256-CBC"
				case "aes-128-gcm":
					cipher = "AES-128-GCM"
				case "aes-256-gcm":
					cipher = "AES-256-GCM"
				case "chacha20-poly1305":
					cipher = "CHACHA20-POLY1305"
				default:
					continue
				}
				if !ciphersSet.Contains(cipher) {
					ciphersSet.Add(cipher)
					ciphers = append(ciphers, cipher)
				}
			}

			if !ciphersSet.Contains("AES-128-GCM") {
				ciphers = append(ciphers, "AES-128-GCM")
			}
			if !ciphersSet.Contains("AES-256-GCM") {
				ciphers = append(ciphers, "AES-256-GCM")
			}
			if !ciphersSet.Contains("CHACHA20-POLY1305") {
				ciphers = append(ciphers, "CHACHA20-POLY1305")
			}

			o.DataCiphers = strings.Join(ciphers, ":")
			break
		case "nobind":
			o.NoBind = true
			break
		case "persist-tun":
			o.PersistTun = true
			break
		case "cipher":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [5]")
				continue
			}

			o.Cipher = lines[1]
			break
		case "auth":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [6]")
				continue
			}

			o.Auth = lines[1]
			break
		case "verb":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [7]")
				continue
			}

			verb, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [8]")
				continue
			}

			o.Verb = verb
			break
		case "mute":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [9]")
				continue
			}

			mute, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [10]")
				continue
			}

			o.Mute = mute
			break
		case "push-peer-info":
			o.PushPeerInfo = true
			break
		case "ping":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [11]")
				continue
			}

			ping, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [12]")
				continue
			}

			o.Ping = ping
			break
		case "ping-restart":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [13]")
				continue
			}

			pingRestart, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [14]")
				continue
			}

			o.PingExit = pingRestart
			break
		case "ping-exit":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [15]")
				continue
			}

			pingExit, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [16]")
				continue
			}

			o.PingExit = pingExit
			break
		case "hand-window":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [17]")
				continue
			}

			handWindow, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [18]")
				continue
			}

			o.HandWindow = handWindow
			break
		case "server-poll-timeout":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [19]")
				continue
			}

			serverPollTimeout, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [20]")
				continue
			}

			o.ServerPollTimeout = serverPollTimeout
			break
		case "reneg-sec":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [21]")
				continue
			}

			renegSec, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [22]")
				continue
			}

			o.RenegSec = renegSec
			break
		case "redirect-gateway":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [35]")
				continue
			}

			switch strings.ToLower(lines[1]) {
			case "local":
				o.RedirectGateway = "local"
				break
			case "autolocal":
				o.RedirectGateway = "autolocal"
				break
			case "def1":
				o.RedirectGateway = "def1"
				break
			case "bypass-dhcp":
				o.RedirectGateway = "bypass-dhcp"
				break
			case "bypass-dns":
				o.RedirectGateway = "bypass-dns"
				break
			case "block-local":
				o.RedirectGateway = "block-local"
				break
			case "ipv6":
				o.RedirectGateway = "ipv6"
				break
			case "!ipv4":
				o.RedirectGateway = "!ipv4"
				break
			default:
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [36]")
				continue
			}

			break
		case "sndbuf":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [23]")
				continue
			}

			sndbuf, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [24]")
				continue
			}

			o.SndBuf = sndbuf
			break
		case "rcvbuf":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [25]")
				continue
			}

			rcvbuf, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [26]")
				continue
			}

			o.RcvBuf = rcvbuf
			break
		case "mssfix":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [37]")
				continue
			}

			mssFix, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [38]")
				continue
			}

			o.MssFix = mssFix
			break
		case "fragment":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [41]")
				continue
			}

			fragment, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [42]")
				continue
			}

			o.Fragment = fragment
			break
		case "tun-mtu":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [39]")
				continue
			}

			tunMtu, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [40]")
				continue
			}

			o.TunMtu = tunMtu
			break
		case "remote-cert-tls":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [27]")
				continue
			}

			switch strings.ToLower(lines[1]) {
			case "server":
				o.RemoteCertTls = "server"
				break
			default:
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [28]")
				continue
			}

			break
		case "comp-lzo":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [29]")
				continue
			}

			switch strings.ToLower(lines[1]) {
			case "yes":
				o.CompLzo = "yes"
				break
			case "no":
				o.CompLzo = "no"
				break
			default:
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [30]")
				continue
			}

			break
		case "block-outside-dns":
			o.BlockOutsideDns = true
			break
		case "compress":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [31]")
				continue
			}

			switch strings.ToLower(lines[1]) {
			case "lzo":
				o.Compress = "lzo"
				break
			case "lz4":
				o.Compress = "lz4"
				break
			default:
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [32]")
				continue
			}

			break
		case "auth-user-pass":
			o.AuthUserPass = true
			break
		case "key-direction":
			if len(lines) != 2 {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [33]")
				continue
			}

			keyDirection, e := strconv.Atoi(lines[1])
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Warn("parser: Configuration line ignored [34]")
				continue
			}

			o.KeyDirection = keyDirection
			break
		}
	}

	if o.Dev == "" {
		o.Dev = "tun"
	}
	if o.DevType == "" {
		o.DevType = "tun"
	}

	o.Remotes = remotes

	return
}
