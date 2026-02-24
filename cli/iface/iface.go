package iface

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/tools/logger"
)

func LoggerFile() (err error) {
	file, err := os.OpenFile(
		"./pritunl-client.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "agent: Failed to create log file"),
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

	model := NewModel()

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
