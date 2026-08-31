package connection

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/godbus/dbus/v5"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/sirupsen/logrus"
)

const (
	nmConnPrefix   = "pritunl-"
	nmFwmarkBase   = 51920
	nmProbeTimeout = 2 * time.Second
	nmConnectWait  = 30 * time.Second
	nmActivateWait = 30 * time.Second

	nmDest          = "org.freedesktop.NetworkManager"
	nmPath          = dbus.ObjectPath("/org/freedesktop/NetworkManager")
	nmSettingsPath  = dbus.ObjectPath("/org/freedesktop/NetworkManager/Settings")
	nmIface         = "org.freedesktop.NetworkManager"
	nmSettingsIface = "org.freedesktop.NetworkManager.Settings"
	nmConnIface     = "org.freedesktop.NetworkManager.Settings.Connection"
	nmDeviceIface   = "org.freedesktop.NetworkManager.Device"
	nmActiveIface   = "org.freedesktop.NetworkManager.Connection.Active"

	// NMSettingsAddConnection2Flags / NMSettingsUpdate2Flags
	nmFlagInMemory         = uint32(0x2)
	nmFlagBlockAutoconnect = uint32(0x20)

	// NMActiveConnectionState
	nmActiveStateActivated    = uint32(2)
	nmActiveStateDeactivating = uint32(3)
	nmActiveStateDeactivated  = uint32(4)

	// NMTernary
	nmTernaryDefault = int32(-1)
	nmTernaryTrue    = int32(1)
)

type nmSettings = map[string]map[string]dbus.Variant
type nmDict = map[string]dbus.Variant

var (
	nmIfaceNumReg = regexp.MustCompile("([0-9]+)$")
)

func nmConnName(iface string) string {
	return nmConnPrefix + iface
}

func nmFwmark(iface string) uint32 {
	match := nmIfaceNumReg.FindStringSubmatch(iface)
	if match != nil && len(match) >= 2 {
		n, err := strconv.Atoi(match[1])
		if err == nil {
			return uint32(nmFwmarkBase + n)
		}
	}
	return uint32(nmFwmarkBase)
}

func nmNewUuid() (uuid string, err error) {
	b := make([]byte, 16)
	_, err = rand.Read(b)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "wg: Failed to generate uuid"),
		}
		return
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	uuid = fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])

	return
}

func generateWgKey() (privateKey, publicKey string, err error) {
	seed := make([]byte, 32)
	_, err = rand.Read(seed)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "wg: Failed to generate private key"),
		}
		return
	}

	seed[0] &= 248
	seed[31] &= 127
	seed[31] |= 64

	priv, err := ecdh.X25519().NewPrivateKey(seed)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "wg: Failed to parse private key"),
		}
		return
	}

	privateKey = base64.StdEncoding.EncodeToString(priv.Bytes())
	publicKey = base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes())

	return
}

func nmBus() (bus *dbus.Conn, err error) {
	bus, err = dbus.SystemBus()
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "wg: Failed to connect to system bus"),
		}
		return
	}

	return
}

func nmErr(err error, msg string) error {
	return &errortypes.RequestError{
		errors.Wrap(err, "wg: NetworkManager "+msg),
	}
}

// nmFindConns returns the paths of WireGuard connections whose id matches.
func nmFindConns(bus *dbus.Conn, match func(id string) bool) (
	paths []dbus.ObjectPath, err error) {

	paths = []dbus.ObjectPath{}

	connPaths := []dbus.ObjectPath{}
	err = bus.Object(nmDest, nmSettingsPath).Call(
		nmSettingsIface+".ListConnections", 0).Store(&connPaths)
	if err != nil {
		err = nmErr(err, "list connections failed")
		return
	}

	for _, pth := range connPaths {
		settings := nmSettings{}
		e := bus.Object(nmDest, pth).Call(
			nmConnIface+".GetSettings", 0).Store(&settings)
		if e != nil {
			continue
		}

		connSet := settings["connection"]
		if connSet == nil {
			continue
		}

		id, _ := connSet["id"].Value().(string)
		typ, _ := connSet["type"].Value().(string)

		if typ == "wireguard" && match(id) {
			paths = append(paths, pth)
		}
	}

	return
}

