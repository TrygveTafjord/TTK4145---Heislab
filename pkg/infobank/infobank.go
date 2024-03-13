package infobank

import (
	"fmt"
	"time"

	"project.com/pkg/assigner"
	"project.com/pkg/elevator"
	"project.com/pkg/network"
	//"project.com/pkg/timer"
)

func Infobank(
	requestUpdateToFSM_ch chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool,
	clearRequestFromFSM_ch chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool, 
	stateUpdateFromFSM_ch chan elevator.State, 
	lightsUpdateToFSM_ch chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool,
	networkUpdateTx_ch chan network.Msg,
	networkUpdateRx_ch chan network.Msg,
	peerUpdate_ch chan network.PeerUpdate,
	obstructionFromFSM_ch chan bool,
) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	periodicUpdate_ch := make(chan bool, 50)

	go elevator.PollButtons(button_ch)
	go PeriodicUpdate(periodicUpdate_ch)

	elevatorMap := make(map[string]ElevatorInfo)
	//hallRequestsMap := make(map[string][4][2]bool)
	initialState := <- stateUpdateFromFSM_ch

	thisElevator := ElevatorInfo{
		Id: "our ID", //from Initialize later
		OrderClearedCounter: 0,
		OrderCounter: 0,
		State: initialState,	

	}

	elevatorMap[thisElevator.Id] = thisElevator

	for {
		select {
		case buttonEvent := <-button_ch:
			thisElevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
			thisElevator.OrderCounter++
			evaluateRequests(elevatorMap, &thisElevator)

			requestUpdateToFSM_ch <- thisElevator.Requests
			lightsUpdateToFSM_ch <- thisElevator.Lights

			
			// msg := network.Msg{
			// 	MsgType:  network.NewOrder,
			// 	Elevator: infobank.thisElevator.Requests,
			// }

			// networkUpdateTx_ch <- msg

		case obstructed := <- obstructionFromFSM_ch:
			thisElevator.State.Obstructed = obstructed
			evaluateRequests(elevatorMap, &thisElevator)
				//kan implementere en cycle her, 
			// msg := ObstructedMsg { 
			// 	Id: thisElevator.Id, 
			// 	Obstructed: obstructed,
			// }				
			// obstructionNetwork_ch <- msg 
		
		case newState := <- stateUpdateFromFSM_ch:

			thisElevator.State = newState
			// msg := StateMsg { 
			// 	Id: thisElevator.Id, 
			// 	State: newState,
			// }

			// newStateNetwork_ch <- msg 
			

			elevatorMap[thisElevator.Id] = thisElevator

		case updatedRequests := <- clearRequestFromFSM_ch:
			thisElevator.Requests = updatedRequests
			elevatorMap[thisElevator.Id] =thisElevator
			
			// msg := RequestClearedMsg { 
			// 	Direction: thisElevator.Dirn, 
			// 	Floor: thisElevator.State.Floor,
			// }

			// requestClearedNetwork_ch <- msg
			
		// case Msg := <-networkUpdateRx_ch:

		// 	switch Msg.MsgType {

		// 	case network.NewOrder:
		// 		handleNewOrder(elevatorMap, &Msg.Elevator, &thisElevator)
		// 		toFSM_ch <- thisElevator

		// 	case network.OrderCompleted:
		// 		handleOrderCompleted(elevatorMap, &Msg.Elevator, &thisElevator)
		// 		toFSM_ch <- thisElevator

		// 	case network.StateUpdate:
		// 		// Midlertidig løsning for packetloss ved slukking av lys, dette kan være kilde til bus, men funker fett nå (merk at vi alltid setter order counter til å være det vi får inn)
		// 		if Msg.Elevator != elevatorMap[Msg.Elevator.Id] {
		// 			handleOrderCompleted(elevatorMap, &Msg.Elevator, &thisElevator)
		// 			toFSM_ch <- thisElevator
		// 		}
		// 		elevatorMap[Msg.Elevator.Id] = Msg.Elevator

		// 	case network.PeriodicMsg:
		// 		SyncronizeAll(thisElevator, elevatorMap, Msg.Elevator, button_ch)

		// 	case network.ObstructedMsg:

		// 		handleNewOrder(elevatorMap, &Msg.Elevator, &thisElevator)

		// 		toFSM_ch <- thisElevator
		// 	}

		// case <-periodicUpdate_ch:
		// 	msg := network.Msg{
		// 		MsgType:  network.PeriodicMsg,
		// 		Elevator: thisElevator,
		// 	}
		// 	networkUpdateTx_ch <- msg

		// case peerUpdate := <-peerUpdate_ch:
		// 	if len(peerUpdate.Lost) != 0 {
		// 		handlePeerupdate(peerUpdate, &thisElevator, &elevatorMap, &hallRequestsMap)
		// 		hallRequestsMap := hallrequestassigner.AssignHallRequests(elevatorMap)
		// 		setLightMatrix(hallRequestsMap, &thisElevator)
		// 		thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
		// 		toFSM_ch <- thisElevator
		// 	}
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

func handlePeerupdate(peerUpdate network.PeerUpdate, thisElevator *ElevatorInfo, elevatorMap *map[string]ElevatorInfo, newAssignmentsMap *map[string][4][2]bool) {

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

func evaluateRequests(elevatorMap map[string]ElevatorInfo, e *ElevatorInfo) {
	elevatorMap[e.Id] = *e
	assignerList := createAssignerInput(elevatorMap)
	hallRequestsMap := assigner.AssignHallRequests(assignerList)

	setElevatorMap(hallRequestsMap, &elevatorMap)
	setElevatorAsignments(elevatorMap, e)
	setLightMatrix(hallRequestsMap, e)
}

func handleNewOrder(elevatorMap map[string]ElevatorInfo, recievedElevator *ElevatorInfo, thisElevator ElevatorInfo) {
	thisElevator.OrderCounter = recievedElevator.OrderCounter

	elevatorMap[recievedElevator.Id] = *recievedElevator
	if recievedElevator.State.Obstructed {
	} else {
	}
	assignerList := createAssignerInput(elevatorMap)
	hallRequestsMap := assigner.AssignHallRequests(assignerList)

	setElevatorMap(hallRequestsMap, &elevatorMap)
	setElevatorAsignments(elevatorMap, &thisElevator)
	setLightMatrix(hallRequestsMap, &thisElevator)

	thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
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

func storeFsmUpdate(elevatorMap map[string]ElevatorInfo, oldState *ElevatorInfo, newState *ElevatorInfo) {
	if newState.Id != oldState.Id {
		fmt.Printf("error: trying to assign values to non similar Id's \n")
		return
	}

	elevatorMap[oldState.Id] = *newState
	*oldState = *newState
}

func calculateMsgType(newState ElevatorInfo, oldState ElevatorInfo) network.MsgType {
	if newState.State.Obstructed != oldState.State.Obstructed {
		return network.ObstructedMsg
	}
	if newState.OrderClearedCounter > oldState.OrderClearedCounter {
		return network.OrderCompleted
	}
	return network.StateUpdate
}

func setLightMatrix(newAssignmentsMap map[string][4][2]bool, e *ElevatorInfo) {
	for _, value := range newAssignmentsMap {
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				e.Lights[i][j] = (e.Lights[i][j] || value[i][j])
				e.Lights[i][elevator.BT_Cab] = e.Requests[i][elevator.BT_Cab]
			}
		}
	}
}

func setElevatorAsignments(elevatorMap map[string]ElevatorInfo, e *ElevatorInfo) {
	e.Requests = elevatorMap[e.Id].Requests
}

func createAssignerInput(elevatorMap map[string]ElevatorInfo) []assigner.AssignerInput {
	assignerList := make([]assigner.AssignerInput, 0, len(elevatorMap))
	for id, value := range elevatorMap {
		a := assigner.AssignerInput {
			Id: 		id,
			Requests: 	value.Requests,
			State: 		value.State,
		}
		assignerList = append(assignerList, a)
	}
	return assignerList
} 

		// 	msg := network.Msg{
		// 		MsgType:  network.PeriodicMsg,
		// 		Elevator: thisElevator,
		// 	}
// type AssingerInput struct {
// 	Id					string
// 	Requests            [elevator.N_FLOORS][elevator.N_BUTTONS]bool
// 	State 				elevator.State
// }
// func saveInfoToFile(e ElevatorInfo) error {
// 	requests := e.Requests
// 	filename := e.Id

// 	file, err := os.Create(filename)
// 	if err != nil {
// 		fmt.Printf("Failed to open file: %v", err)
// 	}

// 	defer file.Close()

// 	writer := csv.NewWriter(file)

// 	var records [][]string
// 	for _, row := range requests {
// 		cabReq := row[2]
// 		records = append(records, []string{boolToString(cabReq)})
// 	}

// 	records = append(records, []string{"OCC:" + strconv.Itoa(e.OrderClearedCounter)})
// 	records = append(records, []string{"OC:" + strconv.Itoa(e.OrderCounter)})
// 	records = append(records, []string{"BH:" + strconv.Itoa(int(e.Behaviour))})
// 	records = append(records, []string{"DIR:" + strconv.Itoa(int(e.Dirn))})

// 	for _, record := range records {
// 		if err := writer.Write(record); err != nil {
// 			return fmt.Errorf("failed to write to CSV: %v", err)
// 		}
// 	}

// 	writer.Flush()
// 	return nil

// }

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
