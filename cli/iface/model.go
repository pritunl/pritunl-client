package iface

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pritunl/pritunl-client/cli/constants"
	"github.com/pritunl/pritunl-client/cli/event"
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/pritunl/tools/logger"
)

const (
	statusTimeout      = 6 * time.Second
	statusErrorTimeout = 12 * time.Second
	splitWidth         = 90
)

var (
	appStyle = lipgloss.NewStyle()

	statusInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Padding(1, 2)
)

type TickMsg time.Time

func TickInterval() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// SyncMsg carries the result of a background profile fetch.
type SyncMsg struct {
	Profiles sprofile.Sprofiles
	Err      error
}

// ActionDoneMsg reports completion of a background service request.
type ActionDoneMsg struct {
	Action  string
	Message string
	Err     error
}

func syncCmd() tea.Cmd {
	return func() tea.Msg {
		sprfls, err := sprofile.GetAll()
		return SyncMsg{
			Profiles: sprfls,
			Err:      err,
		}
	}
}

func actionCmd(action, message string, fn func() error) tea.Cmd {
	return func() tea.Msg {
		return ActionDoneMsg{
			Action:  action,
			Message: message,
			Err:     fn(),
		}
	}
}

// EventMsg carries a service event from the websocket listener.
type EventMsg struct {
	Event *event.Event
}

func waitEventCmd(listener *event.Listener) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-listener.Events()
		if !ok {
			return nil
		}
		return EventMsg{
			Event: evt,
		}
	}
}

// dialogCallback runs on the current model when a dialog closes and may
// return a command to execute.
type dialogCallback func(m *Model, ret int) tea.Cmd

type Model struct {
	listener     *event.Listener
	listDelegate *ListDelegate
	profiles     list.Model
	bindings     KeyMap
	winWidth     int
	winHeight    int
	ready        bool
	syncing      bool
	syncErr      bool
	syncedOnce   bool

	statusMsg  string
	statusErr  bool
	statusTime time.Time
	eventsUp   bool

	showDialog     bool
	dialog         Dialog
	dialogCallback dialogCallback

	showLogs bool
	logs     LogsView

	// Profiles connected from this session, used to surface single
	// sign-on links and device registration keys once they appear.
	watching map[string]bool
	ssoShown map[string]string
	regShown map[string]string
}

func NewModel(listener *event.Listener) Model {
	delegate := &ListDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	delegate.SetSpacing(0)

	lst := list.New([]list.Item{}, delegate, 0, 0)
	lst.Title = "Pritunl Client - Profiles"
	lst.SetShowHelp(false)
	lst.SetFilteringEnabled(false)
	lst.SetShowStatusBar(false)
	lst.DisableQuitKeybindings()

	return Model{
		listener:     listener,
		listDelegate: delegate,
		profiles:     lst,
		bindings:     bindings,
		watching:     map[string]bool{},
		ssoShown:     map[string]string{},
		regShown:     map[string]string{},
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{syncCmd(), TickInterval()}
	if m.listener != nil {
		cmds = append(cmds, waitEventCmd(m.listener))
	}
	return tea.Batch(cmds...)
}

func (m *Model) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusErr = isErr
	m.statusTime = time.Now()
}

// profileName returns the display name for a profile id in the list.
func (m *Model) profileName(id string) string {
	for _, itemInf := range m.profiles.Items() {
		item, ok := itemInf.(ListItem)
		if ok && item.sprfl.Id == id {
			return item.sprfl.FormatedName()
		}
	}
	return ""
}

func (m *Model) selectedProfile() *sprofile.Sprofile {
	item, ok := m.profiles.SelectedItem().(ListItem)
	if !ok {
		return nil
	}
	return item.Sprofile()
}

func (m *Model) openDialog(d Dialog, callback dialogCallback) {
	d.SetSize(min(m.winWidth-4, 70), m.winHeight)
	m.dialog = d
	m.dialogCallback = callback
	m.showDialog = true
}

