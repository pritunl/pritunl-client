package utils

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/sirupsen/logrus"
)

var (
	clientTransport = &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
	}
	client4 = &http.Client{
		Transport: clientTransport,
		Timeout:   12 * time.Second,
	}
	client6 = &http.Client{
		Transport: clientTransport,
		Timeout:   6 * time.Second,
	}
)

type ipResp struct {
	Ip string `json:"ip"`
}

type NopCloser struct {
	io.Reader
}

func (NopCloser) Close() error {
	return nil
}

var httpErrCodes = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
}

func StripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}

func FormatHostPort(hostname string, port int) string {
	if strings.Contains(hostname, ":") && !strings.Contains(hostname, "[") {
		hostname = "[" + hostname + "]"
	}
	return fmt.Sprintf("%s:%d", hostname, port)
}

func GetStatusMessage(code int) string {
	return fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func AbortWithStatus(c *gin.Context, code int) {
	r := render.String{
		Format: GetStatusMessage(code),
	}

	c.Status(code)
	r.WriteContentType(c.Writer)
	c.Writer.WriteHeaderNow()
	r.Render(c.Writer)
	c.Abort()
}

func AbortWithError(c *gin.Context, code int, err error) {
	AbortWithStatus(c, code)
	c.Error(err)
}

func WriteStatus(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, GetStatusMessage(code))
}

func WriteText(w http.ResponseWriter, code int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, text)
}

func WriteUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(401)
	fmt.Fprintln(w, "401 "+msg)
}

func CloneHeader(src http.Header) (dst http.Header) {
	dst = make(http.Header, len(src))
	for k, vv := range src {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		dst[k] = vv2
	}
	return dst
}

func GetLocation(r *http.Request) string {
	host := ""

	switch {
	case r.Header.Get("X-Host") != "":
		host = r.Header.Get("X-Host")
		break
	case r.Host != "":
		host = r.Host
		break
	case r.URL.Host != "":
		host = r.URL.Host
		break
	}

	return "https://" + host
}

func GetPublicAddress4() (addr4 string, err error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "app4.pritunl.com",
		Path:   "/ip",
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		u.String(),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "utils: Request get error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-client")

	res, err := client4.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "utils: Request get error"),
		}
		return
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != 200 {
		err = LogRequestError(res, "")
		return
	}

	data := &ipResp{}
	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "utils: Failed to parse app4 response body"),
		}
		return
	}

	addr4 = data.Ip

	return
}

func GetPublicAddress6() (addr6 string, err error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "app6.pritunl.com",
		Path:   "/ip",
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		u.String(),
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "utils: Request get error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-client")

	res, err := client6.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "utils: Request get error"),
		}
		return
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != 200 {
		err = LogRequestError(res, "")
		return
	}

	data := &ipResp{}
	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "utils: Failed to parse app6 response body"),
		}
		return
	}

	addr6 = data.Ip

	return
}

func ParseRequestError(res *http.Response, message string) (
	fields logrus.Fields, err error) {

	resBody := ""

	data, err := io.ReadAll(res.Body)
	if err == nil {
		resBody = string(data)
	}

	if message == "" {
		message = "request: Bad status from server"
	}

	err = &errortypes.RequestError{
		errors.Newf("request: Bad status %d code from server",
			res.StatusCode),
	}

	fields = logrus.Fields{
		"host":        res.Request.Host,
		"status_code": res.StatusCode,
		"body":        resBody,
		"error":       err,
	}

	return
}

func LogRequestError(res *http.Response, message string) (err error) {
	resBody := ""

	data, err := io.ReadAll(res.Body)
	if err == nil {
		resBody = string(data)
	}

	if message == "" {
		message = "request: Bad status from server"
	}

	err = &errortypes.RequestError{
		errors.Newf("request: Bad status %d code from server",
			res.StatusCode),
	}

	logrus.WithFields(logrus.Fields{
		"url":         res.Request.URL.String(),
		"method":      res.Request.Method,
		"status_code": res.StatusCode,
		"body":        resBody,
		"error":       err,
	}).Error(message)

	return
}
