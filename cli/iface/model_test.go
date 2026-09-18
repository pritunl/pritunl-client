package iface

import (
	"encoding/json"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pritunl/pritunl-client/cli/event"
	"github.com/pritunl/pritunl-client/cli/profile"
	"github.com/pritunl/pritunl-client/cli/sprofile"
)

const testSsoUrl = "https://vpn.example.com/sso/authenticate"

func ssoTestModel() Model {
	m := NewModel(nil)
	m = m.updateSize(tea.WindowSizeMsg{Width: 100, Height: 30})
	m.watching["test-profile"] = true
	return m
}

func ssoTestSync(system bool, url string) SyncMsg {
	return SyncMsg{Profiles: sprofile.Sprofiles{&sprofile.Sprofile{
		Id:     "test-profile",
		Name:   "Test Profile",
		State:  true,
		System: system,
		Profile: &profile.Profile{
			Status: "authenticating",
			SsoUrl: url,
		},
	}}}
}

// Do not execute returned commands: these tests feed service messages directly
// into the model without connecting to a running VPN service.
func ssoTestUpdate(m Model, msg tea.Msg) Model {
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func ssoTestEvent(data string) EventMsg {
	return EventMsg{Event: &event.Event{
		Type: "sso_auth",
		Data: json.RawMessage(data),
	}}
}

func assertSsoDialog(t *testing.T, m Model, url string) {
	t.Helper()
	if !m.showDialog || m.dialog.title != "Single Sign-On Authentication" {
		t.Fatalf("expected SSO dialog, got visible=%v title=%q",
			m.showDialog, m.dialog.title)
	}
	if m.dialog.link != url {
		t.Fatalf("SSO dialog URL = %q, want %q", m.dialog.link, url)
	}
}

func TestSsoAuthUserProfile(t *testing.T) {
	m := ssoTestModel()
	// User profiles receive the URL only in the event, not in profile sync.
	// The browser-open preference must not hide a manually opened link.
	evt := ssoTestEvent(`{"id":"test-profile","url":"` + testSsoUrl + `","open":false}`)
	m = ssoTestUpdate(m, evt)
	m = ssoTestUpdate(m, ssoTestSync(false, ""))
	assertSsoDialog(t, m, testSsoUrl)

	m = ssoTestUpdate(m, DialogCloseMsg{Return: DialogOk})
	m = ssoTestUpdate(m, evt)
	m = ssoTestUpdate(m, ssoTestSync(false, ""))
	if m.showDialog {
		t.Fatal("duplicate SSO event reopened the dialog")
	}

	nextUrl := testSsoUrl + "?attempt=2"
	m = ssoTestUpdate(m, ssoTestEvent(`{"id":"test-profile","url":"`+nextUrl+`"}`))
	m = ssoTestUpdate(m, ssoTestSync(false, ""))
	assertSsoDialog(t, m, nextUrl)
}

func TestSsoAuthWaitsForDialog(t *testing.T) {
	m := ssoTestModel()
	callbackCalled := false
	m.openDialog(NewDialog("Existing dialog", "Keep this dialog"),
		func(m *Model, ret int) tea.Cmd {
			callbackCalled = true
			return nil
		})
	m = ssoTestUpdate(m, ssoTestEvent(`{"id":"test-profile","url":"`+testSsoUrl+`"}`))
	m = ssoTestUpdate(m, ssoTestSync(false, ""))
	if !m.showDialog || m.dialog.title != "Existing dialog" {
		t.Fatal("SSO event replaced the existing dialog")
	}
	m = ssoTestUpdate(m, DialogCloseMsg{Return: DialogOk})
	if !callbackCalled {
		t.Fatal("SSO event replaced the existing dialog callback")
	}
	m = ssoTestUpdate(m, ssoTestSync(false, ""))
	assertSsoDialog(t, m, testSsoUrl)
}

func TestSsoAuthSystemProfile(t *testing.T) {
	for _, eventFirst := range []bool{false, true} {
		name := "sync-first"
		if eventFirst {
			name = "event-first"
		}
		t.Run(name, func(t *testing.T) {
			m := ssoTestModel()
			evt := ssoTestEvent(`{"id":"test-profile","url":"` + testSsoUrl + `"}`)
			if eventFirst {
				m = ssoTestUpdate(m, evt)
			}
			m = ssoTestUpdate(m, ssoTestSync(true, testSsoUrl))
			assertSsoDialog(t, m, testSsoUrl)
			m = ssoTestUpdate(m, DialogCloseMsg{Return: DialogOk})
			m = ssoTestUpdate(m, evt)
			m = ssoTestUpdate(m, ssoTestSync(true, testSsoUrl))
			if m.showDialog {
				t.Fatal("event and profile sync showed the same SSO link twice")
			}
		})
	}
}

func TestSsoAuthIgnoresInvalidOrUnwatchedEvents(t *testing.T) {
	for _, data := range []string{
		``, `null`, `{`, `{"id":"test-profile"}`,
		`{"url":"` + testSsoUrl + `"}`,
		`{"id":"other-profile","url":"` + testSsoUrl + `"}`,
	} {
		t.Run(data, func(t *testing.T) {
			m := ssoTestModel()
			m = ssoTestUpdate(m, ssoTestEvent(data))
			m = ssoTestUpdate(m, ssoTestSync(false, ""))
			if m.showDialog {
				t.Fatal("invalid or unwatched SSO event opened a dialog")
			}
		})
	}
}

func TestSsoAuthDiscardsPendingLinkAfterConnectionEnds(t *testing.T) {
	for _, status := range []string{"connected", "disconnecting", "disconnected"} {
		t.Run(status, func(t *testing.T) {
			m := ssoTestModel()
			m.openMessage("Existing dialog", "Keep this dialog")
			m = ssoTestUpdate(m, ssoTestEvent(`{"id":"test-profile","url":"`+testSsoUrl+`"}`))
			m = ssoTestUpdate(m, EventMsg{Event: &event.Event{
				Type: "update",
				Data: json.RawMessage(`{"id":"test-profile","status":"` + status + `"}`),
			}})
			m = ssoTestUpdate(m, DialogCloseMsg{Return: DialogOk})
			m = ssoTestUpdate(m, ssoTestSync(false, ""))
			if m.showDialog {
				t.Fatal("completed authentication left a pending SSO dialog")
			}
		})
	}
}

func TestSsoAuthDialogClosesWhenConnected(t *testing.T) {
	m := ssoTestModel()
	m = ssoTestUpdate(m, ssoTestEvent(`{"id":"test-profile","url":"`+testSsoUrl+`"}`))
	m = ssoTestUpdate(m, ssoTestSync(false, ""))
	assertSsoDialog(t, m, testSsoUrl)

	m = ssoTestUpdate(m, EventMsg{Event: &event.Event{Type: "connected"}})
	if m.showDialog {
		t.Fatal("SSO dialog remains open after the profile connects")
	}
}
