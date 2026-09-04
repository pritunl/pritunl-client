package tpm

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/sirupsen/logrus"
)

const (
	RequestOpen = "tpm_open"
	RequestSign = "tpm_sign"

	// Time a claimant has to return a result before the claim is
	// released. Must exceed the helper process lifetime enforced by the
	// clients (10 seconds) so a slow but successful helper is not
	// rejected as a non-claimant.
	RequestClaimTtl = 12 * time.Second
	// Interval to broadcast an unclaimed request again. Covers a client
	// whose event socket was reconnecting when the first broadcast was
	// sent, such as shortly after a sleep wake.
	RequestRebroadcast = 3 * time.Second
	// Total time to wait for a result before the operation fails, allows
	// for one claim failover to the other client.
	RequestTimeout = 30 * time.Second
)

type requestEventData struct {
	RequestId string `json:"request_id"`
	KeyData   string `json:"key_data"`
	SignData  string `json:"sign_data,omitempty"`
}

type Result struct {
	KeyData   string
	PublicKey string
	Signature string
}

type Request struct {
	Id        string
	Type      string
	KeyData   string
	SignData  string
	PublicKey string
	Created   time.Time
	Broadcast time.Time
	ClaimedAt time.Time
	ClaimedBy string
	LastError string
	completed bool
	result    *Result
	done      chan struct{}
}

var (
	requests     = map[string]*Request{}
	requestsLock = sync.Mutex{}
)

func (r *Request) fields() logrus.Fields {
	return logrus.Fields{
		"request_id":   r.Id,
		"request_type": r.Type,
		"claimed_by":   r.ClaimedBy,
	}
}

// broadcast sends the request event to all clients, called with the
// requests lock held.
func (r *Request) broadcast() {
	evt := event.Event{
		Type: r.Type,
		Data: &requestEventData{
			RequestId: r.Id,
			KeyData:   r.KeyData,
			SignData:  r.SignData,
		},
	}
	evt.Init()
	r.Broadcast = time.Now()
}

// release clears the claim and broadcasts the request again, called with
// the requests lock held.
func (r *Request) release() {
	r.ClaimedBy = ""
	r.ClaimedAt = time.Time{}
	r.broadcast()
}

func parsePublicKey(pubKey64 string) (key *ecdsa.PublicKey, err error) {
	der, err := base64.StdEncoding.DecodeString(pubKey64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to decode public key"),
		}
		return
	}

	keyInf, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "tpm: Failed to parse public key"),
		}
		return
	}

	key, ok := keyInf.(*ecdsa.PublicKey)
	if !ok {
		err = &errortypes.ParseError{
			errors.New("tpm: Public key is not ECDSA"),
		}
		return
	}

	return
}

// validate checks a submitted result. Open results must carry a parsable
// public key, sign results must verify against the request public key.
// The helper signs with ECDSA P-256 over SHA-256 of the sign data.
func (r *Request) validate(res *Result) (err error) {
	switch r.Type {
	case RequestOpen:
		if res.KeyData == "" || res.PublicKey == "" {
			err = &errortypes.ParseError{
				errors.New("tpm: Open result missing key data"),
			}
			return
		}

		_, err = parsePublicKey(res.PublicKey)
		if err != nil {
			return
		}
	case RequestSign:
		if res.Signature == "" {
			err = &errortypes.ParseError{
				errors.New("tpm: Sign result missing signature"),
			}
			return
		}

		pubKey, e := parsePublicKey(r.PublicKey)
		if e != nil {
			err = e
			return
		}

		sig, e := base64.StdEncoding.DecodeString(res.Signature)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "tpm: Failed to decode signature"),
			}
			return
		}

		signData, e := base64.StdEncoding.DecodeString(r.SignData)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "tpm: Failed to decode sign data"),
			}
			return
		}

		hash := sha256.Sum256(signData)
		if !ecdsa.VerifyASN1(pubKey, hash[:], sig) {
			err = &errortypes.ParseError{
				errors.New("tpm: Signature verification failed"),
			}
			return
		}
	default:
		err = &errortypes.ParseError{
			errors.New("tpm: Unknown request type"),
		}
		return
	}

	return
}

// run registers the request, broadcasts it and waits for a result.
// Unclaimed requests are broadcast again periodically and expired claims
// are released so another client can complete the request.
func run(r *Request) (res *Result, err error) {
	r.Created = time.Now()
	r.done = make(chan struct{})

	requestsLock.Lock()
	requests[r.Id] = r
	r.broadcast()
	requestsLock.Unlock()

	timeout := time.NewTimer(RequestTimeout)
	defer timeout.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.done:
			requestsLock.Lock()
			res = r.result
			requestsLock.Unlock()
			return
		case <-timeout.C:
			requestsLock.Lock()
			delete(requests, r.Id)
			lastErr := r.LastError
			claimed := r.ClaimedBy != ""
			requestsLock.Unlock()

			msg := "tpm: Timeout waiting for client " + r.Type
			if lastErr != "" {
				msg += " client_error=" + lastErr
			} else if !claimed {
				msg += " (no client claimed request)"
			}

			err = &errortypes.RequestError{
				errors.New(msg),
			}
			return
		case <-ticker.C:
			requestsLock.Lock()
			if r.completed {
				requestsLock.Unlock()
				continue
			}

			if r.ClaimedBy != "" {
				if time.Since(r.ClaimedAt) > RequestClaimTtl {
					logrus.WithFields(r.fields()).Warn(
						"tpm: Client claim expired, releasing request")
					r.release()
				}
			} else if time.Since(r.Broadcast) >= RequestRebroadcast {
				r.broadcast()
			}
			requestsLock.Unlock()
		}
	}
}

// Claim attempts to claim the request for the client. Returns 200 when
// claimed, 409 when the request is already claimed and the claim has not
// expired, including by the same client, and 404 when the request is
// unknown or already completed.
func Claim(requestId, clientId string) int {
	requestsLock.Lock()
	defer requestsLock.Unlock()

	r := requests[requestId]
	if r == nil || r.completed {
		return 404
	}

	if r.ClaimedBy != "" && time.Since(r.ClaimedAt) <= RequestClaimTtl {
		return 409
	}

	r.ClaimedBy = clientId
	r.ClaimedAt = time.Now()

	return 200
}

// Complete submits a result or a client error for a claimed request.
// Returns 200 when accepted, 409 when the client is not the current
// claimant, 404 when the request is unknown or already completed and 400
// when the result fails validation. A client error or an invalid result
// releases the claim and broadcasts the request again immediately.
func Complete(requestId, clientId string, res *Result,
	clientErr string) int {

	requestsLock.Lock()
	defer requestsLock.Unlock()

	r := requests[requestId]
	if r == nil || r.completed {
		return 404
	}

	if r.ClaimedBy == "" || r.ClaimedBy != clientId {
		return 409
	}

	if clientErr != "" {
		logrus.WithFields(r.fields()).WithFields(logrus.Fields{
			"error": clientErr,
		}).Warn("tpm: Client request error, releasing request")

		r.LastError = clientErr
		r.release()
		return 200
	}

	err := r.validate(res)
	if err != nil {
		logrus.WithFields(r.fields()).WithFields(logrus.Fields{
			"error": err,
		}).Error("tpm: Invalid client result, releasing request")

		r.LastError = err.Error()
		r.release()
		return 400
	}

	r.result = res
	r.completed = true
	delete(requests, r.Id)
	close(r.done)

	return 200
}
