package tpm

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dropbox/godropbox/errors"
	"github.com/google/go-tpm-tools/client"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

type authInput struct {
	KeyData string `json:"key_data"`
}

type authInput2 struct {
	SignData string `json:"sign_data"`
}

type authOutput struct {
	KeyData   string `json:"key_data"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

type Tpm struct {
	cmd        *exec.Cmd
	stdout     io.ReadCloser
	stderr     *bytes.Buffer
	stdin      io.WriteCloser
	key        *client.Key
	waiter     chan bool
	waiterSet  bool
	exitWaiter chan bool
	privKey64  string
	pubKey64   string
	sig64      string
	readErr    error
	exitErr    error
}

func (t *Tpm) Open(privKey64 string) (err error) {
	t.waiter = make(chan bool, 8)
	t.exitWaiter = make(chan bool, 8)
	t.privKey64 = privKey64

	deviceAuthPth := getDeviceAuthPath()

	t.cmd = command.Command(deviceAuthPth)

	t.stderr = &bytes.Buffer{}
	t.cmd.Stderr = t.stderr

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "tpm: Failed to open stdout"),
		}
		return
	}

	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "tpm: Failed to open stdin"),
		}
		return
	}

	err = t.cmd.Start()
	if err != nil {
		t.Close()
		err = &errortypes.ExecError{
			errors.Wrapf(err, "tpm: Failed to exec device auth"),
		}
		return
	}

	go t.reader()
	go t.wait()

	inputData := &authInput{
		KeyData: t.privKey64,
	}

	inputByt, err := json.Marshal(inputData)
	if err != nil {
		t.Close()
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to marshal input data"),
		}
		return
	}

	err = t.write(inputByt)
	if err != nil {
		t.Close()
		return
	}

	return
}

func (t *Tpm) write(input []byte) (err error) {
	input = append(input, '\n')

	_, err = t.stdin.Write(input)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err,
				"tpm: Failed to write to device auth",
			),
		}
		return
	}

	return
}

func (t *Tpm) wait() {
	defer func() {
		t.exitWaiter <- true

		if !t.waiterSet {
			t.waiterSet = true
			t.waiter <- true
		}
	}()

	t.exitErr = t.cmd.Wait()
	errOutput := t.stderr.String()

	if t.exitErr != nil {
		t.exitErr = &errortypes.WriteError{
			errors.Wrapf(t.exitErr,
				"tpm: Device auth error",
			),
		}

		logrus.WithFields(logrus.Fields{
			"output": errOutput,
			"error":  t.exitErr,
		}).Error("utils: Device auth error")

		return
	}

	return
}

func (t *Tpm) reader() {
	defer t.Close()

	reader := bufio.NewReader(t.stdout)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			t.readErr = &errortypes.ExecError{
				errors.Wrapf(err, "tpm: Failed to read line"),
			}
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("utils: Failed to read line")
			return
		}

		outputData := &authOutput{}

		err = json.Unmarshal(bytes.TrimSpace(line), outputData)
		if err != nil {
			t.readErr = &errortypes.ParseError{
				errors.Wrap(err, "tpm: Failed to unmarshal output data"),
			}
			logrus.WithFields(logrus.Fields{
				"output": string(line),
				"error":  err,
			}).Error("utils: Failed to unmarshal line")
			return
		}

		if outputData.KeyData != "" {
			t.privKey64 = outputData.KeyData
		}
		if outputData.PublicKey != "" {
			t.pubKey64 = outputData.PublicKey
		}
		if outputData.Signature != "" {
			t.sig64 = outputData.Signature
		}

		if !t.waiterSet {
			t.waiterSet = true
			t.waiter <- true
		}
	}
}

func (t *Tpm) Close() {
	defer func() {
		if !t.waiterSet {
			t.waiterSet = true
			t.waiter <- true
		}
	}()

	if t.stdout != nil {
		_ = t.stdout.Close()
	}
	if t.stdin != nil {
		_ = t.stdin.Close()
	}

	return
}

func (t *Tpm) PublicKey() (pubKey64 string, err error) {
	<-t.waiter

	err = t.exitErr
	if err != nil {
		return
	}

	err = t.readErr
	if err != nil {
		return
	}

	pubKey64 = t.pubKey64
	return
}

func (t *Tpm) Sign(data []byte) (privKey64, sig64 string, err error) {
	inputData := &authInput2{
		SignData: base64.StdEncoding.EncodeToString(data),
	}

	input, err := json.Marshal(inputData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to marshal input data"),
		}
		return
	}

	err = t.write(input)
	if err != nil {
		return
	}

	<-t.exitWaiter

	err = t.exitErr
	if err != nil {
		return
	}

	err = t.readErr
	if err != nil {
		return
	}

	privKey64 = t.privKey64
	sig64 = t.sig64

	return
}

func getDeviceAuthPath() string {
	if constants.Development {
		return filepath.Join(utils.GetRootDir(), "..",
			"service_macos", "Pritunl Device Authentication")
	}

	return filepath.Join(string(os.PathSeparator), "Applications",
		"Pritunl.app", "Contents", "Resources",
		"Pritunl Device Authentication")
}
