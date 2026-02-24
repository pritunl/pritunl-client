package setup

import (
	"os"
	"path/filepath"

	"github.com/pritunl/pritunl-client/service/command"
)

func RootDir() string {
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return rootDir
}

func ExecOutput(dir, name string, arg ...string) (output string, err error) {
	cmd := command.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	outputByt, err := cmd.Output()
	if err != nil {
		return
	}
	output = string(outputByt)

	return
}
