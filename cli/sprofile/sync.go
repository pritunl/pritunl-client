package sprofile

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/utils"
	"github.com/pritunl/tools/logger"
)

var (
	syncClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			},
		},
		Timeout: 5 * time.Second,
	}
)

type syncData struct {
	Signature string `json:"signature"`
	Conf      string `json:"conf"`
}

// Sync updates a user profile from the profile sync hosts before
// connecting, matching the desktop client. System profiles are synced by
// the service. Failures are recorded in the sync time and returned.
func (s *Sprofile) Sync() (err error) {
	if s.System || len(s.SyncHosts) == 0 {
		return
	}

	hosts := make([]string, len(s.SyncHosts))
	copy(hosts, s.SyncHosts)
	rand.Shuffle(len(hosts), func(i, j int) {
		hosts[i], hosts[j] = hosts[j], hosts[i]
	})

	conf := ""
	var syncErr error
	for _, host := range hosts {
		if host == "" {
			continue
		}

		conf, syncErr = s.syncRequest(host)
		if syncErr == nil {
			break
		}
	}

	if syncErr != nil {
		logger.WithFields(logger.Fields{
			"profile_id": s.Id,
			"error":      syncErr,
		}).Error("sprofile: Failed to sync profile")

		s.SyncTime = -1
		s.writeConf()
		err = syncErr
		return
	}

	if conf == "" {
		return
	}

	err = s.importSync(conf)
	if err != nil {
		logger.WithFields(logger.Fields{
			"profile_id": s.Id,
			"error":      err,
		}).Error("sprofile: Failed to import profile sync")

		s.SyncTime = -1
		s.writeConf()
		return
	}

	return
}

// syncRequest fetches the signed profile configuration from a sync host,
// an empty configuration is returned when no update is available.
func (s *Sprofile) syncRequest(host string) (conf string, err error) {
	pth := fmt.Sprintf(
		"/key/sync/%s/%s/%s/%s",
		s.OrganizationId,
		s.UserId,
		s.ServerId,
		s.SyncHash,
	)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	nonce, err := utils.RandStr(32)
	if err != nil {
		return
	}

	authStr := strings.Join([]string{
		s.SyncToken,
		timestamp,
		nonce,
		"GET",
		pth,
	}, "&")

	hashFunc := hmac.New(sha512.New, []byte(s.SyncSecret))
	hashFunc.Write([]byte(authStr))
	sig := base64.StdEncoding.EncodeToString(hashFunc.Sum(nil))

	reqUrl := strings.TrimRight(host, "/") + pth + "?ver=2"

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Sync request error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl")
	req.Header.Set("Auth-Token", s.SyncToken)
	req.Header.Set("Auth-Timestamp", timestamp)
	req.Header.Set("Auth-Nonce", nonce)
	req.Header.Set("Auth-Signature", sig)

	resp, err := syncClient.Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Sync request failed"),
		}
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		break
	case 480:
		logger.WithFields(logger.Fields{
			"profile_id": s.Id,
		}).Info("sprofile: Skipping profile sync, requires subscription")
		return
	case 404:
		err = errortypes.RequestError{
			errors.New("sprofile: Failed to sync profile, user not found"),
		}
		return
	case 401:
		err = errortypes.RequestError{
			errors.New(
				"sprofile: Failed to sync profile, authentication failed"),
		}
		return
	default:
		err = errortypes.RequestError{
			errors.Newf("sprofile: Failed to sync profile, status: %d",
				resp.StatusCode),
		}
		return
	}

	data := &syncData{}
	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to parse sync response"),
		}
		return
	}

	if data.Signature == "" || data.Conf == "" {
		return
	}

	hashFuncConf := hmac.New(sha512.New, []byte(s.SyncSecret))
	hashFuncConf.Write([]byte(data.Conf))
	confSig := base64.StdEncoding.EncodeToString(hashFuncConf.Sum(nil))

	if subtle.ConstantTimeCompare(
		[]byte(confSig), []byte(data.Signature)) != 1 {

		err = errortypes.ParseError{
			errors.New("sprofile: Failed to sync profile, signature invalid"),
		}
		return
	}

	conf = data.Conf

	return
}

