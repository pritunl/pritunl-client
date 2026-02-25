package setup

import (
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/pritunl/pritunl-client/service/utils"
)

func TunTapGet(all bool) (adpaters []string, err error) {
	output, err := ExecOutput(
		utils.TunTapPath(),
		utils.TapCtlPath(),
		"list",
	)
	if err != nil {
		return
	}

	adpaters = []string{}
	for _, line := range strings.Split(output, "\n") {
		lines := strings.Fields(line)
		if len(lines) < 2 {
			continue
		}

		name := strings.ToLower(lines[1])

		if !all && name != "ethernet" &&
			name != "local" && name != "pritunl" {

			continue
		}

		adpaters = append(adpaters, lines[0])
	}

	return
}

func TunTapInstall() (err error) {
	cmd := command.Command(
		"pnputil.exe",
		"/add-driver", "ovpn-dco.inf",
		"/install",
	)
	cmd.Dir = utils.TunTapPath()

	err = cmd.Run()
	if err != nil {
		cmd = command.Command(
			"pnputil.exe",
			"/add-driver", "ovpn-dco.inf",
			"/install",
		)
		cmd.Dir = utils.TunTapPath10()

		err = cmd.Run()
		if err != nil {
			err = &errortypes.ExecError{
				errors.Wrap(err, "setup: Driver setup error"),
			}
			return
		}
	}

	return
}

func TunTapClean(all bool) (err error) {
	adapters, err := TunTapGet(all)
	if err != nil {
		return
	}

	for _, adapter := range adapters {
		cmd := command.Command(
			utils.TapCtlPath(),
			"delete",
			adapter,
		)
		cmd.Dir = utils.TunTapPath()

		err = cmd.Run()
		if err != nil {
			err = &errortypes.ExecError{
				errors.Wrap(err, "setup: Driver removal error"),
			}
			return
		}
	}

	return
}
