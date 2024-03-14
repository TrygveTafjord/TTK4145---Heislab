package assigner

import (
	"encoding/json"
	"fmt"

	"project.com/pkg/elevator"
)

//assuming: [up, down, cab] in the 4x3 matrix that is requestst

func AssignHallRequests(assignerList []AssignerInput) map[string][4][2]bool{
	if len(assignerList) == 1 {
		return AssignHallRequestsSingle(assignerList)
	}else{
		return AssignHallRequestsMultiple(assignerList)
	}
}

func AssignHallRequestsSingle(assignerList []AssignerInput) map[string][4][2]bool {

	JSON := CreateJSON(assignerList...)
	
	return HallRequestAssigner(JSON)
}


func AssignHallRequestsMultiple(assignerList []AssignerInput) map[string][4][2]bool {
	healthyElevators:= make(map[string]AssignerInput)
	obstructedElevators := []string{}
	obstructedOrders    := [4][2] bool{}
	emptyRequests       := [4][2] bool{}

	//if obstruction()
	resolveObstrucedElevators(assignerList ,&healthyElevators, &obstructedElevators, &obstructedOrders)
	redistributeObstructedOrders(len(obstructedElevators), &healthyElevators, obstructedOrders)

	JSON := CreateJSON(assignerList...)
	returnMap := HallRequestAssigner(JSON)

	for _,Id := range obstructedElevators{
		returnMap[Id] = emptyRequests
	}

	return returnMap
}


func CreateJSON(elevators ...AssignerInput) []byte {
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

		// Cab requests
		for _, request := range e.Requests {
			cabRequests = append(cabRequests, request[2])
		}

		floor := e.State.Floor // Assuming floor is non-negative.

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


	JSON, err := json.MarshalIndent(masterJSONMap, "", "    ") // "" as prefix and "    " (4 spaces) as indent
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s", err)
		return nil
	}

	// Print the nicely formatted JSON string
	//fmt.Println(string(JSON))

	return JSON
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







func resolveObstrucedElevators(assignerList []AssignerInput, healthyElevators *map[string]AssignerInput, obstructedElevators *[]string, obstructedOrders *[4][2]bool){
	obstructed := *obstructedElevators
	orders := *obstructedOrders
	for _, v := range assignerList {
		if v.State.Obstructed {
			
			obstructed = append(obstructed,v.Id)
			for i := 0; i < elevator.N_FLOORS; i++ {
				for j := 0; j < elevator.N_BUTTONS-1; j++ {
					orders[i][j] = orders[i][j] || v.Requests[i][j]
				}
			}
		} else
		{
		(*healthyElevators)[v.Id] = v
		}
	}

	*obstructedElevators = obstructed
	*obstructedOrders = orders
}


func redistributeObstructedOrders(obstructedElevators int, healthyElevators *map[string]AssignerInput, obstructedOrders [4][2]bool){
	if obstructedElevators!= 0 {
		for id,v := range *healthyElevators{
			tempElev := v
			
			for i := 0; i < elevator.N_FLOORS; i++ {
				for j := 0; j < elevator.N_BUTTONS - 1; j++ {
					tempElev.Requests[i][j] = tempElev.Requests[i][j] || obstructedOrders[i][j]
				}
			}
			(*healthyElevators)[id] = tempElev
			break
		}
	}
}
	
