package netsim

import (
	"math/rand"
)

// NetworkSimulator is used to introduce various real life problems into your network communications
type NetworkSimulator struct {
	latency, jitter, packetLoss, duplicates float64
	active bool

	time float64
	currentIndex int
	packetEntries []PacketEntry
}

// NewNetworkSimulator returns a NetworkSimulator with required data:
// numPackets is the size of the buffer and time is the initial game time (e.g. glfw.GetTime()) in seconds
func NewNetworkSimulator(numPackets int, time float64) *NetworkSimulator {
	return &NetworkSimulator{
		time: time,
		packetEntries: make([]PacketEntry, numPackets),
	}
}

// IsActive returns whether the network simulator is setup to return simulated data
// if not then just send the data on the network
func (n *NetworkSimulator) IsActive() bool {
	return n.active
}

// SetLatency sets the latency of packets (how long to delay sending)
func (n *NetworkSimulator) SetLatency(ms float64) {
	n.latency = ms
	n.updateActive()
}

// SetJitter sets the packet jitter in ms.
func (n *NetworkSimulator) SetJitter(ms float64) {
	n.jitter = ms
	n.updateActive()
}

// SetPacketLoss sets the packet loss percentage.
func (n *NetworkSimulator) SetPacketLoss(percent float64) {
	n.packetLoss = percent
	n.updateActive()
}

// SetDuplicates sets the percentage of chance of duplication.
func (n *NetworkSimulator) SetDuplicates(percent float64) {
	n.duplicates = percent
	n.updateActive()
}

// updateActive updates the active flag whenever network settings are changed
func (n *NetworkSimulator) updateActive() {
	previous := n.active
	n.active = n.latency != 0 || n.jitter != 0 || n.packetLoss != 0 || n.duplicates != 0
	if previous && !n.active {
		n.DiscardPackets()
	}
}

// SendPacket enqueues a packet in the simulator, copying the data from packetData
func (n *NetworkSimulator) SendPacket(to int, packetData []byte) {
	if randomFloat(0, 100) <= n.packetLoss {
		return
	}

	packetEntry := &n.packetEntries[n.currentIndex]

	if packetEntry.packetData != nil {
		packetEntry = &PacketEntry{}
	}

	delay := n.latency / 1000

	if n.jitter > 0 {
		delay += randomFloat(-n.jitter, n.jitter) / 1000
	}

	packetEntry.to = to
	packetEntry.packetData = make([]byte, len(packetData))
	copy(packetEntry.packetData, packetData)
	packetEntry.deliveryTime = n.time + delay
	n.currentIndex = (n.currentIndex+1)%len(n.packetEntries)

	if randomFloat(0, 100) <= n.duplicates {
		nextPacketEntry := n.packetEntries[n.currentIndex]
		nextPacketEntry.to = to
		nextPacketEntry.packetData = make([]byte, len(packetData))
		copy(nextPacketEntry.packetData, packetData)
		nextPacketEntry.deliveryTime = n.time + delay + randomFloat(0, 1)
		n.currentIndex = (n.currentIndex + 1) % len(n.packetEntries)
	}
}

// ReceivePackets dequeues up to maxPackets of packets so they can be sent on the network
func (n *NetworkSimulator) ReceivePackets(maxPackets int) (packetData [][]byte, to []int) {
	if !n.IsActive() {
		return
	}

	packetData = [][]byte{}
	to = []int{}

	for i := 0; i < min(len(n.packetEntries), maxPackets); i++ {
		if n.packetEntries[i].packetData == nil {
			continue
		}

		if n.packetEntries[i].deliveryTime < n.time {
			packetData = append(packetData, n.packetEntries[i].packetData)
			to = append(to, n.packetEntries[i].to)
			n.packetEntries[i].packetData = nil
		}
	}

	return packetData, to
}

// AdvanceTime is used to set the current time in seconds so the delay can be calculated
func (n *NetworkSimulator) AdvanceTime(time float64) {
	n.time = time
}

// DiscardPackets resets all packet entries
func (n *NetworkSimulator) DiscardPackets() {
	for i:=0; i<len(n.packetEntries); i++ {
		packet := &n.packetEntries[i]
		if packet.packetData == nil {
			continue
		}
		n.packetEntries[i] = PacketEntry{}
	}
}

// DiscardClientPackets resets all packets from a single client
func (n *NetworkSimulator) DiscardClientPackets(clientIndex int) {
	for i:=0; i<len(n.packetEntries); i++ {
		packet := &n.packetEntries[i]
		if packet.packetData == nil || packet.to != clientIndex {
			continue
		}
		n.packetEntries[i] = PacketEntry{}
	}
}

// PacketEntry stores metadata associated with a packet
type PacketEntry struct {
	to int
	deliveryTime float64
	packetData []byte
}

func randomFloat(min, max float64) float64 {
	return min + (rand.Float64() * (max - min))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
