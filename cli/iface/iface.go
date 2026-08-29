package iface

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/event"
	"github.com/pritunl/tools/logger"
)

func LoggerFile() (err error) {
	file, err := os.OpenFile(
		filepath.Join(os.TempDir(), "pritunl-client-cli.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "iface: Failed to create log file"),
		}
		return
	}

	logger.Init(
		logger.SetMaxLimit(2*time.Hour),
		logger.SetIcons(true),
	)

	logger.AddHandler(func(record *logger.Record) {
		file.WriteString(record.String())
		file.Sync()
	})

	return
}

func Iface() (err error) {
	err = LoggerFile()
	if err != nil {
		return
	}

	listener := event.NewListener()
	defer listener.Close()

	model := NewModel(listener)

	prog := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		//tea.WithMouseCellMotion(),
	)

	_, err = prog.Run()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "iface: Program run error"),
		}
		return
	}

	return
}
