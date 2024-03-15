package infobank

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"project.com/pkg/assigner"
	"project.com/pkg/elevator"
	"project.com/pkg/network"
	//"project.com/pkg/timer"
)

func Infobank(
	init_ch chan ElevatorInfo,
	requestUpdateToFSM_ch chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool,
	clearRequestFromFSM_ch chan []elevator.ButtonEvent,
	stateUpdateFromFSM_ch chan elevator.State,
	lightsUpdateToFSM_ch chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool,
	obstructionFromFSM_ch chan bool,
	newRequestToNetwork_ch chan network.NewRequest,
	newRequestFromNetwork_ch chan network.NewRequest,
	obstructedToNetwork_ch chan network.Obstructed,
	obstructedFromNetwork_ch chan network.Obstructed,
	stateUpdateToNetwork_ch chan network.StateUpdate,
	stateUpdateFromNetwork_ch chan network.StateUpdate,
	requestClearedToNetwork_ch chan network.RequestCleared,
	requestClearedFromNetwork_ch chan network.RequestCleared,
	peerUpdate_ch chan network.PeerUpdate,

) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	periodicUpdate_ch := make(chan bool, 50)

	go elevator.PollButtons(button_ch)
	go PeriodicUpdate(periodicUpdate_ch)

	elevatorMap := make(map[string]ElevatorInfo)
	//hallRequestsMap := make(map[string][4][2]bool)
	thisElevator := <-init_ch
	inheritedRequests := inheritedRequests(thisElevator)
	for _, reqs := range inheritedRequests {
		button_ch <- reqs
	}

	elevatorMap[thisElevator.Id] = thisElevator

	for {
		select {

		case buttonEvent := <-button_ch:

			thisElevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
			elevatorMap[thisElevator.Id] = thisElevator
			saveInfoToFile(thisElevator)
			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)
			setElevatorMap(hallRequestsMap, &elevatorMap)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
			setLightMatrix(hallRequestsMap, &thisElevator)
			lightsUpdateToFSM_ch <- thisElevator.Lights
			requestUpdateToFSM_ch <- thisElevator.Requests

			msg := network.NewRequest{
				Id:      thisElevator.Id,
				Request: buttonEvent,
			}

			newRequestToNetwork_ch <- msg

		case obstructed := <-obstructionFromFSM_ch:
			if !obstructed {
				removeHallCalls(&thisElevator)
				requestUpdateToFSM_ch <- thisElevator.Requests
			}

			thisElevator.State.OutOfService = obstructed
			evaluateRequests(&elevatorMap, &thisElevator)

			msg := network.Obstructed{
				Id:         thisElevator.Id,
				Obstructed: obstructed,
			}

			obstructedToNetwork_ch <- msg

		case newState := <-stateUpdateFromFSM_ch:

			thisElevator.State = newState
			saveInfoToFile(thisElevator)
			msg := network.StateUpdate{
				Id:    thisElevator.Id,
				State: newState,
			}

			stateUpdateToNetwork_ch <- msg

			elevatorMap[thisElevator.Id] = thisElevator

		case clearedRequests := <-clearRequestFromFSM_ch:

			for _, requests := range clearedRequests {
				thisElevator.Requests[requests.Floor][requests.Button] = false
				thisElevator.Lights[requests.Floor][requests.Button] = false
			}
			thisElevator.Requests[clearedRequests[0].Floor][elevator.BT_Cab] = false
			thisElevator.Lights[clearedRequests[0].Floor][elevator.BT_Cab] = false
			saveInfoToFile(thisElevator)
			elevatorMap[thisElevator.Id] = thisElevator

			msg := network.RequestCleared{
				Id:              thisElevator.Id,
				ClearedRequests: clearedRequests,
			}
			requestClearedToNetwork_ch <- msg

		case msg := <-newRequestFromNetwork_ch:
			updatedElev := elevatorMap[msg.Id]
			updatedElev.Requests[msg.Request.Floor][msg.Request.Button] = true
			elevatorMap[msg.Id] = updatedElev
			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)

			setElevatorMap(hallRequestsMap, &elevatorMap)
			setLightMatrix(hallRequestsMap, &thisElevator)

			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			lightsUpdateToFSM_ch <- thisElevator.Lights
			requestUpdateToFSM_ch <- thisElevator.Requests

		case msg := <-stateUpdateFromNetwork_ch:
			updatedElev := elevatorMap[msg.Id]
			updatedElev.State = msg.State
			elevatorMap[msg.Id] = updatedElev

		case msg := <-requestClearedFromNetwork_ch:

			updatedElev := elevatorMap[msg.Id]
			for _, requests := range msg.ClearedRequests {
				updatedElev.Requests[requests.Floor][requests.Button] = false
				thisElevator.Lights[requests.Floor][requests.Button] = false
			}
			elevatorMap[msg.Id] = updatedElev
			lightsUpdateToFSM_ch <- thisElevator.Lights

		case msg := <-obstructedFromNetwork_ch:
			obstructedElevator := elevatorMap[msg.Id]
			obstructedElevator.State.OutOfService = msg.Obstructed
			elevatorMap[msg.Id] = obstructedElevator
			evaluateRequests(&elevatorMap, &obstructedElevator)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
			requestUpdateToFSM_ch <- thisElevator.Requests

		case <-periodicUpdate_ch:
			msg := network.StateUpdate{
				Id:    thisElevator.Id,
				State: thisElevator.State,
			}
			stateUpdateToNetwork_ch <- msg

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) == 0 {
				break
			}
			handlePeerupdate(peerUpdate, &thisElevator, &elevatorMap)

			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			setLightMatrix(hallRequestsMap, &thisElevator)
			requestUpdateToFSM_ch <- thisElevator.Requests

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) == 0 {
				break
			}
			handlePeerupdate(peerUpdate, &thisElevator, &elevatorMap)
			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
			setLightMatrix(hallRequestsMap, &thisElevator)
			requestUpdateToFSM_ch <- thisElevator.Requests
		}
	}
}

