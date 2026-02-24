package connection

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/update"
	"github.com/sirupsen/logrus"
)

var (
	sprofileShutown = false
)

func ImportSystemProfile(sprfl *sprofile.Sprofile) (
	conn *Connection, err error) {

	prfl := &Profile{
		Id: sprfl.Id,
	}

	prfl.ImportSystemProfile(sprfl)

	conn, err = NewConnection(prfl)
	if err != nil {
		return
	}

	return
}

func SyncSystemProfiles() (err error) {
	sprfls, err := sprofile.GetAll()
	if err != nil {
		return
	}

	conns := GlobalStore.GetAll()

	update := false
	waiter := sync.WaitGroup{}

	for _, sPrfl := range sprfls {
		conn := conns[sPrfl.Id]

		if sPrfl.State {
			if conn == nil {
				conn, err = ImportSystemProfile(sPrfl)
				if err != nil {
					return
				}

				update = true
				waiter.Add(1)

				go func(sPrfl *sprofile.Sprofile) {
					defer func() {
						panc := recover()
						if panc != nil {
							logrus.WithFields(logrus.Fields{
								"trace": string(debug.Stack()),
								"panic": panc,
							}).Error("handlers: Profile start panic")
						}
					}()

					ready := conn.Ready()
					if !ready {
						logrus.WithFields(logrus.Fields{
							"profile_id": conn.Id,
						}).Info("profile: Profile not ready, waiting")
					} else {
						go func() {
							defer func() {
								panc := recover()
								if panc != nil {
									logrus.WithFields(logrus.Fields{
										"trace": string(debug.Stack()),
										"panic": panc,
									}).Error("handlers: Profile start panic")
								}
							}()

							err := conn.Start(Options{
								Interactive: sPrfl.Interactive,
							})
							if err != nil {
								logrus.WithFields(logrus.Fields{
									"profile_id": sPrfl.Id,
									"error":      err,
								}).Error("profile: Failed to start sprofile")
							}
						}()
					}

					waiter.Done()
				}(sPrfl)
			} else if conn.Profile.Mode != sPrfl.LastMode &&
				!(conn.Profile.Mode == "ovpn" && sPrfl.LastMode == "") {

				update = true
				waiter.Add(1)

				go func(sPrfl *sprofile.Sprofile) {
					defer func() {
						panc := recover()
						if panc != nil {
							logrus.WithFields(logrus.Fields{
								"trace": string(debug.Stack()),
								"panic": panc,
							}).Error("handlers: Profile stop panic")
						}
					}()

					conn.Stop()

					conn, err = ImportSystemProfile(sPrfl)
					if err != nil {
						return
					}

					go func() {
						defer func() {
							panc := recover()
							if panc != nil {
								logrus.WithFields(logrus.Fields{
									"trace": string(debug.Stack()),
									"panic": panc,
								}).Error("handlers: Profile start panic")
							}
						}()

						err := conn.Start(Options{
							Interactive: sPrfl.Interactive,
						})
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"profile_id": sPrfl.Id,
								"error":      err,
							}).Error("profile: Failed to start sprofile")
						}
					}()

					waiter.Done()
				}(sPrfl)
			}
		} else if conn != nil {
			update = true
			waiter.Add(1)

			go func() {
				conn.Stop()
				waiter.Done()
			}()
		}
	}

	waiter.Wait()

	if update {
		evt := event.Event{
			Type: "update",
			Data: &Profile{
				Id: "",
			},
		}
		evt.Init()

		if GlobalStore.IsConnected() {
			evt := event.Event{
				Type: "connected",
			}
			evt.Init()
		} else {
			evt := event.Event{
				Type: "disconnected",
			}
			evt.Init()
		}
	}

	return
}

func watchSystemProfiles() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("profile: Watch system profiles panic")
			time.Sleep(5 * time.Second)
			go watchSystemProfiles()
		}
	}()

	time.Sleep(1 * time.Second)
	sprofile.Reload()

	for {
		time.Sleep(2 * time.Second)

		if Shutdown {
			return
		}

		if !GlobalStore.IsActive() {
			_ = update.Check()
		}

		err := SyncSystemProfiles()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("profile: Failed to sync system profiles")
		}
	}
}

func WatchSystemProfiles() {
	go watchSystemProfiles()
}
