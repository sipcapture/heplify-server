package sipparser

import (
	"fmt"
	"strconv"
	"strings"
)

func parseUriHost(u *URI) uriStateFn {
	if len(u.Raw) <= u.atPos {
		u.Error = fmt.Errorf("malformed host part inside URI: %s", u.Raw)
		return nil
	}

	hostPart := u.Raw
	if u.atPos != 0 {
		hostPart = u.Raw[u.atPos+1:]
	}
	if i := strings.IndexByte(hostPart, ';'); i >= 0 {
		hostPart = hostPart[:i]
	}
	if i := strings.IndexByte(hostPart, '?'); i >= 0 {
		hostPart = hostPart[:i]
	}

	// RFC 3261: IPv6 literals must be in brackets, e.g. sip:user@[2001:db8::1]:5060
	if strings.HasPrefix(hostPart, "[") {
		end := strings.IndexByte(hostPart, ']')
		if end <= 1 {
			u.Error = fmt.Errorf("malformed IPv6 host in URI: %s", u.Raw)
			return nil
		}
		u.Host = hostPart[1:end]
		rest := hostPart[end+1:]
		if strings.HasPrefix(rest, ":") && len(rest) > 1 {
			u.Port = rest[1:]
			u.PortInt, _ = strconv.Atoi(u.Port)
		}
		return nil
	}

	// Unbracketed IPv6 (multiple colons) — take whole host; port is ambiguous without brackets.
	if strings.Count(hostPart, ":") > 1 {
		u.Host = hostPart
		return nil
	}

	// IPv4 / hostname with optional :port
	if i := strings.LastIndexByte(hostPart, ':'); i >= 0 {
		port := hostPart[i+1:]
		if port != "" && isDigits(port) {
			u.Host = hostPart[:i]
			u.Port = port
			u.PortInt, _ = strconv.Atoi(port)
			return nil
		}
	}

	u.Host = hostPart
	return nil
}

func isDigits(s string) bool {
	for i := range s {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
