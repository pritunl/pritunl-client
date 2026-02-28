package connection

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/log"
	"github.com/pritunl/pritunl-client/service/parser"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/tuntap"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
)

type Ovpn struct {
	conn           *Connection
	ovpnDir        string
	ovpnPath       string
	parsedPrfl     *parser.Ovpn
	running        int
	connected      bool
	tapIface       string
	managementPort int
	managementPass string
	managementLock sync.Mutex
	authFailed     bool
	lastAuthFailed time.Time
	remotes        parser.Remotes
	cmd            *exec.Cmd
	stdout         io.ReadCloser
	stderr         io.ReadCloser
	outputBuffer   chan string
	outputWait     sync.WaitGroup
}

type AuthData struct {
	Token     string `json:"token"`
	Password  string `json:"password"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
}

func (o *Ovpn) Fields() logrus.Fields {
	remotes := []string{}
	if o.remotes != nil {
		remotes = o.remotes.GetFormatted()
	}

	return logrus.Fields{
		"ovpn_dir":              o.ovpnDir,
		"ovpn_path":             o.ovpnPath,
		"ovpn_running":          o.running,
		"ovpn_connected":        o.connected,
		"ovpn_tap_iface":        o.tapIface,
		"ovpn_management_port":  o.managementPort,
		"ovpn_management_pass":  o.managementPass != "",
		"ovpn_auth_failed":      o.authFailed,
		"ovpn_last_auth_failed": utils.SinceFormatted(o.lastAuthFailed),
		"ovpn_cmd":              o.cmd != nil,
		"ovpn_remotes":          remotes,
	}
}

func (o *Ovpn) Init() {
	o.ovpnDir = GetOvpnDir()
	o.ovpnPath = GetOvpnPath()

	return
}

func (o *Ovpn) GetPublicKey() string {
	return ""
}

func (o *Ovpn) GetReqPrefix() string {
	return "ovpn"
}

func (o *Ovpn) PreConnect() (err error) {
	return
}

func (o *Ovpn) Start() (err error) {
	err = o.conn.Client.Start(o)
	if err != nil {
		return
	}

	return
}

func (o *Ovpn) Connect(data *ConnData) (err error) {
	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	remotes := Remotes{}
	if data.Remote != "" || data.Remote6 != "" {
		foundRemote := false
		foundRemote6 := false
		for _, remote := range o.conn.Data.Remotes {
			if remote.Type != OvpnRemote {
				continue
			}

			if remote.Equal(data.Remote) {
				foundRemote = true
				remotes = append(remotes, remote)
			}
			if remote.Equal(data.Remote6) {
				foundRemote6 = true
				remotes = append(remotes, remote)
			}
		}

		for _, remote := range o.conn.Data.Remotes {
			if remote.Type == OvpnRemote {
				continue
			}

			if !foundRemote && remote.Equal(data.Remote) {
				foundRemote = true
				remotes = append(remotes, remote)
			}
			if !foundRemote6 && remote.Equal(data.Remote6) {
				foundRemote6 = true
				remotes = append(remotes, remote)
			}
		}

		if len(remotes) == 0 {
			if o.conn.Data.DefaultOvpnPort != 0 &&
				o.conn.Data.DefaultOvpnProto != "" {

				logrus.WithFields(o.conn.Fields(logrus.Fields{
					"fixed_remote":  data.Remote,
					"fixed_remote6": data.Remote6,
				})).Error("connection: Failed to match fixed remote, " +
					"using default port protocol")

				remotes = append(remotes, &Remote{
					Addr4: data.Remote,
					Addr6: data.Remote6,
				})
			} else {
				logrus.WithFields(o.conn.Fields(logrus.Fields{
					"fixed_remote":  data.Remote,
					"fixed_remote6": data.Remote6,
				})).Error("connection: Failed to match fixed remote, " +
					"no default port protocol available")

				err = &errortypes.ParseError{
					errors.New("connection: Missing port protocol"),
				}
				return
			}

		}
	} else {
		for _, remote := range o.conn.Data.Remotes {
			if remote.Type != OvpnRemote {
				continue
			}
			remotes = append(remotes, remote)
		}
	}

	o.remotes = remotes.GetParser()

	o.parsedPrfl = parser.Import(
		o.conn.Profile.Data,
		o.remotes,
		o.conn.Profile.DisableGateway,
		o.conn.Profile.DisableDns,
		o.conn.Profile.Dco,
	)

	if runtime.GOOS == "windows" {
		n := GlobalStore.Len()

		err = tuntap.Resize(n)
		if err != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("profile: Failed to resize tuntap adapters")
			err = nil
		}

		if o.conn.State.IsStop() {
			o.conn.State.Close()
			return
		}

		err = tuntap.Configure()
		if err != nil {
			return
		}

		if o.conn.State.IsStop() {
			o.conn.State.Close()
			return
		}
	}

	confPath, err := o.write(data)
	if err != nil {
		return
	}
	o.conn.State.AddPath(confPath)

	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	var authPath string
	// TODO o.conn.Profile.ServerBoxPublicKey != "" ||
	// TODO o.conn.Profile.ServerPublicKey != "" ||
	if (o.conn.Profile.Username != "" && o.conn.Profile.Password != "") ||
		o.parsedPrfl.AuthUserPass ||
		o.conn.Data.HasAuthToken() || data.Token != "" {

		authPath, err = o.writeAuth(data.Token)
		if err != nil {
			return
		}
		o.conn.State.AddPath(authPath)
	}

	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	o.conn.Data.UpdateEvent()

	args := []string{
		"--config", confPath,
		"--verb", "2",
	}

	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	if runtime.GOOS == "windows" {
		o.tapIface = tuntap.Acquire()

		if o.tapIface == "null" {
		} else if o.tapIface != "" {
			args = append(args, "--dev-node", o.tapIface)
		} else {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"tap_size": tuntap.Size(),
			})).Error("connection: Failed to acquire tap")
		}

		if o.conn.State.IsStop() {
			o.conn.State.Close()
			return
		}
	}

	blockPath, err := o.writeBlock()
	if err != nil {
		return
	}
	o.conn.State.AddPath(blockPath)

	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	switch runtime.GOOS {
	case "windows":
		args = append(args, "--script-security", "1")
		break
	case "darwin":
		upPath, e := o.writeUp()
		if e != nil {
			err = e
			return
		}
		o.conn.State.AddPath(upPath)

		downPath, e := o.writeDown()
		if e != nil {
			err = e
			return
		}
		o.conn.State.AddPath(downPath)

		args = append(args, "--script-security", "2",
			"--up", blockPath,
			"--down", blockPath,
			"--route-pre-down", downPath,
			"--tls-verify", blockPath,
			"--ipchange", blockPath,
			"--route-up", upPath,
		)
		break
	case "linux":
		if HasAppArmor() {
			logrus.Info("connection: AppArmor enabled DNS support unavailable")
		} else {
			upPath, e := o.writeUp()
			if e != nil {
				err = e
				return
			}
			o.conn.State.AddPath(upPath)

			downPath, e := o.writeDown()
			if e != nil {
				err = e
				return
			}
			o.conn.State.AddPath(downPath)

			args = append(args, "--script-security", "2",
				"--up", upPath,
				"--down", downPath,
			)
		}

		break
	default:
		panic("profile: Not implemented")
	}

	if authPath != "" {
		args = append(args, "--auth-user-pass", authPath)
	}

	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	cmd := command.Command(GetOvpnPath(), args...)
	cmd.Dir = GetOvpnDir()
	o.cmd = cmd

	o.stdout, err = cmd.StdoutPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "profile: Failed to get stdout"),
		}
		return
	}

	o.stderr, err = cmd.StderrPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "profile: Failed to get stderr"),
		}
		return
	}

	if o.conn.State.IsStop() {
		o.conn.State.Close()
		return
	}

	o.outputBuffer = make(chan string, 100)
	o.outputWait = sync.WaitGroup{}
	o.outputWait.Add(1)

	go o.watchOutput(o.stdout)
	go o.watchOutput(o.stderr)
	go o.parseOutput()

	err = o.cmd.Start()
	if err != nil {
		o.Close()
		err = &errortypes.ExecError{
			errors.Wrap(err, "profile: Failed to start ovpn"),
		}
		return
	}

	o.running = 1
	go o.watchCmd()
	go o.waitCmd()

	return
}

func (o *Ovpn) WatchConnection() (err error) {
	return
}

func (o *Ovpn) Disconnect() {
	o.Close()

	if o.tapIface != "" {
		tuntap.Release(o.tapIface)
	}

	if o.managementPort != 0 {
		ManagementPortRelease(o.managementPort)
	}

}

func (o *Ovpn) write(data *ConnData) (
	pth string, err error) {

	rootDir, err := GetOvpnConfPath()
	if err != nil {
		return
	}

	pth = filepath.Join(rootDir, o.conn.Id)
	prflData := o.parsedPrfl.Export("")

	if runtime.GOOS == "windows" {
		o.managementPort = ManagementPortAcquire()

		managementPassPath, e := o.writeManagementPass()
		if e != nil {
			err = e
			return
		}
		o.conn.State.AddPath(managementPassPath)

		prflData += fmt.Sprintf(
			"management 127.0.0.1 %d %s\n",
			o.managementPort,
			strings.ReplaceAll(managementPassPath, "\\", "\\\\"),
		)
	}

	_ = os.Remove(pth)
	err = ioutil.WriteFile(pth, []byte(prflData), os.FileMode(0600))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write profile"),
		}
		return
	}

	return
}

func (o *Ovpn) writeManagementPass() (pth string, err error) {
	rootDir, err := GetOvpnConfPath()
	if err != nil {
		return
	}

	o.managementPass, err = utils.RandStr(32)
	if err != nil {
		return
	}

	if runtime.GOOS == "windows" {
		pth = filepath.Join(rootDir, o.conn.Id+"-management.txt")
	} else {
		pth = filepath.Join(rootDir, o.conn.Id+"-management")
	}

	_ = os.Remove(pth)
	err = ioutil.WriteFile(pth, []byte(o.managementPass),
		os.FileMode(0600))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write management"),
		}
		return
	}

	return
}

func (o *Ovpn) writeAuth(authToken string) (pth string, err error) {
	rootDir, err := GetOvpnConfPath()
	if err != nil {
		return
	}

	username := o.conn.Profile.Username
	password := o.conn.Profile.Password

	if authToken != "" {
		var serverPubKey [32]byte
		serverPubKeySlic, e := base64.StdEncoding.DecodeString(
			o.conn.Profile.ServerBoxPublicKey)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "profile: Failed to decode server box key"),
			}
			return
		}
		copy(serverPubKey[:], serverPubKeySlic)

		senderPubKey, senderPrivKey, e := box.GenerateKey(rand.Reader)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "profile: Failed to generate nacl key"),
			}
			return
		}

		var nonce [24]byte
		nonceHash := sha256.Sum256(senderPubKey[:])
		copy(nonce[:], nonceHash[:24])

		username = base64.RawStdEncoding.EncodeToString(senderPubKey[:])

		encrypted := box.Seal([]byte{}, []byte(authToken),
			&nonce, &serverPubKey, senderPrivKey)

		ciphertext64 := base64.RawStdEncoding.EncodeToString(encrypted)
		password = "$f$" + ciphertext64
	} else if o.conn.Profile.ServerBoxPublicKey != "" {
		var serverPubKey [32]byte
		serverPubKeySlic, e := base64.StdEncoding.DecodeString(
			o.conn.Profile.ServerBoxPublicKey)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "profile: Failed to decode server box key"),
			}
			return
		}
		copy(serverPubKey[:], serverPubKeySlic)

		tokn, e := o.conn.Data.GetAuthToken()
		if e != nil {
			err = e
			return
		}

		authData := strings.Join([]string{
			tokn.Token,
			fmt.Sprintf("%d", time.Now().Unix()),
			password,
		}, "")

		senderPubKey, senderPrivKey, e := box.GenerateKey(rand.Reader)
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "profile: Failed to generate nacl key"),
			}
			return
		}

		var nonce [24]byte
		nonceHash := sha256.Sum256(senderPubKey[:])
		copy(nonce[:], nonceHash[:24])

		username = base64.RawStdEncoding.EncodeToString(senderPubKey[:])

		encrypted := box.Seal([]byte{}, []byte(authData),
			&nonce, &serverPubKey, senderPrivKey)

		ciphertext64 := base64.RawStdEncoding.EncodeToString(encrypted)
		password = "$x$" + ciphertext64
	} else if o.conn.Profile.ServerPublicKey != "" {
		block, _ := pem.Decode([]byte(o.conn.Profile.ServerPublicKey))

		pub, e := x509.ParsePKCS1PublicKey(block.Bytes)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "profile: Failed to parse public key"),
			}
			return
		}

		nonce, e := utils.RandStr(32)
		if e != nil {
			err = e
			return
		}

		tokn, e := o.conn.Data.GetAuthToken()
		if e != nil {
			err = e
			return
		}

		authData := &AuthData{
			Token:     tokn.Token,
			Password:  password,
			Nonce:     nonce,
			Timestamp: time.Now().Unix(),
		}

		authDataJson, e := json.Marshal(authData)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "profile: Failed to encode auth data"),
			}
			return
		}

		ciphertext, e := rsa.EncryptOAEP(
			sha512.New(),
			rand.Reader,
			pub,
			authDataJson,
			[]byte{},
		)
		if e != nil {
			err = &errortypes.WriteError{
				errors.Wrap(e, "profile: Failed to encrypt auth data"),
			}
			return
		}

		ciphertext64 := base64.StdEncoding.EncodeToString(ciphertext)

		password = "<%=RSA_ENCRYPTED=%>" + ciphertext64
	}

	pth = filepath.Join(rootDir, o.conn.Id+".auth")

	_ = os.Remove(pth)
	err = ioutil.WriteFile(pth, []byte(username+"\n"+password+"\n"),
		os.FileMode(0600))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write profile auth"),
		}
		return
	}

	return
}

func (o *Ovpn) writeUp() (pth string, err error) {
	rootDir, err := GetOvpnConfPath()
	if err != nil {
		return
	}

	if runtime.GOOS == "windows" {
		pth = filepath.Join(rootDir, o.conn.Id+"-up.bat")
	} else {
		pth = filepath.Join(rootDir, o.conn.Id+"-up.sh")
	}

	script := ""
	switch runtime.GOOS {
	case "darwin":
		if o.conn.Profile.DisableDns {
			script = blockScript
		} else if o.conn.Profile.ForceDns {
			DnsForced = true
			script = upDnsScriptDarwin
		} else {
			script = upScriptDarwin
		}
		break
	case "linux":
		if IsOvpn27() {
			if o.conn.Profile.DisableDns {
				script = blockScript
			} else {
				script = resolvedScript27
			}
		} else {
			resolved := true

			resolvData, _ := ioutil.ReadFile("/etc/resolv.conf")
			if resolvData != nil {
				resolvDataStr := string(resolvData)
				if !strings.Contains(resolvDataStr, "systemd-resolved") &&
					!strings.Contains(resolvDataStr, "127.0.0.53") {

					resolved = false
				}
			}

			if o.conn.Profile.DisableDns {
				script = blockScript
			} else if resolved {
				script = resolvedScript
			} else {
				script = resolvScript
			}
		}
		break
	default:
		panic("profile: Not implemented")
	}

	_ = os.Remove(pth)
	err = ioutil.WriteFile(pth, []byte(script), os.FileMode(0755))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write up script"),
		}
		return
	}

	return
}

func (o *Ovpn) writeDown() (pth string, err error) {
	rootDir, err := GetOvpnConfPath()
	if err != nil {
		return
	}

	if runtime.GOOS == "windows" {
		pth = filepath.Join(rootDir, o.conn.Id+"-down.bat")
	} else {
		pth = filepath.Join(rootDir, o.conn.Id+"-down.sh")
	}

	script := ""
	switch runtime.GOOS {
	case "darwin":
		if o.conn.Profile.DisableDns {
			script = blockScript
		} else {
			script = downScriptDarwin
		}
		break
	case "linux":
		if IsOvpn27() {
			if o.conn.Profile.DisableDns {
				script = blockScript
			} else {
				script = resolvedScript27
			}
		} else {
			resolved := true

			resolvData, _ := ioutil.ReadFile("/etc/resolv.conf")
			if resolvData != nil {
				resolvDataStr := string(resolvData)
				if !strings.Contains(resolvDataStr, "systemd-resolved") &&
					!strings.Contains(resolvDataStr, "127.0.0.53") {

					resolved = false
				}
			}

			if o.conn.Profile.DisableDns {
				script = blockScript
			} else if resolved {
				script = resolvedScript
			} else {
				script = resolvScript
			}
		}
		break
	default:
		panic("profile: Not implemented")
	}

	_ = os.Remove(pth)
	err = ioutil.WriteFile(pth, []byte(script), os.FileMode(0755))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write down script"),
		}
		return
	}

	return
}

func (o *Ovpn) writeBlock() (pth string, err error) {
	rootDir, err := GetOvpnConfPath()
	if err != nil {
		return
	}

	script := ""
	if runtime.GOOS == "windows" {
		pth = filepath.Join(rootDir, o.conn.Id+"-block.bat")
		script = blockScriptWindows
	} else {
		pth = filepath.Join(rootDir, o.conn.Id+"-block.sh")
		script = blockScript
	}

	_ = os.Remove(pth)
	err = ioutil.WriteFile(pth, []byte(script), os.FileMode(0755))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "profile: Failed to write block script"),
		}
		return
	}

	return
}

func (o *Ovpn) watchOutput(buffer io.ReadCloser) {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Watch output panic")
		}
	}()

	defer func() {
		_ = buffer.Close()
		// TODO Possibly only end on stdout close
		o.outputBuffer <- ""
	}()

	out := bufio.NewReader(buffer)
	for {
		line, _, err := out.ReadLine()
		if err != nil {
			if err != io.EOF &&
				!strings.Contains(err.Error(), "file already closed") &&
				!strings.Contains(err.Error(), "bad file descriptor") {

				err = &errortypes.ReadError{
					errors.Wrap(err, "profile: Failed to read stdout"),
				}
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("profile: Stdout error")
			}

			return
		}

		lineStr := string(line)
		if lineStr != "" {
			o.outputBuffer <- lineStr
		}
	}
}

func (o *Ovpn) parseOutput() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Parse output panic")
		}
	}()

	defer o.outputWait.Done()

	for {
		line := <-o.outputBuffer
		if line == "" {
			return
		}

		o.parseLine(line)
	}
}

func (o *Ovpn) Close() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Close ovpn panic")
		}
	}()

	o.killCmd()

	stdout := o.stdout
	stderr := o.stderr
	outputBuffer := o.outputBuffer
	if stdout != nil {
		stdout.Close()
	}
	if stderr != nil {
		stderr.Close()
	}
	if outputBuffer != nil {
		outputBuffer <- ""
	}
}

func (o *Ovpn) killCmd() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Kill ovpn cmd panic")
		}
	}()

	cmd := o.cmd
	if cmd == nil || cmd.Process == nil {
		return
	}

	waiter := make(chan bool, 8)
	exited := false

	if o.cmd.ProcessState == nil || !o.cmd.ProcessState.Exited() {
		if runtime.GOOS == "windows" {
			err := o.sendManagementCommand("signal SIGTERM")
			if err != nil {
				err = &errortypes.ExecError{
					errors.Wrap(err, "profile: Management interrupt error"),
				}
				logrus.WithFields(o.conn.Fields(logrus.Fields{
					"error": err,
				})).Error("profile: Management interrupt failed")
			}
		} else {
			err := o.cmd.Process.Signal(os.Interrupt)
			if err != nil {
				err = &errortypes.ExecError{
					errors.Wrap(err, "profile: Interrupt error"),
				}
				logrus.WithFields(o.conn.Fields(logrus.Fields{
					"error": err,
				})).Error("profile: Failed to interrupt ovpn process")
			}
		}
	}

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("utils: Kill command wait panic")
			}
		}()
		cmd.Wait()
		exited = true
		waiter <- true
	}()

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("utils: Kill exit wait panic")
				time.Sleep(1 * time.Second)
				waiter <- true
			}
		}()

		for i := 0; i < 80; i++ {
			time.Sleep(100 * time.Millisecond)
			if exited {
				break
			}
		}

		if !exited {
			logrus.WithFields(o.conn.Fields(nil)).Error(
				"profile: Exit timeout in ovpn process")

			err := cmd.Process.Kill()
			if err != nil {
				err = &errortypes.ExecError{
					errors.Wrap(err, "profile: Kill error"),
				}
				logrus.WithFields(o.conn.Fields(logrus.Fields{
					"error": err,
				})).Error("profile: Failed to kill ovpn process")
			}
			time.Sleep(1 * time.Second)
		}
		waiter <- true
	}()

	<-waiter
	time.Sleep(100 * time.Millisecond)
	cmd.Process.Kill()
}

func (o *Ovpn) parseLine(line string) {
	o.pushOutput(line)

	if o.conn.State.IsStop() {
		return
	}

	if strings.Contains(line, "Initialization Sequence Completed") ||
		strings.Contains(line, "Peer Connection Initiated") {

		o.connected = true
		o.conn.Data.Status = Connected
		o.conn.Data.Timestamp = time.Now().Unix() - 3
		o.conn.Data.UpdateEvent()

		o.conn.Data.ValidateAuthToken()

		sIndex := strings.LastIndex(line, "]")
		eIndex := strings.LastIndex(line, ":")
		if sIndex != -1 && eIndex != -1 {
			o.conn.Data.ServerAddr = strings.TrimSpace(line[sIndex+1 : eIndex])
			o.conn.Data.UpdateEvent()
		}

		go func() {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(o.conn.Fields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					})).Error("profile: Clear DNS cache panic")
				}
			}()

			utils.ClearDNSCache()
		}()
	} else if strings.Contains(line, "Inactivity timeout (--inactive)") {
		o.conn.Data.SendProfileEvent("inactive")
	} else if strings.Contains(line, "Inactivity timeout") ||
		strings.Contains(line, "Connection reset") {

		o.conn.Data.SendProfileEvent("timeout_error")
	} else if strings.Contains(
		line, "Can't assign requested address (code=49)") {

		logrus.WithFields(o.conn.Fields(logrus.Fields{
			"line": line,
		})).Error("connection: Assign address error")

		go func() {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(o.conn.Fields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					})).Error("profile: Kill profile panic")
				}
			}()

			o.killCmd()

			// // TODO Possibly restart all and reset network
			// if !o.conn.State.IsStop() {
			// 	go RestartProfiles()
			// }
		}()
	} else if strings.Contains(line, "AUTH_FAILED") || strings.Contains(
		line, "auth-failure") && !o.authFailed {

		o.authFailed = true
		o.conn.Data.ResetAuthToken()
		o.conn.State.NoReconnect("ovpn_auth_error")
		o.conn.State.SetStop()

		if o.conn.Profile.SystemProfile {
			logrus.WithFields(o.conn.Fields(nil)).Info(
				"connection: Stopping system profile due to " +
					"authentication errors")

			sprofile.Deactivate(o.conn.Profile.Id)
			sprofile.SetAuthErrorCount(o.conn.Profile.Id, 0)
		} else {
			time.Sleep(3 * time.Second)
		}

		if utils.SinceAbs(o.lastAuthFailed) > 5*time.Second {
			o.lastAuthFailed = time.Now()
			o.conn.Data.SendProfileEvent("auth_error")
		}
	} else if strings.Contains(line, "link remote:") {
		sIndex := strings.LastIndex(line, "]")
		eIndex := strings.LastIndex(line, ":")
		if sIndex != -1 && eIndex != -1 {
			o.conn.Data.ServerAddr = strings.TrimSpace(line[sIndex+1 : eIndex])
			o.conn.Data.UpdateEvent()
		}
	} else if strings.Contains(line, "add net") &&
		strings.Contains(line, "gateway") {

		split := strings.Fields(line)
		gwIndex := -1
		for i, s := range split {
			if s == "gateway" {
				gwIndex = i
				break
			}
		}

		if gwIndex != -1 && gwIndex+1 < len(split) &&
			o.conn.Data.ClientAddr == "" {

			o.conn.Data.ClientAddr = split[gwIndex+1]
			o.conn.Data.UpdateEvent()
		}
	} else if strings.Contains(line, "network/local/netmask") {
		eIndex := strings.LastIndex(line, "/")
		line = line[:eIndex]
		sIndex := strings.LastIndex(line, "/") + 1

		o.conn.Data.ClientAddr = line[sIndex:]
		o.conn.Data.UpdateEvent()
	} else if strings.Contains(line, "ifconfig") && strings.Contains(
		line, "netmask") {

		sIndex := strings.Index(line, "ifconfig") + 9
		eIndex := strings.Index(line, "netmask")
		line = line[sIndex:eIndex]

		split := strings.Split(line, " ")
		if len(split) > 2 {
			o.conn.Data.ClientAddr = split[1]
			o.conn.Data.UpdateEvent()
		}
	} else if strings.Contains(line, "ip addr add dev") {
		clientAddr := ""
		sIndex := strings.Index(line, "ip addr add dev") + 16
		eIndex := strings.Index(line, "broadcast")

		if eIndex == -1 {
			ipList := ipReg.FindAllString(line, -1)
			if len(ipList) > 0 {
				clientAddr = ipList[0]
			}
		} else {
			line = line[sIndex:eIndex]
			split := strings.Split(line, " ")

			if len(split) > 1 {
				split := strings.Split(split[1], "/")
				if len(split) > 1 {
					clientAddr = split[0]
				}
			}
		}

		if clientAddr != "" {
			o.conn.Data.ClientAddr = clientAddr
			o.conn.Data.UpdateEvent()
		}
	} else if strings.Contains(line, "net_addr_v4_add:") {
		clientAddr := ""
		line = line[strings.Index(line, "net_addr_v4_add:")+17:]
		line = strings.TrimSpace(line)

		ipList := ipReg.FindAllString(line, -1)
		if len(ipList) > 0 {
			clientAddr = ipList[0]
		}

		if clientAddr != "" {
			o.conn.Data.ClientAddr = clientAddr
			o.conn.Data.UpdateEvent()
		}
	}
}

func (o *Ovpn) pushOutput(output string) {
	output = strings.TrimSpace(output)

	err := log.ProfilePushLog(o.conn.Id, output)
	if err != nil {
		logrus.WithFields(o.conn.Fields(logrus.Fields{
			"output": output,
			"error":  err,
		})).Error("connection: Failed to push profile log output")
	}

	return
}

func (o *Ovpn) sendManagementCommand(cmd string) (err error) {
	o.managementLock.Lock()
	defer o.managementLock.Unlock()

	conn, err := net.DialTimeout(
		"tcp",
		fmt.Sprintf("127.0.0.1:%d", o.managementPort),
		3*time.Second,
	)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Failed to open socket"),
		}
		return
	}
	defer conn.Close()

	go func() {
		for {
			buf := make([]byte, 10000)
			n, e := conn.Read(buf)
			if e != nil || n == 0 {
				break
			}
		}
	}()

	err = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Failed set deadline"),
		}
		return
	}

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", o.managementPass)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Failed to write socket password"),
		}
		return
	}

	time.Sleep(500 * time.Millisecond)

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", cmd)))
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Failed to write socket command"),
		}
		return
	}

	return
}

func (o *Ovpn) watchCmd() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Watch cmd panic")
		}
	}()

	for {
		time.Sleep(3 * time.Second)
		if o.conn.State.IsStop() {
			break
		}
	}
}

func (o *Ovpn) waitCmd() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Wait cmd panic")
		}
	}()

	o.cmd.Wait()
	o.outputWait.Wait()
	o.running = -1

	if runtime.GOOS == "darwin" {
		err := utils.RestoreScutilDns(false)
		if err != nil {
			logrus.WithFields(o.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("profile: Failed to restore DNS")
		}
	}

	o.conn.State.Close()
}
