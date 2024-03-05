package main

import (
	"time"

	"project.com/pkg/elevator"
	
	"project.com/pkg/infobank"
	"project.com/pkg/network"
)

func main() {
	elevator.Init("localhost:15657", 4)

	elevStatusUpdate_ch := make(chan elevator.Elevator, 5)

	networkUpdateTx_ch := make(chan elevator.Elevator, 5)
	networkUpdateRx_ch := make(chan elevator.Elevator, 5)
	updatePeers_ch := make(chan network.PeerUpdate)

	go elevator.FSM(elevStatusUpdate_ch)

	go infobank.Infobank_FSM(elevStatusUpdate_ch, networkUpdateTx_ch, networkUpdateRx_ch)

	go network.Network_fsm(networkUpdateTx_ch, networkUpdateRx_ch, updatePeers_ch)
	//var HRAMatrix [4][2]bool
	//var elevator_list []map[string]interface{}

	/*var reqs [4][3]bool
	reqs[3][1] = true
	var reqs2 [4][3]bool
	reqs2[2][1] = true
	var reqs3 [4][3]bool

	testElevator := elevator.Elevator{Id: "12345", OrderClearedCounter: 3, Floor: 1, Dirn: 0, Requests: reqs, Behaviour: elevator.EB_Idle}
	testElevator1 := elevator.Elevator{Id: "6789", OrderClearedCounter: 2, Floor: 3, Dirn: 0, Requests: reqs2, Behaviour: elevator.EB_Idle}
	testElevator2 := elevator.Elevator{Id: "101112", OrderClearedCounter: 1, Floor: 2, Dirn: 0, Requests: reqs3, Behaviour: elevator.EB_Idle}

	map_of_elevs := make(map[string]elevator.Elevator)

	map_of_elevs["12345"] = testElevator
	map_of_elevs["6789"] = testElevator1
	map_of_elevs["101112"] = testElevator2

	var optimal_path map[string][4][2]bool

	optimal_path = hallrequestassigner.AssignHallRequests(map_of_elevs)
	fmt.Println()
	fmt.Println()
	for key, slice := range optimal_path {
		fmt.Printf("Key: %s, Values: ", key)
		// Iterating through the slice for each key
		for _, pair := range slice {
			// Printing each boolean pair
			fmt.Printf("[%t, %t] ", pair[0], pair[1])
		}
		fmt.Println() // Newline for each key
	}*/
	for {
		time.Sleep(2000 * time.Millisecond)
	}
}
