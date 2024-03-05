package main

import (
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
	"project.com/pkg/network"
)

func main() {
	elevator.Init("localhost:15657", 4)

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
