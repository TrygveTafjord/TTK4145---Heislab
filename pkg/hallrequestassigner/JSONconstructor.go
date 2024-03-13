package hallrequestassigner

import (
	"encoding/json"
	"fmt"

	"project.com/pkg/elevator"
)



//assuming: [up, down, cab] in the 4x3 matrix that is requestst

func AssignHallRequests(elevatorMap map[string]elevator.Elevator) map[string][4][2]bool{
	if len(elevatorMap) == 1 {
		return AssignHallRequestsSingle(elevatorMap)
	}else{
		return AssignHallRequestsMultiple(elevatorMap)
	}
}

func AssignHallRequestsSingle(elevatorMap map[string]elevator.Elevator) map[string][4][2]bool {

	elevatorList := make([]elevator.Elevator, 0, len(elevatorMap))

	for _, v := range elevatorMap {
		elevatorList = append(elevatorList, v)
	}

	JSON := CreateJSON(elevatorList...)
	
	return HallRequestAssigner(JSON)
}


func AssignHallRequestsMultiple(elevatorMap map[string]elevator.Elevator) map[string][4][2]bool {
	healthyElevators:= make(map[string]elevator.Elevator)
	obstructedElevators := []string{}
	obstructedOrders := [4][2] bool{}
	emptyRequests := [4][2] bool{}

	
	resolveObstrucedElevators(elevatorMap,&healthyElevators, &obstructedElevators, &obstructedOrders)
	redistributeObstructedOrders(len(obstructedElevators), &healthyElevators, obstructedOrders)
	
	elevatorList := make([]elevator.Elevator, 0, len(healthyElevators))

	for _, v := range healthyElevators {
		elevatorList = append(elevatorList, v)
	}

	JSON := CreateJSON(elevatorList...)
	returnMap := HallRequestAssigner(JSON)

	for _,Id := range obstructedElevators{
		returnMap[Id] = emptyRequests
	}
	return returnMap
}


func CreateJSON(elevators ...elevator.Elevator) []byte {
	hallRequests := generateHallRequests(elevators)
	auxJSONMap := make(map[string]interface{})

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

		auxJSONMap[e.Id] = stateMap
	}

	masterJSONMap := map[string]interface{}{
		"hallRequests": hallRequests,
		"states":       auxJSONMap,
	}

	/*fmt.Println("Hall Requests Matrix from JSON maker:")
	for i, floorRequests := range hallRequests {
		for j, request := range floorRequests {
			if j == 0 {
				fmt.Printf("%v, [", i)
			}
			fmt.Printf("%t", request)
			if j < len(floorRequests)-1 {
				fmt.Print(", ")
			} else {
				fmt.Println("]")
			}
		}
	}*/

	JSON, err := json.MarshalIndent(masterJSONMap, "", "    ") // "" as prefix and "    " (4 spaces) as indent
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s", err)
		return nil
	}

	// Print the nicely formatted JSON string
	//fmt.Println(string(JSON))

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







func resolveObstrucedElevators(elevatorMap map[string]elevator.Elevator,healthyElevators *map[string]elevator.Elevator, obstructedElevators *[]string, obstructedOrders *[4][2]bool){
	obstructed := *obstructedElevators
	orders := *obstructedOrders
	for _, v := range elevatorMap {
		if v.Obstructed {
			
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


func redistributeObstructedOrders(obstructedElevators int, healthyElevators *map[string]elevator.Elevator, obstructedOrders [4][2]bool){
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
	
