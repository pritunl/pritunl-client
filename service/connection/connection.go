package connection

import (
	"runtime"
	"runtime/debug"

	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

type Options struct {
	Deadline    bool
	Delay       bool
	Interactive bool
	Fork        bool
}

type Connection struct {
	Id      string
	Profile *Profile
	Data    *Data
	State   *State
	Client  *Client
	Ovpn    *Ovpn
	Wg      *Wg
}

func (c *Connection) Init() (err error) {
	c.Ovpn.Init()
	c.Wg.Init()

	return
}

func (c *Connection) Fields(fields ...logrus.Fields) logrus.Fields {
	newFields := logrus.Fields{}

	for _, fieldSet := range fields {
		for key, val := range fieldSet {
			newFields[key] = val
		}
	}

	for key, val := range c.Profile.Fields() {
		newFields[key] = val
	}

	for key, val := range c.Data.Fields() {
		newFields[key] = val
	}

	for key, val := range c.State.Fields() {
		newFields[key] = val
	}

	for key, val := range c.Client.Fields() {
		newFields[key] = val
	}

	for key, val := range c.Ovpn.Fields() {
		newFields[key] = val
	}

	for key, val := range c.Wg.Fields() {
		newFields[key] = val
	}

	return newFields
}

func (c *Connection) Start(opts Options) (err error) {
	if c.State.IsInteractive() && c.Profile.SystemProfile {
		sprfl := sprofile.Get(c.Profile.Id)
		if sprfl != nil {
			sprfl.Interactive = false
		}
	}

	err = c.State.Init(opts)
	if err != nil {
		logrus.WithFields(c.Fields(logrus.Fields{
			"error": err,
		})).Error("connection: Start init error")
		c.State.Close()
		return
	}

	if c.State.IsStop() {
		c.State.Close()
		return
	}

	c.State.SetConnecting()

	conn := GlobalStore.Get(c.Id)
	if conn != nil {
		logrus.WithFields(conn.Fields(nil)).Info(
			"profile: Profile already active, disconnecting")
		conn.StopWait()
	}

	if c.State.IsStop() {
		c.State.Close()
		return
	}

	GlobalStore.Add(c.Id, c)

	if c.State.IsStop() {
		c.State.Close()
		return
	}

	c.Profile.Sync()

	if c.State.IsStop() {
		c.State.Close()
		return
	}

	if c.Profile.Mode == WgMode {
		err = c.Wg.Start()
	} else {
		err = c.Ovpn.Start()
	}
	if err != nil {
		logrus.WithFields(c.Fields(logrus.Fields{
			"error": err,
		})).Error("connection: Start error")
		c.State.Close()
		return
	}

	return
}

func (c *Connection) Restart() {
	c.State.NoReconnect("restart")
	c.StopWait()

	newConn, err := NewConnection(c.Profile)
	if err != nil {
		logrus.WithFields(c.Fields(logrus.Fields{
			"error": err,
		})).Error("profile: Failed to init connection in restart")
		return
	}

	err = newConn.Start(Options{})
	if err != nil {
		logrus.WithFields(c.Fields(logrus.Fields{
			"error": err,
		})).Error("profile: Failed to start connection in restart")
		return
	}
}

func (c *Connection) Stop() {
	c.State.NoReconnect("stop")
	c.Client.Disconnect()
}

func (c *Connection) StopWait() {
	c.State.NoReconnect("stop_wait")
	c.Client.Disconnect()
	c.State.CloseWait()
}

func (c *Connection) StopBackground() {
	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(c.Fields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				})).Error("profile: Stop background panic")
			}
		}()

		c.State.NoReconnect("stop_background")
		c.Client.Disconnect()
		c.State.CloseWait()
	}()
}

func (c *Connection) Ready() bool {
	if c.Profile.DeviceAuth && runtime.GOOS == "darwin" &&
		!config.Config.ForceLocalTpm {

		return event.GetState()
	}
	return true
}

func NewConnection(prfl *Profile) (conn *Connection, err error) {
	prfl.Id = utils.FilterStrN(prfl.Id, 128)

	conn = &Connection{
		Id:      prfl.Id,
		Profile: prfl,
		Data: &Data{
			Id: prfl.Id,
		},
		State:  &State{},
		Client: &Client{},
		Ovpn:   &Ovpn{},
		Wg:     &Wg{},
	}

	conn.Profile.conn = conn
	conn.Data.conn = conn
	conn.State.conn = conn
	conn.Client.conn = conn
	conn.Ovpn.conn = conn
	conn.Wg.conn = conn

	err = conn.Init()
	if err != nil {
		return
	}

	return
}
