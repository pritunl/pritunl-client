package connection

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/constants"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/event"
	"github.com/pritunl/pritunl-client/service/sprofile"
	"github.com/pritunl/pritunl-client/service/tpm"
	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
)

const (
	GlobalTimeoutDirect  = 60 * time.Second
	GlobalTimeoutPreAuth = 180 * time.Second
)

var (
	clientTransport = &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 8 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
	}
	clientInsecure = &http.Client{
		Transport: clientTransport,
		Timeout:   40 * time.Second,
	}
)

type ReqBox struct {
	DeviceId       string   `json:"device_id"`
	DeviceName     string   `json:"device_name"`
	DeviceKey      string   `json:"device_key"`
	DeviceHostname string   `json:"device_hostname"`
	Platform       string   `json:"platform"`
	Version        string   `json:"client_ver"`
	MacAddr        string   `json:"mac_addr"`
	MacAddrs       []string `json:"mac_addrs"`
	Token          string   `json:"token"`
	Nonce          string   `json:"nonce"`
	Password       string   `json:"password"`
	Timestamp      int64    `json:"timestamp"`
	WgPublicKey    string   `json:"wg_public_key"`
	PublicAddress  string   `json:"public_address"`
	PublicAddress6 string   `json:"public_address6"`
	SsoToken       string   `json:"sso_token"`
}

type RespBox struct {
	Mode      string `json:"mode"`
	SsoToken  string `json:"sso_token"`
	SsoUrl    string `json:"sso_url"`
	Data      string `json:"data"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

type EncryptedKeyBox struct {
	Data            string `json:"data"`
	Nonce           string `json:"nonce"`
	PublicKey       string `json:"public_key"`
	Signature       string `json:"signature"`
	DeviceSignature string `json:"device_signature"`
}

type EncryptedRequestData struct {
	Body      []byte
	Token     string
	Timestamp string
	Nonce     string
	Signature string
}

type Cipher struct {
	serverPubKey  *[32]byte
	senderPubKey  *[32]byte
	senderPrivKey *[32]byte
}

type ConnData struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason"`
	RegKey string `json:"reg_key"`

	// ovpn
	Token   string `json:"token"`
	Remote  string `json:"remote"`
	Remote6 string `json:"remote6"`

	// wg
	Configuration *WgConf `json:"configuration"`
}

type PingData struct {
	Status    bool `json:"status"`
	Timestamp int  `json:"timestamp"`
}

type SsoEventData struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type Client struct {
	conn              *Connection
	prov              Provider
	requestCtxLock    sync.Mutex
	requestCtx        *utils.CancelContext
	disconnectLock    sync.Mutex
	disconnect        bool
	disconnected      bool
	disconnectWaiters []chan bool
	startTime         time.Time
}

func (c *Client) Fields() logrus.Fields {
	return logrus.Fields{
		"client_disconnect":         c.disconnect,
		"client_disconnected":       c.disconnected,
		"client_disconnect_waiters": len(c.disconnectWaiters),
		"client_provider":           c.prov != nil,
		"client_startime":           utils.SinceFormatted(c.startTime),
	}
}

func (c *Client) Start(prov Provider) (err error) {
	c.prov = prov
	c.startTime = time.Now()

	GlobalStore.UnsetAuthConnect(c.conn.Id)

	err = c.prov.PreConnect()
	if err != nil {
		c.conn.State.Close()
		return
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	c.conn.Data.UpdateEvent()

	err = c.conn.Data.ParseProfile()
	if err != nil {
		c.conn.State.Close()
		return
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	if c.conn.Profile.Mode == WgMode ||
		c.conn.Profile.DynamicFirewall ||
		c.conn.Profile.SsoAuth ||
		c.conn.Profile.DeviceAuth {

		err = c.connectPreAuth()
		if err != nil {
			c.conn.State.Close()
			return
		}
	} else {
		err = c.connectDirect()
		if err != nil {
			c.conn.State.Close()
			return
		}
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(c.conn.Fields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				})).Error("profile: Watch connection panic")
			}
		}()

		e := c.prov.WatchConnection()
		if e != nil {
			logrus.WithFields(c.conn.Fields(logrus.Fields{
				"error": e,
			})).Error("profile: Watch connection error")
			c.conn.State.Close()
			return
		}
	}()

	return
}

