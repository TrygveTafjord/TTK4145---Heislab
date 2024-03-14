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
	init_ch							chan ElevatorInfo,
	requestUpdateToFSM_ch 			chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool,
	clearRequestFromFSM_ch 			chan []elevator.ButtonEvent, 
	stateUpdateFromFSM_ch 			chan elevator.State, 
	lightsUpdateToFSM_ch 			chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool,
	obstructionFromFSM_ch 			chan bool,
	newRequestToNetwork_ch 			chan network.NewRequest,
	newRequestFromNetwork_ch 		chan network.NewRequest,
	obstructedToNetwork_ch 			chan network.Obstructed,
	obstructedFromNetwork_ch 		chan network.Obstructed,
	stateUpdateToNetwork_ch 		chan network.StateUpdate,
	stateUpdateFromNetwork_ch 		chan network.StateUpdate,
	requestClearedToNetwork_ch 		chan network.RequestCleared,
	requestClearedFromNetwork_ch 	chan network.RequestCleared,
	peerUpdate_ch 					chan network.PeerUpdate,
	
) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	periodicUpdate_ch := make(chan bool, 50)

	go elevator.PollButtons(button_ch)
	go PeriodicUpdate(periodicUpdate_ch)

	elevatorMap := make(map[string]ElevatorInfo)
	thisElevator := <- init_ch

	elevatorMap[thisElevator.Id] = thisElevator

	for {
		select {
		case buttonEvent := <-button_ch:

			thisElevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
			elevatorMap[thisElevator.Id] = thisElevator
			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)
			setElevatorMap(hallRequestsMap, &elevatorMap)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
			setLightMatrix(hallRequestsMap, &thisElevator)
			lightsUpdateToFSM_ch <- thisElevator.Lights
			requestUpdateToFSM_ch <- thisElevator.Requests

			msg := network.NewRequest{
				Id:		 thisElevator.Id,
				Request: buttonEvent,
			}

			newRequestToNetwork_ch <- msg

		case obstructed := <- obstructionFromFSM_ch:
			thisElevator.State.Obstructed = obstructed
			evaluateRequests(elevatorMap, &thisElevator)
				//kan implementere en cycle her, 
			msg := network.Obstructed { 
				Id: thisElevator.Id, 
				Obstructed: obstructed,
			}
			obstructedToNetwork_ch <- msg 
		
		case newState := <- stateUpdateFromFSM_ch:

			thisElevator.State = newState
			msg := network.StateUpdate { 
				Id: thisElevator.Id, 
				State: newState,
			}

			stateUpdateToNetwork_ch <- msg 
			elevatorMap[thisElevator.Id] = thisElevator

		case clearedRequests := <- clearRequestFromFSM_ch:

			for _, requests := range clearedRequests {
				thisElevator.Requests[requests.Floor][requests.Button] = false
				thisElevator.Lights[requests.Floor][requests.Button] = false
			}
			thisElevator.Requests[clearedRequests[0].Floor][elevator.BT_Cab] = false
			thisElevator.Lights[clearedRequests[0].Floor][elevator.BT_Cab] = false
			
			elevatorMap[thisElevator.Id] = thisElevator

			msg := network.RequestCleared { 
				Id:		  		  thisElevator.Id,
				ClearedRequests:  clearedRequests,
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
		
		case msg := <- requestClearedFromNetwork_ch:
			fmt.Printf("recieved cleared request \n")

			updatedElev := elevatorMap[msg.Id]
			for _, requests := range msg.ClearedRequests {
				updatedElev.Requests[requests.Floor][requests.Button] = false
				thisElevator.Lights[requests.Floor][requests.Button] = false
			}
			elevatorMap[msg.Id] = updatedElev
			lightsUpdateToFSM_ch <- thisElevator.Lights
		
		case msg := <- obstructedFromNetwork_ch:
			fmt.Printf("recieved obstructed in infobank from the network! \n")
			updatedElev := elevatorMap[msg.Id]
			updatedElev.State.Obstructed = msg.Obstructed 
			fmt.Printf("Recieved obstruction value: %v \n", msg.Obstructed)
			elevatorMap[msg.Id] = updatedElev

			assignerList := createAssignerInput(elevatorMap)
			hallRequestsMap := assigner.AssignHallRequests(assignerList)
			setElevatorMap(hallRequestsMap, &elevatorMap)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			requestUpdateToFSM_ch <- thisElevator.Requests
		
		case <- periodicUpdate_ch:
			msg := network.StateUpdate { 
				Id: thisElevator.Id, 
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

func evaluateRequests(elevatorMap map[string]ElevatorInfo, e *ElevatorInfo) {
	elevatorMap[e.Id] = *e
	assignerList := createAssignerInput(elevatorMap)
	hallRequestsMap := assigner.AssignHallRequests(assignerList)
	setElevatorMap(hallRequestsMap, &elevatorMap)
	e.Requests = elevatorMap[e.Id].Requests
	setLightMatrix(hallRequestsMap, e)
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
