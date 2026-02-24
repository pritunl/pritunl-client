package connection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/network"
	"github.com/pritunl/pritunl-client/service/platform"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

var (
	lastRetryLogged = time.Time{}
)

type Wg struct {
	conn          *Connection
	lock          sync.Mutex
	wgPath        string
	wgQuickPath   string
	wgUtilPath    string
	wgConfPath    string
	wgConfPath2   string
	connected     bool
	lastHandshake int
	bashPath      string
	publicKey     string
	privateKey    string
	serverPubKey  string
	ssoToken      string
	ssoStart      time.Time
}

type WgConf struct {
	Address       string   `json:"address"`
	Address6      string   `json:"address6"`
	Hostname      string   `json:"hostname"`
	Hostname6     string   `json:"hostname6"`
	Gateway       string   `json:"gateway"`
	Gateway6      string   `json:"gateway6"`
	PingInterval  int      `json:"ping_interval"`
	PingTimeout   int      `json:"ping_timeout"`
	Port          int      `json:"port"`
	Mtu           int      `json:"mtu"`
	WebPort       int      `json:"web_port"`
	WebNoSsl      bool     `json:"web_no_ssl"`
	PublicKey     string   `json:"public_key"`
	Routes        []*Route `json:"routes"`
	Routes6       []*Route `json:"routes6"`
	DnsServers    []string `json:"dns_servers"`
	SearchDomains []string `json:"search_domains"`
}

func (w *Wg) Fields() logrus.Fields {
	return logrus.Fields{
		"wg_path":           w.wgPath,
		"wg_quick_path":     w.wgQuickPath,
		"wg_util_path":      w.wgUtilPath,
		"wg_bash_path":      w.bashPath,
		"wg_conf_path":      w.wgConfPath,
		"wg_conf_path2":     w.wgConfPath2,
		"wg_connected":      w.connected,
		"wg_last_handshake": w.lastHandshake,
		"wg_pub_key":        w.publicKey != "",
		"wg_priv_key":       w.privateKey != "",
		"wg_server_pub_key": w.serverPubKey != "",
		"wg_sso_token":      w.ssoToken != "",
		"wg_sso_start":      w.ssoStart,
	}
}

func (w *Wg) Init() {
	w.wgPath = GetWgPath()
	w.wgQuickPath = GetWgQuickPath()
	w.wgUtilPath = GetWgUtilPath()
	w.bashPath = GetBashPath()
}

func (w *Wg) GetPublicKey() string {
	return w.publicKey
}

func (w *Wg) GetReqPrefix() string {
	return "wg"
}

func (w *Wg) generateKey() (err error) {
	privateKey, err := utils.ExecOutput(w.wgPath, "genkey")
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "wg: Failed to generate private key"),
		}
		return
	}

	publicKey, err := utils.ExecInputOutput(privateKey, w.wgPath, "pubkey")
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "wg: Failed to get public key"),
		}
		return
	}

	w.publicKey = strings.TrimSpace(publicKey)
	w.privateKey = strings.TrimSpace(privateKey)

	return
}

func (w *Wg) Start() (err error) {
	err = w.conn.Client.Start(w)
	if err != nil {
		return
	}

	return
}

func (w *Wg) PreConnect() (err error) {
	err = w.generateKey()
	if err != nil {
		return
	}

	return
}

func (w *Wg) Connect(data *ConnData) (err error) {
	if data.Configuration == nil {
		err = &errortypes.ParseError{
			errors.Wrap(
				err,
				"profile: Authorize returned empty configuration",
			),
		}
		return
	}

	iface := network.InterfaceAcquire()
	if iface == "" {
		err = &errortypes.ReadError{
			errors.New("profile: Failed to acquire interface"),
		}
		return
	}
	w.conn.Data.Iface = iface

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	if w.conn.Profile.DisableGateway {
		routes := []*Route{}
		for _, route := range data.Configuration.Routes {
			if route.Network == "0.0.0.0/0" {
				continue
			}
			routes = append(routes, route)
		}
		data.Configuration.Routes = routes

		routes6 := []*Route{}
		for _, route := range data.Configuration.Routes6 {
			if route.Network == "::/0" {
				continue
			}
			routes6 = append(routes6, route)
		}
		data.Configuration.Routes6 = routes6
	}

	err = w.writeWgConf(data.Configuration)
	if err != nil {
		return
	}

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	err = w.confWg(data.Configuration)
	if err != nil {
		w.conn.Data.SendProfileEvent("configuration_error")

		logrus.WithFields(w.conn.Fields(logrus.Fields{
			"error": err,
		})).Error("profile: Failed to configure wg")

		w.conn.State.Close()
		return
	}

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	w.conn.Data.ValidateAuthToken()

	logrus.WithFields(w.conn.Fields(logrus.Fields{
		"ping_interval":  data.Configuration.PingInterval,
		"ping_timeout":   data.Configuration.PingTimeout,
		"port":           data.Configuration.Port,
		"mtu":            data.Configuration.Mtu,
		"web_port":       data.Configuration.WebPort,
		"web_no_ssl":     data.Configuration.WebNoSsl,
		"dns_servers":    data.Configuration.DnsServers,
		"search_domains": data.Configuration.SearchDomains,
	})).Info("connection: WireGuard configure")

	return
}

