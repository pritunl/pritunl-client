package sprofile

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/utils"
	"github.com/pritunl/tools/logger"
)

const importMaxSize = 3000000

var (
	clientSecure = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 12 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			},
		},
		Timeout: 12 * time.Second,
	}
	clientInsecure = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 12 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			},
		},
		Timeout: 12 * time.Second,
	}
	ip4reg = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	ip6reg = regexp.MustCompile("/\\[[a-fA-F0-9:]*\\]/")
)

// importer collects profile files and imports the OpenVPN profiles among
// them, referenced certificate and key files are inlined. This matches
// the desktop client importer.
type importer struct {
	files map[string]string
}

func newImporter() *importer {
	return &importer{
		files: map[string]string{},
	}
}

func (i *importer) addData(pth, data string) {
	i.files[pth] = data
}

func (i *importer) addPath(pth string) (err error) {
	data, err := os.ReadFile(pth)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to read profile '%s'", pth),
		}
		return
	}

	i.addData(pth, string(data))

	return
}

func (i *importer) addTar(pth string) (err error) {
	tarFile, err := os.Open(pth)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to open tar '%s'", pth),
		}
		return
	}
	defer tarFile.Close()

	tr := tar.NewReader(tarFile)
	for {
		hdr, e := tr.Next()
		if e != nil {
			if e == io.EOF {
				break
			}

			err = errortypes.ReadError{
				errors.Wrap(e, "sprofile: Failed to read tar header"),
			}
			return
		}

		if hdr.Typeflag == tar.TypeDir {
			continue
		}

		data := bytes.NewBuffer(nil)
		_, err = io.Copy(data, tr)
		if err != nil {
			err = errortypes.ReadError{
				errors.Wrap(err, "sprofile: Failed to read tar data"),
			}
			return
		}

		i.addData(hdr.Name, data.String())
	}

	return
}

// readRef returns the content of a file referenced from a profile, either
// from the collected files or relative to the profile path.
func (i *importer) readRef(pth, ref string) (data string, err error) {
	data, ok := i.files[ref]
	if ok {
		return
	}

	refPath := filepath.Join(filepath.Dir(pth), filepath.Clean(ref))
	dataByt, err := os.ReadFile(refPath)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to read profile file '%s'",
				refPath),
		}
		return
	}

	data = string(dataByt)

	return
}

// parse splits a profile file into the profile configuration, from the
// commented json block, and the OpenVPN data with inline keys.
func (i *importer) parse(pth, data string) (prfl *Sprofile,
	ovpnData string, err error) {

	data = strings.ReplaceAll(data, "\r", "")

	fileName := filepath.Base(pth)
	if ext := filepath.Ext(fileName); ext != "" {
		fileName = strings.TrimSuffix(fileName, ext)
	}

	jsonData := ""
	jsonFound := false
	jsonLoaded := false
	keyData := ""

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
			continue
		}

		var tag string
		var ref string
		switch {
		case strings.HasPrefix(line, "ca "):
			tag = "ca"
			ref = strings.TrimSpace(line[3:])
		case strings.HasPrefix(line, "cert "):
			tag = "cert"
			ref = strings.TrimSpace(line[5:])
		case strings.HasPrefix(line, "key "):
			tag = "key"
			ref = strings.TrimSpace(line[4:])
		case strings.HasPrefix(line, "tls-crypt "):
			tag = "tls-crypt"
			ref = strings.TrimSpace(line[10:])
		case strings.HasPrefix(line, "tls-auth "):
			tag = "tls-auth"
			split := strings.Fields(line[9:])
			if len(split) > 1 {
				last := split[len(split)-1]
				if last == "0" || last == "1" {
					keyData += "key-direction " + last + "\n"
					split = split[:len(split)-1]
				}
			}
			ref = strings.Join(split, " ")
		}

		if tag == "" {
			ovpnData += line + "\n"
			continue
		}

		refData, e := i.readRef(pth, ref)
		if e != nil {
			err = e
			return
		}
		keyData += "<" + tag + ">\n" + refData + "</" + tag + ">\n"
	}

	ovpnData = strings.TrimSpace(ovpnData) + "\n" + keyData

	prfl = &Sprofile{}
	if jsonData != "" {
		e := json.Unmarshal([]byte(jsonData), prfl)
		if e != nil {
			logger.WithFields(logger.Fields{
				"path":  pth,
				"error": e,
			}).Error("sprofile: Failed to parse profile configuration")
			prfl = &Sprofile{}
		}
	}

	if prfl.Name == "" && prfl.User == "" && prfl.Server == "" {
		prfl.Name = fileName
	}

	return
}

func randId() (id string, err error) {
	idByte, err := utils.RandBytes(8)
	if err != nil {
		return
	}
	id = hex.EncodeToString(idByte)
	return
}

