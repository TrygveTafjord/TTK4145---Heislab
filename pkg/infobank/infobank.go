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
			//sett cab-lys, dersom det er
			//sørg for at vi har nyeste status til den lokale heisen

			networkUpdateTx_ch <- thisElevator

			elevatorMap[thisElevator.Id] = thisElevator
			newAssignmentsMap := hallrequestassigner.AssignHallRequests(elevatorMap)
			setGlobalLights(newAssignmentsMap, &thisElevator)
			elevatorMap = updateMap(newAssignmentsMap, elevatorMap)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			elevStatusUpdate_ch <- thisElevator

		case newState := <-elevStatusUpdate_ch:
			newState.Id = thisElevator.Id
			elevatorMap[thisElevator.Id] = newState
			thisElevator = newState
			//når trenger vi å gi en ny state?
			networkUpdateTx_ch <- thisElevator

		case recievedElevator := <-networkUpdateRx_ch:

			if recievedElevator.OrderClearedCounter > thisElevator.OrderClearedCounter {
				thisElevator = handleRecievedOrderCompleted(recievedElevator, thisElevator)
				thisElevator.OrderClearedCounter = recievedElevator.OrderClearedCounter
				elevatorMap[thisElevator.Id] = thisElevator
			}
			//eneste som har skjedd er at globallights er oppdatert
			elevatorMap[recievedElevator.Id] = recievedElevator

			//Lag funksjon som sjekker om vi har en ny assignment, dersom det er tilfellet->oppdater fsm
			var newAssignmentsMap map[string][4][2]bool = hallrequestassigner.AssignHallRequests(elevatorMap)

			setGlobalLights(newAssignmentsMap, &thisElevator)

			elevatorMap = updateMap(newAssignmentsMap, elevatorMap)
			thisElevator.Requests = elevatorMap[Ip].Requests

			elevStatusUpdate_ch <- thisElevator
		}
	}
}

func setGlobalLights(newAssignmentsMap map[string][4][2]bool, e *elevator.Elevator) {
	for _, value := range newAssignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.GlobalLights[i][j] = (e.GlobalLights[i][j] || value[i][j])
				e.GlobalLights[i][elevator.BT_Cab] = e.Requests[i][elevator.BT_Cab]
			}
		}
	}
}

func handleRecievedOrderCompleted(recievedElevator elevator.Elevator, thisElevator elevator.Elevator) elevator.Elevator {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			thisElevator.GlobalLights[i][j] = thisElevator.GlobalLights[i][j] && recievedElevator.GlobalLights[i][j]
		}
	}
	return thisElevator
}

func updateMap(newAssignmentsMap map[string][4][2]bool, elevatorMap map[string]elevator.Elevator) map[string]elevator.Elevator {
	returnMap := make(map[string]elevator.Elevator)

	for id, requests := range newAssignmentsMap {
		tempElev := elevatorMap[id]
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				tempElev.Requests[i][j] = requests[i][j]
			}
		}
		returnMap[id] = tempElev
	}
	return returnMap
}
