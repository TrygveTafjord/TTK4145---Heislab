package main

import (
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
	"project.com/pkg/network"
)

func main() {
	//elevator.Init("localhost:15657", 4)
	button_ch := make(chan elevator.ButtonEvent)
	floorSensor_ch := make(chan int)
	stopButton_ch := make(chan bool)
	obstruction_ch := make(chan bool)
	timer_ch := make(chan bool, 5)

	infoUpdate_ch := make(chan elevator.Elevator, 10)
	infoRecieved_ch := make(chan elevator.Elevator, 10)

	peerUpdate_ch := make(chan network.PeerUpdate, 10)
	assigner_ch := make(chan map[string]elevator.Elevator, 10)
	newStatus_ch := make(chan elevator.Elevator, 10)

	newAssignments_ch := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool)

	go elevator.PollFloorSensor(floorSensor_ch)
	go elevator.PollButtons(button_ch)
	go elevator.PollStopButton(stopButton_ch)
	go elevator.PollObstructionSwitch(obstruction_ch)

	go elevator.FSM(newAssignments_ch, floorSensor_ch, stopButton_ch, obstruction_ch, timer_ch)

	go network.Network_fsm(infoUpdate_ch, infoRecieved_ch, peerUpdate_ch)

	go infobank.Infobank_FSM(button_ch, newStatus_ch, infoUpdate_ch, infoRecieved_ch, peerUpdate_ch, assigner_ch)

	Requests := [4][3]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
		{true, true, true},
	}

	e := elevator.Elevator{"Testmelding", 5, 69, elevator.MD_Down, Requests, elevator.EB_DoorOpen, 0.5}

	for {
		infoUpdate_ch <- e
		time.Sleep(2000 * time.Millisecond)
	}

}
