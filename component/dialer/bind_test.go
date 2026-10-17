package dialer

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldPinSourceAddr(t *testing.T) {
	cases := []struct {
		network     string
		destination string
		want        bool
	}{
		{"tcp", "104.16.0.1", true},
		{"tcp4", "104.16.0.1", true},
		{"tcp", "192.168.2.1", true},
		{"tcp", "::ffff:104.16.0.1", true},
		{"tcp6", "::ffff:104.16.0.1", false},
		{"tcp", "2606:4700::1", false},
		{"tcp6", "2606:4700::1", false},
		{"udp", "104.16.0.1", false},
		{"udp4", "104.16.0.1", false},
		{"ip4:icmp", "104.16.0.1", false},
		{"tcp", "127.0.0.1", false},
		{"tcp", "169.254.1.1", false},
		{"tcp", "255.255.255.255", false},
		{"tcp", "0.0.0.0", false},
		{"tcp", "", false},
	}
	for _, c := range cases {
		var destination netip.Addr
		if c.destination != "" {
			destination = netip.MustParseAddr(c.destination)
		}
		assert.Equal(t, c.want, shouldPinSourceAddr(c.network, destination), "%s %s", c.network, c.destination)
	}
}
