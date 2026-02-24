package handlers

import (
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/pritunl/pritunl-client/service/watch"
	"github.com/sirupsen/logrus"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second
	pingWait     = 40 * time.Second
)

var (
	upgrader = websocket.Upgrader{
		HandshakeTimeout: 30 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
)

func eventsGet(c *gin.Context) {
	event.LastPong = time.Now()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(pingWait))
	conn.SetPongHandler(func(x string) (err error) {
		event.LastPong = time.Now()
		conn.SetReadDeadline(time.Now().Add(pingWait))
		return
	})

	list := event.NewListener()

	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		conn.Close()
		list.Close()
	}()

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("events: Panic")
			}
		}()

		defer conn.Close()

		for {
			_, msgByt, err := conn.NextReader()
			if err != nil {
				break
			}

			msg, err := ioutil.ReadAll(msgByt)
			if err != nil {
				continue
			}

			if string(msg) == "awake" {
				event.LastAwake = time.Now()
			}
		}
	}()

	for {
		select {
		case evt, ok := <-list.Listen():
			if !ok {
				conn.WriteControl(websocket.CloseMessage, []byte{},
					time.Now().Add(writeTimeout))
				return
			}

			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			err = conn.WriteJSON(evt)
			if err != nil {
				return
			}
		case <-ticker.C:
			watch.LastPing = time.Now()

			err = conn.WriteControl(websocket.PingMessage, []byte{},
				time.Now().Add(writeTimeout))
			if err != nil {
				return
			}

			connection.Ping = time.Now()
		}
	}
}
