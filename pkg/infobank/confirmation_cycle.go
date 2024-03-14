package infobank

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/network"
	"project.com/pkg/timer"
)

func confirmNewAssignment(
	newRequestToNetwork_ch chan network.NewRequest,
	confirmRequest_ch chan network.Confirm,
	buttonEvent elevator.ButtonEvent,
	numElevators int,
	id string,
) bool {
	

	timeOut_ch := make(chan bool)

	const CONFIRM_TIME float64 = 0.1
	confirmedNodes := make(map[string]bool)
	
	msg := network.NewRequest{
		Id:		 	id,
		Request: 	buttonEvent,
	}

	newRequestToNetwork_ch <- msg

	go timer.Run_timer(CONFIRM_TIME, timeOut_ch)
	ticker := time.NewTicker(20 * time.Millisecond)

	for {
		select {

		case msg := <-confirmRequest_ch:

			if msg.PassWrd != id + fmt.Sprint(buttonEvent.Button) + fmt.Sprint(buttonEvent.Floor) {
				break
			}

			confirmedNodes[msg.Id] = true

			if len(confirmedNodes) == numElevators-1{
				return true
			}

		case <- ticker.C:
			newRequestToNetwork_ch <- msg

		case <- timeOut_ch:
			return false
		}
	}
}



