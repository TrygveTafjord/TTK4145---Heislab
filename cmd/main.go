package main

import (
	//"time"
	"fmt"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
	//"project.com/pkg/network"
)

func main() {
	//elevator.Init("localhost:15657", 4)
	/*button_ch := make(chan elevator.ButtonEvent)
	floorSensor_ch := make(chan int)
	stopButton_ch := make(chan bool)
	obstruction_ch := make(chan bool)
  timer_ch := make(chan bool, 5)
	
	infoUpdate_ch := make(chan elevator.Elevator, 10)
	infoRecieved_ch := make(chan elevator.Elevator, 10)
	peerUpdate_ch := make(chan network.PeerUpdate, 10)

	go elevator.PollFloorSensor(floorSensor_ch)
	go elevator.PollButtons(button_ch)
	go elevator.PollStopButton(stopButton_ch)
	go elevator.PollObstructionSwitch(obstruction_ch)

	go elevator.FSM(button_ch, floorSensor_ch, stopButton_ch, obstruction_ch, timer_ch)

	go network.Network_fsm(infoUpdate_ch, infoRecieved_ch, peerUpdate_ch)

	Requests := [4][3]bool{
		{true, true, true},
		{true, true, true},
		{true, true, true},
		{true, true, true},
	}

	e := elevator.Elevator{"Ole er ikke pedo", 5, 69, elevator.MD_Down, Requests, elevator.EB_DoorOpen, 0.5}

	for {
		infoUpdate_ch <- e
		time.Sleep(2000 * time.Millisecond)
	}*/

	var reqs [4][3]bool 
	reqs[3][1] = true
	//reqs[3][1] = true
	testElevator := elevator.Elevator{
		Floor:     2,
		Dirn:      elevator.MD_Up,
		Requests:  reqs,
		Behaviour: elevator.EB_Moving,
	}
	testElevator1 := elevator.Elevator{
		Floor:     2,
		Dirn:      elevator.MD_Down,
		Requests:  reqs,
		Behaviour: elevator.EB_Moving,
	}
	testElevator2 := elevator.Elevator{
		Floor:     1,
		Dirn:      elevator.MD_Up,
		Requests:  reqs,
		Behaviour: elevator.EB_Idle,
	}

	//var HRAMatrix [4][2]bool
	//var elevator_list []map[string]interface{}
	var result_bytes []byte

	list_of_elevs := []elevator.Elevator{
		testElevator, 
		testElevator1,
		testElevator2,
	}

	result_bytes = hallrequestassigner.CreateJSON(list_of_elevs...)
	var optimal_path map[string][4][2]bool
	optimal_path = hallrequestassigner.HallRequestAssigner(result_bytes)


	fmt.Print(string(result_bytes))
	fmt.Println()
	fmt.Println()
	for key, slice := range optimal_path {
		fmt.Printf("Key: %s, Values: ", key)
		// Iterating through the slice for each key
		for _, pair := range slice {
			// Printing each boolean pair
			fmt.Printf("[%t, %t] ", pair[0], pair[1])
		}
		fmt.Println() // Newline for each key
	}

}
