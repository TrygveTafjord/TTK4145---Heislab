package infobank

import (
	"fmt"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
	"project.com/pkg/network"
)

func Infobank_FSM(
	elevStatusUpdate_ch chan elevator.Elevator,
	networkUpdateTx_ch chan elevator.Elevator,
	networkUpdateRx_ch chan elevator.Elevator,
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

	for {
		select {
		case btn := <-button_ch:

			thisElevator.Requests[btn.Floor][btn.Button] = true
			thisElevator.GlobalLights[btn.Floor][btn.Button] = true

			elevatorMap[thisElevator.Id] = thisElevator
			var newAssignmentsMap map[string][4][2]bool = hallrequestassigner.AssignHallRequests(elevatorMap)
			thisElevator.GlobalLights = setGlobalLights(newAssignmentsMap, thisElevator)

			for i := 0; i < elevator.N_FLOORS; i++ {
				for j := 0; j < elevator.N_BUTTONS-1; j++ {
					thisElevator.Requests[i][j] = newAssignmentsMap["id_1"][i][j] // endre "id_1" til "thisElevator.Id" men i tørr faen ikke røre Ole sin kode
				}
			}
			elevStatusUpdate_ch <- thisElevator
			networkUpdateTx_ch <- thisElevator

		case newState := <-elevStatusUpdate_ch:
			newState.Id = thisElevator.Id
			elevatorMap[thisElevator.Id] = newState
			thisElevator = newState
			networkUpdateTx_ch <- thisElevator

		case recievedElevator := <-networkUpdateRx_ch:

			if recievedElevator.OrderClearedCounter > thisElevator.OrderClearedCounter {
				thisElevator = handleRecievedOrderCompleted(recievedElevator, thisElevator)
			}

			recievedElevator.GlobalLights = thisElevator.GlobalLights

			elevatorMap[recievedElevator.Id] = recievedElevator

			var newAssignmentsMap map[string][4][2]bool = hallrequestassigner.AssignHallRequests(elevatorMap)
			
			thisElevator.GlobalLights = setGlobalLights(newAssignmentsMap, thisElevator)

			for i := 0; i < elevator.N_FLOORS; i++ {
				for j := 0; j < elevator.N_BUTTONS-1; j++ {
					thisElevator.Requests[i][j] = newAssignmentsMap[thisElevator.Id][i][j] // endre "id_1" til "thisElevator.Id"
				}
			}
			elevStatusUpdate_ch <- thisElevator
		}
	}
}

func setGlobalLights(newAssignmentsMap map[string][4][2]bool, e elevator.Elevator) [4][3]bool {
	for _, value := range newAssignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.GlobalLights[i][j] = (e.GlobalLights[i][j] || value[i][j])
			}
		}
	}
	return e.GlobalLights
}

func handleRecievedOrderCompleted(recievedElevator elevator.Elevator, thisElevator elevator.Elevator) elevator.Elevator {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			thisElevator.GlobalLights[i][j] = thisElevator.GlobalLights[i][j] && recievedElevator.GlobalLights[i][j]
		}
	}
	return thisElevator
}