func setElevatorMap(newAssignmentsMap map[string][4][2]bool, elevatorMap *map[string]ElevatorInfo) {

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

func handlePeerupdate(peerUpdate network.PeerUpdate, thisElevator *ElevatorInfo, elevatorMap *map[string]ElevatorInfo) {

	for i := 0; i < len(peerUpdate.Lost); i++ {
		for j := 0; j < elevator.N_FLOORS; j++ {
			for k := 0; k < elevator.N_BUTTONS-1; k++ {
				thisElevator.Requests[j][k] = thisElevator.Requests[j][k] || (*elevatorMap)[peerUpdate.Lost[i]].Requests[j][k]
			}
		}
		delete(*elevatorMap, peerUpdate.Lost[i])
	}
	(*elevatorMap)[thisElevator.Id] = *thisElevator
}

func evaluateRequests(elevatorMap *map[string]ElevatorInfo, e *ElevatorInfo) {
	(*elevatorMap)[e.Id] = *e

	aliveElevators, ordersToBeCleared := removeDeadElevators(*elevatorMap)
	assignmentsMap := transferOrders(ordersToBeCleared, aliveElevators)
	assignerList := createAssignerInput(assignmentsMap)
	hallRequestsMap := assigner.AssignHallRequests(assignerList)
	
	for id, requests := range hallRequestsMap {
		
		tempElev := (*elevatorMap)[id]
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				tempElev.Requests[i][j] = requests[i][j]
			}
		}
		(*elevatorMap)[id] = tempElev
	}

	delete(*elevatorMap, "")

	for ID, elev := range *elevatorMap {
		if elev.State.OutOfService {
			for i := 0; i < elevator.N_FLOORS; i++ {
				for j := 0; j < elevator.N_BUTTONS-1; j++ {
					elev.Requests[i][j] = false
				}
			}
			(*elevatorMap)[ID] = elev
		}
	}

	e.Requests = (*elevatorMap)[e.Id].Requests
	setLightMatrix(hallRequestsMap, e)
}