func nmDeleteConnPath(bus *dbus.Conn, pth dbus.ObjectPath) {
	if pth == "" {
		return
	}

	err := bus.Object(nmDest, pth).Call(nmConnIface+".Delete", 0).Err
	if err != nil {
		if dbusErr, ok := err.(dbus.Error); ok &&
			strings.Contains(dbusErr.Name, "UnknownMethod") ||
			strings.Contains(err.Error(), "UnknownObject") {

			return
		}

		logrus.WithFields(logrus.Fields{
			"path":  pth,
			"error": err,
		}).Warn("connection: Failed to delete NetworkManager connection")
	}
}

func nmDeleteConnName(bus *dbus.Conn, name string) {
	paths, err := nmFindConns(bus, func(id string) bool {
		return id == name
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Warn("connection: Failed to find NetworkManager connection")
		return
	}

	for _, pth := range paths {
		nmDeleteConnPath(bus, pth)
	}
}

func nmClean() {
	bus, err := nmBus()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("connection: NetworkManager unavailable")
		return
	}

	paths, err := nmFindConns(bus, func(id string) bool {
		return strings.HasPrefix(id, nmConnPrefix+"wg")
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("connection: Failed to list NetworkManager connections")
		return
	}

	for _, pth := range paths {
		logrus.WithFields(logrus.Fields{
			"path": pth,
		}).Info("connection: Removing stale NetworkManager connection")

		nmDeleteConnPath(bus, pth)
	}
}

func nmAddrData(addr string) (data nmDict, err error) {
	prefix, e := netip.ParsePrefix(addr)
	if e != nil {
		ip, e2 := netip.ParseAddr(addr)
		if e2 != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "wg: Failed to parse address"),
			}
			return
		}
		prefix = netip.PrefixFrom(ip, ip.BitLen())
	}

	data = nmDict{
		"address": dbus.MakeVariant(prefix.Addr().String()),
		"prefix":  dbus.MakeVariant(uint32(prefix.Bits())),
	}

	return
}

// nmRoutes builds the route configuration. Without custom metrics
// NetworkManager derives routes from the peer AllowedIPs including the
// fwmark based default route handling. Custom metrics require explicit
// routes with default routes placed in the fwmark table so NetworkManager
// still adds the policy routing rules for the endpoint bypass.
func (w *Wg) nmRoutes(allowedIps []string) (peerRoutes bool, fwmark uint32,
	routes4, routes6 []nmDict, autoDef4, autoDef6 int32) {

	autoDef4 = nmTernaryDefault
	autoDef6 = nmTernaryDefault
	routes4 = []nmDict{}
	routes6 = []nmDict{}

	metrics := map[string]int{}
	if w.wgConf != nil {
		for _, route := range w.wgConf.Routes {
			if route.NetGateway || route.Metric == 0 ||
				!routeTracked(route.Network) {

				continue
			}
			metrics[route.Network] = route.Metric
		}
		for _, route := range w.wgConf.Routes6 {
			if route.NetGateway || route.Metric == 0 ||
				!routeTracked(route.Network) {

				continue
			}
			metrics[route.Network] = route.Metric
		}
	}

	hasMetric := false
	for _, network := range allowedIps {
		if metrics[network] != 0 {
			hasMetric = true
			break
		}
	}

	if !hasMetric {
		peerRoutes = true
		return
	}

	fwmark = nmFwmark(w.conn.Data.Iface)

	for _, network := range allowedIps {
		prefix, e := netip.ParsePrefix(network)
		if e != nil {
			continue
		}

		entry := nmDict{
			"dest":   dbus.MakeVariant(prefix.Addr().String()),
			"prefix": dbus.MakeVariant(uint32(prefix.Bits())),
		}

		if !routeTracked(network) {
			entry["table"] = dbus.MakeVariant(fwmark)
			if prefix.Addr().Is6() {
				autoDef6 = nmTernaryTrue
			} else {
				autoDef4 = nmTernaryTrue
			}
		} else if metric := metrics[network]; metric != 0 {
			entry["metric"] = dbus.MakeVariant(uint32(metric))
		}

		if prefix.Addr().Is6() {
			routes6 = append(routes6, entry)
		} else {
			routes4 = append(routes4, entry)
		}
	}

	return
}

