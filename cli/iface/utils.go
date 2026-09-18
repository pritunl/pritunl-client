package iface

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/platform"
	"github.com/pritunl/tools/logger"
)

func renderCol(width int, format string, args ...interface{}) string {
	data := []rune(fmt.Sprintf(format, args...))
	if width <= 0 {
		return ""
	}
	if len(data) <= width {
		return string(data)
	}
	if width < 4 {
		return string(data[:width])
	}
	return string(data[:width-3]) + "..."
}

// openUrl opens the link in the default browser without waiting for the
// browser to exit.
func openUrl(link string) (err error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", link)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", link)
	default:
		cmd = exec.Command("xdg-open", link)
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "iface: Failed to open null device"),
		}
		return
	}
	defer devNull.Close()

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.SysProcAttr = platform.DetachAttr()

	err = cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "iface: Failed to open browser"),
		}
		return
	}

	go func() {
		_ = cmd.Wait()
	}()

	return
}

// copyText copies the text to the system clipboard, the OSC 52 terminal
// clipboard is used when no clipboard utility is available such as in a
// Flatpak or over SSH.
func copyText(text string) tea.Cmd {
	err := clipboard.WriteAll(text)
	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("iface: Clipboard unavailable using terminal clipboard")
		return tea.SetClipboard(text)
	}
	return nil
}
