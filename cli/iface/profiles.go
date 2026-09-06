package iface

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pritunl/pritunl-client/cli/sprofile"
)

const (
	itemFrameHeight = 4
	itemContentLeft = 3
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
	itemTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))
	redStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))
	yellowSytle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fffb00"))

	itemButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3B82F6")).
			Padding(0, 1).
			MarginRight(1)
	itemButtonDangerStyle = itemButtonStyle.
				Background(lipgloss.Color("#EF4444"))
	itemButtonMutedStyle = itemButtonStyle.
				Background(lipgloss.Color("#4B5563"))
)

// itemButton is a rendered profile card button and its column position
// relative to the card content area.
type itemButton struct {
	item  MenuItem
	text  string
	x     int
	width int
}

type ListItem struct {
	sprfl   *sprofile.Sprofile
	profile Profile
}

func NewListItem(sprfl *sprofile.Sprofile) ListItem {
	statusLabel, status := sprfl.FormatedStatus()

	prfl := Profile{
		Id:              sprfl.Id,
		Name:            sprfl.FormatedNameShort(),
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
	return itemTitleStyle.Render(i.profile.Name) +
		itemTypeStyle.Render("  "+i.sprfl.FormatedType())
}

// Buttons returns the card action buttons, the connect and disconnect
// buttons follow the same rules as the menu bar.
func (i ListItem) Buttons() []itemButton {
	items := append(profileActions(i.sprfl),
		MenuItem{Title: "Settings", Key: "s"})

	buttons := []itemButton{}
	x := 0
	for _, item := range items {
		style := itemButtonStyle
		switch item.Key {
		case "d":
			style = itemButtonDangerStyle
		case "s":
			style = itemButtonMutedStyle
		}

		text := style.Render(item.Title)
		width := lipgloss.Width(text)
		buttons = append(buttons, itemButton{
			item:  item,
			text:  text,
			x:     x,
			width: width - style.GetMarginRight(),
		})
		x += width
	}

	return buttons
}

// ButtonsView renders the card button row.
func (i ListItem) ButtonsView() string {
	texts := []string{}
	for _, btn := range i.Buttons() {
		texts = append(texts, btn.text)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, texts...)
}

// ButtonAt returns the key press for the button at the column relative
// to the card content area.
func (i ListItem) ButtonAt(x int) (tea.KeyMsg, bool) {
	for _, btn := range i.Buttons() {
		if x >= btn.x && x < btn.x+btn.width {
			return btn.item.KeyMsg()
		}
	}
	return tea.KeyMsg{}, false
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

	available := min(width-6, 160)
	rightWidth := available / 2
	leftColWidth := available - rightWidth
	leftStyle := itemColStyle.Width(leftColWidth)
	rightStyle := itemColStyle.Width(rightWidth)
	leftWidth := leftColWidth - 1

	left := leftStyle.Render(renderCol(leftWidth, "User: %s", i.profile.User))
	right := rightStyle.Render(i.statusRow(rightWidth))
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, left, right))

	left = leftStyle.Render(renderCol(
		leftWidth, "Server: %s", i.profile.Server))
	right = rightStyle.Render(renderCol(
		rightWidth, "Organization: %s", i.profile.Organization))
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, left, right))

	serverAddr := i.profile.ServerAddress
	if serverAddr == "" {
		serverAddr = "-"
	}
	clientAddr := i.profile.ClientAddress
	if clientAddr == "" {
		clientAddr = "-"
	}

	left = leftStyle.Render(renderCol(
		leftWidth,
		"Server Address: %s",
		serverAddr,
	))
	right = rightStyle.Render(renderCol(
		rightWidth,
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
		return itemFrameHeight + 3
	}
	return itemFrameHeight + 6
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
		listItem.ButtonsView(),
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
