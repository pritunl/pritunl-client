// Miscellaneous utils.
package utils

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/platform"
	"github.com/sirupsen/logrus"
)

var (
	lockedInterfaces set.Set
	networkResetLock sync.Mutex
	macDnsLock       = sync.Mutex{}
)

const (
	PritunlScutilKey = "/Network/Service/Pritunl/DNS"
)

func init() {
	lockedInterfaces = set.NewSet()
}

type Interface struct {
	Id   string
	Name string
}

type Interfaces []*Interface

func (intfs Interfaces) Len() int {
	return len(intfs)
}

func (intfs Interfaces) Swap(i, j int) {
	intfs[i], intfs[j] = intfs[j], intfs[i]
}

func (intfs Interfaces) Less(i, j int) bool {
	return intfs[i].Name < intfs[j].Name
}

func GetTaps() (interfaces []*Interface, err error) {
	interfaces = []*Interface{}

	cmd := command.Command("ipconfig", "/all")

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec ipconfig"),
		}
		return
	}

	buf := bytes.NewBuffer(output)
	scan := bufio.NewReader(buf)

	intName := ""
	intTap := false
	intAddr := ""

	for {
		lineByte, _, e := scan.ReadLine()
		if e != nil {
			if e == io.EOF {
				break
			}
			err = e
			panic(err)
			return
		}
		line := string(lineByte)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Ethernet adapter ") {
			intName = strings.Split(line, "Ethernet adapter ")[1]
			intName = intName[:len(intName)-1]
			intTap = false
			intAddr = ""
		} else if intName != "" {
			if strings.Contains(line, "TAP-Windows Adapter") {
				intTap = true
			} else if strings.Contains(line, "Physical Address") {
				intAddr = strings.Split(line, ":")[1]
				intAddr = strings.TrimSpace(intAddr)
			} else if intTap && intAddr != "" {
				intf := &Interface{
					Id:   intAddr,
					Name: intName,
				}
				interfaces = append(interfaces, intf)
				intName = ""
			}
		}
	}

	sort.Sort(Interfaces(interfaces))

	return
}

func AcquireTap() (intf *Interface, err error) {
	interfaces, err := GetTaps()
	if err != nil {
		return
	}

	for _, intrf := range interfaces {
		if !lockedInterfaces.Contains(intrf.Id) {
			lockedInterfaces.Add(intrf.Id)
			intf = intrf
			return
		}
	}

	return
}

func ReleaseTap(intf *Interface) {
	lockedInterfaces.Remove(intf.Id)
}

func GetScutilKey(typ, key string) (val string, err error) {
	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf("open\nshow %s:%s\nquit\n", typ, key))

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	val = strings.TrimSpace(string(output))

	return
}

func RemoveScutilKey(typ, key string) (err error) {
	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf("open\nremove %s:%s\nquit\n", typ, key))

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

func GetPrimaryService() (serviceId string, err error) {
	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader("open\n" +
		"show State:/Network/Global/IPv4\n" +
		"quit\n",
	)

	output, err := cmd.Output()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "PrimaryService") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				serviceId = strings.TrimSpace(parts[len(parts)-1])
				break
			}
		}
	}

	return
}

