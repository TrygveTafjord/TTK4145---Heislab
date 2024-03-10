package infobank

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
	"project.com/pkg/network"
)

func Infobank_FSM(
	elevStatusUpdate_ch chan elevator.Elevator,
	networkUpdateTx_ch chan network.Msg,
	networkUpdateRx_ch chan network.Msg,
	peerUpdate_ch chan network.PeerUpdate,
) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	periodicUpdate_ch := make(chan bool, 50)

	go elevator.PollButtons(button_ch)
	go PeriodicUpdate(periodicUpdate_ch)

	elevatorMap := make(map[string]elevator.Elevator)
	var thisElevator elevator.Elevator = <-elevStatusUpdate_ch

	elevatorMap[thisElevator.Id] = thisElevator

	for {
		select {
		case btn := <-button_ch:

			thisElevator.Requests[btn.Floor][btn.Button] = true
			thisElevator.OrderCounter++

			msg := network.Msg{
				MsgType:  network.NewOrder,
				Elevator: thisElevator,
			}

			networkUpdateTx_ch <- msg

			handleBtnPress(elevatorMap, &thisElevator)

			elevStatusUpdate_ch <- thisElevator

		case newState := <-elevStatusUpdate_ch:

			storeFsmUpdate(elevatorMap, &thisElevator, &newState)
			msgType := calculateMsgType(newState, thisElevator)

			msg := network.Msg{
				MsgType:  msgType,
				Elevator: thisElevator,
			}

			networkUpdateTx_ch <- msg

		case Msg := <-networkUpdateRx_ch:

			switch Msg.MsgType {

			case network.NewOrder:
				handleNewOrder(elevatorMap, &Msg.Elevator, &thisElevator)
				elevStatusUpdate_ch <- thisElevator

			case network.OrderCompleted:
				handleOrderCompleted(elevatorMap, &Msg.Elevator, &thisElevator)
				elevStatusUpdate_ch <- thisElevator

			case network.StateUpdate:
				// Midlertidig løsning for packetloss ved slukking av lys, dette kan være kilde til bus, men funker fett nå (merk at vi alltid setter order counter til å være det vi får inn)
				if Msg.Elevator != elevatorMap[Msg.Elevator.Id] {
					handleOrderCompleted(elevatorMap, &Msg.Elevator, &thisElevator)
					elevStatusUpdate_ch <- thisElevator
				}
				elevatorMap[Msg.Elevator.Id] = Msg.Elevator
			}

		case <-periodicUpdate_ch:
			msg := network.Msg{
				MsgType:  network.StateUpdate,
				Elevator: thisElevator,
			}
			networkUpdateTx_ch <- msg

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) != 0 {
				hallRequests := make(map[string][4][2]bool)
				handlePeerupdate(peerUpdate, &thisElevator, &elevatorMap, &hallRequests)
				hallRequestsMap := hallrequestassigner.AssignHallRequests(elevatorMap)
				setLightMatrix(hallRequestsMap, &thisElevator)
				thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
				elevStatusUpdate_ch <- thisElevator
				fmt.Printf("\n", elevatorMap)
			}
		}
	}
}

func setElevatorMap(newAssignmentsMap map[string][4][2]bool, elevatorMap *map[string]elevator.Elevator) {

	for id, requests := range newAssignmentsMap {
		tempElev := (*elevatorMap)[id]
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				tempElev.Requests[i][j] = requests[i][j]
			}
		}
		(*elevatorMap)[id] = tempElev
	}
}

func setAllLights(e *elevator.Elevator) {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			elevator.SetButtonLamp(elevator.ButtonType(btn), floor, e.Lights[floor][btn])
		}
	}
}

func PeriodicUpdate(periodicUpdate_ch chan bool) {
	for {
		time.Sleep(1000 * time.Millisecond)
		periodicUpdate_ch <- true
	}
}

func handlePeerupdate(peerUpdate network.PeerUpdate, thisElevator *elevator.Elevator, elevatorMap *map[string]elevator.Elevator, newAssignmentsMap *map[string][4][2]bool) {

	for i := 0; i < len(peerUpdate.Lost); i++ {
		for j := 0; j < elevator.N_FLOORS; j++ {
			for k := 0; k < elevator.N_BUTTONS-1; k++ {
				thisElevator.Requests[j][k] = thisElevator.Requests[j][k] || (*elevatorMap)[peerUpdate.Lost[i]].Requests[j][k]
			}
		}
		delete(*elevatorMap, peerUpdate.Lost[i])
		delete(*newAssignmentsMap, peerUpdate.Lost[i])
		thisElevator.OrderCounter++
	}
	(*elevatorMap)[thisElevator.Id] = *thisElevator

	//Nå må vi redistrubiere requests og fjerne
}

func handleBtnPress(elevatorMap map[string]elevator.Elevator, e *elevator.Elevator) {
	elevatorMap[e.Id] = *e
	hallRequestsMap := hallrequestassigner.AssignHallRequests(elevatorMap)

	setElevatorMap(hallRequestsMap, &elevatorMap)
	setElevatorAsignments(elevatorMap, e)
	setLightMatrix(hallRequestsMap, e)

}

func handleNewOrder(elevatorMap map[string]elevator.Elevator, recievedElevator *elevator.Elevator, thisElevator *elevator.Elevator) {
	thisElevator.OrderCounter = recievedElevator.OrderCounter

	elevatorMap[recievedElevator.Id] = *recievedElevator
	hallRequestsMap := hallrequestassigner.AssignHallRequests(elevatorMap)

	setElevatorMap(hallRequestsMap, &elevatorMap)
	setElevatorAsignments(elevatorMap, thisElevator)
	setLightMatrix(hallRequestsMap, thisElevator)

	thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

}

func handleOrderCompleted(elevatorMap map[string]elevator.Elevator, recievedElevator *elevator.Elevator, thisElevator *elevator.Elevator) {
	elevatorMap[recievedElevator.Id] = *recievedElevator

	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			thisElevator.Lights[i][j] = thisElevator.Lights[i][j] && recievedElevator.Lights[i][j]
		}
	}

	thisElevator.OrderClearedCounter = recievedElevator.OrderClearedCounter

	elevatorMap[thisElevator.Id] = *thisElevator

}

func storeFsmUpdate(elevatorMap map[string]elevator.Elevator, oldState *elevator.Elevator, newState *elevator.Elevator) {
	if newState.Id != oldState.Id {
		fmt.Printf("error: trying to assign values to non similar Id's \n")
		return
	}

	elevatorMap[oldState.Id] = *newState
	*oldState = *newState
}

func calculateMsgType(newState elevator.Elevator, oldState elevator.Elevator) network.MsgType {
	if newState.OrderClearedCounter > oldState.OrderClearedCounter {
		return network.OrderCompleted
	}
	return network.StateUpdate
}

func setLightMatrix(newAssignmentsMap map[string][4][2]bool, e *elevator.Elevator) {
	for _, value := range newAssignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.Lights[i][j] = (e.Lights[i][j] || value[i][j])
				e.Lights[i][elevator.BT_Cab] = e.Requests[i][elevator.BT_Cab]
			}
		}
	}
}

func setElevatorAsignments(elevatorMap map[string]elevator.Elevator, e *elevator.Elevator) {
	e.Requests = elevatorMap[e.Id].Requests
}

//Vi må oppdage om en heis som ikke står i idle og ikke har tom request matrise og har stått stille lenge og behandle det
//Vi må også oppdage om en heis ikke har sendt melding på en stund
// Legg logikk for å melde seg selv av nettverk når vi oppdager vi er fucked
