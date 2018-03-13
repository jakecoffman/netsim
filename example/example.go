package main

import (
	"github.com/jakecoffman/netsim"
	"time"
	"fmt"
	"log"
	"encoding/gob"
	"bytes"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//gob.Register(&Msg{})

	msg1 := NewMsg(9)
	msg2 := &Msg{}

	b, err := msg1.Write()
	if err != nil {
		log.Fatal(err)
	}

	err = msg2.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	if *msg1 != *msg2 {
		log.Fatal("NOT SAME", msg1, msg2)
	}
}

func main() {
	server := make(chan []byte, 100)
	client1 := make(chan []byte, 50)
	client2 := make(chan []byte, 50)
	go client(server, client1, 1)
	go client(server, client2, 2)

	ns := netsim.NewNetworkSimulator(100, getTime())
	ns.SetLatency(50)
	ns.SetPacketLoss(10)
	ns.SetDuplicates(5)
	ns.SetJitter(25)
	tick := time.Tick(16 * time.Millisecond)

	for {
		<-tick

		in:
		for {
			select {
			case packet := <-server:
				msg := &Msg{}
				err := msg.Read(packet)
				if err != nil {
					log.Fatal(err)
				}
				msg.ReceivedAt = time.Now().UnixNano()
				buf, err := msg.Write()
				if err != nil {
					log.Fatal(err)
				}
				ns.SendPacket(msg.ClientID, buf)
			default:
				break in
			}
		}

		ns.AdvanceTime(getTime())

		packets, tos := ns.ReceivePackets(100)
		for i, packet := range packets {
			switch tos[i] {
			case 1:
				client1 <- packet
			case 2:
				client2 <- packet
			default:
				log.Fatal("Unexpected packet", string(packet), tos[i])
			}
		}
	}
}

func client(toServer chan<- []byte, toClient <-chan []byte, clientID int) {
	ns := netsim.NewNetworkSimulator(100, getTime())
	ns.SetLatency(50)

	tick := time.Tick(16 * time.Millisecond)

	for {
		<-tick

		myMsg := NewMsg(clientID)
		buf, err := myMsg.Write()
		if err != nil {
			log.Fatal(err)
		}
		ns.SendPacket(0, buf)
		in:
		for {
			select {
			case packet := <-toClient:
				msg := &Msg{}
				err := msg.Read(packet)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Client", clientID, "RTT", time.Now().Sub(time.Unix(0, msg.SentAt)))
			default:
				break in
			}
		}

		ns.AdvanceTime(getTime())

		packets, tos := ns.ReceivePackets(100)
		for i, packet := range packets {
			switch tos[i] {
			case 0:
				toServer <- packet
			default:
				log.Fatal("Unexpected packet", string(packet), tos[i])
			}
		}
	}
}

func getTime() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}

type Msg struct {
	ClientID   int
	SentAt     int64
	ReceivedAt int64
}

func NewMsg(clientID int) *Msg {
	return &Msg{clientID, time.Now().UnixNano(), 0}
}

func (msg *Msg) Write() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := gob.NewEncoder(buf).Encode(msg)
	return buf.Bytes(), err
}

func (msg *Msg) Read(b []byte) error {
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(msg)
	return err
}