func (w *Wg) WatchConnection() (err error) {
	defer w.conn.State.Close()

	time.Sleep(1 * time.Second)

	interval := w.conn.Data.PingIntervalWg
	if interval == 0 {
		interval = 15
	} else if interval < 5 {
		interval = 5
	}
	interval = interval * 2

	for i := 0; i < 50; i++ {
		if w.conn.State.IsStop() {
			w.conn.State.Close()
			return
		}

		if i%interval == 0 {
			go w.ping()
		}

		err = w.updateHandshake()
		if err != nil {
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("connection: Check handshake status failed")

			w.conn.State.Close()
			return
		}

		if w.conn.State.IsStop() {
			w.conn.State.Close()
			return
		}

		if w.lastHandshake != 0 {
			w.connected = true
			w.conn.Data.Status = Connected
			w.conn.Data.Timestamp = time.Now().Unix() - 3
			w.conn.Data.UpdateEvent()
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	if w.lastHandshake == 0 {
		w.conn.Data.SendProfileEvent("handshake_timeout")

		logrus.WithFields(w.conn.Fields(logrus.Fields{
			"error": err,
		})).Error("connection: WireGuard handshake timeout")

		w.conn.State.Close()
		return
	}

	if !w.conn.Profile.DisableDns && w.conn.Data.DnsServers != nil &&
		len(w.conn.Data.DnsServers) > 0 && runtime.GOOS == "darwin" &&
		!config.Config.DisableWgDns {

		err := utils.SetScutilDns(w.conn.Id,
			w.conn.Data.DnsServers, w.conn.Data.DnsServers)
		if err != nil {
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("connection: Failed to set DNS servers")
		}
	}

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	for {
		if w.conn.State.IsStop() {
			w.conn.State.Close()
			return
		}

		for i := 0; i < interval; i++ {
			time.Sleep(500 * time.Millisecond)
			if w.conn.State.IsStopFast() {
				w.conn.State.Close()
				return
			}
		}

		time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
		if w.conn.State.IsStopFast() {
			w.conn.State.Close()
			return
		}

		start := time.Now()
		var data *PingData
		var final bool
		for i := 0; i < 8; i++ {
			data, final, err = w.ping()
			if err == nil || final || time.Since(start) > 15*time.Second {
				break
			}

			if time.Since(lastRetryLogged) > 30*time.Minute {
				lastRetryLogged = time.Now()
				logrus.WithFields(w.conn.Fields(logrus.Fields{
					"error": err,
				})).Error("connection: Retrying keep alive")
			}

			if w.conn.State.IsStop() {
				w.conn.State.Close()
				return
			}

			time.Sleep(1 * time.Second)
		}
		if err != nil {
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("connection: Keepalive failed")

			w.conn.State.Close()
			return
		}

		if w.conn.State.IsStop() {
			w.conn.State.Close()
			return
		}

		if data == nil || !data.Status {
			logrus.WithFields(w.conn.Fields(nil)).Error(
				"profile: Keepalive missing status")

			w.conn.State.Close()
			return
		}
	}
}

func (w *Wg) updateHandshake() (err error) {
	iface := ""
	if runtime.GOOS == "darwin" {
		iface = w.conn.Data.WgTunIface
	} else {
		iface = w.conn.Data.Iface
	}

	output, err := utils.ExecCombinedOutputLogged(
		[]string{
			"No such device",
			"access interface",
		},
		w.wgPath, "show", iface,
		"latest-handshakes",
	)
	if err != nil {
		return
	}

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if fields[0] == w.serverPubKey {
			lastHandshake, e := strconv.Atoi(fields[1])
			if e != nil {
				continue
			}

			w.lastHandshake = lastHandshake
			return
		}
	}

	w.lastHandshake = 0
	return
}

func (w *Wg) ping() (data *PingData, final bool, err error) {
	scheme := "https"
	if w.conn.Data.WebNoSsl {
		scheme = "http"
	}
	host := fmt.Sprintf("%s:%d", w.conn.Data.GatewayAddr, w.conn.Data.WebPort)
	reqUrl := w.conn.Client.GetUrl(scheme, host, "wg")

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	ciph, reqBx, err := w.conn.Client.InitBox()
	if err != nil {
		return
	}

	ctx := w.conn.Client.GetContext()
	defer ctx.Cancel()

	res, err := w.conn.Client.EncRequest(ctx, "PUT", reqUrl, ciph, reqBx)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode >= 400 && res.StatusCode < 500 {
			final = true
		}

		err = utils.LogRequestError(res,
			"connection: Keepalive failed")
		return
	}

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	respBx := &RespBox{}
	err = json.NewDecoder(res.Body).Decode(respBx)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to parse response body"),
		}
		return
	}

	if w.conn.State.IsStop() {
		w.conn.State.Close()
		return
	}

	data = &PingData{}
	err = w.conn.Client.DecryptRespBox(ciph, respBx, data)
	if err != nil {
		return
	}

	return
}

