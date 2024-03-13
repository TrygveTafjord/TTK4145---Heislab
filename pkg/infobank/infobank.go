package infobank

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
	"project.com/pkg/network"
	//"project.com/pkg/timer"
)

func Infobank_FSM(
	toFSM_ch chan elevator.Elevator,
	fromFSM_ch chan elevator.Elevator,
	networkUpdateTx_ch chan network.Msg,
	networkUpdateRx_ch chan network.Msg,
	peerUpdate_ch chan network.PeerUpdate,
) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	periodicUpdate_ch := make(chan bool, 50)

	go elevator.PollButtons(button_ch)
	go PeriodicUpdate(periodicUpdate_ch)

	elevatorMap := make(map[string]elevator.Elevator)
	hallRequestsMap := make(map[string][4][2]bool)
	thisElevator := <-fromFSM_ch

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

			evaluateRequests(elevatorMap, &thisElevator)

			toFSM_ch <- thisElevator

		case newState := <-fromFSM_ch:
			err := saveCabCallsToFile(newState)
			if err != nil {
				fmt.Printf("Failed to write to CSV file. \n")
			}

			msgType := calculateMsgType(newState, thisElevator)
			storeFsmUpdate(elevatorMap, &thisElevator, &newState)

			if msgType == network.ObstructedMsg {
				evaluateRequests(elevatorMap, &thisElevator)
				toFSM_ch <- thisElevator
			}

			msg := network.Msg{
				MsgType:  msgType,
				Elevator: thisElevator,
			}

			networkUpdateTx_ch <- msg

		case Msg := <-networkUpdateRx_ch:

			switch Msg.MsgType {

			case network.NewOrder:
				handleNewOrder(elevatorMap, &Msg.Elevator, &thisElevator)
				toFSM_ch <- thisElevator

			case network.OrderCompleted:
				handleOrderCompleted(elevatorMap, &Msg.Elevator, &thisElevator)
				toFSM_ch <- thisElevator

			case network.StateUpdate:
				// Midlertidig løsning for packetloss ved slukking av lys, dette kan være kilde til bus, men funker fett nå (merk at vi alltid setter order counter til å være det vi får inn)
				if Msg.Elevator != elevatorMap[Msg.Elevator.Id] {
					handleOrderCompleted(elevatorMap, &Msg.Elevator, &thisElevator)
					toFSM_ch <- thisElevator
				}
				elevatorMap[Msg.Elevator.Id] = Msg.Elevator

			case network.PeriodicMsg:
				SyncronizeAll(thisElevator, elevatorMap, Msg.Elevator, button_ch)

			case network.ObstructedMsg:

				handleNewOrder(elevatorMap, &Msg.Elevator, &thisElevator)

				toFSM_ch <- thisElevator
			}

		case <-periodicUpdate_ch:
			msg := network.Msg{
				MsgType:  network.PeriodicMsg,
				Elevator: thisElevator,
			}
			networkUpdateTx_ch <- msg

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) != 0 {
				handlePeerupdate(peerUpdate, &thisElevator, &elevatorMap, &hallRequestsMap)
				hallRequestsMap := hallrequestassigner.AssignHallRequests(elevatorMap)
				setLightMatrix(hallRequestsMap, &thisElevator)
				thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
				toFSM_ch <- thisElevator
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

func PeriodicUpdate(periodicUpdate_ch chan bool) {
	for {
		time.Sleep(500 * time.Millisecond)
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

}

func evaluateRequests(elevatorMap map[string]elevator.Elevator, e *elevator.Elevator) {
	elevatorMap[e.Id] = *e
	hallRequestsMap := hallrequestassigner.AssignHallRequests(elevatorMap)

	setElevatorMap(hallRequestsMap, &elevatorMap)
	setElevatorAsignments(elevatorMap, e)
	setLightMatrix(hallRequestsMap, e)

}

func handleNewOrder(elevatorMap map[string]elevator.Elevator, recievedElevator *elevator.Elevator, thisElevator *elevator.Elevator) {
	thisElevator.OrderCounter = recievedElevator.OrderCounter

	elevatorMap[recievedElevator.Id] = *recievedElevator
	if recievedElevator.Obstructed {
	} else {
	}
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
	if newState.Obstructed != oldState.Obstructed {
		return network.ObstructedMsg
	}
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

func saveCabCallsToFile(e elevator.Elevator) error {
	requests := e.Requests
	filename := e.Id

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to open file: %v", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	var records [][]string
	for _, row := range requests {
		cabReq := row[2]
		records = append(records, []string{boolToString(cabReq)})
	}

	records = append(records, []string{"OCC:" + strconv.Itoa(e.OrderClearedCounter)})
	records = append(records, []string{"OC:" + strconv.Itoa(e.OrderCounter)})
	records = append(records, []string{"BH:" + strconv.Itoa(int(e.Behaviour))})
	records = append(records, []string{"DIR:" + strconv.Itoa(int(e.Dirn))})

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write to CSV: %v", err)
		}
	}

	writer.Flush()
	return nil

}

func boolToString(value bool) string {

	if value {
		return "true"
	} else {
		return "false"
	}
}
func SyncronizeAll(thisElevator elevator.Elevator, elevatorMap map[string]elevator.Elevator, recievedElevator elevator.Elevator, button_ch chan elevator.ButtonEvent) {
	combinedrequests := thisElevator.Requests

	for _, e := range elevatorMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				combinedrequests[i][j] = combinedrequests[i][j] || e.Requests[i][j]
			}
		}
	}

	if thisElevator.Lights != recievedElevator.Lights || thisElevator.Lights != combinedrequests {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				if thisElevator.Lights[i][j] != recievedElevator.Lights[i][j] {
					button := new(elevator.ButtonEvent)
					button.Floor = i
					button.Button = elevator.ButtonType(j)
					button_ch <- *button
					thisElevator.OrderCounter = recievedElevator.OrderCounter + 1
				} else if combinedrequests[i][j] != thisElevator.Lights[i][j] {
					button := new(elevator.ButtonEvent)
					button.Floor = i
					button.Button = elevator.ButtonType(j)
					button_ch <- *button
					thisElevator.OrderCounter = recievedElevator.OrderCounter + 1

				}
			}
		}
	}
}
