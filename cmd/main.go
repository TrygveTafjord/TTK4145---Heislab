package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/network"
)

type HelloMsg struct {
	Id        string
	Counter   int
	NewInfo   bool
	Floor     int
	Dirn      elevator.MotorDirection
	Requests  [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	Behaviour elevator.ElevatorBehaviour
}

func main() {
	fmt.Printf("ole, men øverst ")
	elevator.Init("localhost:15657", 4)
	fmt.Printf("ole ")

	Button_ch := make(chan elevator.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)
	Timer_ch := make(chan bool, 5)

	go elevator.PollFloorSensor(Floor_sensor_ch)
	go elevator.PollButtons(Button_ch)
	go elevator.PollStopButton(Stop_button_ch)
	go elevator.PollObstructionSwitch(Obstruction_ch)

	go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)

	/* 	for {
		time.Sleep(100 * time.Millisecond)
	} */

	// We make channels for sending and receiving our custom data types
	/* 	helloTx := make(chan network.Msg)
	   	helloRx := make(chan network.Msg)
	   	// ... and start the transmitter/receiver pair on some port
	   	// These functions can take any number of channels! It is also possible to
	   	//
	   	//	start multiple transmitters/receivers on the same port.
	   	go network.TransmitterBcast(16569, helloTx)
	   	go network.ReceiverBcast(16569, helloRx)

	   	// The example message. We just send one of these every second.
	   	go func() {
	   		Requests := [4][3]bool{
	   			{true, true, true},
	   			{true, true, true},
	   			{true, true, true},
	   			{true, true, true},
	   		}
	   		helloMsg := network.Msg{"Dette er id",
	   			69,
	   			true,
	   			420,
	   			elevator.MD_Up,
	   			Requests,
	   			elevator.EB_Idle}
	   		for {
	   			helloTx <- helloMsg
	   			time.Sleep(1 * time.Second)
	   		}
	   	}()

	   	fmt.Println("Started")

	   	for {
	   		time.Sleep(100 * time.Millisecond)
	   	}
	*/

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	Requests := [4][3]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
		{true, true, true},
	}

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := network.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan network.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go network.TransmitterPeers(15647, id, peerTxEnable)
	go network.ReceiverPeers(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go network.TransmitterBcast(20007, helloTx)
	go network.ReceiverBcast(20007, helloRx)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Ole er ikke pedo", 0, true, 69, elevator.MD_Down, Requests, elevator.EB_DoorOpen}
		for {
			helloMsg.Counter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
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
		}
	}
}
