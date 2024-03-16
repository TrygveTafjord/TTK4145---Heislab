package assigner

import (
	"encoding/json"
	"fmt"

	"project.com/pkg/elevator"
)

func AssignHallRequests(assignerList []AssignerInput) map[string][4][2]bool {

	JSON, masterJSONMap := createJSON(assignerList...)
	
	return HallRequestAssigner(JSON, masterJSONMap)
}

func createJSON(elevators ...AssignerInput) ([]byte, map[string]interface{}) {
	hallRequests := generateHallRequests(elevators)
	auxJSONMap := make(map[string]interface{})

	for _, e := range elevators {
		var direction string
		var behaviour string
		var cabRequests []bool

		switch e.State.Dirn {
		case elevator.MD_Up:
			direction = "up"
		case elevator.MD_Down:
			direction = "down"
		case elevator.MD_Stop:
			direction = "stop"
		}

		switch e.State.Behaviour {
		case elevator.EB_DoorOpen:
			behaviour = "doorOpen"
		case elevator.EB_Idle:
			behaviour = "idle"
		case elevator.EB_Moving:
			behaviour = "moving"
		}

		for _, request := range e.Requests {
			cabRequests = append(cabRequests, request[2])
		}

		floor := e.State.Floor

		stateMap := map[string]interface{}{
			"behaviour":   behaviour,
			"floor":       floor,
			"direction":   direction,
			"cabRequests": cabRequests,
		}

		auxJSONMap[e.Id] = stateMap
	}

	masterJSONMap := map[string]interface{}{
		"hallRequests": hallRequests,
		"states":       auxJSONMap,
	}

	JSON, err := json.MarshalIndent(masterJSONMap, "", "    ")
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s", err)
		return nil, masterJSONMap
	}

	return JSON, masterJSONMap
}

func generateHallRequests(elevators []AssignerInput) (resultMatrix [4][2]bool) {

	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			for _, elevator := range elevators {
				resultMatrix[i][j] = resultMatrix[i][j] || elevator.Requests[i][j]
				if resultMatrix[i][j] {
					break
				}
			}
		}
	}
	return resultMatrix
}
