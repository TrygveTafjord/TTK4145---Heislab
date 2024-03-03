package main

import (
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
)

func main() {
	elevator.Init("localhost:15657", 4)

	elevStatusUpdate_ch := make(chan elevator.Elevator, 5)

	go elevator.FSM(elevStatusUpdate_ch)

	go infobank.Infobank_FSM(elevStatusUpdate_ch)

	for {
		time.Sleep(2000 * time.Millisecond)
	}
}
