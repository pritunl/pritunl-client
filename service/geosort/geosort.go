package geosort

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/sirupsen/logrus"
)

var (
	clientTransport = &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // TODO
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
	}
	client = &http.Client{
		Transport: clientTransport,
		Timeout:   10 * time.Second,
	}
)

type GeoSort struct {
	License               string   `json:"license"`
	SourceAddress         string   `json:"source_address"`
	SourceAddress6        string   `json:"source_address6"`
	DestinationAddresses  []string `json:"destination_addresses"`
	DestinationAddresses6 []string `json:"destination_addresses6"`
}

func (g *GeoSort) Sort() (err error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "app.pritunl.com",
		Path:   "/geosort",
	}

	reqData, err := json.Marshal(g)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "geosort: Failed to marshal data"),
		}
		return
	}

	req, err := http.NewRequest(
		"GET",
		u.String(),
		bytes.NewBuffer(reqData),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "geosort: Request put error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-client")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "geosort: Request put error"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = &errortypes.RequestError{
			errors.Newf("geosort: Bad request status %d", resp.StatusCode),
		}
		return
	}

	respData := &GeoSort{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "geosort: Failed to parse response body"),
		}
		return
	}

	origAddr := set.NewSet()
	for _, destAddr := range g.DestinationAddresses {
		origAddr.Add(destAddr)
	}

	destAddrs := []string{}
	newAddr := set.NewSet()
	for _, destAddr := range respData.DestinationAddresses {
		if !origAddr.Contains(destAddr) {
			continue
		}
		newAddr.Add(destAddr)
		destAddrs = append(destAddrs, destAddr)
	}

	origAddr.Subtract(newAddr)
	for destAddrInf := range origAddr.Iter() {
		destAddrs = append(destAddrs, destAddrInf.(string))
	}

	g.DestinationAddresses = destAddrs

	origAddr6 := set.NewSet()
	for _, destAddr6 := range g.DestinationAddresses6 {
		origAddr6.Add(destAddr6)
	}

	destAddrs6 := []string{}
	newAddr6 := set.NewSet()
	for _, destAddr6 := range respData.DestinationAddresses6 {
		if !origAddr6.Contains(destAddr6) {
			continue
		}
		newAddr6.Add(destAddr6)
		destAddrs6 = append(destAddrs6, destAddr6)
	}

	origAddr6.Subtract(newAddr6)
	for destAddrInf6 := range origAddr6.Iter() {
		destAddrs6 = append(destAddrs6, destAddrInf6.(string))
	}

	g.DestinationAddresses6 = destAddrs6

	return
}

func SortRemotes(addr4, addr6 string, remotes []string, license string) (
	newRemotes []string) {

	if addr4 == "" && addr6 == "" {
		newRemotes = remotes
		return
	}

	remotesSet := set.NewSet()
	destAddrs := []string{}
	destAddrsSet := set.NewSet()
	destAddrs6 := []string{}
	destAddrsSet6 := set.NewSet()
	for _, remote := range remotes {
		if remotesSet.Contains(remote) {
			continue
		}
		remotesSet.Add(remote)

		ip := net.ParseIP(remote)
		if ip == nil {
			logrus.WithFields(logrus.Fields{
				"remote": remote,
			}).Warn("profile: Remote not IP address")
			continue
		}

		ipStr := ip.String()
		if ip.To4() == nil {
			if !destAddrsSet6.Contains(ipStr) {
				destAddrsSet6.Add(ipStr)
				destAddrs6 = append(destAddrs6, ipStr)
			}
		} else {
			if !destAddrsSet.Contains(ipStr) {
				destAddrsSet.Add(ipStr)
				destAddrs = append(destAddrs, ipStr)
			}
		}
	}

	geo := &GeoSort{
		License:               license,
		SourceAddress:         addr4,
		SourceAddress6:        addr6,
		DestinationAddresses:  destAddrs,
		DestinationAddresses6: destAddrs6,
	}

	err := geo.Sort()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"remotes": remotes,
			"error":   err,
		}).Error("geosort: Geo sort failed")

		newRemotes = remotes
		return
	}

	newRemotes = geo.DestinationAddresses
	newRemotes = append(newRemotes, geo.DestinationAddresses6...)

	return
}