// importData stores one parsed profile, an existing profile for the same
// user and server is updated instead of adding a duplicate. Profiles the
// server forces to autostart are converted to system profiles.
func (i *importer) importData(pth, data string, system bool) (err error) {
	prfl, ovpnData, err := i.parse(pth, data)
	if err != nil {
		return
	}

	prfl.Id, err = randId()
	if err != nil {
		return
	}
	prfl.System = system
	prfl.Password = ""

	exists := false
	if prfl.OrganizationId != "" && prfl.ServerId != "" &&
		prfl.UserId != "" {

		sprfls, e := GetAll()
		if e != nil {
			if system {
				err = e
				return
			}
			sprfls, e = getAllUser()
			if e != nil {
				err = e
				return
			}
		}

		for _, curPrfl := range sprfls {
			if prfl.OrganizationId != curPrfl.OrganizationId ||
				prfl.ServerId != curPrfl.ServerId ||
				prfl.UserId != curPrfl.UserId {

				continue
			}

			curPrfl.importConf(prfl)

			if curPrfl.System {
				curPrfl.OvpnData = ovpnData
				err = curPrfl.Commit()
				if err != nil {
					return
				}
			} else {
				err = curPrfl.writeConf()
				if err != nil {
					return
				}
				err = curPrfl.writeData(ovpnData)
				if err != nil {
					return
				}
			}

			prfl = curPrfl
			exists = true
			break
		}
	}

	if !exists {
		if system {
			// Autostart is disabled by default on new system profiles
			// unless enforced by the server
			prfl.Disabled = !prfl.ForceConnect
			prfl.OvpnData = ovpnData
			err = prfl.Commit()
			if err != nil {
				return
			}
		} else {
			err = prfl.writeConf()
			if err != nil {
				return
			}
			err = prfl.writeData(ovpnData)
			if err != nil {
				return
			}
		}
	}

	if prfl.ForceConnect && !prfl.System {
		err = prfl.ConvertSystem()
		if err != nil {
			return
		}
	}

	return
}

// run imports every profile file that was collected.
func (i *importer) run(system bool) (err error) {
	count := 0
	for pth, data := range i.files {
		ext := strings.ToLower(filepath.Ext(pth))
		if ext != ".ovpn" && ext != ".conf" {
			continue
		}

		err = i.importData(pth, data, system)
		if err != nil {
			return
		}
		count += 1
	}

	if count == 0 {
		err = errortypes.ParseError{
			errors.New("sprofile: No profiles found to import"),
		}
		return
	}

	return
}

// Import stores profile data as a system profile when system is true,
// otherwise as a user profile in the user profiles directory.
func Import(data string, system bool) (err error) {
	imptr := newImporter()
	imptr.addData("profile.ovpn", data)
	return imptr.run(system)
}

// ImportPath imports a profile from a URI, a .tar archive of profiles or
// a single .ovpn profile file.
func ImportPath(path string, system bool) (err error) {
	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "pritunl://") ||
		strings.HasPrefix(path, "pritunls://") ||
		strings.HasPrefix(path, "pts://") {

		return ImportUri(path, system)
	}

	info, err := os.Stat(path)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to read profile '%s'", path),
		}
		return
	}
	if info.Size() > importMaxSize {
		err = errortypes.ReadError{
			errors.Newf("sprofile: Profile file too large '%s'", path),
		}
		return
	}

	if strings.HasSuffix(strings.ToLower(path), ".tar") {
		return ImportTar(path, system)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrapf(err, "sprofile: Failed to read profile '%s'", path),
		}
		return
	}

	// Detect tar archives without the extension
	if len(data) > 262 && string(data[257:262]) == "ustar" {
		return ImportTar(path, system)
	}

	imptr := newImporter()
	imptr.addData(path, string(data))

	return imptr.run(system)
}

func ImportTar(filename string, system bool) (err error) {
	imptr := newImporter()

	err = imptr.addTar(filename)
	if err != nil {
		return
	}

	return imptr.run(system)
}

func ImportUri(uri string, system bool) (err error) {
	switch {
	case strings.HasPrefix(uri, "pritunl:"):
		uri = strings.Replace(uri, "pritunl:", "https:", 1)
	case strings.HasPrefix(uri, "pritunls:"):
		uri = strings.Replace(uri, "pritunls:", "https:", 1)
	case strings.HasPrefix(uri, "pts:"):
		uri = strings.Replace(uri, "pts:", "https:", 1)
	case strings.HasPrefix(uri, "http:"):
		uri = strings.Replace(uri, "http:", "https:", 1)
	case strings.HasPrefix(uri, "https:"):
		break
	default:
		uri = "https://" + uri
	}
	uri = strings.Replace(uri, "/k/", "/ku/", 1)

	req, err := http.NewRequest(
		"GET",
		uri,
		nil,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "sprofile: Sync profile request error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl")
	req.Header.Set("Accept", "application/json")

	var client *http.Client
	if len(ip4reg.FindAllString(uri, -1)) > 0 ||
		len(ip6reg.FindAllString(uri, -1)) > 0 {

		client = clientInsecure
	} else {
		client = clientSecure
	}

	resp, err := client.Do(req)
	if err != nil {
		err = errortypes.RequestError{
			errors.Wrap(err, "sprofile: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		err = errortypes.RequestError{
			errors.New("sprofile: Invalid or expired profile uri"),
		}
		return
	}

	if resp.StatusCode != 200 {
		err = errortypes.RequestError{
			errors.Newf(
				"sprofile: Unknown profile uri error %d",
				resp.StatusCode,
			),
		}
		return
	}

	data := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "sprofile: Failed to parse uri response body"),
		}
		return
	}

	if len(data) == 0 {
		err = errortypes.ParseError{
			errors.New("sprofile: No data received from server"),
		}
		return
	}

	for name, prflData := range data {
		imptr := newImporter()
		imptr.addData(name, prflData)

		err = imptr.run(system)
		if err != nil {
			return
		}
	}

	return
}
