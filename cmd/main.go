package main

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/network"
)

func main() {
	fmt.Printf("ole, men øverst ")
	elevator.Init("localhost:15657", 4)
	fmt.Printf("ole ")

	Button_ch := make(chan elevator.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)
	Timer_ch := make(chan bool, 5)


	infoUpdate_ch:= make(chan elevator.Elevator)
	info_Recieved_ch:= make(chan elevator.Elevator)
	Peer_update_ch := make(chan string)


	go elevator.PollFloorSensor(Floor_sensor_ch)
	go elevator.PollButtons(Button_ch)
	go elevator.PollStopButton(Stop_button_ch)
	go elevator.PollObstructionSwitch(Obstruction_ch)

	go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)

	go network.Network_fsm(infoUpdate_ch,info_Recieved_ch,Peer_update_ch)


	Requests := [4][3]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
		{true, true, true},
	}


	e := elevator.Elevator{"Ole er ikke pedo", 5,69, elevator.MD_Down, Requests, elevator.EB_DoorOpen,0.5}

	for {
		infoUpdate_ch <- e
		time.Sleep(2000 * time.Millisecond)
	}

}