func handleOrderCompleted(elevatorMap map[string]ElevatorInfo, recievedElevator *ElevatorInfo, thisElevator *ElevatorInfo) {
	elevatorMap[recievedElevator.Id] = *recievedElevator

	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			thisElevator.Lights[i][j] = thisElevator.Lights[i][j] && recievedElevator.Lights[i][j]
		}
	}

	thisElevator.OrderClearedCounter = recievedElevator.OrderClearedCounter

	elevatorMap[thisElevator.Id] = *thisElevator

}

func removeDeadElevators(elevatorMap map[string]ElevatorInfo) (map[string]ElevatorInfo, [elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool) {
	var ordersToBeTransferred [elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool
	assignmentsMap := make(map[string]ElevatorInfo)
	for ID, elev := range elevatorMap {
		if elev.State.OutOfService {
			ordersToBeTransferred = generateHallCalls(elev.Requests)
		} else {
			assignmentsMap[ID] = elev
		}
	}
	return assignmentsMap, ordersToBeTransferred
}

func storeFsmUpdate(elevatorMap map[string]ElevatorInfo, oldState *ElevatorInfo, newState *ElevatorInfo) {
	if newState.Id != oldState.Id {
		return
	}

	elevatorMap[oldState.Id] = *newState
	*oldState = *newState
}

func setLightMatrix(newAssignmentsMap map[string][4][2]bool, e *ElevatorInfo) {
	for _, value := range newAssignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.Lights[i][j] = (e.Lights[i][j] || value[i][j])
			}
			e.Lights[i][elevator.BT_Cab] = e.Requests[i][elevator.BT_Cab]
		}
	}
}

func createAssignerInput(assignerMap map[string]ElevatorInfo) []assigner.AssignerInput {
	var assignerList []assigner.AssignerInput
	for id, elev := range assignerMap {
		a := assigner.AssignerInput{
			Id:       id,
			Requests: elev.Requests,
			State:    elev.State,
		}
		assignerList = append(assignerList, a)
	}
	return assignerList
}

func removeHallCalls(elev *ElevatorInfo) {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			elev.Requests[i][j] = false
		}
	}
}

func transferOrders(ordersToBeTransferred [elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool, assignmentsMap map[string]ElevatorInfo) map[string]ElevatorInfo {
	returnMap := make(map[string]ElevatorInfo)
	for ID, elev := range assignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				if ordersToBeTransferred[i][j] {
					elev.Requests[i][j] = true
				}
			}
		}
		returnMap[ID] = elev
	}
	return returnMap
}

func generateHallCalls(requests [elevator.N_FLOORS][elevator.N_BUTTONS]bool) (ordersToBeTransferred [elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool) {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS-1; j++ {
			ordersToBeTransferred[i][j] = requests[i][j]
		}
	}
	return ordersToBeTransferred
}

func saveInfoToFile(e ElevatorInfo) error {
	requests := e.Requests
	file, err := os.OpenFile(e.Id, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Failed to open file for writing: %v\n", err)
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
	records = append(records, []string{"BH:" + strconv.Itoa(int(e.State.Behaviour))})
	records = append(records, []string{"DIR:" + strconv.Itoa(int(e.State.Dirn))})

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
func SyncronizeAll(thisElevator ElevatorInfo, elevatorMap map[string]ElevatorInfo, recievedElevator ElevatorInfo, button_ch chan elevator.ButtonEvent) {
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

func inheritedRequests(thisElevator ElevatorInfo) []elevator.ButtonEvent {
	var BTNevents []elevator.ButtonEvent
	for floor := 0; floor < 4; floor++ {
		if thisElevator.Requests[floor][elevator.BT_Cab] {
			BTNevents = append(BTNevents, elevator.ButtonEvent{floor, elevator.BT_Cab})
		}
	}
	return BTNevents
}