func (c *Client) globalTimeout(timeout time.Duration) {
	for i := 0; i < int(timeout.Seconds()); i++ {
		time.Sleep(1 * time.Second)
		if c.conn.Data.Status == Connected ||
			c.conn.State.IsStop() {

			break
		}
	}

	if c.conn.State.IsStop() {
		return
	}

	if c.conn.Data.Status != Connected {
		logrus.WithFields(c.conn.Fields(logrus.Fields{
			"global_timeout": timeout.Seconds(),
		})).Error("profile: Global connection timeout")

		c.Disconnect()
	}
}

func (c *Client) connectDirect() (err error) {
	go c.globalTimeout(GlobalTimeoutDirect)

	logrus.WithFields(c.conn.Fields(logrus.Fields{
		"remotes": c.conn.Data.Remotes.GetFormatted(),
	})).Info("connection: Attempting remotes")

	err = c.prov.Connect(&ConnData{})
	if err != nil {
		c.conn.State.Close()
		return
	}

	return
}

func (c *Client) connectPreAuth() (err error) {
	var evt *event.Event
	connErrors := []ConnectionError{}
	final := false
	var data *ConnData

	go c.globalTimeout(GlobalTimeoutPreAuth)

	for _, remote := range c.conn.Data.Remotes {
		logrus.WithFields(c.conn.Fields(logrus.Fields{
			"remote": remote.GetFormatted(),
		})).Info("connection: Attempting remote")

		if c.conn.State.IsStop() {
			c.conn.State.Close()
			return
		}

		data, final, evt, err = c.authorize(remote.Host, "", time.Time{})
		if err != nil {
			connErrors = append(connErrors, ConnectionError{
				Host:  remote.Host,
				Error: err,
			})
			err = nil
			if final {
				break
			}
		} else {
			break
		}

		if c.conn.State.IsStop() {
			c.conn.State.Close()
			return
		}
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	if len(connErrors) > 0 && data == nil {
		if evt != nil {
			evt.Init()
		} else {
			c.conn.Data.SendProfileEvent("connection_error")
		}

		for _, connErr := range connErrors {
			logrus.WithFields(logrus.Fields{
				"host":  connErr.Host,
				"error": connErr.Error,
			}).Error("profile: Connection request failed")
		}

		logrus.Error("profile: All connection requests failed, " +
			"view previous errors for each attempt")

		c.conn.State.Close()

		err = connErrors[len(connErrors)-1].Error
		return
	}

	if data == nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Connection data empty"),
		}
		return
	}

	if !data.Allow {
		logrus.WithFields(c.conn.Fields(logrus.Fields{
			"allow":   data.Allow,
			"reason":  data.Reason,
			"remote":  data.Remote,
			"remote6": data.Remote6,
		})).Info("connection: Authorization failed")

		c.conn.Data.ResetAuthToken()

		if c.conn.State.IsStop() {
			c.conn.State.Close()
			return
		}

		if data.RegKey != "" {
			logrus.WithFields(c.conn.Fields(logrus.Fields{
				"reason": data.Reason,
			})).Error("profile: Device registration required")

			c.conn.Data.RegistrationKey = data.RegKey
			c.conn.State.NoReconnect("client_device_registration")

			if c.conn.Profile.SystemProfile {
				sprofile.Deactivate(c.conn.Profile.Id)

				c.conn.Profile.RegistrationKey = data.RegKey
				sprfl := sprofile.Get(c.conn.Profile.Id)
				if sprfl != nil {
					sprfl.State = false
					sprfl.RegistrationKey = c.conn.Profile.RegistrationKey
					err = sprfl.Commit()
					if err != nil {
						return
					}
				} else {
					logrus.WithFields(c.conn.Fields(nil)).Error(
						"profile: Failed to find system profile")
				}
			}

			c.conn.Data.SendProfileEvent("registration_required")
		} else {
			logrus.WithFields(c.conn.Fields(logrus.Fields{
				"reason": data.Reason,
			})).Error("profile: Failed to authenticate")

			c.conn.State.NoReconnect("client_auth_error")
			c.conn.Data.SendProfileEvent("auth_error")

			if c.conn.Profile.SystemProfile {
				logrus.WithFields(c.conn.Fields(nil)).Error(
					"profile: Stopping system " +
						"profile due to authentication errors")

				sprofile.Deactivate(c.conn.Profile.Id)
				sprofile.SetAuthErrorCount(c.conn.Profile.Id, 0)
			}

			time.Sleep(3 * time.Second)
		}

		c.conn.State.Close()
		return
	} else {
		logrus.WithFields(c.conn.Fields(logrus.Fields{
			"allow":   data.Allow,
			"reason":  data.Reason,
			"remote":  data.Remote,
			"remote6": data.Remote6,
		})).Info("connection: Authorization successful")

		c.conn.Data.RegistrationKey = ""
		if c.conn.Profile.SystemProfile &&
			c.conn.Profile.RegistrationKey != "" {

			c.conn.Profile.RegistrationKey = data.RegKey
			sprfl := sprofile.Get(c.conn.Profile.Id)
			if sprfl != nil {
				sprfl.RegistrationKey = ""
				err = sprfl.Commit()
				if err != nil {
					return
				}
			} else {
				logrus.WithFields(c.conn.Fields(nil)).Error(
					"profile: Failed to find system profile")
			}
		} else {
			c.conn.Data.SendProfileEvent("registration_pass")
		}
	}

	err = c.prov.Connect(data)
	if err != nil {
		c.conn.State.Close()
		return
	}

	return
}