// SyncApply applies a signed sync body sent by the service after a user
// profile connects, the signature is verified with the profile secret.
func (s *Sprofile) SyncApply(body string) (err error) {
	if s.System {
		return
	}

	data := &syncData{}
	err = json.Unmarshal([]byte(body), data)
	if err != nil {
		err = errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to parse gateway sync body"),
		}
		return
	}

	if data.Signature == "" || data.Conf == "" {
		return
	}

	hashFunc := hmac.New(sha512.New, []byte(s.SyncSecret))
	hashFunc.Write([]byte(data.Conf))
	confSig := base64.StdEncoding.EncodeToString(hashFunc.Sum(nil))

	if subtle.ConstantTimeCompare(
		[]byte(confSig), []byte(data.Signature)) != 1 {

		err = errortypes.ParseError{
			errors.New("sprofile: Gateway sync signature invalid"),
		}
		return
	}

	err = s.importSync(data.Conf)
	if err != nil {
		s.SyncTime = -1
		s.writeConf()
		return
	}

	return
}

// importSync applies a synced configuration keeping the device identity
// and keys of the current profile data, matching the desktop client.
func (s *Sprofile) importSync(data string) (err error) {
	curData, err := s.ReadData()
	if err != nil {
		return
	}

	uvId := ""
	uvName := ""
	for _, line := range strings.Split(curData, "\n") {
		if strings.HasPrefix(line, "setenv UV_ID ") {
			uvId = line
		} else if strings.HasPrefix(line, "setenv UV_NAME ") {
			uvName = line
		}
	}

	jsonData := ""
	jsonFound := false
	jsonLoaded := false
	newData := ""

	for _, line := range strings.Split(data, "\n") {
		if !jsonLoaded && !jsonFound && line == "#{" {
			jsonFound = true
			jsonLoaded = true
		}

		if jsonFound && strings.HasPrefix(line, "#") {
			if line == "#}" {
				jsonFound = false
			}
			jsonData += strings.Replace(line, "#", "", 1)
		} else {
			if strings.HasPrefix(line, "setenv UV_ID ") && uvId != "" {
				line = uvId
			} else if strings.HasPrefix(line, "setenv UV_NAME ") &&
				uvName != "" {

				line = uvName
			}

			newData += line + "\n"
		}
	}

	if jsonData != "" {
		conf := &userConf{}
		e := json.Unmarshal([]byte(jsonData), conf)
		if e == nil {
			s.SyncTime = time.Now().Unix()
			s.upsertConf(conf)

			err = s.writeConf()
			if err != nil {
				return
			}
		}
	}

	tlsAuth := ""
	if strings.Contains(curData, "key-direction") &&
		!strings.Contains(newData, "key-direction") {

		tlsAuth += "key-direction 1\n"
	}

	block := extractBlock(curData, "<tls-auth>", "</tls-auth>")
	if block != "" {
		tlsAuth += block + "\n"
	}

	block = extractBlock(curData, "<tls-crypt>", "</tls-crypt>")
	if block != "" {
		tlsAuth += block + "\n"
	}

	cert := ""
	block = extractBlock(newData, "<cert>", "</cert>")
	if block != "" {
		cert = block + "\n"
	}
	if cert == "" {
		block = extractBlock(curData, "<cert>", "</cert>")
		if block != "" {
			cert = block + "\n"
		}
	}

	key := ""
	block = extractBlock(curData, "<key>", "</key>")
	if block != "" {
		key = block + "\n"
	}

	err = s.writeData(newData + tlsAuth + cert + key)
	if err != nil {
		return
	}

	return
}
