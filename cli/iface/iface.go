package iface

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/event"
	"github.com/pritunl/pritunl-client/cli/utils"
	"github.com/pritunl/tools/logger"
)

func LoggerFile() (err error) {
	dataPath := utils.GetDataPath()

	err = os.MkdirAll(dataPath, 0700)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "iface: Failed to create data directory"),
		}
		return
	}

	file, err := os.OpenFile(
		filepath.Join(dataPath, "pritunl.log"),
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
