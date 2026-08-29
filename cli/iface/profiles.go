package iface

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/pritunl/pritunl-client/cli/sprofile"
)

type Profile struct {
	Id              string
	Name            string
	User            string
	Organization    string
	Server          string
	Wg              bool
	Active          bool
	State           string
	RunState        string
	RegistrationKey string
	Connected       bool
	Uptime          int64
	StatusLabel     string
	Status          string
	ServerAddress   string
	ClientAddress   string
}

var (
	itemStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#361da3")).
			Padding(0, 1).
			MarginLeft(1)
	itemSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#a1cdff")).
				Padding(0, 1).
				MarginLeft(1)
	itemColStyle = lipgloss.NewStyle().
			Align(lipgloss.Left)
	itemTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4a8cf7")).
			Bold(true)
	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))
	redStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))
	yellowSytle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fffb00"))
)

type ListItem struct {
	sprfl   *sprofile.Sprofile
	profile Profile
}

func NewListItem(sprfl *sprofile.Sprofile) ListItem {
	statusLabel, status := sprfl.FormatedStatus()

	prfl := Profile{
		Id:              sprfl.Id,
		Name:            sprfl.FormatedName(),
		User:            sprfl.User,
		Organization:    sprfl.Organization,
		Server:          sprfl.Server,
		Wg:              sprfl.Wg,
		Active:          sprfl.State,
		State:           sprfl.FormatedState(),
		RunState:        sprfl.FormatedRunState(),
		RegistrationKey: sprfl.RegistrationKey,
		StatusLabel:     statusLabel,
		Status:          status,
	}

	if sprfl.Profile != nil {
		prfl.Connected = sprfl.Profile.ClientAddr != ""
		prfl.Uptime = sprfl.Profile.Uptime()
		prfl.ServerAddress = sprfl.Profile.ServerAddr
		prfl.ClientAddress = sprfl.Profile.ClientAddr
	}

	return ListItem{
		sprfl:   sprfl,
		profile: prfl,
	}
}

func (i ListItem) Profile() Profile {
	return i.profile
}

func (i ListItem) Sprofile() *sprofile.Sprofile {
	return i.sprfl
}

func (i ListItem) FilterValue() string {
	return i.profile.Name
}

func (i ListItem) Title() string {
	return itemTitleStyle.Render(i.profile.Name)
}

func (i ListItem) Body(width int) string {
	rows := []string{}

	colWidth := width - 6
	style := itemColStyle.Width(colWidth)

	row := style.Render(renderCol(colWidth, "User: %s", i.profile.User))
	rows = append(rows, row)

	row = style.Render(i.statusRow(colWidth))
	rows = append(rows, row)

	row = style.Render(renderCol(colWidth, "Server: %s", i.profile.Server))
	rows = append(rows, row)
	row = renderCol(colWidth, "Organization: %s", i.profile.Organization)
	rows = append(rows, row)

	serverAddr := i.profile.ServerAddress
	if serverAddr == "" {
		serverAddr = "-"
	}
	clientAddr := i.profile.ClientAddress
	if clientAddr == "" {
		clientAddr = "-"
	}

	row = style.Render(renderCol(
		colWidth,
		"Server Address: %s",
		serverAddr,
	))
	rows = append(rows, row)
	row = style.Render(renderCol(
		colWidth,
		"Client Address: %s",
		clientAddr,
	))
	rows = append(rows, row)

	return strings.Join(rows, "\n")
}

func (i ListItem) BodySplit(width int) string {
	rows := []string{}

	colWidth := min((width-6)/2, 60)
	style := itemColStyle.Width(colWidth)

	leftWidth := colWidth - 1

	left := style.Render(renderCol(leftWidth, "User: %s", i.profile.User))

	right := style.Render(i.statusRow(colWidth))

	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, left, right))

	left = style.Render(renderCol(leftWidth, "Server: %s", i.profile.Server))
	right = renderCol(colWidth, "Organization: %s", i.profile.Organization)
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, left, right))

	serverAddr := i.profile.ServerAddress
	if serverAddr == "" {
		serverAddr = "-"
	}
	clientAddr := i.profile.ClientAddress
	if clientAddr == "" {
		clientAddr = "-"
	}

	left = style.Render(renderCol(
		leftWidth,
		"Server Address: %s",
		serverAddr,
	))
	right = style.Render(renderCol(
		colWidth,
		"Client Address: %s",
		clientAddr,
	))
	rows = append(rows, lipgloss.JoinHorizontal(
		lipgloss.Left, left, right))

	return strings.Join(rows, "\n")
}

type ListDelegate struct {
	list.DefaultDelegate
	width int
	split bool
}

func (d *ListDelegate) SetWidth(w int) {
	d.width = w
}

func (d ListDelegate) Height() int {
	if d.split {
		return 6
	}
	return 9
}

func (d *ListDelegate) SetSplit(x bool) {
	d.split = x
}

func (d ListDelegate) Render(w io.Writer, model list.Model,
	index int, item list.Item) {

	listItem, ok := item.(ListItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == model.Index() {
		style = itemSelectedStyle
	} else {
		style = itemStyle
	}

	var body string
	if d.split {
		body = listItem.BodySplit(d.width)
	} else {
		body = listItem.Body(d.width)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		listItem.Title(),
		body,
	)

	fmt.Fprint(w, style.Render(content))
}

func (i ListItem) statusRow(colWidth int) string {
	if i.profile.RegistrationKey != "" && !i.profile.Connected {
		label := "Registration Required"
		return fmt.Sprintf(
			"%s: %s",
			yellowSytle.Render(label),
			yellowSytle.Bold(true).Render(renderCol(
				colWidth-len(label)-2, "%s", i.profile.RegistrationKey)),
		)
	}

	return fmt.Sprintf(
		"%s: %s",
		i.profile.StatusLabel,
		i.statusStyle().Render(renderCol(
			colWidth-len(i.profile.StatusLabel)-2, "%s", i.profile.Status)),
	)
}

func (i ListItem) statusStyle() lipgloss.Style {
	if !i.profile.Active {
		return redStyle
	}
	if i.profile.Connected && i.profile.StatusLabel == "Online For" {
		return greenStyle
	}
	return yellowSytle
}
