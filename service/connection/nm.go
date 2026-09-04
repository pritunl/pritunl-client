package connection

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"net"
	"net/netip"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/sirupsen/logrus"
)

const (
	nmConnPrefix     = "pritunl-"
	nmFwmarkBase     = 51920
	nmProbeTimeout   = 2 * time.Second
	nmConnectWait    = 30 * time.Second
	nmActivateWait   = 30 * time.Second
	nmDeactivateWait = 2 * time.Second
	nmSignalBuffer   = 64

	nmDest          = "org.freedesktop.NetworkManager"
	nmPath          = dbus.ObjectPath("/org/freedesktop/NetworkManager")
	nmSettingsPath  = dbus.ObjectPath("/org/freedesktop/NetworkManager/Settings")
	nmIface         = "org.freedesktop.NetworkManager"
	nmSettingsIface = "org.freedesktop.NetworkManager.Settings"
	nmConnIface     = "org.freedesktop.NetworkManager.Settings.Connection"
	nmDeviceIface   = "org.freedesktop.NetworkManager.Device"
	nmActiveIface   = "org.freedesktop.NetworkManager.Connection.Active"
	nmStateSignal   = nmActiveIface + ".StateChanged"

	// user.data keys marking connections owned by this client. NM
	// requires a namespace dot in every key.
	nmUserKeyManaged = "pritunl.managed"
	nmUserKeyProfile = "pritunl.profile"

	// NMSettingsAddConnection2Flags / NMSettingsUpdate2Flags
	nmFlagInMemory         = uint32(0x2)
	nmFlagVolatile         = uint32(0x10)
	nmFlagBlockAutoconnect = uint32(0x20)

	// NMActiveConnectionState
	nmActiveStateActivated    = uint32(2)
	nmActiveStateDeactivating = uint32(3)
	nmActiveStateDeactivated  = uint32(4)

	// Synthetic state delivered to watchers when the system bus connection
	// is lost. NetworkManager tears down every dbus-client bound activation
	// when the bus name disappears, so the connection is effectively
	// deactivated even though the signal can no longer be received.
	nmActiveStateBusLost = ^uint32(0)

	// NMTernary
	nmTernaryDefault = int32(-1)
	nmTernaryTrue    = int32(1)

	dbusErrUnknownMethod = "org.freedesktop.DBus.Error.UnknownMethod"
	dbusErrUnknownObject = "org.freedesktop.DBus.Error.UnknownObject"
	nmErrUnknownConn     = "org.freedesktop.NetworkManager.Settings." +
		"Connection.Error.UnknownConnection"
	nmErrNotActive     = "org.freedesktop.NetworkManager.ConnectionNotActive"
	nmErrUnknownActive = "org.freedesktop.NetworkManager.UnknownConnection"
)

type nmSettings = map[string]map[string]dbus.Variant
type nmDict = map[string]dbus.Variant

var (
	nmIfaceNumReg = regexp.MustCompile("([0-9]+)$")

	// Fixed namespace for deriving deterministic UUIDv5 connection ids from
	// profile ids. Never change this value, startup garbage collection
	// relies on it to identify connections left by previous runs.
	nmUuidNamespace = uuid.MustParse("5d3e1f7a-9c2b-4f8e-a6d1-7b0c4e9f2a13")
)

// nmBusConn is the single system bus connection shared by every
// NetworkManager operation for the lifetime of the process. Activations are
// requested with bind-activation=dbus-client which ties them to the unique
// name of this connection, so it must never be closed while a tunnel is up.
// Signal subscriptions for active connection state changes are made on the
// same connection and dispatched to registered watchers by object path.
type nmBusConn struct {
	lock      sync.Mutex
	conn      *dbus.Conn
	watchLock sync.Mutex
	watchers  map[dbus.ObjectPath]func(path dbus.ObjectPath, state uint32)
}

var nmBusGlobal = &nmBusConn{
	watchers: map[dbus.ObjectPath]func(path dbus.ObjectPath, state uint32){},
}

func nmConnName(iface string) string {
	return nmConnPrefix + iface
}