func SetPrimaryServiceDns(serviceId string, domains []string) (err error) {
	input := fmt.Sprintf("open\n"+
		"get State:/Network/Service/%s/DNS\n"+
		"d.add SearchDomains * %s\n"+
		"set State:/Network/Service/%s/DNS\n"+
		"set Setup:/Network/Service/%s/DNS\n"+
		"quit\n",
		serviceId, strings.Join(domains, " "), serviceId, serviceId)

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(input)

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

func SetScutilDns(connId string, addresses, domains []string) (err error) {
	logrus.Info("utils: Configure DNS")

	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	input := ""
	if domains == nil || len(domains) == 0 {
		input = fmt.Sprintf("open\n"+
			"d.init\n"+
			"d.add ServerAddresses * %s\n"+
			"d.add SupplementalMatchDomains * \"\"\n"+
			"remove State:%s\n"+
			"remove Setup:%s\n"+
			"set State:%s\n"+
			"set Setup:%s\n"+
			"set State:/Network/Pritunl/Connection/%s\n"+
			"quit\n",
			strings.Join(addresses, " "),
			PritunlScutilKey, PritunlScutilKey, PritunlScutilKey,
			PritunlScutilKey, connId)
	} else {
		input = fmt.Sprintf("open\n"+
			"d.init\n"+
			"d.add ServerAddresses * %s\n"+
			"d.add SearchDomains * %s\n"+
			"d.add SupplementalMatchDomains * \"\"\n"+
			"remove State:%s\n"+
			"remove Setup:%s\n"+
			"set State:%s\n"+
			"set Setup:%s\n"+
			"set State:/Network/Pritunl/Connection/%s\n"+
			"quit\n",
			strings.Join(addresses, " "), strings.Join(domains, " "),
			PritunlScutilKey, PritunlScutilKey, PritunlScutilKey,
			PritunlScutilKey, connId)
	}

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(input)

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	if len(domains) > 0 {
		primaryServiceId, e := GetPrimaryService()
		if e != nil {
			logrus.WithFields(logrus.Fields{
				"error": e,
			}).Warning("utils: Failed to get primary service for DNS")
		} else if primaryServiceId != "" {
			err = SetPrimaryServiceDns(primaryServiceId, domains)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Warning("utils: Failed to set primary service DNS")
				err = nil
			}
		}
	}

	return
}

func ClearScutilDns(connId string) (err error) {
	logrus.Info("utils: Clearing DNS state")

	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf("open\n"+
			"remove State:/Network/Pritunl/Connection/%s\n"+
			"quit\n", connId))

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

func CopyScutilKey(typ, src, dst string) (err error) {
	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf("open\n"+
			"get %s:%s\n"+
			"set %s:%s\n"+
			"quit\n", typ, src, typ, dst))

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

type ScutilKey struct {
	Type string
	Key  string
}

func CopyScutilMultiKey(typ, src string, dsts ...*ScutilKey) (err error) {
	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	stdin := fmt.Sprintf("open\nget %s:%s\n", typ, src)
	for _, dst := range dsts {
		stdin += fmt.Sprintf("set %s:%s\n", dst.Type, dst.Key)
	}
	stdin += "quit\n"

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(stdin)

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

func CopyClearScutilMultiKey(typ, src string, dsts ...*ScutilKey) (
	err error) {

	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	stdin := fmt.Sprintf("open\nget %s:%s\n", typ, src)
	for _, dst := range dsts {
		if dst.Type == "State" {
			stdin += fmt.Sprintf("remove %s:%s\n", dst.Type, dst.Key)
		}
		stdin += fmt.Sprintf("set %s:%s\n", dst.Type, dst.Key)
	}
	stdin += "quit\n"

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader(stdin)

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

func RestoreScutilDns(force bool) (err error) {
	if runtime.GOOS != "darwin" {
		return
	}

	logrus.Info("utils: Restore DNS")

	connIds, err := GetScutilConnIds()
	if err != nil {
		return
	}
	connected := len(connIds) != 0

	restoreKey := ""
	if connected && !force {
		restoreKey = fmt.Sprintf(
			"/Network/Pritunl/Connection/%s", connIds[0])
	}

	if restoreKey != "" {
		err = CopyClearScutilMultiKey(
			"State", restoreKey,
			&ScutilKey{
				Type: "State",
				Key:  PritunlScutilKey,
			},
			&ScutilKey{
				Type: "Setup",
				Key:  PritunlScutilKey,
			},
		)
		if err != nil {
			return
		}
	} else {
		err = RemoveScutilKey("State", PritunlScutilKey)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "State",
				"key":  PritunlScutilKey,
			}).Error("utils: Failed to clear DNS service")
			return
		}
		err = RemoveScutilKey("Setup", PritunlScutilKey)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "Setup",
				"key":  PritunlScutilKey,
			}).Error("utils: Failed to clear DNS service")
			return
		}
	}

	ClearDNSCache()

	return
}

