package iface

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/pritunl/tools/logger"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 0, 0, 0)

	menuBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3B82F6"))

	menuItemActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Background(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

type TickMsg time.Time

func TickInterval() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type Model struct {
	listDelegate *ListDelegate
	profiles     list.Model
	menu         []MenuItem
	menuCur      int
	bindings     KeyMap
	help         help.Model
	winWidth     int
	winHeight    int
	ready        bool

	showDialog     bool
	dialog         Dialog
	dialogResult   string
	dialogCallback func(returnVal int)
}

func (m Model) Init() tea.Cmd {
	return TickInterval()
}

func (m *Model) ConnectCallback(prompts []sprofile.Prompt,
	callback func(sprofile.PromptValues), err error) {

	if err != nil {
		logger.WithFields(logger.Fields{
			"error": err,
		}).Error("iface: Failed to sync profiles")
		return
	}

	if len(prompts) == 0 {
		return
	}

	if prompts[0].Type == sprofile.PromptLink {
		width := max(len(prompts[0].Value)+4, 60)

		m.showDialog = true
		m.dialog = NewDialog(
			"Single Sign-On Authentication",
			prompts[0].Label+"\n"+prompts[0].Value,
			&OptionButton{
				Label:  "Close",
				Return: 1,
			},
		)
		m.dialog.SetSize(min(m.winWidth-10, width), min(m.winHeight-10, 20))
		m.dialogCallback = nil
		return
	}

	opts := []Option{}
	optsMap := map[string]*OptionText{}

	for _, prompt := range prompts {
		opt := &OptionText{
			Label:       prompt.Label,
			Placeholder: prompt.Placeholder,
			Value:       prompt.Value,
		}
		optsMap[prompt.Key] = opt
		opts = append(opts, opt)
	}

	opts = append(opts,
		&OptionButton{
			Label:  "Cancel",
			Return: 1,
		},
		&OptionButton{
			Label:  "Connect",
			Return: 2,
		},
	)

	m.showDialog = true
	m.dialog = NewDialog(
		"Profile Connect",
		"Authentication Required",
		opts...,
	)
	m.dialog.SetSize(min(m.winWidth-10, 60), min(m.winHeight-10, 20))
	m.dialogCallback = func(returnVal int) {
		if returnVal != 2 {
			return
		}

		values := sprofile.PromptValues{}

		for key, opt := range optsMap {
			values[key] = opt.GetValue()
		}

		callback(values)
	}
}

func (m *Model) Connect(mode string) {
	prflItemInf := m.profiles.SelectedItem()
	prflItem, ok := prflItemInf.(ListItem)
	if !ok {
		return
	}
	prfl := prflItem.Profile()
	if prfl.Active {
		return
	}

	sprofile.StartCallback(prfl.Id, mode, m.ConnectCallback)
}

func (m *Model) Disconnect() {
	prflItemInf := m.profiles.SelectedItem()
	prflItem, ok := prflItemInf.(ListItem)
	if !ok {
		return
	}
	prfl := prflItem.Profile()
	logger.WithFields(logger.Fields{
		"active": prfl.Active,
	}).Info("iface: Disconnect")
	if !prfl.Active {
		return
	}

	err := sprofile.Stop(prfl.Id)
	if err != nil {
		return
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showDialog {
			var dialogCmd tea.Cmd
			m.dialog, dialogCmd = m.dialog.Update(msg)
			return m, dialogCmd
		}

		switch {
		case key.Matches(msg, m.bindings.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.bindings.Connect):
			m.Connect("ovpn")
			return m, nil
		case key.Matches(msg, m.bindings.ConnectWg):
			m.Connect("wg")
			return m, nil
		case key.Matches(msg, m.bindings.Disconnect):
			m.Disconnect()
			return m, nil
		case key.Matches(msg, m.bindings.Import):
			logger.Info("iface: Import")
			return m, nil
		case key.Matches(msg, m.bindings.Logs):
			logger.Info("iface: Logs")
			return m, nil
		case key.Matches(msg, m.bindings.Settings):
			logger.Info("iface: Settings")
			m.showDialog = true
			m.dialog = NewDialog(
				"Settings",
				"Configure application settings",
				&OptionText{
					Label:       "Name",
					Placeholder: "Enter name...",
					Value:       "",
				},
				&OptionText{
					Label:       "Test",
					Placeholder: "Enter test...",
					Value:       "",
				},
				&OptionText{
					Label:       "Server",
					Placeholder: "Enter server...",
					Value:       "",
				},
				&OptionButton{
					Label:  "Cancel",
					Return: 1,
				},
				&OptionButton{
					Label:  "Save",
					Return: 2,
				},
			)
			m.dialog.SetSize(min(m.winWidth-10, 60), min(m.winHeight-10, 20))
			return m, nil
		}
	case tea.WindowSizeMsg:
		marginX, marginY := appStyle.GetFrameSize()

		m.winWidth = msg.Width - marginX
		m.winHeight = msg.Height - marginY
		m.listDelegate.SetSplit(m.winWidth >= 90)
		m.listDelegate.SetWidth(msg.Width - marginX)
		m.profiles.SetSize(msg.Width-marginX, msg.Height-marginY)

		m.ready = true
		return m, nil

	case DialogCloseMsg:
		m.showDialog = false

		if m.dialogCallback != nil {
			m.dialogCallback(msg.Return)
		}

		switch msg.Return {
		case 2:
			m.dialogResult = "Settings saved"
			logger.Info("iface: Settings saved")
		default:
			m.dialogResult = "Settings cancelled"
			logger.Info("iface: Settings cancelled")
		}

		return m, nil
	case TickMsg:
		err := m.Sync()
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("iface: Failed to sync profiles")
		}

		return m, TickInterval()
	}

	profiles, cmd := m.profiles.Update(msg)
	m.profiles = profiles

	return m, cmd
}

