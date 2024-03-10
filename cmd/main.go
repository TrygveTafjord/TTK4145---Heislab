package main

import (
	"fmt"
	"os"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
	"project.com/pkg/network"
)

func main() {
	args := os.Args
	fmt.Println(args)
	port := args[1]
	elevator.Init("localhost:"+port, 4)

	id, err := network.LocalIP()
	if err != nil {
		fmt.Printf("could not get IP")
	}

	elevStatusUpdate_ch := make(chan elevator.Elevator, 50)
	networkUpdateTx_ch := make(chan network.Msg, 50)
	networkUpdateRx_ch := make(chan network.Msg, 50)
	peerUpdate_ch := make(chan network.PeerUpdate,50)



	go elevator.FSM(elevStatusUpdate_ch)
	go infobank.Infobank_FSM(elevStatusUpdate_ch, networkUpdateTx_ch, networkUpdateRx_ch, peerUpdate_ch)
	go network.Network_fsm(networkUpdateTx_ch, networkUpdateRx_ch, peerUpdate_ch)

	elevator.ElevatorInit(elevStatusUpdate_ch, id)

	for {
		time.Sleep(2000 * time.Millisecond)
	}
}
