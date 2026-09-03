// Package event streams service events from the /events websocket.
package event

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pritunl/tools/logger"
)

const (
	// Synthetic event types emitted by the listener itself.
	ServiceConnected    = "service_connected"
	ServiceDisconnected = "service_disconnected"

	reconnectDelay = 1 * time.Second
	reconnectMax   = 10 * time.Second
)

type Event struct {
	Id        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// TpmData is the data attached to the secure enclave device
// authentication request events tpm_open and tpm_sign.
type TpmData struct {
	RequestId string `json:"request_id"`
	KeyData   string `json:"key_data"`
	SignData  string `json:"sign_data"`
}

// Tpm returns the secure enclave request data or nil when the event does
// not carry it.
func (e *Event) Tpm() *TpmData {
	if len(e.Data) == 0 || string(e.Data) == "null" {
		return nil
	}

	data := &TpmData{}
	err := json.Unmarshal(e.Data, data)
	if err != nil {
		return nil
	}

	if data.RequestId == "" {
		return nil
	}

	return data
}

// ProfileData is the subset of connection data attached to profile events.
type ProfileData struct {
	Id              string `json:"id"`
	Mode            string `json:"mode"`
	Status          string `json:"status"`
	ServerAddr      string `json:"server_addr"`
	ClientAddr      string `json:"client_addr"`
	RegistrationKey string `json:"registration_key"`
	SsoUrl          string `json:"sso_url"`
	Url             string `json:"url"`
}

// Profile parses the event data as profile data, nil when not present.
func (e *Event) Profile() *ProfileData {
	if len(e.Data) == 0 || string(e.Data) == "null" {
		return nil
	}

	data := &ProfileData{}
	err := json.Unmarshal(e.Data, data)
	if err != nil {
		return nil
	}

	if data.Id == "" {
		return nil
	}

	return data
}

type Listener struct {
	events chan *Event
	closed chan struct{}
	once   sync.Once
	lock   sync.Mutex
	conn   *websocket.Conn
}

// NewListener starts a background connection to the service events
// websocket that reconnects until Close is called.
func NewListener() (l *Listener) {
	l = &Listener{
		events: make(chan *Event, 64),
		closed: make(chan struct{}),
	}

	go l.run()

	return
}

// Events returns the channel of received events, it is closed by Close.
func (l *Listener) Events() <-chan *Event {
	return l.events
}

func (l *Listener) isClosed() bool {
	select {
	case <-l.closed:
		return true
	default:
		return false
	}
}

func (l *Listener) send(evt *Event) {
	select {
	case l.events <- evt:
	case <-l.closed:
	}
}

func (l *Listener) run() {
	defer close(l.events)

	delay := reconnectDelay
	connected := false
	loggedErr := false

	for !l.isClosed() {
		conn, err := wsDial("/events")
		if err != nil {
			if connected {
				connected = false
				l.send(&Event{Type: ServiceDisconnected})
			}
			if !loggedErr {
				loggedErr = true
				logger.WithFields(logger.Fields{
					"error": err,
				}).Warn("event: Failed to connect to service events")
			}

			select {
			case <-time.After(delay):
			case <-l.closed:
				return
			}
			delay = min(delay*2, reconnectMax)
			continue
		}

		l.lock.Lock()
		if l.isClosed() {
			l.lock.Unlock()
			wsClose(conn)
			return
		}
		l.conn = conn
		l.lock.Unlock()

		delay = reconnectDelay
		loggedErr = false
		connected = true
		l.send(&Event{Type: ServiceConnected})

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			evt := &Event{}
			err = json.Unmarshal(msg, evt)
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Warn("event: Failed to parse event")
				continue
			}

			l.send(evt)
		}

		l.lock.Lock()
		l.conn = nil
		l.lock.Unlock()
		wsClose(conn)

		if l.isClosed() {
			return
		}

		connected = false
		l.send(&Event{Type: ServiceDisconnected})

		select {
		case <-time.After(reconnectDelay):
		case <-l.closed:
			return
		}
	}
}

// Close stops the listener and closes the events channel.
func (l *Listener) Close() {
	l.once.Do(func() {
		close(l.closed)

		l.lock.Lock()
		conn := l.conn
		l.conn = nil
		l.lock.Unlock()

		if conn != nil {
			wsClose(conn)
		}
	})
}
