package network

import (
	"fmt"

	"time"

	"project.com/pkg/elevator"
)

func Network_fsm(infoUpdate chan elevator.Elevator, external_info chan elevator.Elevator, UpdatePeers chan PeerUpdate) {

	Requests := [4][3]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
		{true, true, true},
	}

	id, err := LocalIP()
	if err != nil {
		fmt.Printf("could not get IP")
	}

	peerUpdateCh := make(chan PeerUpdate)
	peerTxEnable := make(chan bool)
	networkTx := make(chan elevator.Elevator)
	networkRx := make(chan elevator.Elevator)

	go TransmitterPeers(15647, id, peerTxEnable)
	go ReceiverPeers(15647, peerUpdateCh)
	go TransmitterBcast(20007, networkTx)
	go ReceiverBcast(20007, networkRx)

	periodicMsg := elevator.Elevator{id, 0, 69, elevator.MD_Down, Requests, elevator.EB_DoorOpen, 0.8}
	go func() {

		for {
			periodicMsg.OrderClearedCounter++
			networkTx <- periodicMsg
			time.Sleep(5 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			UpdatePeers <- p
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-networkRx:
			fmt.Printf("Received: %#v\n", a)
			external_info <- a

		case i := <-infoUpdate:
			periodicMsg = i
			networkTx <- i
		}
	}

	
}