func RefreshScutilDns() (err error) {
	if runtime.GOOS != "darwin" {
		return
	}

	logrus.Info("utils: Refresh DNS")

	err = CopyClearScutilMultiKey(
		"State", PritunlScutilKey,
		&ScutilKey{
			Type: "State",
			Key:  PritunlScutilKey,
		},
		&ScutilKey{
			Type: "Setup",
			Key:  PritunlScutilKey,
		},
	)
	if err != nil {
		return
	}

	ClearDNSCacheFast()

	return
}

func GetScutilConnIds() (ids []string, err error) {
	ids = []string{}

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader("open\nlist\nquit\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	for _, line := range strings.Split(string(output), "\n") {
		if !strings.Contains(line, "State:/Network/Pritunl/Connection/") {
			continue
		}

		spl := strings.Split(line, "State:/Network/Pritunl/Connection/")
		if len(spl) == 2 {
			ids = append(ids, strings.TrimSpace(spl[1]))
		}
	}

	return
}

func ClearScutilConnKeys() (err error) {
	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	connIds, err := GetScutilConnIds()
	if err != nil {
		return
	}

	remove := ""

	for _, connId := range connIds {
		remove += fmt.Sprintf(
			"remove State:/Network/Pritunl/Connection/%s\n", connId)
	}

	if remove == "" {
		return
	}

	cmd := command.Command("/usr/sbin/scutil")
	cmd.Stdin = strings.NewReader("open\n" + remove + "quit\n")

	err = cmd.Run()
	if err != nil {
		err = &CommandError{
			errors.Wrap(err, "utils: Failed to exec scutil"),
		}
		return
	}

	return
}

