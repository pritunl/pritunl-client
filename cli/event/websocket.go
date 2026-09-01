package event

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/websocket"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/service"
)

const (
	dialTimeout  = 10 * time.Second
	readTimeout  = 90 * time.Second
	writeTimeout = 10 * time.Second
)

var (
	unixDialer = &websocket.Dialer{
		HandshakeTimeout: dialTimeout,
		NetDialContext: func(ctx context.Context, _, _ string) (
			net.Conn, error) {

			dialer := &net.Dialer{Timeout: dialTimeout}
			return dialer.DialContext(ctx, "unix", service.GetUnixSocket())
		},
	}
	tcpDialer = &websocket.Dialer{
		HandshakeTimeout: dialTimeout,
	}
)

func wsDial(path string) (conn *websocket.Conn, err error) {
	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	header := http.Header{}
	header.Set("Auth-Key", authKey)
	header.Set("User-Agent", "pritunl")

	var dialer *websocket.Dialer
	var url string
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		dialer = unixDialer
		url = "ws://unix" + path
	} else {
		dialer = tcpDialer
		url = "ws://" + service.TcpAddress + path
	}

	conn, resp, err := dialer.Dial(url, header)
	if err != nil {
		if resp != nil {
			err = &errortypes.RequestError{
				errors.Wrapf(err, "event: Websocket handshake error %d",
					resp.StatusCode),
			}
		} else {
			err = &errortypes.RequestError{
				errors.Wrap(err, "event: Failed to connect to service"),
			}
		}
		return
	}

	// The service pings every 30 seconds, drop the connection when
	// pings stop arriving so the listener reconnects.
	conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetPingHandler(func(data string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return conn.WriteControl(websocket.PongMessage, []byte(data),
			time.Now().Add(writeTimeout))
	})

	return
}

func wsClose(conn *websocket.Conn) {
	conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(writeTimeout),
	)
	conn.Close()
}