func (m Model) renderMenu() string {
	var menuItems []string

	menu := []MenuItem{}

	curProfileInf := m.profiles.SelectedItem()
	if curProfile, ok := curProfileInf.(ListItem); ok {
		if m.listDelegate.split {
			if curProfile.profile.Active {
				menu = append(menu, MenuItem{
					Title: "Disconnect",
					Key:   "d",
				})
			} else {
				menu = append(menu, MenuItem{
					Title: "Connect OpenVPN",
					Key:   "c",
				})
				if curProfile.Profile().Wg {
					menu = append(menu, MenuItem{
						Title: "Connect WireGuard",
						Key:   "w",
					})
				}
			}
		} else {
			if curProfile.profile.Active {
				menu = append(menu, MenuItem{
					Title: "Disconnect",
					Key:   "d",
				})
			} else {
				menu = append(menu, MenuItem{
					Title: "Connect OVPN",
					Key:   "c",
				})
				if curProfile.Profile().Wg {
					menu = append(menu, MenuItem{
						Title: "Connect WG",
						Key:   "w",
					})
				}
			}
		}

	}

	menu = append(menu, []MenuItem{
		{Title: "Import", Key: "i"},
		{Title: "Logs", Key: "l"},
		{Title: "Settings", Key: "s"},
	}...)

	for i, item := range menu {
		menuText := fmt.Sprintf("%s (%s)", item.Title, item.Key)
		if i == m.menuCur {
			menuItems = append(menuItems,
				menuItemActiveStyle.Render(menuText))
		} else {
			menuItems = append(menuItems,
				menuItemStyle.Render(menuText))
		}
	}

	return menuBarStyle.Width(m.winWidth).Render(
		lipgloss.JoinHorizontal(lipgloss.Left, menuItems...))
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	listView := m.profiles.View()

	parts := strings.SplitN(listView, "\n", 2)
	if len(parts) == 2 {
		listView = menuBarStyle.Width(m.winWidth).Render(
			m.profiles.Title) + "\n" + parts[1]
	}

	mainView := appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			listView,
			m.renderMenu(),
		),
	)

	if m.showDialog {
		return lipgloss.Place(
			m.winWidth,
			m.winHeight,
			lipgloss.Center,
			lipgloss.Center,
			m.dialog.View(),
		)
	}

	return mainView
}

func (m *Model) Sync() (err error) {
	items := []list.Item{}

	sprfls, err := sprofile.GetAll()
	if err != nil {
		return
	}

	for _, sprfl := range sprfls {
		statusLabel, status := sprfl.FormatedStatus()

		if sprfl.Profile != nil {
			items = append(items, ListItem{
				profile: Profile{
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
					Connected:       sprfl.Profile.ClientAddr != "",
					Uptime:          sprfl.Profile.Uptime(),
					StatusLabel:     statusLabel,
					Status:          status,
					ServerAddress:   sprfl.Profile.ServerAddr,
					ClientAddress:   sprfl.Profile.ClientAddr,
				},
			})
		} else {
			items = append(items, ListItem{
				profile: Profile{
					Id:              sprfl.Id,
					Name:            sprfl.FormatedName(),
					User:            sprfl.User,
					Organization:    sprfl.Organization,
					Server:          sprfl.Server,
					Wg:              sprfl.Wg,
					Active:          false,
					State:           sprfl.FormatedState(),
					RunState:        sprfl.FormatedRunState(),
					RegistrationKey: sprfl.RegistrationKey,
					Uptime:          0,
					StatusLabel:     statusLabel,
					Status:          status,
					ServerAddress:   "",
					ClientAddress:   "",
				},
			})
		}
	}

	m.profiles.SetItems(items)

	return
}

func NewModel() Model {
	delegate := &ListDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	delegate.SetSpacing(0)

	lst := list.New([]list.Item{}, delegate, 0, 0)
	lst.Title = "Pritunl Client - Profiles"
	lst.SetShowHelp(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowStatusBar(false)

	model := Model{
		listDelegate: delegate,
		profiles:     lst,
		menuCur:      -1,
		bindings:     bindings,
		help:         help.New(),
		ready:        false,
		showDialog:   false,
		dialog:       Dialog{},
		dialogResult: "",
	}

	_ = model.Sync()

	return model
}
