package netsim

import (
	"testing"
)

func Test_NetworkSimulator(t *testing.T) {
	ns := NewNetworkSimulator(10, 0)
	ns.SetLatency(100)
	ns.SendPacket(1, []byte("Hello, world!"))
	ns.AdvanceTime(.1) // in seconds, so .1 = 100ms
	packets, tos := ns.ReceivePackets(1)

	if len(packets) != 0 || len(tos) != 0 {
		t.Fatal("Expected 0 got", len(packets), len(tos))
	}

	ns.AdvanceTime(.11)
	packets, tos = ns.ReceivePackets(1)

	if len(packets) != 1 || len(tos) != 1 {
		t.Fatal("Expected 1 got", len(packets), len(tos))
	}

	if string(packets[0]) != "Hello, world!" {
		t.Fatal("Response ", string(packets[0]))
	}
}
