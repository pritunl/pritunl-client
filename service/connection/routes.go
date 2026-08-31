package connection

import (
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/sirupsen/logrus"
)

var GlobalRoutes = &RouteStore{
	owners: map[string]string{},
	conns:  map[string]*routeConn{},
	order:  []string{},
}

type routeConn struct {
	id     string
	wg     *Wg
	routes map[string]*Route
	ipv6   set.Set
	wanted []string
	active set.Set
}

type RouteStore struct {
	lock   sync.Mutex
	owners map[string]string
	conns  map[string]*routeConn
	order  []string
}

func routeTracked(network string) bool {
	return network != "0.0.0.0/0" && network != "::/0"
}

func (s *RouteStore) Claim(wg *Wg, routes []*Route,
	routes6 []*Route) (allowed []string) {

	s.lock.Lock()
	defer s.lock.Unlock()

	id := wg.conn.Id

	s.removeLocked(id)

	rc := &routeConn{
		id:     id,
		wg:     wg,
		routes: map[string]*Route{},
		ipv6:   set.NewSet(),
		wanted: []string{},
		active: set.NewSet(),
	}

	add := func(route *Route, ipv6 bool) {
		network := route.Network

		if !routeTracked(network) {
			allowed = append(allowed, network)
			return
		}

		if _, exists := rc.routes[network]; exists {
			return
		}
		rc.routes[network] = route
		rc.wanted = append(rc.wanted, network)
		if ipv6 {
			rc.ipv6.Add(network)
		}

		owner := s.owners[network]
		if owner != "" && owner != id {
			logrus.WithFields(wg.conn.Fields(logrus.Fields{
				"network":  network,
				"owner_id": owner,
			})).Info("connection: Excluding duplicate route")
			return
		}

		s.owners[network] = id
		rc.active.Add(network)
		allowed = append(allowed, network)
	}

	allowed = []string{}
	for _, route := range routes {
		add(route, false)
	}
	for _, route := range routes6 {
		add(route, true)
	}

	s.conns[id] = rc
	s.order = append(s.order, id)

	return
}

func (s *RouteStore) Release(wg *Wg) {
	s.lock.Lock()
	defer s.lock.Unlock()

	id := wg.conn.Id

	rc := s.conns[id]
	if rc == nil || rc.wg != wg {
		return
	}

	released := s.removeLocked(id)

	for _, network := range released {
		for _, otherId := range s.order {
			other := s.conns[otherId]
			if other == nil {
				continue
			}

			route := other.routes[network]
			if route == nil || other.active.Contains(network) {
				continue
			}

			s.owners[network] = otherId
			other.active.Add(network)

			logrus.WithFields(other.wg.conn.Fields(logrus.Fields{
				"network":     network,
				"released_id": id,
			})).Info("connection: Reassigning released route")

			other.wg.addAllowedIp(
				route,
				other.ipv6.Contains(network),
				other.allowedIps(),
			)
			break
		}
	}
}

func (s *RouteStore) removeLocked(id string) (released []string) {
	released = []string{}

	rc := s.conns[id]
	if rc == nil {
		return
	}

	for _, network := range rc.wanted {
		if s.owners[network] == id {
			delete(s.owners, network)
			released = append(released, network)
		}
	}

	delete(s.conns, id)

	order := []string{}
	for _, otherId := range s.order {
		if otherId != id {
			order = append(order, otherId)
		}
	}
	s.order = order

	return
}

func (rc *routeConn) allowedIps() (allowed []string) {
	allowed = []string{}

	if rc.wg.conn.Data.Routes != nil {
		for _, route := range rc.wg.conn.Data.Routes {
			if route.NetGateway {
				continue
			}
			if !routeTracked(route.Network) ||
				rc.active.Contains(route.Network) {

				allowed = append(allowed, route.Network)
			}
		}
	}

	if rc.wg.conn.Data.Routes6 != nil {
		for _, route := range rc.wg.conn.Data.Routes6 {
			if route.NetGateway {
				continue
			}
			if !routeTracked(route.Network) ||
				rc.active.Contains(route.Network) {

				allowed = append(allowed, route.Network)
			}
		}
	}

	return
}
