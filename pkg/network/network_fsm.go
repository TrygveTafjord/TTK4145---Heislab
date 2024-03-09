package network

import (
	"fmt"

	//"time"

	"project.com/pkg/elevator"
)

func Network_fsm(networkUpdateTx_ch chan elevator.Elevator, networkUpdateRx_ch chan elevator.Elevator, peerUpdate_ch chan PeerUpdate) {

	id, err := LocalIP()
	if err != nil {
		fmt.Printf("could not get IP")
	}

	peerUpdateCh := make(chan PeerUpdate,5)
	peerTxEnable := make(chan bool,5)
	networkTx := make(chan elevator.Elevator, 5)
	networkRx := make(chan elevator.Elevator, 5)

	

	go TransmitterPeers(15650, id, peerTxEnable)
	go ReceiverPeers(15650, peerUpdateCh)
	go TransmitterBcast(20025, networkTx)
	go ReceiverBcast(20025, networkRx)

	for {
		select {
		case p := <-peerUpdateCh:
			peerUpdate_ch <- p
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			
		case a := <-networkRx:
			if a.Id != id {

				networkUpdateRx_ch <- a

			}

		case i := <-networkUpdateTx_ch:
			networkTx <- i
		}
	}
}