// nmSettings builds the NetworkManager connection profile equivalent of
// the wg-quick configuration.
func (w *Wg) nmSettings(allowedIps []string) (settings nmSettings, err error) {
	data := w.wgConf
	if data == nil {
		err = &errortypes.ParseError{
			errors.New("wg: Missing WireGuard configuration"),
		}
		return
	}

	peerRoutes, fwmark, routes4, routes6, autoDef4, autoDef6 :=
		w.nmRoutes(allowedIps)

	peer := nmDict{
		"public-key": dbus.MakeVariant(data.PublicKey),
		"endpoint": dbus.MakeVariant(
			net.JoinHostPort(data.Hostname, strconv.Itoa(data.Port))),
		"allowed-ips": dbus.MakeVariant(allowedIps),
	}

	wgSet := nmDict{
		"private-key":            dbus.MakeVariant(w.privateKey),
		"private-key-flags":      dbus.MakeVariant(uint32(0)),
		"peers":                  dbus.MakeVariant([]nmDict{peer}),
		"peer-routes":            dbus.MakeVariant(peerRoutes),
		"fwmark":                 dbus.MakeVariant(fwmark),
		"ip4-auto-default-route": dbus.MakeVariant(autoDef4),
		"ip6-auto-default-route": dbus.MakeVariant(autoDef6),
	}
	if data.Mtu != 0 {
		wgSet["mtu"] = dbus.MakeVariant(uint32(data.Mtu))
	}

	connSet := nmDict{
		"id":             dbus.MakeVariant(w.nmName),
		"uuid":           dbus.MakeVariant(w.nmUuid),
		"type":           dbus.MakeVariant("wireguard"),
		"interface-name": dbus.MakeVariant(w.conn.Data.Iface),
		"autoconnect":    dbus.MakeVariant(false),
	}

	dns4 := []uint32{}
	dns6 := [][]byte{}
	searchDomains := []string{}
	hasDns := false
	if !w.conn.Profile.DisableDns {
		for _, server := range data.DnsServers {
			addr, e := netip.ParseAddr(server)
			if e != nil {
				continue
			}

			hasDns = true
			if addr.Is4() {
				b := addr.As4()
				// ipv4.dns is au with in_addr_t in native representation
				dns4 = append(dns4, binary.NativeEndian.Uint32(b[:]))
			} else {
				b := addr.As16()
				dns6 = append(dns6, b[:])
			}
		}

		searchDomains = append(searchDomains, data.SearchDomains...)
	}

	addr4, err := nmAddrData(data.Address)
	if err != nil {
		return
	}

	ip4Set := nmDict{
		"method":       dbus.MakeVariant("manual"),
		"address-data": dbus.MakeVariant([]nmDict{addr4}),
	}
	if len(dns4) > 0 {
		ip4Set["dns"] = dbus.MakeVariant(dns4)
	}
	if len(searchDomains) > 0 {
		ip4Set["dns-search"] = dbus.MakeVariant(searchDomains)
	}
	if hasDns {
		// Match wg-quick behavior of routing all DNS queries through the
		// tunnel when DNS servers are provided.
		ip4Set["dns-priority"] = dbus.MakeVariant(int32(-1))
	}
	if len(routes4) > 0 {
		ip4Set["route-data"] = dbus.MakeVariant(routes4)
	}

	ip6Set := nmDict{}
	if data.Address6 != "" {
		addr6, e := nmAddrData(data.Address6)
		if e != nil {
			err = e
			return
		}

		ip6Set["method"] = dbus.MakeVariant("manual")
		ip6Set["address-data"] = dbus.MakeVariant([]nmDict{addr6})
		if len(dns6) > 0 {
			ip6Set["dns"] = dbus.MakeVariant(dns6)
		}
		if len(searchDomains) > 0 {
			ip6Set["dns-search"] = dbus.MakeVariant(searchDomains)
		}
		if hasDns {
			ip6Set["dns-priority"] = dbus.MakeVariant(int32(-1))
		}
		if len(routes6) > 0 {
			ip6Set["route-data"] = dbus.MakeVariant(routes6)
		}
	} else {
		ip6Set["method"] = dbus.MakeVariant("disabled")
	}

	settings = nmSettings{
		"connection": connSet,
		"wireguard":  wgSet,
		"ipv4":       ip4Set,
		"ipv6":       ip6Set,
	}

	return
}

