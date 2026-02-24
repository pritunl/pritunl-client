package watch

import (
	"fmt"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pritunl/pritunl-client/cli/constants"
	"github.com/pritunl/pritunl-client/cli/errortypes"
	"github.com/pritunl/pritunl-client/cli/sprofile"
)

func refresh() {
	for {
		grid := termui.NewGrid()

		termW, termH := termui.TerminalDimensions()
		grid.SetRect(0, 0, termW, termH)

		table := widgets.NewTable()

		sprfls, err := sprofile.GetAll()
		if err != nil {
			termui.Clear()

			para := widgets.NewParagraph()
			para.Title = fmt.Sprintf(
				"Pritunl Client v%s - %s",
				constants.Version,
				time.Now().Format("01/02/2006 15:04:05"),
			)
			para.Text = ""

			grid.Set(termui.NewRow(1, termui.NewCol(1, para)))

			termui.Render(grid)

			time.Sleep(200 * time.Millisecond)

			para.Text = err.Error()
			termui.Render(grid)

			time.Sleep(1 * time.Second)
		} else {
			rows := [][]string{
				[]string{
					"ID",
					"Name",
					"Online For",
					"Server Address",
					"Client Address",
				},
			}

			for _, sprfl := range sprfls {
				if sprfl.Profile != nil {
					rows = append(rows, []string{
						sprfl.Id,
						sprfl.FormatedName(),
						sprfl.Profile.FormatedTime(),
						sprfl.Profile.ServerAddr,
						sprfl.Profile.ClientAddr,
					})
				} else {
					rows = append(rows, []string{
						sprfl.Id,
						sprfl.FormatedName(),
						"Disconnected",
						"-",
						"-",
					})
				}
			}

			table.Title = fmt.Sprintf(
				"Pritunl Client v%s - %s",
				constants.Version,
				time.Now().Format("01/02/2006 15:04:05"),
			)
			table.RowSeparator = true
			table.FillRow = false
			table.TextStyle = termui.NewStyle(termui.ColorWhite)
			table.TextAlignment = termui.AlignCenter
			table.Rows = rows

			grid.Set(termui.NewRow(1, termui.NewCol(1, table)))

			termui.Render(grid)
		}

		time.Sleep(1 * time.Second)
	}
}

func Init() (err error) {
	err = termui.Init()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, ""),
		}
		return
	}
	defer termui.Close()

	go refresh()

	uiEvents := termui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}