func (c *Client) GetUrl(scheme, host, handle string) *url.URL {
	reqPath := fmt.Sprintf(
		"/key/%s/%s/%s/%s",
		handle,
		c.conn.Profile.OrgId,
		c.conn.Profile.UserId,
		c.conn.Profile.ServerId,
	)

	return &url.URL{
		Scheme: scheme,
		Host:   ParseAddress(host),
		Path:   reqPath,
	}
}

func (c *Client) InitBox() (ciph *Cipher, reqBx *ReqBox, err error) {
	var serverPubKey [32]byte
	serverPubKeySlic, err := base64.StdEncoding.DecodeString(
		c.conn.Profile.ServerBoxPublicKey)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to decode server box key"),
		}
		return
	}
	copy(serverPubKey[:], serverPubKeySlic)

	senderPubKey, senderPrivKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Failed to generate nacl key"),
		}
		return
	}

	ciph = &Cipher{
		serverPubKey:  &serverPubKey,
		senderPubKey:  senderPubKey,
		senderPrivKey: senderPrivKey,
	}

	macAddrs, err := c.conn.Data.GetMacAddrs()
	if err != nil {
		return
	}

	pltfrm := ""
	switch runtime.GOOS {
	case "linux":
		pltfrm = "linux"
		break
	case "windows":
		pltfrm = "win"
		break
	case "darwin":
		pltfrm = "mac"
		break
	default:
		pltfrm = "unknown"
		break
	}

	reqBx = &ReqBox{
		DeviceId:       c.conn.Data.DeviceId,
		DeviceName:     c.conn.Data.DeviceName,
		DeviceHostname: c.conn.Data.Hostname,
		Platform:       pltfrm,
		Version:        constants.Version,
		MacAddr:        c.conn.Data.MacAddr,
		MacAddrs:       macAddrs,
		Timestamp:      time.Now().Unix(),
		PublicAddress:  c.conn.Data.PublicAddr,
		PublicAddress6: c.conn.Data.PublicAddr6,
		WgPublicKey:    c.prov.GetPublicKey(),
	}

	return
}

