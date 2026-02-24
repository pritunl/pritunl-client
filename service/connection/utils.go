package connection

import (
	"fmt"
	"net"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/tuntap"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

var (
	ipReg             = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	profileReg        = regexp.MustCompile(`[^a-z0-9_\- ]+`)
	restartLock       sync.Mutex
	cachedPublicAddr4 = ""
	cachedPublicAddr6 = ""
)

var safeChars = set.NewSet(
	'a',
	'b',
	'c',
	'd',
	'e',
	'f',
	'g',
	'h',
	'i',
	'j',
	'k',
	'l',
	'm',
	'n',
	'o',
	'p',
	'q',
	'r',
	's',
	't',
	'u',
	'v',
	'w',
	'x',
	'y',
	'z',
	'A',
	'B',
	'C',
	'D',
	'E',
	'F',
	'G',
	'H',
	'I',
	'J',
	'K',
	'L',
	'M',
	'N',
	'O',
	'P',
	'Q',
	'R',
	'S',
	'T',
	'U',
	'V',
	'W',
	'X',
	'Y',
	'Z',
	'0',
	'1',
	'2',
	'3',
	'4',
	'5',
	'6',
	'7',
	'8',
	'9',
	'-',
	'.',
	':',
	'[',
	']',
)

func FilterHostStr(s string, n int) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > n {
		s = s[:n]
	}

	ns := ""
	for _, c := range s {
		if safeChars.Contains(c) {
			ns += string(c)
		}
	}

	return ns
}

func ParseAddress(input string) (addr string) {
	input = FilterHostStr(input, 256)

	endBracketIndex := strings.LastIndex(input, "]")
	if strings.HasPrefix(input, "[") && endBracketIndex != -1 {
		addr = input[1:endBracketIndex]
		if strings.Contains(addr, ":") {
			ip := net.ParseIP(addr)
			if ip != nil {
				addr = "[" + ip.String() + "]"
			}
		}

		colonIndex := strings.LastIndex(input, ":")
		if colonIndex > endBracketIndex {
			port, _ := strconv.Atoi(input[colonIndex+1:])
			if port != 0 && port != 443 {
				addr += fmt.Sprintf(":%d", port)
			}
		}

		return
	}

	if strings.Contains(input, ":") {
		ip := net.ParseIP(input)
		if ip != nil {
			addr = "[" + ip.String() + "]"
			return
		}

		colonIndex := strings.LastIndex(input, ":")
		addr = input[:colonIndex]
		if strings.Contains(addr, ":") {
			ip := net.ParseIP(addr)
			if ip != nil {
				addr = "[" + ip.String() + "]"
			}
		}

		port, _ := strconv.Atoi(input[colonIndex+1:])
		if port != 0 && port != 443 {
			addr += fmt.Sprintf(":%d", port)
		}

		return
	}

	addr = input
	return
}

func RestartProfiles(clean bool) (err error) {
	restartLock.Lock()
	defer restartLock.Unlock()

	conns := GlobalStore.GetAll()
	prfls := []*Profile{}

	for _, conn := range conns {
		if conn.State.IsReconnect() {
			prfls = append(prfls, conn.Profile)
		}
		conn.StopBackground()
	}

	for _, conn := range conns {
		conn.StopWait()
	}

	if clean && runtime.GOOS == "windows" {
		if config.Config.DisableNetClean {
			logrus.Info("utils: Network clean disabled")
		} else {
			err = tuntap.Clean()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("utils: Failed to clear interfaces")
				err = nil
			}
		}
	}

	for _, prfl := range prfls {
		go func(prfl *Profile) {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					}).Error("handlers: Profile start panic")
				}
			}()

			newConn, e := NewConnection(prfl)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"profile_id": prfl.Id,
					"error":      e,
				}).Error("profile: Failed to init connection in restart all")
				return
			}

			e = newConn.Start(Options{})
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"profile_id": prfl.Id,
					"error":      e,
				}).Error("profile: Failed to start connection in restart all")
				return
			}
		}(prfl)
	}

	return
}

func Clean() (err error) {
	if runtime.GOOS != "windows" {
		return
	}

	for i := 0; i < 10; i++ {
		_, _ = utils.ExecCombinedOutput(
			"sc.exe", "stop", fmt.Sprintf("WireGuardTunnel$pritunl%d", i),
		)
		time.Sleep(100 * time.Millisecond)
		_, _ = utils.ExecCombinedOutput(
			"sc.exe", "delete", fmt.Sprintf("WireGuardTunnel$pritunl%d", i),
		)
	}

	return
}

func GetPublicAddress4() (addr4 string, err error) {
	if GlobalStore.IsConnected() && cachedPublicAddr4 != "" {
		logrus.Info("connection: Using cached public address")
		addr4 = cachedPublicAddr4
		return
	}

	addr4, err = utils.GetPublicAddress4()
	if err != nil {
		if cachedPublicAddr4 != "" {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("connection: Failed to get public address using cache")
			addr4 = cachedPublicAddr4
			err = nil
			return
		}
		return
	}
	cachedPublicAddr4 = addr4

	return
}

func GetPublicAddress6() (addr6 string, err error) {
	if GlobalStore.IsConnected() && cachedPublicAddr6 != "" {
		logrus.Info("connection: Using cached public address6")
		addr6 = cachedPublicAddr6
		return
	}

	addr6, err = utils.GetPublicAddress6()
	if err != nil {
		if cachedPublicAddr6 != "" {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("connection: Failed to get public address6 using cache")
			addr6 = cachedPublicAddr6
			err = nil
			return
		}
		return
	}
	cachedPublicAddr6 = addr6

	return
}

func NetworkManagerSupport() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	_, err := user.Lookup(NmOvpnUser)
	if err != nil {
		return false
	}

	_, err = user.LookupGroup(NmOvpnUser)
	if err != nil {
		return false
	}

	return true
}

func HasAppArmor() bool {
	exists, err := utils.ExistsFile("/usr/sbin/apparmor_status")
	if err != nil {
		return false
	}

	if !exists {
		return false
	}

	cmd := command.Command("/usr/sbin/apparmor_status")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	if strings.Contains(string(output), "openvpn") {
		return true
	}

	return false
}

func GetOvpnVer() (major, minor int, err error) {
	output, err := exec.Command("openvpn", "--version").Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok && len(output) > 0 {
			err = nil
		} else {
			err = &errortypes.ExecError{
				errors.Wrap(
					err,
					"connection: Failed to check openvpn version",
				),
			}
			return
		}
	}

	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		err = &errortypes.ParseError{
			errors.New("connection: Failed to parse openvpn version output"),
		}
		return
	}

	verStr := strings.Split(fields[1], "_")[0]
	parts := strings.Split(verStr, ".")
	if len(parts) < 2 {
		err = &errortypes.ParseError{
			errors.New("connection: Failed to parse openvpn version"),
		}
		return
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		err = &errortypes.ParseError{
			errors.New("connection: Failed to parse openvpn major version"),
		}
		return
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		err = &errortypes.ParseError{
			errors.New("connection: Failed to parse openvpn minor version"),
		}
		return
	}

	return
}

func IsOvpn27() bool {
	major, minor, err := GetOvpnVer()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("connection: Failed to read openvpn version")
		return false
	}

	if major == 2 && minor >= 7 {
		return true
	}

	if major > 2 {
		return true
	}

	return false
}