func ResetNetworking() {
	logrus.Info("utils: Reseting networking")

	networkResetLock.Lock()
	defer networkResetLock.Unlock()

	switch runtime.GOOS {
	case "windows":
		command.Command("netsh", "interface", "ip", "delete",
			"destinationcache").Run()
		command.Command("ipconfig", "/release").Run()
		command.Command("ipconfig", "/renew").Run()
		command.Command("arp", "-d", "*").Run()
		command.Command("nbtstat", "-R").Run()
		command.Command("nbtstat", "-RR").Run()
		command.Command("ipconfig", "/flushdns").Run()
		command.Command("nbtstat", "/registerdns").Run()
		break
	case "darwin":
		cmd := command.Command("/usr/sbin/networksetup", "-getcurrentlocation")

		output, err := cmd.CombinedOutput()
		if err != nil {
			err = &CommandError{
				errors.Wrap(err, "utils: Failed to get network location"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("utils: Reset networking error")
			return
		}

		location := strings.TrimSpace(string(output))

		if location == "pritunl-reset" {
			return
		}

		err = command.Command(
			"/usr/sbin/networksetup",
			"-createlocation",
			"pritunl-reset",
		).Run()
		if err != nil {
			err = &CommandError{
				errors.Wrap(err, "utils: Failed to create network location"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("utils: Reset networking error")
		}

		err = command.Command(
			"/usr/sbin/networksetup",
			"-switchtolocation",
			"pritunl-reset",
		).Run()
		if err != nil {
			err = &CommandError{
				errors.Wrap(err, "utils: Failed to set network location"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("utils: Reset networking error")
		}

		err = command.Command(
			"/usr/sbin/networksetup",
			"-switchtolocation",
			location,
		).Run()
		if err != nil {
			err = &CommandError{
				errors.Wrap(err, "utils: Failed to set network location"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("utils: Reset networking error")
		}

		err = command.Command(
			"/usr/sbin/networksetup",
			"-deletelocation",
			"pritunl-reset",
		).Run()
		if err != nil {
			err = &CommandError{
				errors.Wrap(err, "utils: Failed to delete network location"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("utils: Reset networking error")
		}
		break
	case "linux":
		output, _ := ExecOutput("/usr/bin/nmcli", "networking")
		if strings.Contains(output, "enabled") {
			command.Command("/usr/bin/nmcli", "connection", "reload").Run()
			command.Command("/usr/bin/nmcli", "networking", "off").Run()
			command.Command("/usr/bin/nmcli", "networking", "on").Run()
		}
		break
	default:
		panic("profile: Not implemented")
	}
}

func ClearDns() {
	if runtime.GOOS != "darwin" {
		return
	}

	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	logrus.Info("utils: Clearing DNS")

	output, err := ExecCombinedOutputLogged(
		nil,
		"/usr/sbin/networksetup",
		"-listallnetworkservices",
	)
	if err != nil {
		return
	}

	for _, netService := range strings.Split(output, "\n") {
		if netService == "" || strings.Contains(netService, "asterisk") {
			continue
		}

		_, _ = ExecCombinedOutputLogged(
			nil,
			"/usr/sbin/networksetup",
			"-setdnsservers",
			netService,
			"Empty",
		)
	}

	command.Command("dscacheutil", "-flushcache").Run()
	command.Command("killall", "-HUP", "mDNSResponder").Run()
}

func ResetDns() {
	if runtime.GOOS != "darwin" {
		return
	}

	logrus.Info("utils: Resetting DNS")

	networkResetLock.Lock()
	defer networkResetLock.Unlock()

	_ = RefreshScutilDns()
}

func ClearDNSCache() {
	if runtime.GOOS != "darwin" {
		return
	}

	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	switch runtime.GOOS {
	case "windows":
		command.Command("ipconfig", "/flushdns").Run()
		go func() {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					}).Error("utils: Clear DNS cache panic")
				}
			}()

			for i := 0; i < 3; i++ {
				time.Sleep(1 * time.Second)
				command.Command("ipconfig", "/flushdns").Run()
			}
		}()
		break
	case "darwin":
		command.Command("dscacheutil", "-flushcache").Run()
		command.Command("killall", "-HUP", "mDNSResponder").Run()
		go func() {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					}).Error("utils: Clear DNS cache panic")
				}
			}()

			for i := 0; i < 3; i++ {
				time.Sleep(1 * time.Second)
				command.Command("dscacheutil", "-flushcache").Run()
				command.Command("killall", "-HUP", "mDNSResponder").Run()
			}
		}()
		break
	case "linux":
		command.Command("systemd-resolve", "--flush-caches").Run()
		command.Command("resolvectl", "--flush-caches").Run()
		go func() {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					}).Error("utils: Clear DNS cache panic")
				}
			}()

			for i := 0; i < 3; i++ {
				time.Sleep(1 * time.Second)
				command.Command("systemd-resolve", "--flush-caches").Run()
			}
		}()
		break
	default:
		panic("profile: Not implemented")
	}
}

func ClearDNSCacheFast() {
	macDnsLock.Lock()
	defer macDnsLock.Unlock()

	switch runtime.GOOS {
	case "windows":
		command.Command("ipconfig", "/flushdns").Run()
		break
	case "darwin":
		command.Command("dscacheutil", "-flushcache").Run()
		command.Command("killall", "-HUP", "mDNSResponder").Run()
		break
	case "linux":
		command.Command("systemd-resolve", "--flush-caches").Run()
		command.Command("resolvectl", "--flush-caches").Run()
		break
	default:
		panic("profile: Not implemented")
	}
}

func Uuid() (id string) {
	idByte := make([]byte, 16)

	_, err := rand.Read(idByte)
	if err != nil {
		err = &IoError{
			errors.Wrap(err, "utils: Failed to get random data"),
		}
		panic(err)
	}

	id = hex.EncodeToString(idByte[:])

	return
}

func GetRootDir() (pth string) {
	pth, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	return
}

func TunTapPath() string {
	if constants.Development {
		pth := filepath.Join(GetRootDir(), "..", "tuntap_win", "tuntap_amd64")

		exists, _ := ExistsDir(pth)
		if exists {
			return pth
		}
	}

	return filepath.Join(GetRootDir(), "tuntap")
}

