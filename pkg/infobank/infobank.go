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

func Infobank_FSM(newStatus_ch chan elevator.Elevator, infoUpdate_ch chan elevator.Elevator, externalInfo chan elevator.Elevator, peerUpdate_ch chan network.PeerUpdate, assigner_ch chan map[string][4][2]bool) {
	elevatorList := make(map[string]elevator.Elevator)
	elevatorTimes := make(map[string]float64)
	this_elevator := new(elevator.Elevator)

	for {
		select {
		case this := <-newStatus_ch:
			elevatorList[this.Id] = this
			infoUpdate_ch <- this
			hallrequestassigner.AssignHallRequests(assigner_ch, elevatorList)
			*this_elevator = this

		case external := <-externalInfo:
			if external.Completed_order_counter == this_elevator.Completed_order_counter {
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
			//si fra til FSM om DENNE heisen
			newAssignments_ch <- newAssignmentsMap["id_1"] //må endre så ID i NewAssigmentMap er IP, som i elevatorList
			updateInformation(newAssignmentsMap, elevatorList) //oppdater elevatorList med de nye hall-requestene
			//si fra til Network om de andre heisene? lagre informasjonen et sted? 
			//"or'e sammen matrisene for master matrix"
		}

		for {
			infoUpdate_ch <- *this_elevator
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

func updateInformation(newAssignmentsMap map[string][4][2]bool, elevatorMap map[string]elevator.Elevator){
	//not finished
	fmt.Print("information was (not really) updated")
	//for key, value := range newAssignmentsMap {
	//}
}