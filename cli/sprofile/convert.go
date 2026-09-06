package sprofile

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
)

// ConvertSystem moves a user profile to the service as a system profile.
// The profile is disconnected first and the user profile files are
// removed once the service has stored the profile.
func (s *Sprofile) ConvertSystem() (err error) {
	if s.System {
		return
	}

	_ = s.Disconnect()

	data, err := s.ReadData()
	if err != nil {
		return
	}

	sprfl := *s
	sprfl.System = true
	sprfl.OvpnData = data
	sprfl.Password = ""
	sprfl.Profile = nil

	// Autostart is disabled by default on new system profiles unless
	// enforced by the server
	sprfl.Disabled = !s.ForceConnect

	err = serviceCall("PUT", "/sprofile", &sprfl)
	if err != nil {
		return
	}

	_ = serviceCall("DELETE", "/log/"+s.Id, nil)

	err = s.removeUser()
	if err != nil {
		return
	}

	s.System = true
	s.OvpnData = data

	return
}

// ConvertUser moves a system profile to the user profiles directory. The
// profile is disconnected and removed from the service first.
func (s *Sprofile) ConvertUser() (err error) {
	if !s.System {
		return
	}

	if s.ForceConnect {
		err = errortypes.ParseError{
			errors.New("sprofile: Autostart enforced by server"),
		}
		return
	}

	_ = s.Disconnect()

	err = serviceCall("DELETE", "/sprofile/"+s.Id, nil)
	if err != nil {
		return
	}

	s.System = false
	s.State = false

	err = s.writeConf()
	if err != nil {
		return
	}

	err = s.writeData(s.OvpnData)
	if err != nil {
		return
	}

	return
}
