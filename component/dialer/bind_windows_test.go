package dialer

import (
	"net"
	"net/netip"
	"testing"

	"github.com/metacubex/mihomo/component/iface"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upInterfaceWithIPv4 returns an up, non-loopback interface that has a
// non-link-local IPv4 address, or skips the test.
func upInterfaceWithIPv4(t *testing.T) *iface.Interface {
	ifaces, err := iface.Interfaces()
	require.NoError(t, err)
	for _, ifaceObj := range ifaces {
		if ifaceObj.Flags&net.FlagUp == 0 || ifaceObj.Flags&net.FlagLoopback != 0 {
			continue
		}
		if _, err := ifaceObj.PickIPv4Addr(netip.Addr{}); err == nil {
			return ifaceObj
		}
	}
	t.Skip("no up interface with an IPv4 address")
	return nil
}

func TestBindIfaceToDialerPinsTCPv4Source(t *testing.T) {
	ifaceObj := upInterfaceWithIPv4(t)

	dialer := &net.Dialer{}
	require.NoError(t, bindIfaceToDialer(ifaceObj.Name, dialer, "tcp", netip.MustParseAddr("104.16.0.1")))
	assert.NotNil(t, dialer.ControlContext)

	local, ok := dialer.LocalAddr.(*net.TCPAddr)
	require.True(t, ok, "LocalAddr = %#v", dialer.LocalAddr)
	assert.Equal(t, 0, local.Port)
	localAddr, _ := netip.AddrFromSlice(local.IP)
	found := false
	for _, prefix := range ifaceObj.Addresses {
		if prefix.Addr() == localAddr.Unmap() {
			found = true
		}
	}
	assert.True(t, found, "%s is not an address of %s", localAddr, ifaceObj.Name)
}

func TestBindIfaceToDialerLeavesOtherSourcesAlone(t *testing.T) {
	ifaceObj := upInterfaceWithIPv4(t)

	cases := []struct {
		network     string
		destination string
	}{
		{"udp", "104.16.0.1"},
		{"tcp", "2606:4700::1"},
		{"ip4:icmp", "104.16.0.1"},
	}
	for _, c := range cases {
		dialer := &net.Dialer{}
		require.NoError(t, bindIfaceToDialer(ifaceObj.Name, dialer, c.network, netip.MustParseAddr(c.destination)))
		assert.NotNil(t, dialer.ControlContext, "%s %s", c.network, c.destination)
		assert.Nil(t, dialer.LocalAddr, "%s %s", c.network, c.destination)
	}

	// A caller-provided local address is kept.
	preset := &net.TCPAddr{IP: net.IPv4(10, 0, 0, 1)}
	dialer := &net.Dialer{LocalAddr: preset}
	require.NoError(t, bindIfaceToDialer(ifaceObj.Name, dialer, "tcp", netip.MustParseAddr("104.16.0.1")))
	assert.Same(t, preset, dialer.LocalAddr)
}
