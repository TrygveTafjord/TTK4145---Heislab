package diagnostics

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
)


func Diagnostics(updateFromFSM_ch chan elevator.Elevator, setObstructedState_ch chan bool){
	timeInSameStateWhileOrders := 0
	currentState := <-updateFromFSM_ch
	prevState 	 := currentState

	selfCheck_ch := make(chan bool)
	go PeriodicCheck(selfCheck_ch)

	for{
		select{
			case updatedElevator := <- updateFromFSM_ch:
				prevState = currentState
				currentState = updatedElevator

			case <-selfCheck_ch:
				if hasRequest(currentState) && currentState.State.Behaviour == prevState.State.Behaviour {
					timeInSameStateWhileOrders += 1
				} else {
					timeInSameStateWhileOrders = 0
				}

				diagnose := selfDiagnose(currentState, prevState, timeInSameStateWhileOrders)
								
				switch diagnose {
				
					case Healthy:
						break
					case Obstructed:
						setObstructedState_ch <- true
					case Reinitialize:
						fmt.Printf("Reinitialize!\n")
				}
				prevState = currentState
			}
		}
}

func PeriodicCheck(selfCheck_ch chan bool) {
	for {
		time.Sleep(1000 * time.Millisecond)
		selfCheck_ch <- true
	}
}

func selfDiagnose(currentState elevator.Elevator, prevState elevator.Elevator, timeInSameStateWhileOrders int) Diagnose {
	
	if timeInSameStateWhileOrders > 0 && currentState.State.Behaviour == elevator.EB_Idle{
		return Reinitialize

	} else if timeInSameStateWhileOrders > 10 && elevator.GetObstruction() {
		return Obstructed

	} else if timeInSameStateWhileOrders > 15 && !elevator.GetObstruction() {
		return Reinitialize

	}
	return Healthy
}



func hasRequest(e elevator.Elevator) bool {
	for i := 0; i < elevator.N_FLOORS; i++ {
		for j := 0; j < elevator.N_BUTTONS; j++ {
			if e.Requests[i][j] {
				return true
			}
		}
	}
	return false
}
// func Selfdiagnose(currentState elevator.Elevator, prevState elevator.Elevator, obstruction bool, standstill *int) Diagnose {


// 	if hasRequest(currentState) && currentState.State.Behaviour == prevState.State.Behaviour {

// 		switch currentState.State.Behaviour {

// 		case elevator.EB_Idle:
// 			return Problem

// 		case elevator.EB_DoorOpen:
// 			if currentState.State.Floor == prevState.State.Floor {
// 				*standstill += 1
// 			}
// 		case elevator.EB_Moving:
// 			if currentState.State.Floor == prevState.State.Floor {
// 				*standstill += 1
// 			}
// 		}
// 		*prevState = *elevator

// 		if *standstill > 10 && obstruction {
// 			return Obstructed
// 		} else if *standstill == 20 && !obstruction {
// 			return Problem
// 		}

// 	} else if obstruction {
// 		return Unchanged

// 	} else {
// 		*standstill = 0
// 	}
// 	return Healthy
// }

