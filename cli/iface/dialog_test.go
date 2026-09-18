package iface

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestSsoDialogLinkRenderingAndCopy(t *testing.T) {
	url := "https://vpn.example.com/key/request?state=" + strings.Repeat("abc123", 30)
	for _, width := range []int{40, 80, 300} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			m := ssoTestModel()
			m = ssoTestUpdate(m, ssoTestSync(true, url))
			d := m.dialog
			d.SetSize(width, 30)
			view := d.View()
			if !strings.Contains(view, ansi.SetHyperlink(url)) {
				t.Fatal("rendered hyperlink does not target the complete URL")
			}
			plain := ansi.Strip(view)
			linkLines := 0
			for _, line := range strings.Split(plain, "\n") {
				if strings.Contains(line, "Open SSO link") {
					linkLines++
				}
				if ansi.StringWidth(line) > width {
					t.Fatalf("dialog exceeds width %d: %q", width, line)
				}
			}
			if linkLines != 1 {
				t.Fatalf("URL occupies %d lines, want one", linkLines)
			}
			if !strings.Contains(plain, "c: copy link") {
				t.Fatal("missing copy shortcut hint")
			}
			if strings.Contains(plain, "abc123") {
				t.Fatal("visible URL should use a short label")
			}
			d, cmd := d.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
			if cmd == nil || d.copyStatus != "Copying link…" {
				t.Fatal("copy shortcut did not start a clipboard command")
			}
		})
	}
}

func TestOpenLinkCommand(t *testing.T) {
	cmd := openLinkCmd("https://vpn.example.com/key/request?state=test")
	if cmd == nil {
		t.Fatal("open link command is nil")
	}
}

func TestOpenURLCommandRejectsUnsafeURLs(t *testing.T) {
	for _, raw := range []string{"", "file:///tmp/test", "https://"} {
		if err := (&openURLCommand{url: raw}).Run(); err == nil {
			t.Fatalf("URL %q was accepted", raw)
		}
	}
}

func TestSsoLinkHasMouseRegion(t *testing.T) {
	d := NewDialog("SSO", "Authenticate")
	d.link = "https://vpn.example.com/key/request?state=test"
	_, regions := d.render()
	for _, region := range regions {
		if region.index == dialogRegionLink {
			if region.w != len("Open SSO link") || region.h != 1 {
				t.Fatalf("unexpected SSO link mouse region: %+v", region)
			}
			return
		}
	}
	t.Fatal("SSO link has no mouse region")
}

func TestCopyLinkResult(t *testing.T) {
	url := "https://vpn.example.com/key/request?state=unwrapped-token"
	for _, fail := range []bool{false, true} {
		t.Run(strconv.FormatBool(fail), func(t *testing.T) {
			d := NewDialog("SSO", "Authenticate")
			d.link = url
			cmd := copyLinkCmd(url, func(text string) error {
				if text != url {
					t.Fatalf("clipboard received %q, want full URL %q", text, url)
				}
				if fail {
					return errors.New("clipboard unavailable")
				}
				return nil
			})
			d, _ = d.Update(cmd())
			want := "Link copied"
			if fail {
				want = "Copy failed: clipboard unavailable"
			}
			if !strings.Contains(ansi.Strip(d.View()), want) {
				t.Fatalf("dialog does not report %q", want)
			}
			d.copyStatus = ""
			d, _ = d.Update(linkCopiedMsg{url: "https://other.example.com"})
			if d.copyStatus != "" {
				t.Fatal("copy result from an earlier dialog changed this dialog")
			}
		})
	}
}

func TestDialogWithoutLinkDoesNotCopy(t *testing.T) {
	d := NewDialog("Message", "No authentication link")
	_, cmd := d.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if cmd != nil {
		t.Fatal("copy shortcut must only apply to dialogs with a link")
	}
	if strings.Contains(d.View(), "c: copy link") {
		t.Fatal("ordinary dialog shows copy shortcut")
	}
}
