package infobank

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/hallrequestassigner"
	"project.com/pkg/network"
	"project.com/pkg/timer"
)

var confirmOtherNodesTime float64 = 2

func Infobank_FSM(
	button_ch chan elevator.ButtonEvent,
	newStatus_ch chan elevator.Elevator,
	infoUpdate_ch chan elevator.Elevator,
	externalInfo chan elevator.Elevator,
	peerUpdate_ch chan network.PeerUpdate,
	assigner_ch chan map[string][elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool) {

	elevatorList := make(map[string]elevator.Elevator)
	elevatorTimes := make(map[string]float64)
	thisElevator := new(elevator.Elevator)

	for {
		select {
		case btn := <-button_ch:
			thisElevator.Requests[btn.Floor][btn.Button] = true
			elevatorList[thisElevator.Id] = *thisElevator
			infoUpdate_ch <- elevatorList[thisElevator.Id]  
			hallrequestassigner.AssignHallRequests(assigner_ch, elevatorList)

		case newState := <-newStatus_ch:

			if newState.Requests != thisElevator.Requests {
				newState.OrderClearedCounter++
			}

			elevatorList[newState.Id] = newState
			infoUpdate_ch <- newState  // Lag funksjonalitet for å ta imot counter
			hallrequestassigner.AssignHallRequests(assigner_ch, elevatorList)
			*thisElevator = newState

		case external := <-externalInfo:
			if external.OrderClearedCounter == thisElevator.OrderClearedCounter {
				//Her må request synkroniseres på en eller annen måte
			} else {
				elevatorList[external.Id] = external

			}
			hallrequestassigner.AssignHallRequests(assigner_ch, elevatorList)
			elevatorTimes[external.Id] = timer.Get_wall_time() + confirmOtherNodesTime

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) != 0 {
				for i := 0; i < len(peerUpdate.Lost); i++ {
					delete(elevatorList, peerUpdate.Lost[i])
					delete(elevatorTimes, peerUpdate.Lost[i])
				}
			} else if peerUpdate.New != "" {
				elevatorList[peerUpdate.New] = *new(elevator.Elevator) //Her brytes kanskje noen lover
				elevatorTimes[peerUpdate.New] = timer.Get_wall_time() + confirmOtherNodesTime
			}
			hallrequestassigner.AssignHallRequests(assigner_ch, elevatorList)
		case newAssignmentsMap := <- assigner_ch:
			//do shit
			fmt.Print("im doing stuff because orders were reassigned")
			fmt.Printf("newAssignmentsMap: %v\n", newAssignmentsMap)
		}

		for {
			infoUpdate_ch <- *thisElevator
			currentTime := timer.Get_wall_time()

			for id, Times := range elevatorTimes {
				if Times < currentTime {
					delete(elevatorList, id)
					delete(elevatorTimes, id)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
