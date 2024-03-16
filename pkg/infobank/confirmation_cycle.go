package infobank

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/network"
	"project.com/pkg/timer"
)

func confirmCycleNewAssignment(
	newRequestToNetwork_ch chan network.NewRequest,
	confirmRequest_ch chan network.Confirm,
	buttonEvent elevator.ButtonEvent,
	numElevators int,
	id string,
) bool {

	timeOut_ch := make(chan bool)

	const confirmTime float64 = 0.1
	confirmedNodes := make(map[string]bool)

	msg := network.NewRequest{
		Id:      id,
		Request: buttonEvent,
	}

	newRequestToNetwork_ch <- msg

	go timer.Run_timer(confirmTime, timeOut_ch)
	ticker := time.NewTicker(5 * time.Millisecond)

	for {
		select {

		case msg := <-confirmRequest_ch:

			if msg.PassWrd != id+fmt.Sprint(buttonEvent.Button)+fmt.Sprint(buttonEvent.Floor) {
				break
			}

			confirmedNodes[msg.Id] = true

			if len(confirmedNodes) == numElevators-1 {
				return true
			}

		case <-ticker.C:
			newRequestToNetwork_ch <- msg

		case <-timeOut_ch:
			return false
		}
	}
}

func confirmCycleObstructionState(
	obstructedToNetwork_ch chan network.Obstructed,
	confirmRequest_ch chan network.Confirm,
	obstruction bool,
	numElevators int,
	id string,
) {

	timeOut_ch := make(chan bool)

	const confirmTime float64 = 0.15
	confirmedNodes := make(map[string]bool)

	msg := network.Obstructed{
		Id:         id,
		Obstructed: obstruction,
	}

	obstructedToNetwork_ch <- msg

	go timer.Run_timer(confirmTime, timeOut_ch)
	ticker := time.NewTicker(5 * time.Millisecond)

	for {
		select {

		case msg := <-confirmRequest_ch:

			if msg.PassWrd != id {
				break
			}

			confirmedNodes[msg.Id] = true

			if len(confirmedNodes) == numElevators-1 {
				return
			}

		case <-ticker.C:
			obstructedToNetwork_ch <- msg

		case <-timeOut_ch:
			return
		}
	}
}
