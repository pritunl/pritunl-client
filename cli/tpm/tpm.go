package tpm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/constants"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/service"
	"github.com/pritunl/pritunl-client/cli/utils"
	"github.com/pritunl/tools/logger"
)

const (
	RequestOpen = "tpm_open"
	RequestSign = "tpm_sign"

	// Helper process lifetime, must be below the service claim ttl.
	procTimeout = 10 * time.Second
	// Upper bound on concurrent helper processes.
	maxProcs = 4
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

type claimData struct {
	ClientId string `json:"client_id"`
}

type resultData struct {
	ClientId  string `json:"client_id"`
	KeyData   string `json:"key_data,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Signature string `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

type proc struct {
	requestId string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stderr    *bytes.Buffer
	lock      sync.Mutex
	done      bool
}

var (
	// Identifies this process to the service as the claimant of a
	// request.
	clientId  = utils.Uuid()
	procs     = map[string]*proc{}
	procsLock = sync.Mutex{}
)

func getDeviceAuthPath() string {
	if constants.Development {
		return filepath.Join(utils.GetRootDir(), "..",
			"service_macos", "Pritunl Device Authentication")
	}

	return filepath.Join(string(os.PathSeparator), "Applications",
		"Pritunl.app", "Contents", "Resources",
		"Pritunl Device Authentication")
}

func post(path string, data interface{}) (status int, err error) {
	body, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to marshal request"),
		}
		return
	}

	authKey, err := service.GetAuthKey()
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", service.GetAddress()+path,
		bytes.NewReader(body))
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "tpm: Request failed"),
		}
		return
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		req.Host = "unix"
	}
	req.Header.Set("Auth-Key", authKey)
	req.Header.Set("User-Agent", "pritunl")
	req.Header.Set("Content-Type", "application/json")

	resp, err := service.GetClient().Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "tpm: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	status = resp.StatusCode

	return
}

// claim claims the request with the service, false when another client
// holds the claim or the request is already completed.
func claim(requestId string) bool {
	status, err := post("/tpm/request/"+requestId+"/claim", &claimData{
		ClientId: clientId,
	})
	if err != nil {
		logger.WithFields(logger.Fields{
			"request_id": requestId,
			"error":      err,
		}).Error("tpm: Claim request error")
		return false
	}

	if status == 200 {
		return true
	}

	// 409 claimed by another client, 404 already completed or abandoned
	// by the service.
	if status != 409 && status != 404 {
		logger.WithFields(logger.Fields{
			"request_id": requestId,
			"status":     status,
		}).Error("tpm: Claim request error")
	}

	return false
}

func complete(requestId string, result *resultData) {
	result.ClientId = clientId

	status, err := post("/tpm/request/"+requestId, result)
	if err != nil {
		logger.WithFields(logger.Fields{
			"request_id": requestId,
			"error":      err,
		}).Error("tpm: Result request error")
		return
	}

	if status != 200 {
		logger.WithFields(logger.Fields{
			"request_id": requestId,
			"status":     status,
		}).Error("tpm: Result request error")
	}
}

// fail logs the error and reports it to the service so the claim is
// released immediately instead of waiting for the claim ttl. Only the
// first failure or result for a request is reported.
func (p *proc) fail(err error) {
	p.lock.Lock()
	if p.done {
		p.lock.Unlock()
		return
	}
	p.done = true
	p.lock.Unlock()

	logger.WithFields(logger.Fields{
		"request_id": p.requestId,
		"error":      err,
	}).Error("tpm: Secure enclave error")

	complete(p.requestId, &resultData{
		Error: err.Error(),
	})
}

func (p *proc) result(result *resultData) {
	p.lock.Lock()
	if p.done {
		p.lock.Unlock()
		return
	}
	p.done = true
	p.lock.Unlock()

	complete(p.requestId, result)
}

func (p *proc) isDone() bool {
	p.lock.Lock()
	done := p.done
	p.lock.Unlock()
	return done
}

func (p *proc) write(data interface{}) (err error) {
	input, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to marshal input"),
		}
		return
	}
	input = append(input, '\n')

	// A write to a helper that already exited returns EPIPE here
	// instead of raising a signal.
	_, err = p.stdin.Write(input)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "tpm: Failed to write to secure enclave process"),
		}
		return
	}

	return
}

