package main

import (
	"fmt"

	"OTP.com/Heis2e/pkg/elevio"
)

func main() {
	elevio.Init("localhost:15657", 4)

	// noen channels
	Button_ch := make(chan elevio.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)

	go elevio.PollingGoRoutine(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch)

	// noen goroutines
	for {
		select {
		case a := <-Button_ch:
			fmt.Printf("Knapp: %+v\n", a)
		case a := <-Floor_sensor_ch:
			fmt.Printf("Gulv: %+v\n", a)
		case a := <-Stop_button_ch:
			fmt.Printf("Stop: %+v\n", a)
		case a := <-Obstruction_ch:
			fmt.Printf("Obstruksjon: %+v\n", a)
		}
	}

}