func nmProfileUuid(profileId string) string {
	return uuid.NewSHA1(nmUuidNamespace, []byte(profileId)).String()
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

// nmBus returns the shared system bus connection, connecting on first use.
// A connection that was lost (proxy restart) is replaced, which is safe
// because NetworkManager already tore down every activation bound to the
// old unique name.
func nmBus() (bus *dbus.Conn, err error) {
	b := nmBusGlobal

	b.lock.Lock()
	defer b.lock.Unlock()

	if b.conn != nil {
		if b.conn.Connected() {
			bus = b.conn
			return
		}

		logrus.Warn("connection: System bus connection lost, reconnecting")
		b.conn = nil
	}

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "wg: Failed to connect to system bus"),
		}
		return
	}

	err = conn.AddMatchSignal(
		dbus.WithMatchInterface(nmActiveIface),
		dbus.WithMatchMember("StateChanged"),
	)
	if err != nil {
		_ = conn.Close()
		err = &errortypes.RequestError{
			errors.Wrap(err, "wg: Failed to subscribe to NetworkManager signals"),
		}
		return
	}

	sigCh := make(chan *dbus.Signal, nmSignalBuffer)
	conn.Signal(sigCh)
	go b.dispatch(sigCh)

	b.conn = conn
	bus = conn

	logrus.WithFields(logrus.Fields{
		"unique_name": conn.Names(),
	}).Info("connection: System bus connected")

	return
}

// dispatch delivers active connection state changes to watchers. The
// channel is closed by godbus when the underlying connection is lost, at
// which point every watcher is told the activation is gone.
func (b *nmBusConn) dispatch(sigCh chan *dbus.Signal) {
	for sig := range sigCh {
		if sig == nil || sig.Name != nmStateSignal || len(sig.Body) < 1 {
			continue
		}

		state, ok := sig.Body[0].(uint32)
		if !ok {
			continue
		}

		b.watchLock.Lock()
		fn := b.watchers[sig.Path]
		b.watchLock.Unlock()

		if fn != nil {
			go fn(sig.Path, state)
		}
	}

	b.watchLock.Lock()
	watchers := b.watchers
	b.watchers = map[dbus.ObjectPath]func(
		path dbus.ObjectPath, state uint32){}
	b.watchLock.Unlock()

	if len(watchers) > 0 {
		logrus.WithFields(logrus.Fields{
			"watchers": len(watchers),
		}).Warn("connection: System bus lost with active NetworkManager " +
			"connections")
	}

	for pth, fn := range watchers {
		go fn(pth, nmActiveStateBusLost)
	}
}

func nmWatchActive(pth dbus.ObjectPath,
	fn func(path dbus.ObjectPath, state uint32)) {

	b := nmBusGlobal
	b.watchLock.Lock()
	b.watchers[pth] = fn
	b.watchLock.Unlock()
}

func nmUnwatchActive(pth dbus.ObjectPath) {
	b := nmBusGlobal
	b.watchLock.Lock()
	delete(b.watchers, pth)
	b.watchLock.Unlock()
}

func nmErr(err error, msg string) error {
	return &errortypes.RequestError{
		errors.Wrap(err, "wg: NetworkManager "+msg),
	}
}

func nmErrName(err error) string {
	if dbusErr, ok := err.(dbus.Error); ok {
		return dbusErr.Name
	}
	if dbusErr, ok := err.(*dbus.Error); ok && dbusErr != nil {
		return dbusErr.Name
	}
	return ""
}

func nmIsUnknownMethod(err error) bool {
	return nmErrName(err) == dbusErrUnknownMethod
}

// nmIsGone reports errors meaning the connection or activation no longer
// exists, which callers deleting or deactivating treat as success.
func nmIsGone(err error) bool {
	switch nmErrName(err) {
	case dbusErrUnknownMethod, dbusErrUnknownObject, nmErrUnknownConn,
		nmErrNotActive, nmErrUnknownActive:

		return true
	}
	return false
}

func nmCurrentUser() string {
	usr, err := user.Current()
	if err != nil || usr == nil || usr.Username == "" {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("connection: Failed to get current user for " +
			"NetworkManager permissions")
		return ""
	}
	return usr.Username
}