func (m *Model) openMessage(title, message string) {
	m.openDialog(NewDialog(
		title,
		message,
		&OptionButton{
			Label:  "Close",
			Return: DialogOk,
		},
	), nil)
}

func (m *Model) openError(action string, err error) {
	logger.WithFields(logger.Fields{
		"action": action,
		"error":  err,
	}).Error("iface: Action failed")

	m.openMessage(action+" Failed", errorMessage(err))
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	// Drop the wrapped stack trace from dropbox errors
	if i := strings.Index(msg, "\n"); i != -1 {
		msg = msg[:i]
	}

	return msg
}

func (m *Model) Connect(mode string) tea.Cmd {
	sprfl := m.selectedProfile()
	if sprfl == nil {
		return nil
	}
	if sprfl.State {
		m.setStatus("Profile already active", true)
		return nil
	}

	if mode == "wg" && !sprfl.Wg {
		m.setStatus("WireGuard not available for profile", true)
		return nil
	}
	if mode == "ovpn" && sprfl.HideOvpn {
		mode = "wg"
	}
	mode = sprfl.ResolveMode(mode)

	prompts := sprofile.PasswordPrompts(sprfl)
	if len(prompts) == 0 && sprfl.PreConnectMsg == "" {
		return m.connectCmd(sprfl, mode, "")
	}

	opts := []Option{}
	optsMap := map[string]*OptionText{}

	for _, prompt := range prompts {
		opt := &OptionText{
			Label:       prompt.Label,
			Placeholder: prompt.Placeholder,
			Value:       prompt.Value,
			Password:    true,
		}
		optsMap[prompt.Key] = opt
		opts = append(opts, opt)
	}

	opts = append(opts,
		&OptionButton{
			Label:  "Cancel",
			Return: DialogCancel,
		},
		&OptionButton{
			Label:  "Connect",
			Return: DialogOk,
		},
	)

	message := sprfl.PreConnectMsg
	if len(prompts) > 0 {
		if message != "" {
			message += "\n\n"
		}
		message += "Authentication required"
	}

	m.openDialog(NewDialog(
		"Connect "+sprfl.FormatedName(),
		message,
		opts...,
	), func(m *Model, ret int) tea.Cmd {
		if ret != DialogOk {
			return nil
		}

		values := sprofile.PromptValues{}
		for key, opt := range optsMap {
			values[key] = opt.GetValue()
		}

		return m.connectCmd(sprfl, mode, sprofile.BuildPassword(values))
	})

	return nil
}

func (m *Model) connectCmd(sprfl *sprofile.Sprofile, mode,
	password string) tea.Cmd {

	m.watching[sprfl.Id] = true
	m.setStatus("Connecting "+sprfl.FormatedName(), false)

	return actionCmd("Connect", "", func() error {
		return sprfl.Connect(mode, password)
	})
}

func (m *Model) Disconnect() tea.Cmd {
	sprfl := m.selectedProfile()
	if sprfl == nil {
		return nil
	}
	if !sprfl.State {
		m.setStatus("Profile not active", true)
		return nil
	}

	delete(m.watching, sprfl.Id)
	m.setStatus("Disconnecting "+sprfl.FormatedName(), false)

	return actionCmd("Disconnect", "", func() error {
		return sprfl.Disconnect()
	})
}

func (m *Model) Import() {
	uriOpt := &OptionText{
		Label:       "Profile URI or file path",
		Placeholder: "pritunl://... or /path/to/profile.tar",
	}

	m.openDialog(NewDialog(
		"Import Profile",
		"Enter a profile URI from the Pritunl server or the path to a "+
			"downloaded .tar or .ovpn profile file.",
		uriOpt,
		&OptionButton{
			Label:  "Cancel",
			Return: DialogCancel,
		},
		&OptionButton{
			Label:  "Import",
			Return: DialogOk,
		},
	), func(m *Model, ret int) tea.Cmd {
		if ret != DialogOk {
			return nil
		}

		path := strings.TrimSpace(uriOpt.GetValue())
		if path == "" {
			m.setStatus("Profile URI or path required", true)
			return nil
		}

		m.setStatus("Importing profile", false)

		return actionCmd("Import", "Profile imported", func() error {
			return sprofile.ImportPath(path)
		})
	})
}

