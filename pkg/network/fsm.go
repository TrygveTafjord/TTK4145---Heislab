package network

import (
	/*
	   	"flag"

	   "fmt"
	   "net/mail"
	   "os"
	*/
	"fmt"

	"project.com/pkg/elevator"
)

//"project.com/pkg/elevator"

const port_number int = 20007;




func fsm() {


	Msg_ch   	 	  := make(chan Msg)
	Confirm_ch   	  := make(chan string)
	ButtonEvent_ch    := make(chan elevator.ButtonEvent)
	ImAliveCounter_ch := make(chan string)

	go TransmitterBcast(port_number, ImAliveCounter_ch, Msg_ch)
	go ReceiverBcast(port_number, ImAliveCounter_ch, Msg_ch)
	

	orderCounter := 0

	for{
		select{
		case msg := <- Msg_ch:	
			if msg.NewInfo {

				//UppdateInfobank(msg)
				//UppdateNodeStatus(msg)

				orderCounter += 1;
				localIP, err := LocalIP()
				if err != nil {
					fmt.Println(err)
					localIP = "DISCONNECTED"
				}	
				Confirm_ch <- localIP		

				fmt.Print("new info recieved")
			}
		case btn := <- ButtonEvent_ch: 
			//msg := createMsg(); 
			//Msg_ch <- msg
			//enter_confirm_fsm()
			fmt.Print("Buttonevent happened: ", btn)
		
		case <- ImAliveCounter_ch:
	
			//msg := createMsg(); 
			TransmitterBcast(port_number, )
		
		}

		}

	}





	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
/* 	var id string
	flag.S		localIP, err := LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}ess ID)
/* 	if id == "" {
		localIP, err := LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	} */

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
/* 	peerUpdateCh := make(chan PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go TransmitterPeers(15647, id, peerTxEnable)
	go ReceiverPeers(15647, peerUpdateCh)
 */
	// We make channels for sending and receiving our custom data types
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	//go TransmitterBcast(16569, helloTx)
	//go ReceiverBcast(16569, helloRx)

	// The example message. We just send one of these every second.
	//go func() {
	//	helloMsg := HelloMsg{"Hello from " + id, 0}
	//	for {
	//		helloMsg.Iter++
	//		helloTx <- helloMsg
	//		time.Sleep(1 * time.Second)
	//	}
	//}()

	//fmt.Println("Started")
	//for {
	//	select {
	//	case p := <-peerUpdateCh:
	//		fmt.Printf("Peer update:\n")
	//		fmt.Printf("  Peers:    %q\n", p.Peers)
	//		fmt.Printf("  New:      %q\n", p.New)
	//		fmt.Printf("  Lost:     %q\n", p.Lost)
	//	
	//	case a := <-helloRx:
	//		fmt.Printf("Received: %#v\n", a)
	//	}
	//}

