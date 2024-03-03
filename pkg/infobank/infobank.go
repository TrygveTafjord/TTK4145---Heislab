package infobank

import (
	"fmt"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
	"project.com/pkg/network"
)

func Infobank_FSM(
	elevStatusUpdate_ch chan elevator.Elevator,
) {

	button_ch := make(chan elevator.ButtonEvent, 10)
	go elevator.PollButtons(button_ch)

	elevatorMap := make(map[string]elevator.Elevator)
	var thisElevator elevator.Elevator

	//wait for initial status update (potential bug: what if no status update is received?)
	thisElevator = <-elevStatusUpdate_ch

	Ip, e := network.LocalIP()
	if e != nil {
		fmt.Printf("could not get IP")
	}

	thisElevator.Id = Ip
	elevatorMap[thisElevator.Id] = thisElevator

	numBtnPresses := 0
	for {
		select {
		case btn := <-button_ch:
			numBtnPresses++
			thisElevator.Requests[btn.Floor][btn.Button] = true
			elevatorMap[thisElevator.Id] = thisElevator
			var newAssignmentsMap map[string][4][2]bool = hallrequestassigner.AssignHallRequests(elevatorMap)

			for i := 0; i < elevator.N_FLOORS; i++ {
				for j := 0; j < elevator.N_BUTTONS-1; j++ {
					thisElevator.Requests[i][j] = newAssignmentsMap["id_1"][i][j] // endre "id_1" til "thisElevator.Id" men i tørr faen ikke røre Ole sin kode
				}
			}
			elevStatusUpdate_ch <- thisElevator

		case newState := <-elevStatusUpdate_ch:
			if newState.Requests != thisElevator.Requests {
				newState.OrderClearedCounter++
			}
			newState.Id = thisElevator.Id
			elevatorMap[thisElevator.Id] = newState
			thisElevator = newState
		}
	}
}
