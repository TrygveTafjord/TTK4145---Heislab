package main

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/network"
)

func main() {
	elevator.Init("localhost:15657", 4)

	Button_ch := make(chan elevator.ButtonEvent)
	Floor_sensor_ch := make(chan int)
	Stop_button_ch := make(chan bool)
	Obstruction_ch := make(chan bool)
	Timer_ch := make(chan bool, 5)

	go elevator.PollFloorSensor(Floor_sensor_ch)
	go elevator.PollButtons(Button_ch)
	go elevator.PollStopButton(Stop_button_ch)
	go elevator.PollObstructionSwitch(Obstruction_ch)
	go elevator.FSM(Button_ch, Floor_sensor_ch, Stop_button_ch, Obstruction_ch, Timer_ch)

	/* 	for {
		time.Sleep(100 * time.Millisecond)
	} */

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan network.Msg)
	helloRx := make(chan network.Msg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//
	//	start multiple transmitters/receivers on the same port.
	go network.TransmitterBcast(16569, helloTx)
	go network.ReceiverBcast(16569, helloRx)

	// The example message. We just send one of these every second.
	go func() {
		Requests := [4][3]bool{
			{true, true, true},
			{true, true, true},
			{true, true, true},
			{true, true, true},
		}
		helloMsg := network.Msg{"Dette er id",
			69,
			true,
			420,
			elevator.MD_Up,
			Requests,
			elevator.EB_Idle}
		for {
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")

}
