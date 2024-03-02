package elevator

import (
	"time"

	"project.com/pkg/network"
	"project.com/pkg/timer"

)
var confirmOtherNodesTime float64 = 2

func Infobank_FSM(newStatus_ch chan Elevator, infoUpdate_ch chan Elevator, externalInfo chan Elevator, peerUpdate_ch chan network.PeerUpdate, assigner_ch chan map[string]Elevator) {
	elevatorList := make(map[string]Elevator)
	elevatorTimes := make(map[string]float64)
	this_elevator := new(Elevator)
	
	for {
		select {
		case this := <-newStatus_ch:
			elevatorList[this.Id] = this
			infoUpdate_ch <- this
			assigner_ch <- elevatorList
			*this_elevator = this

		case external := <-externalInfo:
			if (external.Completed_order_counter == this_elevator.Completed_order_counter){
				//Her må request synkroniseres på en eller annen måte
			} else
			{
				elevatorList[external.Id] = external
				
			}
			assigner_ch <- elevatorList
			elevatorTimes[external.Id] = timer.Get_wall_time() + confirmOtherNodesTime

		case peerUpdate := <-peerUpdate_ch:
			if len(peerUpdate.Lost) != 0{
				for i := 0; i < len(peerUpdate.Lost); i++ {
					delete(elevatorList, peerUpdate.Lost[i])
					delete(elevatorTimes,peerUpdate.Lost[i])
				}
			} else if peerUpdate.New != "" {
				elevatorList[peerUpdate.New] = *new(Elevator) //Her brytes kanskje noen lover
				elevatorTimes[peerUpdate.New] = timer.Get_wall_time() + confirmOtherNodesTime
			}
			assigner_ch <- elevatorList
		}



		for {
			infoUpdate_ch <- *this_elevator
			currentTime := timer.Get_wall_time()
			
			for id, Times := range elevatorTimes {
				if Times < currentTime {
					delete(elevatorList, id)
					delete(elevatorTimes,id)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
