package setup

import (
	"fmt"
	"os"

	"github.com/pritunl/pritunl-client/service/command"
)

func Uninstall() {
	cmd := command.Command("sc.exe", "stop", "pritunl")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	cmd = command.Command("sc.exe", "delete", "pritunl")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	err := TunTapClean(false)
	if err != nil {
		fmt.Println(err.Error())
	}
}