func TapCtlPath() string {
	if constants.Development {
		pth := filepath.Join(GetRootDir(), "..",
			"openvpn_win", "openvpn_amd64", "tapctl.exe")

		exists, _ := ExistsFile(pth)
		if exists {
			return pth
		}
	}

	return filepath.Join(GetRootDir(), "openvpn", "tapctl.exe")
}

func GetWinDrive() string {
	systemDrv := os.Getenv("SYSTEMDRIVE")
	if systemDrv == "" {
		return "C:\\"
	}
	return systemDrv + "\\"
}

func GetAuthPath() (pth string) {
	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(GetWinDrive(), "ProgramData", "Pritunl")

		_ = platform.MkdirReadSecure(pth)

		pth = filepath.Join(pth, "auth")
		break
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "run", "pritunl.auth")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}

func GetLogPath() (pth string) {
	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(GetWinDrive(), "ProgramData", "Pritunl")

		_ = platform.MkdirReadSecure(pth)

		pth = filepath.Join(pth, "pritunl-client.log")
		break
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "log", "pritunl-client.log")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}

func GetLogPath2() (pth string) {
	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join(GetWinDrive(), "ProgramData", "Pritunl")

		_ = platform.MkdirReadSecure(pth)

		pth = filepath.Join(pth, "pritunl-client.log.1")
		break
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "log", "pritunl-client.log.1")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}

func InitTempDir() (err error) {
	if runtime.GOOS != "windows" {
		pth := filepath.Join(string(filepath.Separator), "tmp", "pritunl")

		_ = os.RemoveAll(pth)
		err = platform.MkdirSecure(pth)
		if err != nil {
			err = &IoError{
				errors.Wrap(err, "utils: Failed to create temp directory"),
			}
			return
		}
	}

	return
}

func GetTempDir() (pth string, err error) {
	if runtime.GOOS == "windows" {
		pth = filepath.Join(GetWinDrive(), "ProgramData", "Pritunl", "Temp")
		err = platform.MkdirSecure(pth)
		if err != nil {
			err = &IoError{
				errors.Wrap(
					err, "utils: Failed to create temp directory"),
			}
			return
		}
	} else {
		pth = filepath.Join(string(filepath.Separator), "tmp", "pritunl")
		err = platform.MkdirSecure(pth)
		if err != nil {
			err = &IoError{
				errors.Wrap(
					err, "utils: Failed to create temp directory"),
			}
			return
		}
	}

	return
}

func GetPidPath() (pth string) {
	switch runtime.GOOS {
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "run", "pritunl.pid")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}

func PidInit() (err error) {
	if runtime.GOOS == "windows" {
		return
	}

	pth := GetPidPath()
	pid := 0

	data, err := ioutil.ReadFile(pth)
	if err != nil {
		if !os.IsNotExist(err) {
			err = errortypes.ReadError{
				errors.Wrapf(err, "utils: Failed to read %s", pth),
			}
			return
		}
		err = nil
	} else {
		pidStr := strings.TrimSpace(string(data))
		if pidStr != "" {
			pid, _ = strconv.Atoi(pidStr)
		}
	}

	_ = os.Remove(pth)
	err = ioutil.WriteFile(
		pth,
		[]byte(strconv.Itoa(os.Getpid())),
		0644,
	)
	if err != nil {
		err = errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to write pid"),
		}
		return
	}

	if pid != 0 {
		proc, e := os.FindProcess(pid)
		if e == nil {
			proc.Signal(os.Interrupt)

			done := false

			go func() {
				defer func() {
					recover()
				}()

				time.Sleep(5 * time.Second)

				if done {
					return
				}
				proc.Kill()
			}()

			proc.Wait()
			done = true

			time.Sleep(2 * time.Second)
		}
	}

	return
}

func GetHostname() (hostname string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to get hostname"),
		}
		return
	}

	return
}
