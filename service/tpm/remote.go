package tpm

import (
	"encoding/base64"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
)

// Remote performs secure enclave operations through a user session
// client, see request.go for the claim protocol. Each operation runs a
// separate helper process on the client, the enclave key is passed by
// its data representation so no state is held between operations.
type Remote struct {
	privKey64 string
	pubKey64  string
}

// Open loads the enclave key, generating one when privKey64 is empty.
// The public key of a stored key is cached in the config so a stored key
// needs no client round trip, the key is loaded by the sign request.
func (t *Remote) Open(privKey64 string) (err error) {
	if privKey64 != "" && config.Config.EnclavePublicKey != "" {
		t.privKey64 = privKey64
		t.pubKey64 = config.Config.EnclavePublicKey
		return
	}

	id, err := utils.RandStr(16)
	if err != nil {
		return
	}

	res, err := run(&Request{
		Id:      id,
		Type:    RequestOpen,
		KeyData: privKey64,
	})
	if err != nil {
		return
	}

	t.privKey64 = res.KeyData
	t.pubKey64 = res.PublicKey

	return
}

func (t *Remote) Close() {
}

func (t *Remote) PublicKey() (pubKey64 string, err error) {
	pubKey64 = t.pubKey64
	return
}

func (t *Remote) Sign(data []byte) (privKey64, sig64 string, err error) {
	if t.privKey64 == "" || t.pubKey64 == "" {
		err = &errortypes.RequestError{
			errors.New("tpm: Sign before open"),
		}
		return
	}

	id, err := utils.RandStr(16)
	if err != nil {
		return
	}

	res, err := run(&Request{
		Id:        id,
		Type:      RequestSign,
		KeyData:   t.privKey64,
		SignData:  base64.StdEncoding.EncodeToString(data),
		PublicKey: t.pubKey64,
	})
	if err != nil {
		return
	}

	privKey64 = t.privKey64
	sig64 = res.Signature

	return
}
