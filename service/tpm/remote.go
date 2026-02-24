package tpm

import (
	"encoding/base64"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/utils"
)

var (
	callers     = map[string]*Remote{}
	callersLock = sync.Mutex{}
)

type tpmEventData struct {
	Id         string `json:"id"`
	PrivateKey string `json:"private_key"`
	SignData   string `json:"sign_data"`
}

type Remote struct {
	callerId  string
	callerErr string
	privKey64 string
	pubKey64  string
	sig64     string
}

func (t *Remote) Open(privKey64 string) (err error) {
	t.callerId, err = utils.RandStr(16)
	if err != nil {
		return
	}

	callersLock.Lock()
	callers[t.callerId] = t
	callersLock.Unlock()

	evt := event.Event{
		Type: "tpm_open",
		Data: &tpmEventData{
			Id:         t.callerId,
			PrivateKey: privKey64,
		},
	}
	evt.Init()

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if t.pubKey64 != "" || t.callerErr != "" {
			break
		}
	}

	if t.callerErr != "" {
		err = &errortypes.RequestError{
			errors.New("tpm: Client TPM error " + t.callerErr),
		}
		return
	}

	if t.pubKey64 == "" {
		err = &errortypes.RequestError{
			errors.New("tpm: Timeout waiting for client TPM open"),
		}
		return
	}

	return
}

func (t *Remote) Close() {
	evt := event.Event{
		Type: "tpm_close",
		Data: &tpmEventData{
			Id: t.callerId,
		},
	}
	evt.Init()

	callersLock.Lock()
	delete(callers, t.callerId)
	callersLock.Unlock()
}

func (t *Remote) PublicKey() (pubKey64 string, err error) {
	pubKey64 = t.pubKey64
	return
}

func (t *Remote) Sign(data []byte) (privKey64, sig64 string, err error) {
	evt := event.Event{
		Type: "tpm_sign",
		Data: &tpmEventData{
			Id:       t.callerId,
			SignData: base64.StdEncoding.EncodeToString(data),
		},
	}
	evt.Init()

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if t.sig64 != "" || t.callerErr != "" {
			break
		}
	}

	if t.callerErr != "" {
		err = &errortypes.RequestError{
			errors.New("tpm: Client TPM error " + t.callerErr),
		}
		return
	}

	if t.sig64 == "" {
		err = &errortypes.RequestError{
			errors.New("tpm: Timeout waiting for client TPM sign"),
		}
		return
	}

	privKey64 = t.privKey64
	sig64 = t.sig64

	return
}

func RemoteCallback(callerId, pubKey, privKey, signature, error string) {
	callersLock.Lock()
	caller := callers[callerId]
	callersLock.Unlock()

	if caller == nil {
		return
	}

	if pubKey != "" {
		caller.pubKey64 = pubKey
	}
	if privKey != "" {
		caller.privKey64 = privKey
	}
	if signature != "" {
		caller.sig64 = signature
	}
	if error != "" {
		caller.callerErr = error
	}
}