func (w *Wg) writeWgConf(data *WgConf) (err error) {
	allowedIps := []string{}
	if data.Routes != nil {
		for _, route := range data.Routes {
			if w.conn.Profile.DisableGateway && route.Network == "0.0.0.0/0" {
				continue
			}

			if route.NetGateway {
				// TODO wireguard exclude
				//allowedIps = append(allowedIps, "!"+route.Network)
			} else {
				allowedIps = append(allowedIps, route.Network)
			}
		}
	}
	if data.Routes6 != nil {
		for _, route := range data.Routes6 {
			if w.conn.Profile.DisableGateway && route.Network == "::/0" {
				continue
			}

			if route.NetGateway {
				// TODO wireguard exclude
				//allowedIps = append(allowedIps, "!"+route.Network)
			} else {
				allowedIps = append(allowedIps, route.Network)
			}
		}
	}

	addr := data.Address
	if data.Address6 != "" {
		addr += "," + data.Address6
	}

	templData := WgConfData{
		Address:    addr,
		PrivateKey: w.privateKey,
		PublicKey:  data.PublicKey,
		AllowedIps: strings.Join(allowedIps, ","),
		Endpoint:   fmt.Sprintf("%s:%d", data.Hostname, data.Port),
	}

	if data.Mtu != 0 {
		templData.HasMtu = true
		templData.Mtu = data.Mtu
	}

	if !w.conn.Profile.DisableDns && len(data.DnsServers) > 0 &&
		runtime.GOOS != "darwin" {

		templData.HasDns = true
		templData.DnsServers = strings.Join(data.DnsServers, ",")
	}

	if !w.conn.Profile.DisableDns && len(data.SearchDomains) > 0 &&
		runtime.GOOS != "darwin" {

		templData.HasDns = true
		if templData.DnsServers != "" {
			templData.DnsServers += ","
		}
		templData.DnsServers += strings.Join(data.SearchDomains, ",")
	}

	output := &bytes.Buffer{}
	err = WgConfTempl.Execute(output, templData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to exec wg template"),
		}
		return
	}

	rootDir, rootDir2, err := GetWgConfDir()
	if err != nil {
		return
	}

	if rootDir != "" {
		err = platform.MkdirSecure(rootDir)
		if err != nil {
			return
		}
	}

	if rootDir2 != "" {
		err = platform.MkdirSecure(rootDir2)
		if err != nil {
			return
		}
	}

	w.wgConfPath = filepath.Join(rootDir, w.conn.Data.Iface+".conf")
	w.conn.State.AddPath(w.wgConfPath)

	_ = os.Remove(w.wgConfPath)
	err = ioutil.WriteFile(
		w.wgConfPath,
		[]byte(output.String()),
		os.FileMode(0600),
	)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write wg conf"),
		}
		return
	}

	if rootDir2 != "" {
		w.wgConfPath2 = filepath.Join(rootDir2, w.conn.Data.Iface+".conf")
		w.conn.State.AddPath(w.wgConfPath2)

		_ = os.Remove(w.wgConfPath2)
		err = ioutil.WriteFile(
			w.wgConfPath2,
			[]byte(output.String()),
			os.FileMode(0600),
		)
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrap(err, "profile: Failed to write wg conf"),
			}
			return
		}
	}

	return
}

