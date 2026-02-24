package token

import (
	"time"

	"github.com/pritunl/pritunl-client/service/utils"
	"github.com/sirupsen/logrus"
)

type Token struct {
	Profile            string    `json:"profile"`
	ServerPublicKey    string    `json:"-"`
	ServerBoxPublicKey string    `json:"-"`
	Token              string    `json:"-"`
	Timestamp          time.Time `json:"timestamp"`
	Ttl                int       `json:"ttl"`
	Valid              bool      `json:"valid"`
}

func (t *Token) Init() (err error) {
	logrus.WithFields(logrus.Fields{
		"ttl":     t.Profile,
		"profile": t.Profile,
	}).Info("token: Token init")

	t.Valid = false

	token, err := utils.RandStrComplex(16)
	if err != nil {
		return
	}

	t.Token = token
	t.Timestamp = time.Now()

	return
}

func (t *Token) Reset() (err error) {
	logrus.WithFields(logrus.Fields{
		"profile": t.Profile,
	}).Info("token: Token reset")

	t.Valid = false

	token, err := utils.RandStrComplex(16)
	if err != nil {
		return
	}

	t.Token = token
	t.Timestamp = time.Now()

	return
}

func (t *Token) Update() (expired bool, err error) {
	if utils.SinceAbs(t.Timestamp) > time.Duration(t.Ttl)*time.Second {
		expired = true

		err = t.Init()
		if err != nil {
			return
		}
	}

	return
}
