package router

import (
	"context"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/handlers"
)

type Router struct {
	server *http.Server
}

func (r *Router) runSock() (err error) {
	_ = os.Remove("/var/run/pritunl.sock")

	listener, err := net.Listen("unix", "/var/run/pritunl.sock")
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "main: Failed to create unix socket"),
		}
		return
	}

	err = os.Chmod("/var/run/pritunl.sock", 0777)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "main: Failed to chmod unix socket"),
		}
		return
	}

	err = r.server.Serve(listener)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "main: Server listen error"),
		}
		return
	}

	return
}

func (r *Router) runTcp() (err error) {
	err = r.server.ListenAndServe()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "main: Server listen error"),
		}
		return
	}

	return
}

func (r *Router) Run() (err error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		err = r.runTcp()
		if err != nil {
			return
		}
	} else {
		err = r.runSock()
		if err != nil {
			return
		}
	}

	return
}

func (r *Router) Shutdown() {
	defer func() {
		recover()
	}()

	webCtx, webCancel := context.WithTimeout(
		context.Background(),
		1*time.Second,
	)
	defer webCancel()

	_ = r.server.Shutdown(webCtx)
	_ = r.server.Close()
}

func (r *Router) Init() {
	router := gin.New()
	handlers.Register(router)

	r.server = &http.Server{
		Addr:           "127.0.0.1:9770",
		Handler:        router,
		ReadTimeout:    300 * time.Second,
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 4096,
	}

	return
}
