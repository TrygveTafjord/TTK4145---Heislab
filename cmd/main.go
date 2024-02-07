package main

import (
	"time"

	"OTP.com/Heis2e/pkg/elevio"
	"OTP.com/Heis2e/pkg/fsm"
)

func main() {
	elevio.Init("localhost:15657", 4)


	// noen channels
	Button_ch := make(chan elevio.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)

	go elevio.PollFloorSensor(Floor_sensor_ch)
	go elevio.PollButtons(Button_ch)
	go elevio.PollStopButton(Stop_button_ch)
	go elevio.PollObstructionSwitch(Obstruction_ch)
	go fsm.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch)
	for {
		time.Sleep(100 * time.Millisecond)
	}

}