// nmConnInfo is the subset of connection settings used for identification.
type nmConnInfo struct {
	path    dbus.ObjectPath
	id      string
	uuid    string
	typ     string
	managed bool
	profile string
}

// nmListConns reads the identifying settings of every visible connection.
// Connections that cannot be read (permissions) are skipped.
func nmListConns(bus *dbus.Conn) (conns []*nmConnInfo, err error) {
	conns = []*nmConnInfo{}

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

		info := &nmConnInfo{
			path: pth,
		}
		info.id, _ = connSet["id"].Value().(string)
		info.uuid, _ = connSet["uuid"].Value().(string)
		info.typ, _ = connSet["type"].Value().(string)

		if userSet := settings["user"]; userSet != nil {
			data, _ := userSet["data"].Value().(map[string]string)
			info.managed = data[nmUserKeyManaged] == "1"
			info.profile = data[nmUserKeyProfile]
		}

		conns = append(conns, info)
	}

	return
}

// nmDeleteConnPath deletes a settings connection. A connection that is
// already gone is success, volatile connections vanish on their own.
func nmDeleteConnPath(bus *dbus.Conn, pth dbus.ObjectPath) {
	if pth == "" {
		return
	}

	err := bus.Object(nmDest, pth).Call(nmConnIface+".Delete", 0).Err
	if err != nil && !nmIsGone(err) {
		logrus.WithFields(logrus.Fields{
			"path":  pth,
			"error": err,
		}).Warn("connection: Failed to delete NetworkManager connection")
	}
}

// nmDeleteConnUuid removes any existing connection with the uuid so a new
// connection with the same deterministic uuid can be added.
func nmDeleteConnUuid(bus *dbus.Conn, connUuid string) {
	if connUuid == "" {
		return
	}

	pth := dbus.ObjectPath("")
	err := bus.Object(nmDest, nmSettingsPath).Call(
		nmSettingsIface+".GetConnectionByUuid", 0, connUuid).Store(&pth)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"uuid": connUuid,
		"path": pth,
	}).Info("connection: Replacing existing NetworkManager connection")

	nmDeleteConnPath(bus, pth)
}

func nmDeleteConnName(bus *dbus.Conn, name string) {
	conns, err := nmListConns(bus)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Warn("connection: Failed to find NetworkManager connection")
		return
	}

	for _, info := range conns {
		if info.typ == "wireguard" && info.id == name {
			nmDeleteConnPath(bus, info.path)
		}
	}
}

// nmDeactivate deactivates an active connection with a short timeout. A
// volatile connection is deleted by NetworkManager once deactivated.
func nmDeactivate(bus *dbus.Conn, activePath dbus.ObjectPath) {
	if activePath == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), nmDeactivateWait)
	defer cancel()

	err := bus.Object(nmDest, nmPath).CallWithContext(
		ctx, nmIface+".DeactivateConnection", 0, activePath).Err
	if err != nil && !nmIsGone(err) {
		logrus.WithFields(logrus.Fields{
			"active_path": activePath,
			"error":       err,
		}).Warn("connection: Failed to deactivate NetworkManager connection")
	}
}