func (w *Wg) nmWaitActive(bus *dbus.Conn, activePath dbus.ObjectPath) (
	err error) {

	obj := bus.Object(nmDest, activePath)
	start := time.Now()

	for time.Since(start) < nmActivateWait {
		if w.conn.State.IsStop() {
			err = &errortypes.RequestError{
				errors.New("wg: Connection stopped during activation"),
			}
			return
		}

		variant, e := obj.GetProperty(nmActiveIface + ".State")
		if e != nil {
			err = nmErr(e, "active connection lost")
			return
		}

		state, _ := variant.Value().(uint32)
		switch state {
		case nmActiveStateActivated:
			return
		case nmActiveStateDeactivating, nmActiveStateDeactivated:
			err = &errortypes.RequestError{
				errors.New("wg: NetworkManager activation failed"),
			}
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	err = &errortypes.RequestError{
		errors.New("wg: NetworkManager activation timeout"),
	}
	return
}

func (w *Wg) nmActivate(bus *dbus.Conn, connPath dbus.ObjectPath) (
	err error) {

	for i := 0; i < 3; i++ {
		if i != 0 {
			time.Sleep(500 * time.Millisecond)
		}

		if w.conn.State.IsStop() {
			err = &errortypes.RequestError{
				errors.New("wg: Connection stopped during activation"),
			}
			return
		}

		activePath := dbus.ObjectPath("")
		err = bus.Object(nmDest, nmPath).Call(
			nmIface+".ActivateConnection", 0,
			connPath, dbus.ObjectPath("/"), dbus.ObjectPath("/"),
		).Store(&activePath)
		if err != nil {
			err = nmErr(err, "activate connection failed")
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Warn("connection: NetworkManager activate failed")
			continue
		}

		err = w.nmWaitActive(bus, activePath)
		if err != nil {
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Warn("connection: NetworkManager activation failed")
			continue
		}

		return
	}

	return
}

// nmApply adds the WireGuard profile to NetworkManager as an in-memory
// connection and activates it, replacing any existing connection for the
// interface. Caller must hold w.lock.
func (w *Wg) nmApply(allowedIps []string) (err error) {
	bus, err := nmBus()
	if err != nil {
		return
	}

	iface := w.conn.Data.Iface
	w.nmName = nmConnName(iface)

	if w.nmUuid == "" {
		w.nmUuid, err = nmNewUuid()
		if err != nil {
			return
		}
	}

	if w.nmConnPath != "" {
		nmDeleteConnPath(bus, dbus.ObjectPath(w.nmConnPath))
		w.nmConnPath = ""
	}
	nmDeleteConnName(bus, w.nmName)

	settings, err := w.nmSettings(allowedIps)
	if err != nil {
		return
	}

	connPath := dbus.ObjectPath("")
	result := nmDict{}
	err = bus.Object(nmDest, nmSettingsPath).Call(
		nmSettingsIface+".AddConnection2", 0,
		settings,
		nmFlagInMemory|nmFlagBlockAutoconnect,
		nmDict{},
	).Store(&connPath, &result)
	if err != nil {
		err = nmErr(err, "add connection failed")
		return
	}
	w.nmConnPath = string(connPath)

	err = w.nmActivate(bus, connPath)
	if err != nil {
		nmDeleteConnPath(bus, connPath)
		w.nmConnPath = ""
		return
	}

	logrus.WithFields(w.conn.Fields(logrus.Fields{
		"nm_name": w.nmName,
		"nm_path": w.nmConnPath,
	})).Info("connection: NetworkManager WireGuard connection active")

	return
}

func (w *Wg) confWgNm() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err = w.nmApply(w.allowedIps)
	if err != nil {
		return
	}

	return
}

// nmReapply updates the peer allowed IPs and routes in place using
// Update2 followed by a device Reapply which NetworkManager supports for
// WireGuard peers without restarting the tunnel. Caller must hold w.lock.
func (w *Wg) nmReapply(allowedIps []string) (err error) {
	if w.nmConnPath == "" {
		err = w.nmApply(allowedIps)
		if err != nil {
			return
		}
		w.allowedIps = allowedIps
		return
	}

	bus, err := nmBus()
	if err != nil {
		return
	}

	settings, err := w.nmSettings(allowedIps)
	if err != nil {
		return
	}

	result := nmDict{}
	err = bus.Object(nmDest, dbus.ObjectPath(w.nmConnPath)).Call(
		nmConnIface+".Update2", 0,
		settings,
		nmFlagInMemory|nmFlagBlockAutoconnect,
		nmDict{},
	).Store(&result)
	if err != nil {
		err = nmErr(err, "update connection failed")
		return
	}

	devPath := dbus.ObjectPath("")
	err = bus.Object(nmDest, nmPath).Call(
		nmIface+".GetDeviceByIpIface", 0, w.conn.Data.Iface,
	).Store(&devPath)
	if err != nil {
		err = nmErr(err, "get device failed")
		return
	}

	err = bus.Object(nmDest, devPath).Call(
		nmDeviceIface+".Reapply", 0,
		nmSettings{}, uint64(0), uint32(0),
	).Err
	if err != nil {
		err = nmErr(err, "reapply failed")
		return
	}

	w.allowedIps = allowedIps

	return
}

func (w *Wg) clearWgNm() {
	w.lock.Lock()
	defer w.lock.Unlock()

	bus, err := nmBus()
	if err != nil {
		logrus.WithFields(w.conn.Fields(logrus.Fields{
			"error": err,
		})).Warn("connection: NetworkManager unavailable")
		return
	}

	if w.nmConnPath != "" {
		nmDeleteConnPath(bus, dbus.ObjectPath(w.nmConnPath))
		w.nmConnPath = ""
	}

	if w.conn.Data.Iface != "" {
		nmDeleteConnName(bus, nmConnName(w.conn.Data.Iface))
	}
}

// nmProbe checks tunnel reachability with a TCP connection to the gateway
// web port. Handshake state is unavailable without CAP_NET_ADMIN.
func (w *Wg) nmProbe() bool {
	if w.conn.Data.GatewayAddr == "" || w.conn.Data.WebPort == 0 {
		return false
	}

	host := net.JoinHostPort(
		w.conn.Data.GatewayAddr,
		strconv.Itoa(w.conn.Data.WebPort),
	)

	conn, err := net.DialTimeout("tcp", host, nmProbeTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()

	return true
}

func (w *Wg) nmWaitConnected() {
	start := time.Now()

	for time.Since(start) < nmConnectWait {
		if w.conn.State.IsStop() {
			return
		}

		if w.nmProbe() {
			w.lastHandshake = int(time.Now().Unix())
			return
		}

		time.Sleep(500 * time.Millisecond)
	}
}
