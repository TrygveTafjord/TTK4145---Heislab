package hallreqass

import (
	"encoding/json"
	"fmt"

	"project.com/pkg/elevator"
)

// func updateForeignElevs(foreignElevator elevator.Elevator){

// }

//assuming: [up, down, cab] in the 4x3 matrix that is requestst

func CreateStateMap(e elevator.Elevator) (states map[string]interface{}) {
	var direction string
	var behaviour string
	var floor int
	var cabRequests []bool
	//add hallrequestst at a later point - that is common for all elevators, this function is for individual elevators

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

	//cab requests
	for _, floor := range e.Requests {
		cabRequests = append(cabRequests, floor[2])
	}

	floor = e.Floor //NB has to be non-negative - is it always?

	stateMap := map[string]interface{}{
		"behaviour":   behaviour,
		"floor":       floor,
		"direction":   direction,
		"cabRequests": cabRequests,
	}

	return stateMap

}

func CreateMasterJSON(e elevator.Elevator,
	e1JSON map[string]interface{},
	e2JSON map[string]interface{},
	e3JSON map[string]interface{}) (JSON []byte) {

	var hallRequests [][]bool
	for _, floor := range e.Requests {
		hallRequests = append(hallRequests, floor[:2])
	}
	//currently, the loop above just gets the hallreq info from one of the elevators
	//at a later point, this should probably be changed to a list of the hallreqs from the infobank
	//should be fine for now, the assumption is that they all have the same info anyway

	auxJSONMap := map[string]interface{}{
		"id_1": e1JSON,
		"id_2": e2JSON,
		"id_3": e3JSON,
	}

	masterJSONMap := map[string]interface{}{
		"hallRequests": hallRequests,
		"states":       auxJSONMap,
	}

	JSON, err := json.Marshal(masterJSONMap)
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s", err)
	}

	return JSON
}

// func updateLocalInfoFromJSON(){

// }
