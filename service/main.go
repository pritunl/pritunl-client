package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-client/service/auth"
	"github.com/pritunl/pritunl-client/service/autoclean"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/connection"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/logger"
	"github.com/pritunl/pritunl-client/service/router"
	"github.com/pritunl/pritunl-client/service/setup"
	"github.com/pritunl/pritunl-client/service/tuntap"
	"github.com/pritunl/pritunl-client/service/update"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/pritunl/pritunl-client/service/watch"
	"github.com/pritunl/pritunl-client/service/winsvc"
	"github.com/sirupsen/logrus"
)

func main() {
	install := flag.Bool("install", false, "run post install")
	uninstall := flag.Bool("uninstall", false, "run pre uninstall")
	clean := flag.Bool("clean", false, "clean up tuntap adapters")
	devPtr := flag.Bool("dev", false, "development mode")
	flag.Parse()

	if *install {
		setup.Install()
		return
	}

	if *uninstall {
		setup.Uninstall()
		return
	}

	if *clean {
		err := setup.TunTapClean(true)
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	if *devPtr {
		constants.Development = true
	}

	err := config.Load()
	if err != nil {
		panic(err)
	}

	err = utils.PidInit()
	if err != nil {
		panic(err)
	}

	err = utils.InitTempDir()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("main: Failed to init temp dir")
		panic(err)
	}

	if runtime.GOOS == "darwin" {
		output, err := utils.ExecOutput("uname", "-r")
		if err == nil {
			macosVersion, err := strconv.Atoi(strings.Split(output, ".")[0])
			if err == nil && macosVersion < 20 {
				constants.Macos10 = true
			}
		}
	}

	logger.Init()

	logrus.WithFields(logrus.Fields{
		"version": constants.Version,
	}).Info("main: Service starting")

	go update.Check()

	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("main: Panic")
			time.Sleep(5 * time.Second)
			panic(panc)
		}
	}()

	err = auth.Init()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("main: Failed to init auth")
		panic(err)
	}

	err = autoclean.CheckAndClean()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("main: Failed to run check and clean")
		panic(err)
	}

	if runtime.GOOS == "windows" {
		if config.Config.DisableNetClean {
			logrus.Info("main: Network clean disabled")
		} else {
			err = tuntap.Clean()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("main: Failed to clear interfaces")
				err = nil
			}
		}
	}

	gin.SetMode(gin.ReleaseMode)

	watch.StartWatch()

	err = connection.Clean()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("main: Failed to clean profiles")
		panic(err)
	}

	routr := &router.Router{}
	routr.Init()

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("utils: Main router panic")
			}

			time.Sleep(1 * time.Second)
		}()

		for {
			if constants.Interrupt {
				return
			}

			err = routr.Run()
			if constants.Interrupt {
				return
			}
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("main: Server error")
			} else {
				logrus.Error("main: Unexpected server close")
			}

			newRoutr := &router.Router{}
			newRoutr.Init()
			routr = newRoutr

			time.Sleep(1 * time.Second)
		}
	}()

	connection.WatchSystemProfiles()

	if winsvc.IsWindowsService() {
		service := winsvc.New()

		err = service.Run()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("main: Service error")
			panic(err)
		}
	} else {
		sig := make(chan os.Signal, 100)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
	}

	evt := &event.Event{
		Id:   utils.Uuid(),
		Type: "shutdown",
	}
	evt.Init()

	time.Sleep(500 * time.Millisecond)

	constants.Interrupt = true

	routr.Shutdown()

	time.Sleep(100 * time.Millisecond)

	connection.SetShutdown()

	conns := connection.GlobalStore.GetAll()
	for _, conn := range conns {
		conn.StopBackground()
	}

	for _, conn := range conns {
		conn.StopWait()
	}

	if runtime.GOOS == "darwin" {
		_ = utils.ClearScutilConnKeys()
		_ = utils.RestoreScutilDns(true)
	}

	time.Sleep(750 * time.Millisecond)
}
