package connection

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

var GlobalStore = &Store{
	conns:      map[string]*Connection{},
	conditions: map[string]*Condition{},
	stops:      map[string]time.Time{},
}

type Condition struct {
	AuthReconnect bool `json:"auth_reconnect"`
}

type StoreData struct {
	Id string `json:"id"`
	*Data
	*Condition
}

type Store struct {
	dnsForced      bool
	lock           sync.RWMutex
	conditionsLock sync.Mutex
	stopsLock      sync.Mutex
	conns          map[string]*Connection
	conditions     map[string]*Condition
	stops          map[string]time.Time
}

func (s *Store) cleanState() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("profile: Clean state panic")
		}
	}()

	if runtime.GOOS == "darwin" && len(s.conns) == 0 {
		err := utils.ClearScutilConnKeys()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("connection: Failed to clear scutil connection keys")
		}

		if s.dnsForced {
			utils.ClearDns()
			s.dnsForced = false
		}
	}
}

func (s *Store) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.conns)
}

func (s *Store) IsActive() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.conns) > 0
}

func (s *Store) IsConnected() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, conn := range s.conns {
		if conn.Data.Status == Connected {
			return true
		}
	}

	return false
}

func (s *Store) SetStop(prflId string) {
	prflId = utils.FilterId(prflId)

	s.stopsLock.Lock()
	defer s.stopsLock.Unlock()

	s.stops[prflId] = time.Now()
}

func (s *Store) IsStop(prflId string) bool {
	prflId = utils.FilterId(prflId)

	s.stopsLock.Lock()
	defer s.stopsLock.Unlock()

	stop := s.stops[prflId]
	if !stop.IsZero() {
		if utils.SinceAbs(stop) < 6*time.Second {
			return true
		}
		delete(s.stops, prflId)
	}
	return false
}

func (s *Store) Add(prflId string, conn *Connection) {
	prflId = utils.FilterId(prflId)

	s.lock.Lock()
	c := s.conns[prflId]
	if c == nil {
		s.conns[prflId] = conn
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()

	logrus.WithFields(c.Fields(nil)).Error(
		"connection: Overwriting stored connection")
	c.StopWait()

	s.lock.Lock()
	c = s.conns[prflId]
	if c != nil {
		c.State.SetStop()
	}
	s.conns[prflId] = conn
	s.lock.Unlock()

	logrus.WithFields(conn.Fields(nil)).Error(
		"connection: Overwrote stored connection")

	return
}

func (s *Store) Remove(prflId string, conn *Connection) {
	prflId = utils.FilterId(prflId)

	s.lock.Lock()
	defer s.lock.Unlock()

	c := s.conns[prflId]
	if c == conn {
		delete(s.conns, prflId)
	} else {
		logrus.WithFields(c.Fields(nil)).Error(
			"connection: Attempting to delete active connection")
		logrus.WithFields(conn.Fields(nil)).Error(
			"connection: Attempted to delete active connection")
	}

	go func() {
		s.lock.RLock()
		defer s.lock.RUnlock()

		s.cleanState()
	}()

	return
}

func (s *Store) GetIds() (ids []string) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ids = []string{}
	for connId := range s.conns {
		ids = append(ids, connId)
	}

	return
}

func (s *Store) Get(prflId string) (conn *Connection) {
	prflId = utils.FilterId(prflId)

	s.lock.RLock()
	defer s.lock.RUnlock()

	s.cleanState()
	conn = s.conns[prflId]

	return
}

func (s *Store) GetData(prflId string) (prfl *Data) {
	prflId = utils.FilterId(prflId)

	s.lock.RLock()
	defer s.lock.RUnlock()

	s.cleanState()
	conn := s.conns[prflId]
	if conn != nil {
		prfl = conn.Data
	}

	return
}

func (s *Store) SetAuthConnect(prflId string) {
	s.conditionsLock.Lock()
	defer s.conditionsLock.Unlock()

	condition := s.conditions[prflId]
	if condition != nil {
		condition.AuthReconnect = true
	} else {
		s.conditions[prflId] = &Condition{
			AuthReconnect: true,
		}
	}
}

func (s *Store) UnsetAuthConnect(prflId string) {
	s.conditionsLock.Lock()
	defer s.conditionsLock.Unlock()

	condition := s.conditions[prflId]
	if condition != nil {
		condition.AuthReconnect = false
	}
}

func (s *Store) GetAll() (conns map[string]*Connection) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	conns = map[string]*Connection{}

	s.cleanState()
	for _, conn := range s.conns {
		conns[conn.Id] = conn
	}

	return
}

func (s *Store) GetAllData() (prfls map[string]*StoreData) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	prfls = map[string]*StoreData{}

	s.cleanState()
	for _, conn := range s.conns {
		prfls[conn.Id] = &StoreData{
			Id:   conn.Id,
			Data: conn.Data,
		}
	}

	s.conditionsLock.Lock()
	defer s.conditionsLock.Unlock()
	for prflId, condition := range s.conditions {
		data := prfls[prflId]
		if data != nil {
			data.Condition = condition
		} else {
			prfls[prflId] = &StoreData{
				Id:        prflId,
				Condition: condition,
			}
		}
	}

	return
}

func (s *Store) GetAllId() (connIds set.Set) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	connIds = set.NewSet()

	s.cleanState()
	for _, conn := range s.conns {
		connIds.Add(conn.Id)
	}

	return
}

func (s *Store) SetDnsForced() {
	s.lock.Lock()
	s.dnsForced = true
	s.lock.Unlock()
}

func (s *Store) RestoreDns(force bool) (err error) {
	if runtime.GOOS != "darwin" {
		return
	}

	logrus.Info("connection: Restore DNS")

	state, err := utils.GetScutilState()
	if err != nil {
		return
	}

	activeIds := set.NewSet()
	for _, connId := range s.GetIds() {
		activeIds.Add(connId)
	}

	activeConnIds := []string{}
	staleConnIds := []string{}
	for _, connId := range state.Connections {
		if activeIds.Contains(connId) {
			activeConnIds = append(activeConnIds, connId)
		} else {
			staleConnIds = append(staleConnIds, connId)
		}
	}

	if len(staleConnIds) > 0 {
		logrus.WithFields(logrus.Fields{
			"conn_ids": staleConnIds,
		}).Warning("connection: Removing stale DNS connection keys")

		err = utils.RemoveScutilConnKeys(staleConnIds)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("connection: Failed to remove stale DNS connection keys")
		}
	}

	connected := len(activeConnIds) != 0

	restoreKey := ""
	if connected && !force {
		restoreKey = fmt.Sprintf(
			"/Network/Pritunl/Connection/%s", activeConnIds[0])
	}

	if restoreKey != "" {
		err = utils.CopyClearScutilMultiKey(
			"State", restoreKey,
			&utils.ScutilKey{
				Type: "State",
				Key:  utils.PritunlScutilKey,
			},
		)
		if err != nil {
			return
		}
	} else {
		err = utils.RemoveScutilKey("State", utils.PritunlScutilKey)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "State",
				"key":  utils.PritunlScutilKey,
			}).Error("connection: Failed to clear DNS service")
			return
		}
		err = utils.RemoveScutilKey("Setup", utils.PritunlScutilKey)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "Setup",
				"key":  utils.PritunlScutilKey,
			}).Error("connection: Failed to clear DNS service")
			return
		}
	}

	if !connected {
		err = utils.RestoreSearchDomains(state.PrimaryService)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("connection: Failed to restore primary DNS")
			err = nil
		}
	}

	utils.ClearDNSCache()

	return
}
