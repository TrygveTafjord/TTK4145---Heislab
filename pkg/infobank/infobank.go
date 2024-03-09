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
	networkUpdateTx_ch chan elevator.Elevator,
	networkUpdateRx_ch chan elevator.Elevator,
	peerUpdate_ch chan network.PeerUpdate,
) {

	button_ch := make(chan elevator.ButtonEvent, 50)
	go elevator.PollButtons(button_ch)

	elevatorMap := make(map[string]elevator.Elevator)

	var thisElevator elevator.Elevator
	thisElevator = <-elevStatusUpdate_ch
	networkUpdateTx_ch <- thisElevator

	elevatorMap[thisElevator.Id] = thisElevator
	var hallRequestsMap map[string][4][2]bool

	periodicUpdate_ch:= make(chan bool,50)
	go PeriodicUpdate(periodicUpdate_ch)


	for {
		select {
		case btn := <-button_ch:
			//Pass information about new assignment to other nodes
			thisElevator.Requests[btn.Floor][btn.Button] = true
			thisElevator.OrderCounter++
			networkUpdateTx_ch <- thisElevator

			//Update map to contain information about this new assignment
			elevatorMap[thisElevator.Id] = thisElevator
			hallrequestassigner.AssignHallRequests(elevatorMap, &hallRequestsMap)
			setLights(hallRequestsMap, &thisElevator)
			updateMapWithNewAssignments(hallRequestsMap, &elevatorMap)

			//Pass information about the new assignment distribution to our local elevator
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests

			//Pass information about the newly distributed assignments to FSM, through our local elevator
			elevStatusUpdate_ch <- thisElevator

		case newState := <-elevStatusUpdate_ch:
			//Potensiell bug -> Ordercounter og ClearOrderCounter får feil verdi
			elevatorMap[thisElevator.Id] = newState
			thisElevator = newState
			networkUpdateTx_ch <- thisElevator

		case recievedElevator := <-networkUpdateRx_ch:
			elevatorMap[recievedElevator.Id] = recievedElevator

			if recievedElevator.OrderClearedCounter > thisElevator.OrderClearedCounter {
				thisElevator = handleRecievedOrderCompleted(recievedElevator, thisElevator)
				hallrequestassigner.AssignHallRequests(elevatorMap, &hallRequestsMap)
				setLights(hallRequestsMap, &thisElevator)
				thisElevator.OrderClearedCounter = recievedElevator.OrderClearedCounter
				elevatorMap[thisElevator.Id] = thisElevator
				elevStatusUpdate_ch <- thisElevator

			} else if recievedElevator.OrderCounter > thisElevator.OrderCounter {
				thisElevator.OrderCounter = recievedElevator.OrderCounter
				elevatorMap[thisElevator.Id] = thisElevator
				hallrequestassigner.AssignHallRequests(elevatorMap, &hallRequestsMap)

				setLights(hallRequestsMap, &thisElevator)

				updateMapWithNewAssignments(hallRequestsMap, &elevatorMap)
				thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
				elevStatusUpdate_ch <- thisElevator
			} else{ 
			hallrequestassigner.AssignHallRequests(elevatorMap, &hallRequestsMap)
			setLights(hallRequestsMap, &thisElevator)
			updateMapWithNewAssignments(hallRequestsMap, &elevatorMap)
			thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
			elevStatusUpdate_ch <- thisElevator
			}
		case <- periodicUpdate_ch:
			networkUpdateTx_ch <- thisElevator

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) != 0{
				handlePeerupdate(peerUpdate,&thisElevator,&elevatorMap, &hallRequestsMap)
				hallrequestassigner.AssignHallRequests(elevatorMap, &hallRequestsMap)
				setLights(hallRequestsMap, &thisElevator)
				thisElevator.Requests = elevatorMap[thisElevator.Id].Requests
				elevStatusUpdate_ch <- thisElevator
				fmt.Printf("\n", elevatorMap)
			}
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

func updateMapWithNewAssignments(newAssignmentsMap map[string][4][2]bool, elevatorMap *map[string]elevator.Elevator) {
	returnMap := make(map[string]elevator.Elevator)

	for id, requests := range newAssignmentsMap {
		tempElev := (*elevatorMap)[id]
		for i := 0; i < elevator.N_FLOORS; i++ {
			for j := 0; j < elevator.N_BUTTONS-1; j++ {
				tempElev.Requests[i][j] = requests[i][j]
			}
		}
		returnMap[id] = tempElev
	}
	*elevatorMap = returnMap
}
func handleNewOrder(elevatorMap map[string]elevator.Elevator, thisElevator *elevator.Elevator) {

}

func setAllLights(e *elevator.Elevator) {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			elevator.SetButtonLamp(elevator.ButtonType(btn), floor, e.Lights[floor][btn])
		}
	}
}


func PeriodicUpdate(periodicUpdate_ch chan bool){
	for{
		time.Sleep(5000 * time.Millisecond)
		periodicUpdate_ch <- true
	}
}


func handlePeerupdate(peerUpdate network.PeerUpdate, thisElevator *elevator.Elevator, elevatorMap *map[string]elevator.Elevator, newAssignmentsMap *map[string][4][2]bool){
	fmt.Printf("\n Her skal skal vi overføre requestst!")
	fmt.Printf("\n Denne heisens requests er: ", thisElevator.Requests)
	fmt.Printf("\n Heisen vi mister sin   er: ", (*elevatorMap)[peerUpdate.Lost[0]].Requests)


	for i := 0; i < len(peerUpdate.Lost); i++ {
		for j := 0; j < elevator.N_FLOORS; j++ {
			for k := 0; k < elevator.N_BUTTONS-1; k++ {
				thisElevator.Requests[j][k] = thisElevator.Requests[j][k] || (*elevatorMap)[peerUpdate.Lost[i]].Requests[j][k]
			}
		}
		delete(*elevatorMap, peerUpdate.Lost[i])
		delete(*newAssignmentsMap,peerUpdate.Lost[i])
		//fmt.Printf("\n" , thisElevator.OrderCounter)
		thisElevator.OrderCounter++ 
	}
	(*elevatorMap)[thisElevator.Id] = *thisElevator
	fmt.Printf("\n Den samlede matrisen er nå: ", thisElevator.Requests)

	//Nå må vi redistrubiere requests og fjerne
}


//Vi må oppdage om en heis som ikke står i idle og ikke har tok request matris har stått stille lenge og behandle det
//Vi må også oppdage om en heis ikke har sendt melding på en stund
// Legg logikk for å melde seg selv av nettverk når vi oppdager vi er fucked
