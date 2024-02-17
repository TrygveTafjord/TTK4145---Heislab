package main

import (
	"time"

	"OTP.com/Heis2e/pkg/elevator"
)

func main() {
	elevator.Init("localhost:15657", 4)

	Button_ch := make(chan elevator.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)
	Timer_ch := make(chan bool,5)

	go elevator.PollFloorSensor(Floor_sensor_ch)
	go elevator.PollButtons(Button_ch)
	go elevator.PollStopButton(Stop_button_ch)
	go elevator.PollObstructionSwitch(Obstruction_ch)
	go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)

	for {
		time.Sleep(100 * time.Millisecond)
	}

}
