package utils

import (
	"crypto/rand"
	"encoding/ascii85"
	"encoding/base64"
	"encoding/hex"
	"math"
	"math/big"
	mathrand "math/rand"
	"regexp"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
)

var (
	randRe = regexp.MustCompile("[^a-zA-Z0-9]+")
)

func RandStr(n int) (str string, err error) {
	for i := 0; i < 10; i++ {
		input, e := RandBytes(int(math.Ceil(float64(n) * 1.25)))
		if e != nil {
			err = e
			return
		}

		output := base64.RawStdEncoding.EncodeToString(input)
		output = randRe.ReplaceAllString(output, "")

		if len(output) < n {
			continue
		}

		str = output[:n]
		break
	}

	if str == "" {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "utils: Random generate error"),
		}
		return
	}

	return
}

func RandStrComplex(n int) (str string, err error) {
	for i := 0; i < 10; i++ {
		input, e := RandBytes(int(math.Ceil(float64(n) * 1.4)))
		if e != nil {
			err = e
			return
		}

		output := make([]byte, ascii85.MaxEncodedLen(len(input)))
		_ = ascii85.Encode(output, input)

		if len(string(output)) < n {
			continue
		}

		str = string(output)[:n]
		break
	}

	if str == "" {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "utils: Random complex generate error"),
		}
		return
	}

	return
}

func RandId() (id string, err error) {
	randByt, err := RandBytes(8)
	if err != nil {
		return
	}

	id = hex.EncodeToString(randByt)
	return
}

func RandBytes(size int) (bytes []byte, err error) {
	bytes = make([]byte, size)
	_, err = rand.Read(bytes)
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "utils: Random read error"),
		}
		return
	}

	return
}

func init() {
	n, err := rand.Int(rand.Reader, big.NewInt(9223372036854775806))
	if err != nil {
		panic(err)
	}

	mathrand.Seed(n.Int64())
}