func (c *Client) authorize(host string, ssoToken string,
	ssoStart time.Time) (data *ConnData, final bool,
	evt *event.Event, err error) {

	tokn, err := c.conn.Data.GetAuthToken()
	if err != nil {
		return
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	ciph, reqBx, err := c.InitBox()
	if err != nil {
		return
	}

	reqBx.Password = c.conn.Profile.Password
	reqBx.Token = tokn.Token
	reqBx.Nonce = tokn.Nonce
	reqBx.SsoToken = ssoToken

	handle := ""
	if ssoToken != "" || (c.conn.Profile.SsoAuth && tokn.Validated) {
		handle = c.prov.GetReqPrefix() + "_wait"
	} else {
		handle = c.prov.GetReqPrefix()
	}
	reqUrl := c.GetUrl("https", host, handle)

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	ctx := c.GetContext()
	defer ctx.Cancel()

	res, err := c.EncRequest(ctx, "POST", reqUrl, ciph, reqBx)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 428 && ssoToken != "" {
		if time.Since(ssoStart) > SingleSignOnTimeout {
			evt = &event.Event{
				Type: "timeout_error",
				Data: c.conn.Data,
			}

			err = &errortypes.RequestError{
				errors.Newf("connection: Single sign-on timeout"),
			}
			return
		}

		data, _, evt, err = c.authorize(host, ssoToken, ssoStart)
		if err != nil {
			return
		}

		if c.conn.State.IsStop() {
			c.conn.State.Close()
			return
		}

		final = true
		return
	}

	if res.StatusCode == 429 {
		evt = &event.Event{
			Type: "offline_error",
			Data: c.conn.Data,
		}

		err = &errortypes.RequestError{
			errors.New("profile: Server is offline"),
		}
		return
	}

	if res.StatusCode != 200 {
		err = utils.LogRequestError(
			res, "connection: Failed to complete authorize")
		return
	}

	respBx := &RespBox{}
	err = json.NewDecoder(res.Body).Decode(respBx)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to parse response body"),
		}
		return
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	if respBx.SsoUrl != "" && respBx.SsoToken != "" && ssoToken == "" {
		if !c.conn.State.IsInteractive() {
			logrus.WithFields(c.conn.Fields(nil)).Info(
				"connection: Stopping non-interactive single sign-on")

			GlobalStore.SetAuthConnect(c.conn.Id)
			evt2 := &event.Event{
				Type: "wakeup",
			}
			evt2.Init()

			c.conn.State.NoReconnect("client_auth_error")
			c.conn.Data.SendProfileEvent("sso_interactive")

			if c.conn.Profile.SystemProfile {
				logrus.WithFields(c.conn.Fields(nil)).Error(
					"profile: Stopping system " +
						"profile due to non-interactive single sign-on")

				sprofile.Deactivate(c.conn.Profile.Id)
				sprofile.SetAuthErrorCount(c.conn.Profile.Id, 0)
			}

			c.conn.State.Close()
			return
		}

		evt2 := &event.Event{
			Type: "sso_auth",
			Data: &SsoEventData{
				Id:  c.conn.Profile.Id,
				Url: respBx.SsoUrl,
			},
		}
		evt2.Init()

		if c.conn.Profile.SystemProfile {
			c.conn.Data.SsoUrl = respBx.SsoUrl
		}

		c.conn.Data.Status = "authenticating"
		c.conn.Data.UpdateEvent()

		data, _, evt, err = c.authorize(
			host, respBx.SsoToken, time.Now())
		if err != nil {
			return
		}

		if c.conn.State.IsStop() {
			c.conn.State.Close()
			return
		}

		if c.conn.Profile.SystemProfile {
			c.conn.Data.SsoUrl = ""
		}

		final = true
		return
	} else if ssoToken != "" {
		c.conn.Data.Status = "connecting"
		c.conn.Data.UpdateEvent()
	}

	if c.conn.State.IsStop() {
		c.conn.State.Close()
		return
	}

	data = &ConnData{}
	err = c.DecryptRespBox(ciph, respBx, data)
	if err != nil {
		return
	}

	return
}

