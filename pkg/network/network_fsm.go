package network

func Network(
	initialize_ch 				  chan string,
	newRequestToInfobank_ch 	  chan NewRequest,
	newRequestFromInfobank_ch     chan NewRequest,
	confirmationToInfobank_ch     chan Confirm,
	confirmationFromInfobank_ch   chan Confirm,
	obstructedToInfobank_ch       chan Obstructed,
	obstructedFromInfobank_ch     chan Obstructed,
	stateUpdateToInfobank_ch      chan StateUpdate,
	stateUpdateFromInfobank_ch    chan StateUpdate,
	requestClearedToInfobank_ch   chan RequestCleared,
	requestClearedFromInfobank_ch chan RequestCleared,
	periodicInfobankToNetwork_ch  chan Periodic,
	periodicNetworkToInfobank_ch  chan Periodic,
	peerUpdate_ch 				  chan PeerUpdate) {

	id := <-initialize_ch

	const (
		buffSize = 5
	)

	newRequestTx_ch 	:= make(chan NewRequest, buffSize)
	newRequestRx_ch 	:= make(chan NewRequest, buffSize)
	confirmRequestTx_ch := make(chan Confirm, buffSize)
	confirmRequestRx_ch := make(chan Confirm, buffSize)
	obstructedTx_ch 	:= make(chan Obstructed, buffSize)
	obstructedRx_ch 	:= make(chan Obstructed, buffSize)
	stateUpdateTx_ch 	:= make(chan StateUpdate, buffSize)
	stateUpdateRx_ch 	:= make(chan StateUpdate, buffSize)
	requestClearedTx_ch := make(chan RequestCleared, buffSize)
	requestClearedRx_ch := make(chan RequestCleared, buffSize)
	periodicTx_ch 		:= make(chan Periodic, buffSize)
	periodicRx_ch 		:= make(chan Periodic, buffSize)
  	peerUpdateCh 		:= make(chan PeerUpdate, buffSize)
	peerTxEnable 		:= make(chan bool, buffSize)

	go TransmitterPeers(15653, id, peerTxEnable)
	go ReceiverPeers(15653, peerUpdateCh)
	go TransmitterBcast(20029, newRequestTx_ch, confirmRequestTx_ch, obstructedTx_ch, stateUpdateTx_ch, requestClearedTx_ch, periodicTx_ch)
	go ReceiverBcast(20029, newRequestRx_ch, confirmRequestRx_ch, obstructedRx_ch, stateUpdateRx_ch, requestClearedRx_ch, periodicRx_ch)


	for {
		select {
		case p := <-peerUpdateCh:
			peerUpdate_ch <- p

		case msg := <-newRequestRx_ch:
			if msg.Id != id {
				newRequestToInfobank_ch <- msg
			}
		case msg := <-newRequestFromInfobank_ch:
			newRequestTx_ch <- msg

		case msg := <-stateUpdateRx_ch:
			if msg.Id != id {
				stateUpdateToInfobank_ch <- msg
			}
		case msg := <-stateUpdateFromInfobank_ch:
			stateUpdateTx_ch <- msg

		case msg := <-requestClearedFromInfobank_ch:
			requestClearedTx_ch <- msg

		case msg := <-requestClearedRx_ch:
			if msg.Id != id {
				requestClearedToInfobank_ch <- msg
			}
		case msg := <-obstructedFromInfobank_ch:
			obstructedTx_ch <- msg

		case msg := <-obstructedRx_ch:
			if msg.Id != id {
				obstructedToInfobank_ch <- msg
			}
		case msg := <-confirmRequestRx_ch:
			if msg.Id != id {
				confirmationToInfobank_ch <- msg
			}
		case msg := <-confirmationFromInfobank_ch:
			confirmRequestTx_ch <- msg

		case msg := <-periodicRx_ch:
			if msg.Id != id {
				periodicNetworkToInfobank_ch <- msg
			}
		case msg := <-periodicInfobankToNetwork_ch:
			periodicTx_ch <- msg

		}
	}
}
