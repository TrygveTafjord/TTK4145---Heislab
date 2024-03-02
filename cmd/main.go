package main

import (
	"time"
	"fmt"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
)

func main() {
	//elevator.Init("localhost:15657", 4)

	/*Button_ch := make(chan elevator.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)
	Timer_ch := make(chan bool,5)

	go elevator.PollFloorSensor(Floor_sensor_ch)
	go elevator.PollButtons(Button_ch)
	go elevator.PollStopButton(Stop_button_ch)
	go elevator.PollObstructionSwitch(Obstruction_ch)
	go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)*/
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

	E1map := hallrequestassigner.CreateStateMap(testElevator)
	E2map := hallrequestassigner.CreateStateMap(testElevator1)
	E3map := hallrequestassigner.CreateStateMap(testElevator2)

	jsonBytes := hallrequestassigner.CreateMasterJSON(testElevator, E1map, E2map, E3map)

	returnmap := new(map[string][][2]bool)
	*returnmap = hallrequestassigner.HallRequestAssigner(jsonBytes); 

	fmt.Printf("output: \n")
	for k, v := range *returnmap {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}

}