func (w *Wg) confWg(data *WgConf) (err error) {
	w.conn.Data.ClientAddr = data.Address
	w.conn.Data.ServerAddr = data.Hostname
	w.conn.Data.GatewayAddr = data.Gateway
	w.conn.Data.GatewayAddr6 = data.Gateway6
	w.conn.Data.PingIntervalWg = data.PingInterval
	w.conn.Data.PingTimeoutWg = data.PingTimeout
	w.conn.Data.WebPort = data.WebPort
	w.conn.Data.WebNoSsl = data.WebNoSsl
	w.conn.Data.DnsServers = data.DnsServers
	w.conn.Data.SearchDomains = data.SearchDomains
	w.conn.Data.Routes = data.Routes
	w.conn.Data.Routes6 = data.Routes6

	w.serverPubKey = data.PublicKey

	switch runtime.GOOS {
	case "darwin":
		err = w.confWgMac()
		break
	case "windows":
		err = w.confWgWin()
		break
	case "linux":
		err = w.confWgLinux()
		break
	default:
		panic("profile: Not implemented")
	}
	if err != nil {
		return
	}

	w.applyRouteMetrics(data)

	return
}

func (w *Wg) confWgLinux() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for i := 0; i < 3; i++ {
		_, _ = utils.ExecCombinedOutput(
			w.wgQuickPath, "down", w.conn.Data.Iface,
		)

		if i == 0 {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			w.wgQuickPath,
			"up", w.conn.Data.Iface,
		)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	return
}

func (w *Wg) confWgMac() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	output := ""
	for i := 0; i < 3; i++ {
		_, _ = utils.ExecCombinedOutput(
			w.bashPath, w.wgQuickPath, "down", w.conn.Data.Iface,
		)

		if i == 0 {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}

		output, err = utils.ExecCombinedOutputLogged(
			nil,
			w.bashPath,
			w.wgQuickPath,
			"up", w.conn.Data.Iface,
		)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	tunIface := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "INFO") {
			match := wgIfaceMacReg.FindStringSubmatch(line)
			if match != nil && len(match) >= 2 {
				tunIface = match[1]
			}
		}
	}

	if tunIface == "" {
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Interface for") {
				lines := strings.Fields(line)
				tunIface = lines[len(lines)-1]
			}
		}
	}

	if tunIface == "" {
		err = &errortypes.ParseError{
			errors.New("profile: Failed to parse wg interface output"),
		}
		return
	}
	w.conn.Data.WgTunIface = tunIface

	return
}

func (w *Wg) confWgWin() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for i := 0; i < 3; i++ {
		_, _ = utils.ExecCombinedOutput(
			"sc.exe", "stop", fmt.Sprintf(
				"WireGuardTunnel$%s", w.conn.Data.Iface),
		)
		time.Sleep(100 * time.Millisecond)
		_, _ = utils.ExecCombinedOutput(
			"sc.exe", "delete", fmt.Sprintf(
				"WireGuardTunnel$%s", w.conn.Data.Iface),
		)

		if i == 0 {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			GetWgUtilPath(),
			"/installtunnelservice", w.wgConfPath,
		)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	return
}

func (w *Wg) applyRouteMetrics(data *WgConf) {
	switch runtime.GOOS {
	case "linux":
		w.applyRouteMetricsLinux(data)
		break
	case "windows":
		w.applyRouteMetricsWin(data)
		break
	}
}

