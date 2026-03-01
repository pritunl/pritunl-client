package connection

import (
	"os"
	"sync"
	"time"

	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

type State struct {
	conn               *Connection
	startTime          time.Time
	id                 string
	stop               bool
	lastStopCheckStack string
	lastStopCheckTime  time.Time
	deadline           bool
	delay              bool
	interactive        bool
	noReconnect        bool
	closed             bool
	systemInteractive  bool
	closeWaiters       []chan bool
	closeWaitersLock   sync.Mutex
	tempPaths          []string
}

func (s *State) Fields() logrus.Fields {
	return logrus.Fields{
		"state_id":                 s.id,
		"state_time":               s.startTime,
		"state_stop":               s.stop,
		"state_deadline":           s.deadline,
		"state_delay":              s.delay,
		"state_no_reconnect":       s.noReconnect,
		"state_interactive":        s.interactive,
		"state_system_interactive": s.systemInteractive,
		"state_closed":             s.closed,
		"state_closed_waiters":     len(s.closeWaiters),
		"state_temp_paths":         s.tempPaths,
	}
}

func (s *State) Init(opts Options) (err error) {
	s.id, err = utils.RandId()
	if err != nil {
		return
	}

	s.deadline = opts.Deadline
	s.delay = opts.Delay
	s.interactive = opts.Interactive
	s.startTime = time.Now()
	s.tempPaths = []string{}

	go s.stopWatch()

	return
}

func (s *State) PreStart() {
	if s.delay {
		time.Sleep(3 * time.Second)
	}
}

func (s *State) IsReconnect() bool {
	if GlobalStore.IsStop(s.conn.Id) {
		return false
	}
	return !s.noReconnect && s.conn.Profile.Reconnect
}

func (s *State) IsInteractive() bool {
	return s.interactive || s.systemInteractive
}

func (s *State) NoReconnect(reason string) {
	if s.noReconnect {
		return
	}
	logrus.WithFields(s.conn.Fields(logrus.Fields{
		"reason": reason,
	})).Info("connection: Stopping reconnect")
	s.noReconnect = true
}

func (s *State) stopWatch() {
	for {
		time.Sleep(1 * time.Second)
		if s.closed {
			return
		}

		if time.Since(s.lastStopCheckTime) > 3*time.Minute {
			logrus.WithFields(s.conn.Fields(logrus.Fields{
				"last_stop_check": s.lastStopCheckTime.Format(
					"2006-01-02 15:04:05"),
				"trace": s.lastStopCheckStack,
			})).Info("state: Detected dead state")
			s.lastStopCheckTime = time.Now()
		}
	}
}

func (s *State) SetStop() {
	// if s.stop {
	// 	logrus.WithFields(s.conn.Fields(nil)).Info(
	// 		"state: Profile already in stop")
	// 	return
	// }
	s.stop = true
}

func (s *State) IsStop() bool {
	trace := utils.GetStackTrace()

	s.lastStopCheckTime = time.Now()
	s.lastStopCheckStack = trace

	if Shutdown || s.stop {
		return true
	}
	return false
}

func (s *State) IsStopFast() bool {
	if Shutdown || s.stop {
		return true
	}
	return false
}

func (s *State) SetConnecting() {
	logrus.WithFields(logrus.Fields{
		"profile_id":       s.conn.Profile.Id,
		"mode":             s.conn.Profile.Mode,
		"dynamic_firewall": s.conn.Profile.DynamicFirewall,
		"device_auth":      s.conn.Profile.DeviceAuth,
		"disable_gateway":  s.conn.Profile.DisableGateway,
		"disable_dns":      s.conn.Profile.DisableDns,
		"dco":              s.conn.Profile.Dco,
		"debug_output":     s.conn.Profile.DebugOutput,
		"geo_sort":         s.conn.Profile.GeoSort,
		"force_connect":    s.conn.Profile.ForceConnect,
		"force_dns":        s.conn.Profile.ForceDns,
		"sso_auth":         s.conn.Profile.SsoAuth,
		"reconnect":        s.conn.Profile.Reconnect,
	}).Info("profile: Connecting")

	s.conn.Data.Status = Connecting

	return
}

func (s *State) AddPath(pth string) {
	s.tempPaths = append(s.tempPaths, pth)
}

func (s *State) RemovePaths() {
	paths := s.tempPaths
	for _, pth := range paths {
		os.Remove(pth)
	}
}

func (s *State) CloseWait() {
	waiter := make(chan bool, 8)

	s.closeWaitersLock.Lock()
	if s.closed {
		s.closeWaitersLock.Unlock()
		return
	}
	s.closeWaiters = append(s.closeWaiters, waiter)
	s.closeWaitersLock.Unlock()

	<-waiter
	time.Sleep(50 * time.Millisecond)
}

func (s *State) Close() {
	s.conn.Client.Disconnect()

	s.closeWaitersLock.Lock()
	if s.closed {
		s.closeWaitersLock.Unlock()
		if LogClose {
			logrus.WithFields(s.conn.Fields(logrus.Fields{
				"trace": utils.GetStackTrace(),
			})).Info("connection: Connection already closed")
		}
		return
	}
	s.closed = true

	if LogClose {
		logrus.WithFields(s.conn.Fields(logrus.Fields{
			"trace": utils.GetStackTrace(),
		})).Info("connection: Connection closed")
	}

	GlobalStore.Remove(s.conn.Id, s.conn)

	if s.closeWaiters != nil {
		for _, waiter := range s.closeWaiters {
			waiter <- true
		}
	}
	s.closeWaiters = nil

	s.closeWaitersLock.Unlock()

	s.conn.Client.Disconnected()
}
