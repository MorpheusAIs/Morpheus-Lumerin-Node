package videojob

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strings"
)

// Two URLs come from the other side: the caller's first-frame image_url, which
// the provider hands to the vendor, and the MP4 link the provider relays from
// the vendor's status, which the consumer hands to its caller. Both must be
// public https URLs, so neither can point into a private network.

var (
	ErrRemoteURL     = errors.New("refused remote url")
	errAddrNotPublic = errors.New("address is not public")
)

// ValidateRemoteURL accepts an absolute https URL with a host and no
// credentials in it.
func ValidateRemoteURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRemoteURL, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be https", ErrRemoteURL)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("%w: missing host", ErrRemoteURL)
	}
	if u.User != nil {
		return fmt.Errorf("%w: credentials in url", ErrRemoteURL)
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if addr, err := netip.ParseAddr(host); err == nil {
		if !isPublicAddr(addr) {
			return fmt.Errorf("%w: %s", ErrRemoteURL, errAddrNotPublic)
		}
		return nil
	}
	if !isPublicHostname(host) {
		return fmt.Errorf("%w: host %q is not public", ErrRemoteURL, host)
	}
	return nil
}

// isPublicHostname refuses names that are local by convention and the
// numeric spellings net/url does not parse as addresses but a browser dials
// (127.1, 2130706433, 0x7f000001, 0177.0.0.1). A public name has a dot and
// a label that is not a number.
func isPublicHostname(host string) bool {
	if host == "localhost" || host == "localhost.localdomain" || !strings.Contains(host, ".") {
		return false
	}
	// A browser maps Unicode spellings (fullwidth digits, an ideographic
	// stop) onto ASCII before it dials; a public vendor host is ASCII (IDNs
	// arrive as punycode).
	for _, r := range host {
		if r > 0x7f {
			return false
		}
	}
	for _, suffix := range []string{".localhost", ".localdomain", ".local", ".internal", ".home.arpa"} {
		if strings.HasSuffix(host, suffix) {
			return false
		}
	}
	public := false
	for _, label := range strings.Split(host, ".") {
		if label == "" {
			return false
		}
		if !numericLabel(label) {
			public = true
		}
	}
	return public
}

// numericLabel reports whether a hostname label is a decimal, octal or hex
// number, the forms an IPv4 shorthand is written in. A bare "0x" is zero.
func numericLabel(label string) bool {
	digits, hexa := label, false
	if strings.HasPrefix(label, "0x") {
		digits, hexa = label[2:], true
	}
	if digits == "" {
		return hexa
	}
	for _, r := range digits {
		if r >= '0' && r <= '9' {
			continue
		}
		if hexa && r >= 'a' && r <= 'f' {
			continue
		}
		return false
	}
	return true
}

func isPublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	// A zone names a local interface, and the prefix checks below do not see
	// through it.
	if addr.Zone() != "" {
		return false
	}
	if !addr.IsValid() || addr.IsUnspecified() || addr.IsLoopback() || addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() || addr.IsMulticast() {
		return false
	}
	for _, p := range nonPublicPrefixes {
		if p.Contains(addr) {
			return false
		}
	}
	return true
}

var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	// IPv6 forms that embed an IPv4 address (NAT64, 6to4) can reach a
	// private IPv4 host through a translator.
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("2001::/32"), // Teredo
	// IPv4-compatible IPv6 (::a.b.c.d) and the deprecated site-local range.
	netip.MustParsePrefix("::/96"),
	netip.MustParsePrefix("fec0::/10"),
}
