package utils

import (
	"bytes"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/command"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/sirupsen/logrus"
)

func Exec(name string, arg ...string) (err error) {
	cmd := command.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	return
}

func ExecInput(dir, input, name string, arg ...string) (err error) {
	cmd := command.Command(name, arg...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err,
				"utils: Failed to get stdin in exec '%s'", name),
		}
		return
	}

	if dir != "" {
		cmd.Dir = dir
	}

	err = cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	var wrErr error
	go func() {
		defer func() {
			wrErr = stdin.Close()
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to close stdin in exec '%s'",
						name,
					),
				}
			}
		}()

		_, wrErr = io.WriteString(stdin, input)
		if wrErr != nil {
			wrErr = &errortypes.ExecError{
				errors.Wrapf(
					wrErr,
					"utils: Failed to write stdin in exec '%s'",
					name,
				),
			}
			return
		}
	}()

	err = cmd.Wait()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	if wrErr != nil {
		return
	}

	return
}

func ExecInputOutput(input, name string, arg ...string) (
	output string, err error) {

	cmd := command.Command(name, arg...)

	stdout := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to get stdin in exec '%s'", name),
		}
		return
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	var wrErr error
	go func() {
		defer func() {
			wrErr = stdin.Close()
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to close stdin in exec '%s'",
						name,
					),
				}
			}
		}()

		_, wrErr = io.WriteString(stdin, input)
		if wrErr != nil {
			wrErr = &errortypes.ExecError{
				errors.Wrapf(
					wrErr,
					"utils: Failed to write stdin in exec '%s'",
					name,
				),
			}
			return
		}
	}()

	err = cmd.Wait()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	if wrErr != nil {
		return
	}

	output = string(stdout.Bytes())

	return
}

func ExecInputOutputCombindLogged(input, name string, arg ...string) (
	output string, err error) {

	cmd := command.Command(name, arg...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to get stdin in exec '%s'", name),
		}
		return
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	var wrErr error
	go func() {
		defer func() {
			wrErr = stdin.Close()
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to close stdin in exec '%s'",
						name,
					),
				}
			}
		}()

		_, wrErr = io.WriteString(stdin, input)
		if wrErr != nil {
			wrErr = &errortypes.ExecError{
				errors.Wrapf(
					wrErr,
					"utils: Failed to write stdin in exec '%s'",
					name,
				),
			}
			return
		}
	}()

	err = cmd.Wait()

	output = stdout.String()
	errOutput := stderr.String()

	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output":       output,
			"error_output": errOutput,
			"cmd":          name,
			"arg":          arg,
			"error":        err,
		}).Error("utils: Process exec error")

		return
	}

	if wrErr != nil {
		logrus.WithFields(logrus.Fields{
			"output":       output,
			"error_output": errOutput,
			"cmd":          name,
			"arg":          arg,
			"error":        wrErr,
		}).Error("utils: Process exec error")

		return
	}

	output = string(stdout.Bytes())

	return
}

func ExecOutput(name string, arg ...string) (output string, err error) {
	cmd := command.Command(name, arg...)
	cmd.Stderr = os.Stderr

	outputByt, err := cmd.Output()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}
	output = string(outputByt)

	return
}

func ExecOutputLogged(ignores []string, name string, arg ...string) (
	output string, err error) {

	cmd := command.Command(name, arg...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	output = stdout.String()
	errOutput := stderr.String()

	if err != nil && ignores != nil {
		for _, ignore := range ignores {
			if strings.Contains(output, ignore) ||
				strings.Contains(errOutput, ignore) {

				err = nil
				break
			}
		}
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output":       output,
			"error_output": errOutput,
			"cmd":          name,
			"arg":          arg,
			"error":        err,
		}).Error("utils: Process exec error")
		return
	}

	return
}

func ExecCombinedOutput(name string, arg ...string) (
	output string, err error) {

	cmd := command.Command(name, arg...)

	outputByt, err := cmd.CombinedOutput()
	if outputByt != nil {
		output = string(outputByt)
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	return
}

func ExecCombinedOutputLogged(ignores []string, name string, arg ...string) (
	output string, err error) {

	cmd := command.Command(name, arg...)

	outputByt, err := cmd.CombinedOutput()
	if outputByt != nil {
		output = string(outputByt)
	}

	if err != nil && ignores != nil {
		for _, ignore := range ignores {
			if strings.Contains(output, ignore) {
				err = nil
				output = ""
				break
			}
		}
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output": output,
			"cmd":    name,
			"arg":    arg,
			"error":  err,
		}).Error("utils: Process exec error")
		return
	}

	return
}

func ExecWaitTimeout(proc *os.Process, timeout time.Duration) {
	defer func() {
		panc := recover()
		if panc != nil {
			logrus.WithFields(logrus.Fields{
				"trace": string(debug.Stack()),
				"panic": panc,
			}).Error("utils: Panic")
		}
	}()

	waiter := make(chan bool, 2)

	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("utils: Panic")
			}
		}()
		proc.Wait()
		waiter <- true
	}()
	go func() {
		defer func() {
			panc := recover()
			if panc != nil {
				logrus.WithFields(logrus.Fields{
					"trace": string(debug.Stack()),
					"panic": panc,
				}).Error("utils: Panic")
			}
		}()
		time.Sleep(timeout)
		go func() {
			defer func() {
				panc := recover()
				if panc != nil {
					logrus.WithFields(logrus.Fields{
						"trace": string(debug.Stack()),
						"panic": panc,
					}).Error("utils: Panic")
				}
			}()
			proc.Kill()
			proc.Kill()
			proc.Kill()
		}()
		time.Sleep(1 * time.Second)
		waiter <- true
	}()

	<-waiter
}
