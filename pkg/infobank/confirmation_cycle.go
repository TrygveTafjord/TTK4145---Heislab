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
		Id:      id,
		Request: buttonEvent,
	}

	newRequestToNetwork_ch <- msg

	go timer.Run_timer(CONFIRM_TIME, timeOut_ch)
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
			fmt.Printf("sent a request!\n")
			newRequestToNetwork_ch <- msg

		case <-timeOut_ch:
			return false
		}
	}
}

// func confirmRemovedAssignments(
// 	requestClearedToNetwork_ch chan network.RequestCleared,
// 	confirmRequest_ch chan network.Confirm,
// 	buttonEvent []elevator.ButtonEvent,
// 	numElevators int,
// 	id string,
// ) bool {

// 	timeOut_ch := make(chan bool)

// 	const CONFIRM_TIME float64 = 0.1
// 	confirmedNodes := make(map[string]bool)

// 	msg := network.RequestCleared{
// 		Id:              id,
// 		ClearedRequests: buttonEvent,
// 	}

// 	requestClearedToNetwork_ch <- msg

// 	go timer.Run_timer(CONFIRM_TIME, timeOut_ch)
// 	ticker := time.NewTicker(20 * time.Millisecond)

// 	for {
// 		select {

// 		case msg := <-confirmRequest_ch:

// 			if msg.PassWrd != id+fmt.Sprint(buttonEvent[0].Button)+fmt.Sprint(buttonEvent[0].Floor) {
// 				break
// 			}

// 			confirmedNodes[msg.Id] = true

// 			if len(confirmedNodes) == numElevators-1 {
// 				return true
// 			}

// 		case <-ticker.C:
// 			requestClearedToNetwork_ch <- msg

// 		case <-timeOut_ch:
// 			return false
// 		}
// 	}
// }

// func msgConfirmed(
// 	network_ch chan interface{},
// 	confirmRequest_ch chan network.Confirm,
// 	numElevators int,
// 	passWord int,
// 	msgToNetwork interface{},
// ) bool {

// 	timeOut_ch := make(chan bool)

// 	const CONFIRM_TIME float64 = 0.1
// 	confirmedNodes := make(map[string]bool)

// 	network_ch <- msgToNetwork

// 	go timer.Run_timer(CONFIRM_TIME, timeOut_ch)
// 	ticker := time.NewTicker(20 * time.Millisecond)

// 	for {
// 		select {

// 		case msg := <-confirmRequest_ch:

// 			if msg.PassWrd != passWord {
// 				break
// 			}

// 			confirmedNodes[msg.Id] = true

// 			if len(confirmedNodes) == numElevators-1{
// 				return true
// 			}

// 		case <- ticker.C:
// 			network_ch <- msgToNetwork

// 		case <- timeOut_ch:
// 			return false
// 		}
// 	}
// }
