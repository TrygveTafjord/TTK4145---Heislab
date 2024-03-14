package network

import (
	"fmt"
	//"time"
)

func Network(
	initialize_ch					chan string,	
	newRequestToInfobank_ch 		chan NewRequest,
	newRequestFromInfobank_ch 		chan NewRequest,
	obstructedToInfobank_ch 		chan Obstructed,
	obstructedFromInfobank_ch 		chan Obstructed,
	stateUpdateToInfobank_ch 		chan StateUpdate,
	stateUpdateFromInfobank_ch 		chan StateUpdate,
	requestClearedToInfobank_ch 	chan RequestCleared,
	requestClearedFromInfobank_ch   chan RequestCleared,
	peerUpdate_ch 					chan PeerUpdate,) {


	id := <- initialize_ch	

	const (BUFF_SIZE = 5)

	newRequestTx_ch 	:= make(chan NewRequest, BUFF_SIZE)
	newRequestRx_ch 	:= make(chan NewRequest, BUFF_SIZE)
	obstructedTx_ch 	:= make(chan Obstructed, BUFF_SIZE)
	obstructedRx_ch 	:= make(chan Obstructed, BUFF_SIZE)
	stateUpdateTx_ch 	:= make(chan StateUpdate, BUFF_SIZE)
	stateUpdateRx_ch 	:= make(chan StateUpdate, BUFF_SIZE)
	requestClearedTx_ch := make(chan RequestCleared, BUFF_SIZE)
	requestClearedRx_ch := make(chan RequestCleared, BUFF_SIZE)

	peerUpdateCh := make(chan PeerUpdate, 5)
	peerTxEnable := make(chan bool, 5)

	go TransmitterPeers(15653, id, peerTxEnable)
	go ReceiverPeers(15653, peerUpdateCh)
	go TransmitterBcast(20029, newRequestTx_ch, obstructedTx_ch, stateUpdateTx_ch, requestClearedTx_ch)
	go ReceiverBcast(20029, newRequestRx_ch, obstructedRx_ch, stateUpdateRx_ch, requestClearedRx_ch)

	for {
		select {
		case p := <-peerUpdateCh:
			peerUpdate_ch <- p
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case msg := <-newRequestRx_ch:
			if msg.Id != id {
				newRequestToInfobank_ch <- msg
			}
		case msg := <- newRequestFromInfobank_ch:
			newRequestTx_ch <- msg
		
		case msg := <- stateUpdateRx_ch:
			if msg.Id != id {
				stateUpdateToInfobank_ch <- msg
			}
		case msg := <- stateUpdateFromInfobank_ch:
			stateUpdateTx_ch <- msg
		
		case msg := <- requestClearedFromInfobank_ch:
			requestClearedTx_ch <- msg 
		
		case msg := <- requestClearedRx_ch:
			if msg.Id != id {
				requestClearedToInfobank_ch <- msg
			}
		case msg := <- obstructedFromInfobank_ch:
			obstructedTx_ch <- msg

		case msg := <- obstructedRx_ch:
			if msg.Id != id {
				obstructedToInfobank_ch <- msg
			}
			} 
	}
}