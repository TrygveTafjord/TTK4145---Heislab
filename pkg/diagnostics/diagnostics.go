package diagnostics

import (
	"os"
	"time"

	"project.com/pkg/elevator"
)

func Diagnostics(updateFromFSM_ch chan elevator.Elevator, obstructionDiagnose_ch chan bool) {
	timeInSameStateWhileOrders := 0
	currentState := <-updateFromFSM_ch
	prevState := currentState

	selfCheck_ch := make(chan bool)
	go PeriodicCheck(selfCheck_ch)

	for {
		select {
		case updatedElevator := <-updateFromFSM_ch:
			prevState = currentState
			currentState = updatedElevator

		case <-selfCheck_ch:
			if hasRequest(currentState) && currentState.State == prevState.State && !currentState.State.OutOfService {
				timeInSameStateWhileOrders++
			} else {
				timeInSameStateWhileOrders = 0
			}

			diagnose := selfDiagnose(currentState, timeInSameStateWhileOrders)

			switch diagnose {

			case Obstructed:
				obstructionDiagnose_ch <- true

			case Reinitialize:
				os.Exit(1)
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

func selfDiagnose(currentState elevator.Elevator, timeInSameStateWhileOrders int) Diagnose {

	if timeInSameStateWhileOrders > 0 && currentState.State.Behaviour == elevator.EB_Idle {
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
