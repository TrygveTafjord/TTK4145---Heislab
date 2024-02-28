package main

import (
	//"time"

	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"project.com/pkg/elevator"
	"project.com/pkg/hallreqass"
)

type Output struct {
	ResultField1 [][]bool
	ResultField2 [][]bool
	ResultField3 [][]bool
}

func main() {
	// elevator.Init("localhost:15657", 4)

	// Button_ch := make(chan elevator.ButtonEvent)
	// Floor_sensor_ch := make(chan int)
	// Stop_button_ch := make(chan bool)
	// Obstruction_ch := make(chan bool)
	// Timer_ch := make(chan bool,5)

	var reqs [4][3]bool
	testElevator := elevator.Elevator{
		Floor:     2,
		Dirn:      elevator.MD_Up,
		Requests:  reqs,
		Behaviour: elevator.EB_Moving,
	}
	testElevator1 := elevator.Elevator{
		Floor:     3,
		Dirn:      elevator.MD_Up,
		Requests:  reqs,
		Behaviour: elevator.EB_Moving,
	}
	testElevator2 := elevator.Elevator{
		Floor:     1,
		Dirn:      elevator.MD_Down,
		Requests:  reqs,
		Behaviour: elevator.EB_Idle,
	}

	E1map := hallreqass.CreateStateMap(testElevator)
	E2map := hallreqass.CreateStateMap(testElevator1)
	E3map := hallreqass.CreateStateMap(testElevator2)

	jsonData := hallreqass.CreateMasterJSON(testElevator, E1map, E2map, E3map)

	// Prepare the D program command
	cmd := exec.Command("./hall_request_assigner") // Adjust the executable name/path as necessary

	// Set the stdin to be our input JSON
	cmd.Stdin = bytes.NewReader(jsonData)

	// Capture the stdout for the JSON output
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Run the D program
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running D program: %v", err)
	}

	// The D program has completed, and its JSON output should now be in stdout
	// Unmarshal the JSON output into an Output struct
	var outputData Output
	if err := json.Unmarshal(stdout.Bytes(), &outputData); err != nil {
		fmt.Printf("Error unmarshaling output JSON: %v", err)
	}

	// Use the output data as needed
	fmt.Printf("Received output: %+v\n", outputData)

	// var prettyJSON bytes.Buffer

	// // Use json.Indent to format the JSON
	// err := json.Indent(&prettyJSON, jsonData, "", "    ")
	// if err != nil {
	// 	fmt.Println("Error pretty-printing JSON:", err)
	// 	return
	// }

	// // Print the pretty-printed JSON
	// fmt.Println(prettyJSON.String())

	// go elevator.PollFloorSensor(Floor_sensor_ch)
	// go elevator.PollButtons(Button_ch)
	// go elevator.PollStopButton(Stop_button_ch)
	// go elevator.PollObstructionSwitch(Obstruction_ch)
	// go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)

	// for {
	// 	time.Sleep(100 * time.Millisecond)
	// }

}
