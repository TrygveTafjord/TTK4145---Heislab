package network

import (
	"fmt"

	//"time"
	
)

func Network_fsm(networkUpdateTx_ch chan Msg, networkUpdateRx_ch chan Msg, peerUpdate_ch chan PeerUpdate) {

	id, err := LocalIP()
	if err != nil {
		fmt.Printf("could not get IP")
	}

	peerUpdateCh := make(chan PeerUpdate,5)
	peerTxEnable := make(chan bool,5)
	networkTx := make(chan Msg, 5)
	networkRx := make(chan Msg, 5)

	

	go TransmitterPeers(15651, id, peerTxEnable)
	go ReceiverPeers(15651, peerUpdateCh)
	go TransmitterBcast(20026, networkTx)
	go ReceiverBcast(20026, networkRx)

	for {
		select {
		case p := <-peerUpdateCh:
			peerUpdate_ch <- p
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			
		case a := <-networkRx:
			if a.Elevator.Id != id {

				networkUpdateRx_ch <- a

			}

		case i := <-networkUpdateTx_ch:
			networkTx <- i
		}
	}
}
