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
	thisElevator.OrderClearedCounter = 0
	thisElevator.OrderCounter = 0
	elevatorMap[thisElevator.Id] = thisElevator

	for {
		select {
		case btn := <-button_ch:

			thisElevator.Requests[btn.Floor][btn.Button] = true
			thisElevator.OrderCounter++
			// sørg for at vi har nyeste status fra andre heiser, dette krever hver gang vi har en statusoppdatering lokalt-> send det til alle andre
			networkUpdateTx_ch <- thisElevator

			//Wrap inn i update map funksjon
			elevatorMap[thisElevator.Id] = thisElevator
			newAssignmentsMap := hallrequestassigner.AssignHallRequests(elevatorMap)
			setLights(newAssignmentsMap, &thisElevator)
			elevatorMap = updateMap(newAssignmentsMap, elevatorMap)
			//wrap inn i update map funksjo

			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			elevStatusUpdate_ch <- thisElevator

		case newState := <-elevStatusUpdate_ch:

			//Oppdater bare det som fsm skal ha kjennskap om
			//Potensiell bug -> Ordercounter og ClearOrderCounter får feil verdi
			newState.Id = thisElevator.Id
			elevatorMap[thisElevator.Id] = newState
			thisElevator = newState
			networkUpdateTx_ch <- thisElevator

		case recievedElevator := <-networkUpdateRx_ch:

			//Sjekk om vi har fjernet en ny ordre og håndter det, FSM må vite om at vi har fjernet lys, men trenger vi da å få informasjon om at FSM har fjernet lys, ikke egt så det kan endre. Det blir ikke nødvendig hvis bugs av at vi oppdaterer infobank om at vi har fjernet lys?

			elevatorMap[recievedElevator.Id] = recievedElevator

			if recievedElevator.OrderClearedCounter > thisElevator.OrderClearedCounter {
				thisElevator = handleRecievedOrderCompleted(recievedElevator, thisElevator)
				thisElevator.OrderClearedCounter = recievedElevator.OrderClearedCounter
				elevatorMap[thisElevator.Id] = thisElevator
				elevStatusUpdate_ch <- thisElevator
			}

			//Er det noen tilfeller hvor vi ikke ønsker å oppdatere elevatormappet? Vi må annta at den inkommende meldingen har nyeste status om seg selv, men hva med f.eks global-lights?
			//Merk at vi gjør akkuratt det samme her som når vi får ett nytt knappetrykk
			//Lag funksjon som sjekker om vi har en ny assignment ., dersom det er tilfellet->oppdater fsm og øk ordercounter
			if recievedElevator.OrderCounter > thisElevator.OrderCounter {
				thisElevator.OrderCounter = recievedElevator.OrderCounter
				elevatorMap[thisElevator.Id] = thisElevator
				newAssignmentsMap := hallrequestassigner.AssignHallRequests(elevatorMap)
				setLights(newAssignmentsMap, &thisElevator)
				elevatorMap = updateMap(newAssignmentsMap, elevatorMap)
				thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
				elevStatusUpdate_ch <- thisElevator
			}

			//Legg inn logikk for å oppdatere periodisk oversikt (om noen har falt ut)
		}
	}
}

func setLights(newAssignmentsMap map[string][4][2]bool, e *elevator.Elevator) {
	for _, value := range newAssignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.Lights[i][j] = (e.Lights[i][j] || value[i][j])
				e.Lights[i][elevator.BT_Cab] = e.Requests[i][elevator.BT_Cab]
			}
		}
	}
}

func handleRecievedOrderCompleted(recievedElevator elevator.Elevator, thisElevator elevator.Elevator) elevator.Elevator {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			thisElevator.Lights[i][j] = thisElevator.Lights[i][j] && recievedElevator.Lights[i][j]
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
func handleNewOrder(elevatorMap map[string]elevator.Elevator, thisElevator *elevator.Elevator) {

}
