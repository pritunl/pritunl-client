// For OS X to detect removal of Pritunl.app and auto uninstall all files.
package autoclean

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

const (
	pathSep = string(os.PathSeparator)
)

var (
	cleanLock = sync.Mutex{}
)

func clean() (err error) {
	command.Command("kextunload", "-b", "net.sf.tuntaposx.tap").Run()
	command.Command("kextunload", "-b", "net.sf.tuntaposx.tun").Run()

	paths := []string{
		filepath.Join(pathSep, "private", "var", "db", "receipts",
			"com.pritunl.pkg.Pritunl.bom"),
		filepath.Join(pathSep, "private", "var", "db", "receipts",
			"com.pritunl.pkg.Pritunl.plist"),
		filepath.Join(pathSep, "Library", "LaunchAgents",
			"com.pritunl.client.plist"),
		filepath.Join(pathSep, "Library", "LaunchDaemons",
			"com.pritunl.service.plist"),
	}

	for _, path := range paths {
		err = os.RemoveAll(path)
		if err != nil {
			err = &RemoveError{
				errors.Wrap(err, "autoclean: Failed to remove file"),
			}
		}
	}

	return
}

// Check for Pritunl.app and uninstall if missing
func CheckAndClean() (err error) {
	cleanLock.Lock()
	defer cleanLock.Unlock()

	root := utils.GetRootDir()
	if runtime.GOOS != "darwin" ||
		root != "/Applications/Pritunl.app/Contents/Resources" {

		return
	}

	path := filepath.Join(pathSep, "Applications", "Pritunl.app")
	if _, e := os.Stat(path); !os.IsNotExist(e) {
		return
	}

	err = clean()
	if err != nil {
		return
	}

	os.Exit(0)

	return
}

// Watch for Pritunl.app removal for next 10 minutes and uninstall if missing
func CheckAndCleanWatch() {
	root := utils.GetRootDir()
	if runtime.GOOS != "darwin" ||
		root != "/Applications/Pritunl.app/Contents/Resources" {

		return
	}

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("autoclean: Check and clean panic")
			}
		}()

		for i := 0; i < 200; i++ {
			time.Sleep(3 * time.Second)

			err := CheckAndClean()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("autoclean: Failed to run check and clean")
				return
			}
		}
	}()
}