func (c *Client) encryptReqBox(method string, reqPath string,
	ciph *Cipher, reqBx *ReqBox) (encReqData *EncryptedRequestData,
	err error) {

	if c.conn.Profile.ServerBoxPublicKey == "" {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Server box public key not set"),
		}
		return
	}

	var tp tpm.TpmCaller
	if runtime.GOOS == "darwin" && !config.Config.ForceLocalTpm {
		tp = &tpm.Remote{}
	} else {
		tp = &tpm.Tpm{}
	}

	if c.conn.Profile.DeviceAuth && method == "POST" {
		err = tp.Open(config.Config.EnclavePrivateKey)
		if err != nil {
			return
		}
		defer tp.Close()

		deviceKey, e := tp.PublicKey()
		if e != nil {
			err = e
			return
		}

		reqBx.DeviceKey = deviceKey
	}

	boxData, err := json.Marshal(reqBx)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to marshal wg key box"),
		}
		return
	}

	var nonce [24]byte
	nonceSl := make([]byte, 24)
	_, err = rand.Read(nonceSl)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "profile: Failed to generate nacl nonce"),
		}
		return
	}
	copy(nonce[:], nonceSl)

	encrypted := box.Seal([]byte{}, boxData,
		&nonce, ciph.serverPubKey, ciph.senderPrivKey)

	nonce64 := base64.StdEncoding.EncodeToString(nonceSl)
	ciphertext64 := base64.StdEncoding.EncodeToString(encrypted)
	senderPubKey64 := base64.StdEncoding.EncodeToString(ciph.senderPubKey[:])

	encBox := &EncryptedKeyBox{
		Data:      ciphertext64,
		Nonce:     nonce64,
		PublicKey: senderPubKey64,
	}

	userPrivKeyBlock, _ := pem.Decode([]byte(c.conn.Data.PrivateKey))
	if userPrivKeyBlock == nil {
		err = &errortypes.ParseError{
			errors.New("profile: Failed to decode private key"),
		}
		return
	}

	userPrivKey, err := x509.ParsePKCS1PrivateKey(userPrivKeyBlock.Bytes)
	if err != nil {
		userPrivKeyInf, e := x509.ParsePKCS8PrivateKey(
			userPrivKeyBlock.Bytes)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "profile: Failed to parse private key"),
			}
			return
		}

		userPrivKey = userPrivKeyInf.(*rsa.PrivateKey)
	}

	reqHash := sha512.Sum512([]byte(strings.Join([]string{
		encBox.Data,
		encBox.Nonce,
		encBox.PublicKey,
	}, "&")))

	rsaSig, err := rsa.SignPSS(
		rand.Reader,
		userPrivKey,
		crypto.SHA512,
		reqHash[:],
		&rsa.PSSOptions{
			SaltLength: 0,
			Hash:       crypto.SHA512,
		},
	)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to rsa sign data"),
		}
		return
	}

	encBox.Signature = base64.StdEncoding.EncodeToString(rsaSig)

	if c.conn.Profile.DeviceAuth && method == "POST" {
		privKey64 := ""
		privKey64, encBox.DeviceSignature, err = tp.Sign(reqHash[:])
		if err != nil {
			return
		}

		if privKey64 != "" {
			config.Config.EnclavePrivateKey = privKey64

			err = config.Save()
			if err != nil {
				return
			}
		}
	}

	encData, err := json.Marshal(encBox)
	if err != nil {
		return
	}

	encReqData = &EncryptedRequestData{
		Body: encData,
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	authNonce, err := utils.RandStr(32)
	if err != nil {
		return
	}

	authData := []string{
		c.conn.Profile.SyncToken,
		timestamp,
		authNonce,
		method,
		reqPath,
		encBox.Data,
		encBox.Nonce,
		encBox.PublicKey,
		encBox.Signature,
	}

	if encBox.DeviceSignature != "" {
		authData = append(authData, encBox.DeviceSignature)
	}

	authStr := strings.Join(authData, "&")

	hashFunc := hmac.New(sha512.New, []byte(c.conn.Profile.SyncSecret))
	hashFunc.Write([]byte(authStr))
	rawSignature := hashFunc.Sum(nil)
	sig := base64.StdEncoding.EncodeToString(rawSignature)

	encReqData.Token = c.conn.Profile.SyncToken
	encReqData.Timestamp = timestamp
	encReqData.Nonce = authNonce
	encReqData.Signature = sig

	return
}

func (c *Client) DecryptRespBox(ciph *Cipher, respBx *RespBox,
	dest interface{}) (err error) {

	respHashFunc := hmac.New(sha512.New, []byte(c.conn.Profile.SyncSecret))
	respHashFunc.Write([]byte(respBx.Data + "&" + respBx.Nonce))
	respRawSignature := respHashFunc.Sum(nil)
	respSig := base64.StdEncoding.EncodeToString(respRawSignature)

	if subtle.ConstantTimeCompare([]byte(respSig),
		[]byte(respBx.Signature)) != 1 {

		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Response signature invalid"),
		}
		return
	}

	respCiphertext, err := base64.StdEncoding.DecodeString(respBx.Data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to parse response data"),
		}
		return
	}

	var respNonce [24]byte
	respNonceSl, err := base64.StdEncoding.DecodeString(respBx.Nonce)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "profile: Failed to parse response nonce"),
		}
		return
	}
	copy(respNonce[:], respNonceSl)

	respPlaintext, ok := box.Open([]byte{}, respCiphertext,
		&respNonce, ciph.serverPubKey, ciph.senderPrivKey)

	if !ok {
		err = &errortypes.ParseError{
			errors.New("profile: Failed to decrypt response"),
		}
		return
	}

	err = json.Unmarshal(respPlaintext, dest)
	if err != nil {
		err = &errortypes.ParseError{
			errors.New("profile: Failed to parse response"),
		}
		return
	}

	return
}