func (w *Wg) applyRouteMetricsLinux(data *WgConf) {
	time.Sleep(200 * time.Millisecond)

	iface := w.conn.Data.Iface

	if data.Routes != nil {
		for _, route := range data.Routes {
			if route.Metric == 0 || route.NetGateway {
				continue
			}
			if route.Network == "0.0.0.0/0" {
				continue
			}

			utils.ExecCombinedOutput(
				"ip", "-4", "route", "del",
				route.Network, "dev", iface,
			)

			_, err := utils.ExecCombinedOutputLogged(
				nil,
				"ip", "-4", "route", "add",
				route.Network, "dev", iface,
				"metric", strconv.Itoa(route.Metric),
			)
			if err != nil {
				logrus.WithFields(w.conn.Fields(logrus.Fields{
					"network": route.Network,
					"metric":  route.Metric,
					"error":   err,
				})).Warn("connection: Failed to set IPv4 route metric")
			}
		}
	}

	if data.Routes6 != nil {
		for _, route := range data.Routes6 {
			if route.Metric == 0 || route.NetGateway {
				continue
			}
			if route.Network == "::/0" {
				continue
			}

			utils.ExecCombinedOutput(
				"ip", "-6", "route", "del",
				route.Network, "dev", iface,
			)

			_, err := utils.ExecCombinedOutputLogged(
				nil,
				"ip", "-6", "route", "add",
				route.Network, "dev", iface,
				"metric", strconv.Itoa(route.Metric),
			)
			if err != nil {
				logrus.WithFields(w.conn.Fields(logrus.Fields{
					"network": route.Network,
					"metric":  route.Metric,
					"error":   err,
				})).Warn("connection: Failed to set IPv6 route metric")
			}
		}
	}
}

func (w *Wg) applyRouteMetricsWin(data *WgConf) {
	time.Sleep(300 * time.Millisecond)

	iface := w.conn.Data.Iface

	if data.Routes != nil {
		for _, route := range data.Routes {
			if route.Metric == 0 || route.NetGateway {
				continue
			}
			if route.Network == "0.0.0.0/0" {
				continue
			}

			_, err := utils.ExecCombinedOutputLogged(
				nil,
				"netsh", "interface", "ipv4", "set", "route",
				route.Network, iface,
				fmt.Sprintf("metric=%d", route.Metric),
			)
			if err != nil {
				logrus.WithFields(w.conn.Fields(logrus.Fields{
					"network": route.Network,
					"metric":  route.Metric,
					"error":   err,
				})).Warn("connection: Failed to set IPv4 route metric")
			}
		}
	}

	if data.Routes6 != nil {
		for _, route := range data.Routes6 {
			if route.Metric == 0 || route.NetGateway {
				continue
			}
			if route.Network == "::/0" {
				continue
			}

			_, err := utils.ExecCombinedOutputLogged(
				nil,
				"netsh", "interface", "ipv6", "set", "route",
				route.Network, iface,
				fmt.Sprintf("metric=%d", route.Metric),
			)
			if err != nil {
				logrus.WithFields(w.conn.Fields(logrus.Fields{
					"network": route.Network,
					"metric":  route.Metric,
					"error":   err,
				})).Warn("connection: Failed to set IPv6 route metric")
			}
		}
	}
}

func (w *Wg) clearWgLinux() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.conn.Data.Iface != "" {
		utils.ExecCombinedOutputLogged(
			[]string{
				"does not exist",
				"is not a",
			},
			w.wgQuickPath,
			"down", w.conn.Data.Iface,
		)
	}
}

func (w *Wg) clearWgMac() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.conn.Data.Iface != "" {
		utils.ExecCombinedOutputLogged(
			[]string{
				"is not a",
			},
			w.bashPath,
			w.wgQuickPath,
			"down", w.conn.Data.Iface,
		)
	}
}

func (w *Wg) clearWgWin() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.conn.Data.Iface != "" {
		_, _ = utils.ExecCombinedOutput(
			"sc.exe", "stop",
			fmt.Sprintf("WireGuardTunnel$%s", w.conn.Data.Iface),
		)
		time.Sleep(100 * time.Millisecond)
		_, _ = utils.ExecCombinedOutput(
			"sc.exe", "delete",
			fmt.Sprintf("WireGuardTunnel$%s", w.conn.Data.Iface),
		)
	}
}

func (w *Wg) clearWg() {
	switch runtime.GOOS {
	case "linux":
		w.clearWgLinux()
		break
	case "darwin":
		w.clearWgMac()
		break
	case "windows":
		w.clearWgWin()
		break
	}

	network.InterfaceRelease(w.conn.Data.Iface)
}

func (w *Wg) Disconnect() {
	w.clearWg()

	return
}
