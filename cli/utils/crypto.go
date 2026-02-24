package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"regexp"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
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

func RandMacAddr() (addr string, err error) {
	bytes := make([]byte, 6)
	_, err = rand.Read(bytes)
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "utils: Random read error"),
		}
		return
	}

	addr = strings.ToUpper(fmt.Sprintf("%x", bytes))
	return
}

func GenerateRsaKey() (encodedPriv, encodedPub []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to generate rsa key"),
		}
		return
	}

	blockPriv := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	encodedPriv = pem.EncodeToMemory(blockPriv)

	bytesPub, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "utils: Failed to marshal rsa public key"),
		}
		return
	}

	blockPub := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesPub,
	}
	encodedPub = pem.EncodeToMemory(blockPub)

	return
}

func init() {
	n, err := rand.Int(rand.Reader, big.NewInt(9223372036854775806))
	if err != nil {
		panic(err)
	}

	mathrand.Seed(n.Int64())
}
