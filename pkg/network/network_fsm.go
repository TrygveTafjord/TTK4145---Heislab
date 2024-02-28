package network

import (
	"fmt"

	"time"

	"project.com/pkg/elevator"
)

func Network_fsm(infoUpdate chan Msg, Send chan Msg, UpdatePeers chan string) {

	Requests := [4][3]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
		{true, true, true},
	}
	counter := 0

	id, err := LocalIP()
	if err != nil {
		fmt.Printf("could not get IP")
	}

	peerUpdateCh := make(chan PeerUpdate)
	peerTxEnable := make(chan bool)
	helloTx := make(chan Msg)
	helloRx := make(chan Msg)

	go TransmitterPeers(15647, id, peerTxEnable)
	go ReceiverPeers(15647, peerUpdateCh)
	go TransmitterBcast(20007, helloTx)
	go ReceiverBcast(20007, helloRx)

	go func() {

		helloMsg := Msg{id, 0, true, 69, elevator.MD_Down, Requests, elevator.EB_DoorOpen}
		for {
			helloMsg.Counter++
			helloTx <- helloMsg
			time.Sleep(2 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
			//send_to_infobank()

		case i := <-infoUpdate:
			Hello := Msg{id, counter,true, i.Floor, i.Dirn, i.Requests, i.Behaviour}
			helloTx <- Hello
		}
	}

}