func (m *Model) Logs() tea.Cmd {
	sprfl := m.selectedProfile()
	if sprfl == nil {
		return nil
	}

	m.logs = NewLogsView(sprfl, m.winWidth, m.winHeight)
	m.showLogs = true

	return fetchLogsCmd(sprfl)
}

func (m *Model) Settings() {
	sprfl := m.selectedProfile()
	if sprfl == nil {
		return
	}

	nameOpt := &OptionText{
		Label:       "Name",
		Placeholder: sprfl.FormatedName(),
		Value:       sprfl.Name,
	}
	autostartOpt := &OptionToggle{
		Label: "Autostart",
		Value: !sprfl.Disabled,
	}
	gatewayOpt := &OptionToggle{
		Label: "Disable Default Gateway",
		Value: sprfl.DisableGateway,
	}
	dnsOpt := &OptionToggle{
		Label: "Disable DNS",
		Value: sprfl.DisableDns,
	}
	ipv6Opt := &OptionToggle{
		Label: "Disable IPv6",
		Value: sprfl.DisableIpv6,
	}
	dcoOpt := &OptionToggle{
		Label: "Data Channel Offload",
		Value: sprfl.Dco,
	}
	debugOpt := &OptionToggle{
		Label: "Debug Output",
		Value: sprfl.DebugOutput,
	}
	forceDnsOpt := &OptionToggle{
		Label: "Force DNS Configuration",
		Value: sprfl.ForceDns,
	}

	message := ""
	if sprfl.ForceConnect {
		message = "Autostart is enforced by the server"
	}

	m.openDialog(NewDialog(
		"Settings "+sprfl.FormatedName(),
		message,
		nameOpt,
		autostartOpt,
		gatewayOpt,
		dnsOpt,
		ipv6Opt,
		dcoOpt,
		debugOpt,
		forceDnsOpt,
		&OptionButton{
			Label:  "Cancel",
			Return: DialogCancel,
		},
		&OptionButton{
			Label:  "Save",
			Return: DialogOk,
		},
	), func(m *Model, ret int) tea.Cmd {
		if ret != DialogOk {
			return nil
		}

		updated := *sprfl
		updated.Name = strings.TrimSpace(nameOpt.GetValue())
		updated.Disabled = !autostartOpt.GetValue()
		updated.DisableGateway = gatewayOpt.GetValue()
		updated.DisableDns = dnsOpt.GetValue()
		updated.DisableIpv6 = ipv6Opt.GetValue()
		updated.Dco = dcoOpt.GetValue()
		updated.DebugOutput = debugOpt.GetValue()
		updated.ForceDns = forceDnsOpt.GetValue()

		if updated.ForceConnect && updated.Disabled {
			m.setStatus("Autostart enforced by server", true)
			return nil
		}

		m.setStatus("Saving settings", false)

		return actionCmd("Save Settings", "Settings saved", func() error {
			return updated.Commit()
		})
	})
}