func (p *proc) kill() {
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Signal(os.Interrupt)
	}
}

func (p *proc) timeout() {
	time.Sleep(procTimeout)

	if p.cmd.ProcessState != nil {
		return
	}

	logger.WithFields(logger.Fields{
		"request_id": p.requestId,
	}).Error("tpm: Secure enclave process timed out")

	p.kill()
}

// handleLine processes one JSON output line from the helper. An open
// request completes on the first line and closes stdin so the helper
// exits, a sign request completes on the line carrying the signature.
func (p *proc) handleLine(typ string, line []byte) {
	output := &authOutput{}

	err := json.Unmarshal(line, output)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to parse secure enclave output"),
		}
		p.fail(err)
		return
	}

	if typ == RequestOpen {
		p.result(&resultData{
			KeyData:   output.KeyData,
			PublicKey: output.PublicKey,
		})
		_ = p.stdin.Close()
	} else if output.Signature != "" {
		p.result(&resultData{
			Signature: output.Signature,
		})
	}
}

// run spawns the helper, feeds it the request and reports the result.
// Stdout is read to EOF before the process is reaped, Wait must not run
// while the pipe is still being read or the final line can be lost.
func run(typ, requestId, keyData, signData string) {
	procsLock.Lock()
	count := len(procs)
	procsLock.Unlock()

	if count >= maxProcs {
		err := &errortypes.ExecError{
			errors.Newf("tpm: Too many secure enclave processes count=%d",
				count),
		}
		logger.WithFields(logger.Fields{
			"request_id": requestId,
			"error":      err,
		}).Error("tpm: Secure enclave error")
		complete(requestId, &resultData{
			Error: err.Error(),
		})
		return
	}

	cmd := exec.Command(getDeviceAuthPath())

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	p := &proc{
		requestId: requestId,
		cmd:       cmd,
		stderr:    stderr,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.fail(&errortypes.ExecError{
			errors.Wrap(err, "tpm: Failed to open stdout"),
		})
		return
	}

	p.stdin, err = cmd.StdinPipe()
	if err != nil {
		p.fail(&errortypes.ExecError{
			errors.Wrap(err, "tpm: Failed to open stdin"),
		})
		return
	}

	err = cmd.Start()
	if err != nil {
		p.fail(&errortypes.ExecError{
			errors.Wrap(err, "tpm: Secure enclave exec error"),
		})
		return
	}

	procsLock.Lock()
	procs[requestId] = p
	procsLock.Unlock()

	go p.timeout()

	// Both inputs are written up front, the helper reads them in order.
	err = p.write(&authInput{
		KeyData: keyData,
	})
	if err == nil && typ == RequestSign {
		err = p.write(&authInput2{
			SignData: signData,
		})
	}
	if err != nil {
		p.fail(err)
		p.kill()
	}

	reader := bufio.NewReader(stdout)
	for {
		line, e := reader.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if len(line) > 0 && !p.isDone() {
			p.handleLine(typ, line)
		}
		if e != nil {
			break
		}
	}

	err = cmd.Wait()

	procsLock.Lock()
	if procs[requestId] == p {
		delete(procs, requestId)
	}
	procsLock.Unlock()

	if !p.isDone() {
		if err == nil {
			err = errors.New("tpm: Secure enclave process exited " +
				"without result")
		}
		p.fail(&errortypes.ExecError{
			errors.Wrapf(err, "tpm: Secure enclave exec error output=%q",
				stderr.String()),
		})
	}
}

// Handle processes a tpm_open or tpm_sign event. The claim and helper
// run in the background so the interface is not blocked.
func Handle(typ, requestId, keyData, signData string) {
	if requestId == "" {
		logger.WithFields(logger.Fields{
			"event_type": typ,
		}).Error("tpm: Secure enclave event missing request id")
		return
	}

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logger.WithFields(logger.Fields{
					"request_id": requestId,
					"panic":      panc,
				}).Error("tpm: Secure enclave handler panic")
			}
		}()

		if !claim(requestId) {
			return
		}

		run(typ, requestId, keyData, signData)
	}()
}
