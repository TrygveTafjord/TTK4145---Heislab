package main

import (
	//"time"

	//"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"project.com/pkg/elevator"
	"project.com/pkg/hallreqass"
)

func main() {
	// elevator.Init("localhost:15657", 4)

	// Button_ch := make(chan elevator.ButtonEvent)
	// Floor_sensor_ch := make(chan int)
	// Stop_button_ch := make(chan bool)
	// Obstruction_ch := make(chan bool)
	// Timer_ch := make(chan bool,5)

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

	E1map := hallreqass.CreateStateMap(testElevator)
	E2map := hallreqass.CreateStateMap(testElevator1)
	E3map := hallreqass.CreateStateMap(testElevator2)

	jsonBytes := hallreqass.CreateMasterJSON(testElevator, E1map, E2map, E3map)

	ret, err := exec.Command("./hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	// Print the pretty-printed JSON

	// go elevator.PollFloorSensor(Floor_sensor_ch)
	// go elevator.PollButtons(Button_ch)
	// go elevator.PollStopButton(Stop_button_ch)
	// go elevator.PollObstructionSwitch(Obstruction_ch)
	// go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)

	// for {
	// 	time.Sleep(100 * time.Millisecond)
	// }

}
