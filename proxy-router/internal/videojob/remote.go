package videojob

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
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
	if addr, err := netip.ParseAddr(u.Hostname()); err == nil && !isPublicAddr(addr) {
		return fmt.Errorf("%w: %s", ErrRemoteURL, errAddrNotPublic)
	}
	return nil
}

func isPublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
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
}
