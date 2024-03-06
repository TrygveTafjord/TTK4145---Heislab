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

	elevStatusUpdate_ch := make(chan elevator.Elevator, 5)

	networkUpdateTx_ch := make(chan elevator.Elevator, 5)
	networkUpdateRx_ch := make(chan elevator.Elevator, 5)
	updatePeers_ch := make(chan network.PeerUpdate)

	go elevator.FSM(elevStatusUpdate_ch)
	go infobank.Infobank_FSM(elevStatusUpdate_ch, networkUpdateTx_ch, networkUpdateRx_ch)
	go network.Network_fsm(networkUpdateTx_ch, networkUpdateRx_ch, updatePeers_ch)

	for {
		time.Sleep(2000 * time.Millisecond)
	}
}
