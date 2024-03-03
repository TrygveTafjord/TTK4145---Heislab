package hallrequestassigner

import (
	"encoding/json"
	"fmt"

	"project.com/pkg/elevator"
)

//assuming: [up, down, cab] in the 4x3 matrix that is requestst

func AssignHallRequests(elevatorMap map[string]elevator.Elevator) map[string][4][2]bool {

	elevatorList := make([]elevator.Elevator, 0, len(elevatorMap))

	for _, v := range elevatorMap {
		elevatorList = append(elevatorList, v)
	}

	JSON := CreateJSON(elevatorList...)
	newAssignments := HallRequestAssigner(JSON)
	return newAssignments
}

func CreateJSON(elevators ...elevator.Elevator) []byte {
	var stateMaps []map[string]interface{}
	hallRequests := generateHallRequests(elevators)

	for _, e := range elevators {
		var direction string
		var behaviour string
		var cabRequests []bool

		switch e.Dirn {
		case elevator.MD_Up:
			direction = "up"
		case elevator.MD_Down:
			direction = "down"
		case elevator.MD_Stop:
			direction = "stop"
		}

		switch e.Behaviour {
		case elevator.EB_DoorOpen:
			behaviour = "doorOpen"
		case elevator.EB_Idle:
			behaviour = "idle"
		case elevator.EB_Moving:
			behaviour = "moving"
		}

		// Cab requests
		for _, request := range e.Requests {
			cabRequests = append(cabRequests, request[2])
		}

		floor := e.Floor // Assuming floor is non-negative.

		stateMap := map[string]interface{}{
			"behaviour":   behaviour,
			"floor":       floor,
			"direction":   direction,
			"cabRequests": cabRequests,
		}

		stateMaps = append(stateMaps, stateMap)
	}

	auxJSONMap := make(map[string]interface{})

	for i, stateMaps := range stateMaps {
		auxJSONMap[fmt.Sprintf("id_%d", i+1)] = stateMaps
	}

	masterJSONMap := map[string]interface{}{
		"hallRequests": hallRequests,
		"states":       auxJSONMap,
	}

	JSON, err := json.Marshal(masterJSONMap)
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s", err)
		return nil
	}

	return JSON
}

func generateHallRequests(elevators []elevator.Elevator) (resultMatrix [4][2]bool) {

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
