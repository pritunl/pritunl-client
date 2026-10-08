package utils

import (
	"net"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dropbox/godropbox/container/set"
)

const (
	idSafeLimit     = 256
	domainSafeLimit = 1024
	base64SafeLimit = 32768
	pemSafeLimit    = 32768
)

var nameSafeChar = set.NewSet(
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'-', '=', '_', '@', '.', ':', '/', '#', '*', '+', '?', '^', '~',
)

var domainSafeChar = set.NewSet(
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'-', '.',
)

var idSafeChar = set.NewSet(
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', '-', '_',
)

var base64SafeChar = set.NewSet(
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'+', '/', '=', '-', '_',
)

var pemSafeChar = set.NewSet(
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'+', '/', '=', '-', ' ', '\n',
)

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func isDigits(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func FilterStr(s string, n int) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > n {
		s = s[:n]
	}

	if s == "self" {
		s = "invalid-name"
	}

	var ns strings.Builder
	for _, c := range s {
		if nameSafeChar.Contains(c) {
			ns.WriteString(string(c))
		}
	}

	return ns.String()
}

func FilterId(s string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > idSafeLimit {
		s = s[:idSafeLimit]
	}

	var ns strings.Builder
	for _, c := range s {
		if idSafeChar.Contains(c) {
			ns.WriteString(string(c))
		}
	}

	return ns.String()
}

func FilterDomain(s string) string {
	if len(s) == 0 {
		return ""
	}

	if s == "self" {
		s = "invalid-name"
	}

	if len(s) > domainSafeLimit {
		s = s[:domainSafeLimit]
	}

	port := ""
	host, prt, err := net.SplitHostPort(s)
	if err == nil && isDigits(prt) {
		s = host
		port = ":" + prt
	}

	if ip := net.ParseIP(strings.Trim(s, "[]")); ip != nil {
		if port != "" && ip.To4() == nil {
			return "[" + ip.String() + "]" + port
		}
		return ip.String() + port
	}

	var ns strings.Builder
	for _, c := range s {
		if domainSafeChar.Contains(c) {
			ns.WriteString(string(c))
		}
	}

	s = ns.String()
	s = strings.TrimPrefix(s, ".")
	s = strings.TrimSuffix(s, ".")
	if s == "" {
		return ""
	}

	return s + port
}

func FilterBase64(s string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > base64SafeLimit {
		s = s[:base64SafeLimit]
	}

	var ns strings.Builder
	for _, c := range s {
		if base64SafeChar.Contains(c) {
			ns.WriteString(string(c))
		}
	}

	return ns.String()
}

func FilterPem(s string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > pemSafeLimit {
		s = s[:pemSafeLimit]
	}

	s = normalizeNewlines(s)

	var ns strings.Builder
	for _, c := range s {
		if pemSafeChar.Contains(c) {
			ns.WriteString(string(c))
		}
	}

	return ns.String()
}

func FilterText(s string, n int) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > n {
		s = s[:n]
	}

	s = normalizeNewlines(s)

	var ns strings.Builder
	for _, c := range s {
		if c == utf8.RuneError {
			continue
		}
		if c == '\n' || c == '\t' {
			ns.WriteRune(c)
			continue
		}
		if unicode.IsControl(c) {
			continue
		}
		ns.WriteRune(c)
	}

	return ns.String()
}

func FilterOpen(s string) string {
	if len(s) == 0 {
		return ""
	}

	return s
}