// nmClean runs at startup before any connection is created and removes
// connections left behind by an unclean exit: anything carrying the
// pritunl.managed marker, anything with the deterministic uuid of a known
// profile and legacy connections from older builds matched by name.
func nmClean() {
	bus, err := nmBus()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("connection: NetworkManager unavailable")
		return
	}

	knownUuids := map[string]bool{}
	sprfls, err := sprofile.GetAll()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("connection: Failed to load profiles for NetworkManager clean")
		err = nil
	} else {
		for _, sprfl := range sprfls {
			knownUuids[nmProfileUuid(sprfl.Id)] = true
		}
	}

	conns, err := nmListConns(bus)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("connection: Failed to list NetworkManager connections")
		return
	}

	for _, info := range conns {
		reason := ""
		switch {
		case info.managed:
			reason = "managed"
		case knownUuids[info.uuid]:
			reason = "profile_uuid"
		case info.typ == "wireguard" &&
			strings.HasPrefix(info.id, nmConnPrefix+"wg"):

			reason = "legacy_name"
		default:
			continue
		}

		logrus.WithFields(logrus.Fields{
			"path":    info.path,
			"id":      info.id,
			"uuid":    info.uuid,
			"profile": info.profile,
			"reason":  reason,
		}).Info("connection: Removing stale NetworkManager connection")

		nmDeleteConnPath(bus, info.path)
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

	// User owned connection so it can be modified and deleted by the same
	// user under settings.modify.own without a polkit prompt.
	if username := nmCurrentUser(); username != "" {
		connSet["permissions"] = dbus.MakeVariant(
			[]string{"user:" + username + ":"})
	}

	userSet := nmDict{
		"data": dbus.MakeVariant(map[string]string{
			nmUserKeyManaged: "1",
			nmUserKeyProfile: w.conn.Profile.Id,
		}),
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
		"user":       userSet,
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

// nmAddConnLegacy adds an in-memory connection on NetworkManager versions
// without AddAndActivateConnection2. Such connections are not volatile and
// must be deleted explicitly.
func nmAddConnLegacy(bus *dbus.Conn, settings nmSettings) (
	connPath dbus.ObjectPath, err error) {

	result := nmDict{}
	err = bus.Object(nmDest, nmSettingsPath).Call(
		nmSettingsIface+".AddConnection2", 0,
		settings,
		nmFlagInMemory|nmFlagBlockAutoconnect,
		nmDict{},
	).Store(&connPath, &result)
	if err == nil {
		return
	}
	if !nmIsUnknownMethod(err) {
		err = nmErr(err, "add connection failed")
		return
	}

	err = bus.Object(nmDest, nmSettingsPath).Call(
		nmSettingsIface+".AddConnectionUnsaved", 0,
		settings,
	).Store(&connPath)
	if err != nil {
		err = nmErr(err, "add unsaved connection failed")
		return
	}

	return
}

// nmAddActivate adds and activates the connection in one call. The
// connection is volatile so NetworkManager deletes it once deactivated and
// the activation is bound to this D-Bus client so NetworkManager
// deactivates it when the process or sandbox disappears. Falls back to a
// separate add and activate when the method is unavailable.
func nmAddActivate(bus *dbus.Conn, settings nmSettings) (connPath,
	activePath dbus.ObjectPath, volatile bool, err error) {

	opts := nmDict{
		"persist":         dbus.MakeVariant("volatile"),
		"bind-activation": dbus.MakeVariant("dbus-client"),
	}

	result := nmDict{}
	err = bus.Object(nmDest, nmPath).Call(
		nmIface+".AddAndActivateConnection2", 0,
		settings,
		dbus.ObjectPath("/"),
		dbus.ObjectPath("/"),
		opts,
	).Store(&connPath, &activePath, &result)
	if err == nil {
		volatile = true
		return
	}
	if !nmIsUnknownMethod(err) {
		err = nmErr(err, "add and activate connection failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"error": err,
	}).Warn("connection: NetworkManager AddAndActivateConnection2 " +
		"unavailable, using legacy add and activate without " +
		"automatic cleanup")

	connPath, err = nmAddConnLegacy(bus, settings)
	if err != nil {
		return
	}

	err = bus.Object(nmDest, nmPath).Call(
		nmIface+".ActivateConnection", 0,
		connPath, dbus.ObjectPath("/"), dbus.ObjectPath("/"),
	).Store(&activePath)
	if err != nil {
		nmDeleteConnPath(bus, connPath)
		connPath = ""
		err = nmErr(err, "activate connection failed")
		return
	}

	return
}

// nmRelease deactivates the current activation and forgets it. Volatile
// connections are deleted by NetworkManager on deactivation, legacy
// connections are deleted explicitly. Caller must hold w.lock.
func (w *Wg) nmRelease(bus *dbus.Conn) {
	activePath := dbus.ObjectPath(w.nmActivePath)
	connPath := dbus.ObjectPath(w.nmConnPath)
	volatile := w.nmVolatile

	w.nmActivePath = ""
	w.nmConnPath = ""
	w.nmVolatile = false

	if activePath != "" {
		nmUnwatchActive(activePath)
		nmDeactivate(bus, activePath)
	}

	if connPath != "" && !volatile {
		nmDeleteConnPath(bus, connPath)
	}
}

// nmStateChanged handles StateChanged signals for the active connection.
// Once NetworkManager reports the activation deactivated the volatile
// connection no longer exists, so the stored paths are dropped without
// any further D-Bus calls and the connection is closed.
func (w *Wg) nmStateChanged(pth dbus.ObjectPath, state uint32) {
	if state != nmActiveStateDeactivated && state != nmActiveStateBusLost {
		return
	}

	w.lock.Lock()
	if w.nmActivePath != string(pth) {
		w.lock.Unlock()
		return
	}

	if state == nmActiveStateBusLost && !w.nmVolatile {
		w.lock.Unlock()
		logrus.WithFields(w.conn.Fields(logrus.Fields{
			"active_path": pth,
		})).Warn("connection: System bus lost, NetworkManager " +
			"connection state unknown")
		return
	}

	nmUnwatchActive(pth)
	w.nmActivePath = ""
	w.nmConnPath = ""
	w.lock.Unlock()

	logrus.WithFields(w.conn.Fields(logrus.Fields{
		"active_path": pth,
		"state":       state,
	})).Info("connection: NetworkManager connection deactivated")

	w.conn.State.Close()
}

// nmApply adds the WireGuard profile to NetworkManager as a volatile
// connection bound to this D-Bus client and activates it, replacing any
// existing connection for the profile. Caller must hold w.lock.
func (w *Wg) nmApply(allowedIps []string) (err error) {
	bus, err := nmBus()
	if err != nil {
		return
	}

	iface := w.conn.Data.Iface
	w.nmName = nmConnName(iface)
	w.nmUuid = nmProfileUuid(w.conn.Profile.Id)

	w.nmRelease(bus)
	nmDeleteConnName(bus, w.nmName)

	settings, err := w.nmSettings(allowedIps)
	if err != nil {
		return
	}

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

		// A deterministic uuid replaces any leftover from a previous
		// run or a failed attempt still being torn down.
		nmDeleteConnUuid(bus, w.nmUuid)

		connPath, activePath, volatile, e := nmAddActivate(bus, settings)
		if e != nil {
			err = e
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Warn("connection: NetworkManager activate failed")
			continue
		}

		w.nmConnPath = string(connPath)
		w.nmActivePath = string(activePath)
		w.nmVolatile = volatile
		nmWatchActive(activePath, w.nmStateChanged)

		err = w.nmWaitActive(bus, activePath)
		if err != nil {
			logrus.WithFields(w.conn.Fields(logrus.Fields{
				"error": err,
			})).Warn("connection: NetworkManager activation failed")
			w.nmRelease(bus)
			continue
		}

		logrus.WithFields(w.conn.Fields(logrus.Fields{
			"nm_name":        w.nmName,
			"nm_uuid":        w.nmUuid,
			"nm_path":        w.nmConnPath,
			"nm_active_path": w.nmActivePath,
			"nm_volatile":    w.nmVolatile,
		})).Info("connection: NetworkManager WireGuard connection active")

		return
	}

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

	// Keep the volatile flag, IN_MEMORY on Update2 would clear it and
	// leave the connection behind after deactivation.
	flags := nmFlagInMemory | nmFlagBlockAutoconnect
	if w.nmVolatile {
		flags = nmFlagVolatile | nmFlagBlockAutoconnect
	}

	result := nmDict{}
	err = bus.Object(nmDest, dbus.ObjectPath(w.nmConnPath)).Call(
		nmConnIface+".Update2", 0,
		settings,
		flags,
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

// clearWgNm handles user disconnect and normal quit. The activation is
// deactivated synchronously with a short timeout so the interface goes
// away promptly, correctness does not depend on it since NetworkManager
// tears down the dbus-client bound volatile connection on its own.
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

	volatile := w.nmVolatile
	w.nmRelease(bus)

	if !volatile && w.conn.Data.Iface != "" {
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