func (c *Client) GetContext() (ctx *utils.CancelContext) {
	ctx = utils.NewCancelContext()
	ctx.OnCancel(func() {
		c.requestCtxLock.Lock()
		if c.requestCtx == ctx {
			c.requestCtx = nil
		}
		c.requestCtxLock.Unlock()
	})

	c.requestCtxLock.Lock()
	c.requestCtx = ctx
	c.requestCtxLock.Unlock()

	return
}

func (c *Client) CancelRequest() {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(c.conn.Fields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			})).Error("profile: Cancel request panic")
		}
	}()

	c.requestCtxLock.Lock()
	ctx := c.requestCtx
	c.requestCtxLock.Unlock()

	if ctx != nil {
		ctx.Cancel()
	}
}

func (c *Client) EncRequest(ctx context.Context, method string,
	reqUrl *url.URL, ciph *Cipher, reqBx *ReqBox) (
	resp *http.Response, err error) {

	encReqData, err := c.encryptReqBox(method, reqUrl.Path, ciph, reqBx)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		reqUrl.String(),
		bytes.NewBuffer(encReqData.Body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "profile: Request put error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-client")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", encReqData.Token)
	req.Header.Set("Auth-Timestamp", encReqData.Timestamp)
	req.Header.Set("Auth-Nonce", encReqData.Nonce)
	req.Header.Set("Auth-Signature", encReqData.Signature)

	resp, err = clientInsecure.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "profile: Request put error"),
		}
		return
	}

	return
}

func (c *Client) Disconnect() {
	c.disconnectLock.Lock()
	if c.disconnected {
		c.disconnectLock.Unlock()
		return
	}
	if c.disconnect {
		waiter := make(chan bool, 8)
		c.disconnectWaiters = append(c.disconnectWaiters, waiter)
		c.disconnectLock.Unlock()
		<-waiter
		return
	}
	c.disconnect = true
	c.disconnectLock.Unlock()

	c.conn.State.SetStop()

	logrus.WithFields(c.conn.Fields(nil)).Error(
		"connection: Disconnecting")

	c.conn.Data.Status = "disconnecting"
	c.conn.Data.UpdateEvent()

	c.CancelRequest()

	delay := 5*time.Second - utils.SinceAbs(c.startTime)
	if delay > 0 && delay <= 5*time.Second {
		time.Sleep(1*time.Second + delay)
	} else {
		time.Sleep(1 * time.Second)
	}

	if c.prov != nil {
		c.prov.Disconnect()
	}

	time.Sleep(1 * time.Second)

	if runtime.GOOS == "darwin" && !config.Config.DisableWgDns {
		err := utils.ClearScutilDns(c.conn.Id)
		if err != nil {
			logrus.WithFields(c.conn.Fields(logrus.Fields{
				"error": err,
			})).Error("profile: Failed to clear scutil DNS")
		}
	}

	c.conn.State.RemovePaths()

	c.conn.Data.Status = "disconnected"
	c.conn.Data.Clear()
	c.conn.Data.UpdateEvent()

	c.disconnectLock.Lock()
	c.disconnected = true
	if c.disconnectWaiters != nil {
		for _, waiter := range c.disconnectWaiters {
			waiter <- true
		}
	}
	c.disconnectWaiters = nil
	c.disconnectLock.Unlock()

	return
}

func (c *Client) Disconnected() {
	if c.conn.State.IsReconnect() {
		logrus.WithFields(c.conn.Fields(nil)).Info(
			"profile: Disconnected with restart")
		go c.conn.Restart()
	} else {
		logrus.WithFields(c.conn.Fields(nil)).Info(
			"profile: Disconnected without restart")
	}
}

type Provider interface {
	GetPublicKey() string
	GetReqPrefix() string
	PreConnect() (err error)
	Connect(data *ConnData) (err error)
	WatchConnection() (err error)
	Disconnect()
}
