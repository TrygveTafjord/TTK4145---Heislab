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
	sendConfirmation_ch chan network.Confirm,
	recieveConfirmation_ch chan network.Confirm,
	obstructedToNetwork_ch chan network.Obstructed,
	obstructedFromNetwork_ch chan network.Obstructed,
	stateUpdateToNetwork_ch chan network.StateUpdate,
	stateUpdateFromNetwork_ch chan network.StateUpdate,
	requestClearedToNetwork_ch chan network.RequestCleared,
	requestClearedFromNetwork_ch chan network.RequestCleared,
	periodicInfobankToNetwork_ch chan network.Periodic,
	periodicNetworkToInfobank_ch chan network.Periodic,
	peerUpdate_ch chan network.PeerUpdate,

) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	periodicUpdate_ch := make(chan bool, 50)

	go elevator.PollButtons(button_ch)
	go PeriodicUpdate(periodicUpdate_ch)

	elevatorMap := make(map[string]ElevatorInfo)

	thisElevator := <-init_ch
	inheritedRequests := inheritedRequests(thisElevator)
	for _, reqs := range inheritedRequests {
		button_ch <- reqs
	}

	elevatorMap[thisElevator.Id] = thisElevator

	for {
		select {

		case buttonEvent := <-button_ch:

			if len(elevatorMap) > 1 {
				if !confirmNewAssignment(newRequestToNetwork_ch, recieveConfirmation_ch, buttonEvent, len(elevatorMap), thisElevator.Id) {
					break
				}
			}

			thisElevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
			distributeRequests(&elevatorMap, &thisElevator)

			setLightMatrix(elevatorMap, &thisElevator)

			saveInfoToFile(thisElevator)

			lightsUpdateToFSM_ch <- thisElevator.Lights
			requestUpdateToFSM_ch <- thisElevator.Requests

		case obstructed := <-obstructionFromFSM_ch:

			thisElevator.State.OutOfService = obstructed

			distributeRequests(&elevatorMap, &thisElevator)

			confirmObstructionState(obstructedToNetwork_ch, recieveConfirmation_ch, obstructed, len(elevatorMap), thisElevator.Id)

			// msg := network.Obstructed{
			// 	Id:         thisElevator.Id,
			// 	Obstructed: obstructed,
			// }
			//confirmObstruction(obstructedToNetwork_ch, recieveConfirmation_ch, obstructed, len(elevatorMap), thisElevator.Id)
			// obstructedToNetwork_ch <- msg

		case newState := <-stateUpdateFromFSM_ch:

			thisElevator.State = newState
			elevatorMap[thisElevator.Id] = thisElevator

			saveInfoToFile(thisElevator)

			msg := network.StateUpdate{
				Id:    thisElevator.Id,
				State: newState,
			}

			stateUpdateToNetwork_ch <- msg

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

			confirmation := network.Confirm{
				Id:      thisElevator.Id,
				PassWrd: msg.Id + fmt.Sprint(msg.Request.Button) + fmt.Sprint(msg.Request.Floor),
			}
			sendConfirmation_ch <- confirmation

			updatedElev := elevatorMap[msg.Id]
			updatedElev.Requests[msg.Request.Floor][msg.Request.Button] = true
			elevatorMap[msg.Id] = updatedElev

			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)
			setElevatorMap(hallRequestsMap, &elevatorMap)

			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			setLightMatrix(elevatorMap, &thisElevator)

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

			confirmation := network.Confirm{
				Id:      thisElevator.Id,
				PassWrd: msg.Id,
			}

			sendConfirmation_ch <- confirmation

			updatedElev := elevatorMap[msg.Id]
			updatedElev.State.OutOfService = msg.Obstructed
			elevatorMap[msg.Id] = updatedElev

			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)

			elevatorMap[msg.Id] = updatedElev
			setElevatorMap(hallRequestsMap, &elevatorMap)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			requestUpdateToFSM_ch <- thisElevator.Requests

		case <-periodicUpdate_ch:

			msg := network.Periodic{
				Id:       thisElevator.Id,
				State:    thisElevator.State,
				Requests: thisElevator.Requests,
			}

			periodicInfobankToNetwork_ch <- msg

		case msg := <-periodicNetworkToInfobank_ch:

			recievedElevator := elevatorMap[msg.Id]
			recievedElevator.Requests = msg.Requests
			recievedElevator.State = msg.State

			if msg.Requests == elevatorMap[msg.Id].Requests {
				elevatorMap[msg.Id] = recievedElevator
				break
			}

			syncronizeLights(msg.Requests, msg.Id, elevatorMap, &thisElevator.Lights)
			elevatorMap[msg.Id] = recievedElevator

			lightsUpdateToFSM_ch <- thisElevator.Lights

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) == 0 {
				break
			}
			removeLostPeers(peerUpdate, &thisElevator, &elevatorMap)

			distributeRequests(&elevatorMap, &thisElevator)

			setLightMatrix(elevatorMap, &thisElevator)

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
		time.Sleep(300 * time.Millisecond)
		periodicUpdate_ch <- true
	}
}

func removeLostPeers(peerUpdate network.PeerUpdate, thisElevator *ElevatorInfo, elevatorMap *map[string]ElevatorInfo) {

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

func distributeRequests(elevatorMap *map[string]ElevatorInfo, e *ElevatorInfo) {
	(*elevatorMap)[e.Id] = *e
	assignerList := createAssignerInput((*elevatorMap))
	hallRequestsMap := assigner.AssignHallRequests(assignerList)
	setElevatorMap(hallRequestsMap, elevatorMap)
	e.Requests = (*elevatorMap)[e.Id].Requests
}

/*func removeDeadElevators(elevatorMap map[string]ElevatorInfo) (assignmentsMap map[string]ElevatorInfo, *[elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool ) {
	var ordersToBeTransferred *[elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool
	for id, elev := range elevatorMap {
		if elev.State.OutOfService {
			*ordersToBeTransferred = generateHallCalls(elev.Requests)
		}else{
			assignmentsMap[id] = elev
		}
	}
	return assignmentsMap, ordersToBeTransferred
}*/

func setLightMatrix(elevatorMap map[string]ElevatorInfo, e *ElevatorInfo) {
	for _, value := range elevatorMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.Lights[i][j] = (e.Lights[i][j] || value.Requests[i][j])
			}
			e.Lights[i][elevator.BT_Cab] = e.Requests[i][elevator.BT_Cab]
		}
	}
}

func syncronizeLights(requests [elevator.N_FLOORS][elevator.N_BUTTONS]bool, id string, elevatorMap map[string]ElevatorInfo, lights *[elevator.N_FLOORS][elevator.N_BUTTONS]bool) {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS; j++ {
			if requests[i][j] != elevatorMap[id].Requests[i][j] {
				lights[i][j] = requests[i][j]
			}
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

func inheritedRequests(thisElevator ElevatorInfo) []elevator.ButtonEvent {
	var BTNevents []elevator.ButtonEvent
	for floor := 0; floor < 4; floor++ {
		if thisElevator.Requests[floor][elevator.BT_Cab] {
			BTNevents = append(BTNevents, elevator.ButtonEvent{Floor: floor, Button: elevator.BT_Cab})
		}
	}
	return BTNevents
}