func (m *Model) Remove() {
	sprfl := m.selectedProfile()
	if sprfl == nil {
		return
	}

	m.openDialog(NewDialog(
		"Remove Profile",
		fmt.Sprintf("Remove profile %s? This cannot be undone.",
			sprfl.FormatedName()),
		&OptionButton{
			Label:  "Cancel",
			Return: DialogCancel,
		},
		&OptionButton{
			Label:  "Remove",
			Return: DialogOk,
		},
	), func(m *Model, ret int) tea.Cmd {
		if ret != DialogOk {
			return nil
		}

		m.setStatus("Removing "+sprfl.FormatedName(), false)

		return actionCmd("Remove", "Profile removed", func() error {
			return sprfl.Remove()
		})
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateSize(msg), nil

	case TickMsg:
		cmds := []tea.Cmd{TickInterval()}
		if !m.syncing {
			m.syncing = true
			cmds = append(cmds, syncCmd())
		}
		if m.showLogs {
			cmds = append(cmds, fetchLogsCmd(m.logs.sprfl))
		}
		return m, tea.Batch(cmds...)

	case SyncMsg:
		m.syncing = false
		return m.updateSync(msg)

	case EventMsg:
		return m.updateEvent(msg.Event)

	case ActionDoneMsg:
		if msg.Err != nil {
			m.setStatus(msg.Action+" failed", true)
			m.openError(msg.Action, msg.Err)
			return m, nil
		}
		if msg.Message != "" {
			m.setStatus(msg.Message, false)
		}
		m.syncing = true
		return m, syncCmd()

	case DialogCloseMsg:
		m.showDialog = false
		callback := m.dialogCallback
		m.dialogCallback = nil
		if callback != nil {
			return m, callback(&m, msg.Return)
		}
		return m, nil

	case LogsCloseMsg:
		m.showLogs = false
		return m, nil

	case LogsClearMsg:
		sprfl := msg.Sprofile
		m.openDialog(NewDialog(
			"Clear Logs",
			fmt.Sprintf("Clear the log output for %s?",
				sprfl.FormatedName()),
			&OptionButton{
				Label:  "Cancel",
				Return: DialogCancel,
			},
			&OptionButton{
				Label:  "Clear",
				Return: DialogOk,
			},
		), func(m *Model, ret int) tea.Cmd {
			if ret != DialogOk {
				return nil
			}
			return clearLogsCmd(sprfl)
		})
		return m, nil

	case LogsMsg:
		if m.showLogs {
			var cmd tea.Cmd
			m.logs, cmd = m.logs.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.showDialog {
		var cmd tea.Cmd
		m.dialog, cmd = m.dialog.Update(msg)
		return m, cmd
	}

	if m.showLogs {
		var cmd tea.Cmd
		m.logs, cmd = m.logs.Update(msg)
		return m, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.bindings.Quit):
			return m, tea.Quit
		case key.Matches(keyMsg, m.bindings.Connect):
			return m, m.Connect("ovpn")
		case key.Matches(keyMsg, m.bindings.ConnectWg):
			return m, m.Connect("wg")
		case key.Matches(keyMsg, m.bindings.Disconnect):
			return m, m.Disconnect()
		case key.Matches(keyMsg, m.bindings.Import):
			m.Import()
			return m, nil
		case key.Matches(keyMsg, m.bindings.Logs):
			return m, m.Logs()
		case key.Matches(keyMsg, m.bindings.Settings):
			m.Settings()
			return m, nil
		case key.Matches(keyMsg, m.bindings.Remove):
			m.Remove()
			return m, nil
		case key.Matches(keyMsg, m.bindings.Refresh):
			return m, m.resync()
		}
	}

	profiles, cmd := m.profiles.Update(msg)
	m.profiles = profiles

	return m, cmd
}

func (m *Model) resync() tea.Cmd {
	if m.syncing {
		return nil
	}
	m.syncing = true
	return syncCmd()
}

// updateEvent mirrors the event handling in the desktop client, single
// sign-on links are surfaced from profile sync instead of the sso event.
func (m Model) updateEvent(evt *event.Event) (tea.Model, tea.Cmd) {
	if evt == nil {
		return m, nil
	}

	cmds := []tea.Cmd{}
	if m.listener != nil {
		cmds = append(cmds, waitEventCmd(m.listener))
	}

	data := evt.Profile()
	name := ""
	if data != nil {
		name = m.profileName(data.Id)
	}

	profileMsg := func(msg string) string {
		if name != "" {
			return msg + " on " + name
		}
		return msg
	}

	switch evt.Type {
	case event.ServiceConnected:
		if !m.eventsUp {
			m.eventsUp = true
			if m.syncErr {
				m.setStatus("Service connected", false)
			}
		}
		cmds = append(cmds, m.resync())
	case event.ServiceDisconnected:
		m.eventsUp = false
		m.setStatus("Service connection lost, reconnecting", true)
	case "update", "connected", "disconnected", "profile_sync",
		"wakeup", "registration_pass":

		cmds = append(cmds, m.resync())
	case "registration_required":
		if data != nil && data.RegistrationKey != "" &&
			m.regShown[data.Id] != data.RegistrationKey {

			m.regShown[data.Id] = data.RegistrationKey
			m.openMessage(
				"Device Registration Required",
				"This device must be approved by an administrator "+
					"before connecting"+profileMsg("")+". Provide the "+
					"registration key below to the administrator:\n\n"+
					data.RegistrationKey,
			)
		}
		m.setStatus(profileMsg("Device registration required"), true)
		cmds = append(cmds, m.resync())
	case "auth_error":
		m.setStatus(profileMsg("Failed to authenticate"), true)
		if !m.showDialog {
			m.openMessage(
				"Authentication Failed",
				profileMsg("Failed to authenticate")+
					". Check the password or passcode and try again.",
			)
		}
		cmds = append(cmds, m.resync())
	case "inactive":
		m.setStatus(profileMsg("Disconnected due to inactivity"), true)
		cmds = append(cmds, m.resync())
	case "timeout_error":
		m.setStatus(profileMsg("Connection timed out"), true)
		cmds = append(cmds, m.resync())
	case "offline_error":
		m.setStatus(profileMsg("Server is offline"), true)
		cmds = append(cmds, m.resync())
	case "connection_error":
		m.setStatus(profileMsg("Failed to connect"), true)
		cmds = append(cmds, m.resync())
	case "handshake_timeout":
		m.setStatus(profileMsg("Handshake timeout"), true)
		cmds = append(cmds, m.resync())
	case "configuration_error":
		m.setStatus(profileMsg("Configuration error"), true)
		cmds = append(cmds, m.resync())
	case "shutdown":
		m.setStatus("Pritunl service is shutting down", true)
	case "sso_auth", "sso_interactive", "tpm_open", "tpm_sign", "tpm_close":
		// Single sign-on links are shown from profile sync, TPM events
		// are handled by the desktop client only.
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateSize(msg tea.WindowSizeMsg) Model {
	marginX, marginY := appStyle.GetFrameSize()

	m.winWidth = msg.Width - marginX
	m.winHeight = msg.Height - marginY

	m.listDelegate.SetSplit(m.winWidth >= splitWidth)
	m.listDelegate.SetWidth(m.winWidth)

	// Title bar, status line and menu bar surround the list
	m.profiles.SetSize(m.winWidth, max(m.winHeight-2, 1))

	if m.showDialog {
		m.dialog.SetSize(min(m.winWidth-4, 70), m.winHeight)
	}
	if m.showLogs {
		m.logs.SetSize(m.winWidth, m.winHeight)
	}

	m.ready = true
	return m
}

func (m Model) updateSync(msg SyncMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		if !m.syncErr {
			logger.WithFields(logger.Fields{
				"error": msg.Err,
			}).Error("iface: Failed to sync profiles")
		}
		m.syncErr = true
		m.setStatus("Service unavailable: "+errorMessage(msg.Err), true)
		return m, nil
	}
	if m.syncErr {
		m.setStatus("Service connected", false)
	}
	m.syncErr = false
	m.syncedOnce = true

	items := []list.Item{}
	for _, sprfl := range msg.Profiles {
		items = append(items, NewListItem(sprfl))
	}

	index := m.profiles.Index()
	m.profiles.SetItems(items)
	if len(items) > 0 && index >= len(items) {
		m.profiles.Select(len(items) - 1)
	}

	if m.showDialog {
		return m, nil
	}

	// Surface single sign-on links and device registration keys for
	// profiles connected from this session.
	for _, sprfl := range msg.Profiles {
		if !m.watching[sprfl.Id] {
			continue
		}

		if sprfl.Profile != nil && sprfl.Profile.SsoUrl != "" &&
			m.ssoShown[sprfl.Id] != sprfl.Profile.SsoUrl {

			m.ssoShown[sprfl.Id] = sprfl.Profile.SsoUrl
			m.openMessage(
				"Single Sign-On Authentication",
				"Open the link below in a browser to complete "+
					"authentication:\n\n"+sprfl.Profile.SsoUrl,
			)
			return m, nil
		}

		if sprfl.RegistrationKey != "" &&
			m.regShown[sprfl.Id] != sprfl.RegistrationKey {

			m.regShown[sprfl.Id] = sprfl.RegistrationKey
			m.openMessage(
				"Device Registration Required",
				"This device must be approved by an administrator "+
					"before connecting. Provide the registration key "+
					"below to the administrator:\n\n"+
					sprfl.RegistrationKey,
			)
			return m, nil
		}

		if !sprfl.State {
			delete(m.watching, sprfl.Id)
		}
	}

	return m, nil
}

func (m Model) menuItems() []MenuItem {
	menu := []MenuItem{}

	sprfl := m.selectedProfile()
	if sprfl != nil {
		if sprfl.State {
			menu = append(menu, MenuItem{Title: "Disconnect", Key: "d"})
		} else {
			if sprfl.HideOvpn {
				menu = append(menu,
					MenuItem{Title: "Connect WireGuard", Key: "c"})
			} else {
				menu = append(menu,
					MenuItem{Title: "Connect OpenVPN", Key: "c"})
				if sprfl.Wg {
					menu = append(menu,
						MenuItem{Title: "Connect WireGuard", Key: "w"})
				}
			}
		}
		menu = append(menu,
			MenuItem{Title: "Logs", Key: "l"},
			MenuItem{Title: "Settings", Key: "s"},
			MenuItem{Title: "Remove", Key: "r"},
		)
	}

	menu = append(menu,
		MenuItem{Title: "Import", Key: "i"},
		MenuItem{Title: "Quit", Key: "q"},
	)

	return menu
}

func (m Model) renderStatus() string {
	timeout := statusTimeout
	if m.statusErr {
		timeout = statusErrorTimeout
	}
	if m.statusMsg != "" && time.Since(m.statusTime) < timeout {
		text := renderCol(m.winWidth-1, "%s", m.statusMsg)
		if m.statusErr {
			return " " + statusErrorStyle.Render(text)
		}
		return " " + statusInfoStyle.Render(text)
	}

	if m.syncErr {
		return " " + statusErrorStyle.Render(renderCol(
			m.winWidth-1, "%s", "Unable to reach the Pritunl service"))
	}

	return ""
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.showLogs && !m.showDialog {
		return appStyle.Render(m.logs.View())
	}

	title := menuBarStyle.Width(m.winWidth).Render(fmt.Sprintf(
		" %s v%s", m.profiles.Title, constants.Version))

	var body string
	if len(m.profiles.Items()) == 0 {
		var text string
		if !m.syncedOnce {
			text = "Loading profiles..."
		} else {
			text = "No profiles, press i to import a profile"
		}
		body = lipgloss.NewStyle().Height(max(m.winHeight-3, 1)).Render(
			emptyStyle.Render(text))
	} else {
		listView := m.profiles.View()

		// Replace the list title line with the full width title bar
		parts := strings.SplitN(listView, "\n", 2)
		if len(parts) == 2 {
			body = parts[1]
		} else {
			body = listView
		}
	}

	mainView := appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			body,
			m.renderStatus(),
			renderMenuBar(m.winWidth, m.menuItems()),
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
