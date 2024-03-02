The HRA module works like this: 

You send a list (not map! has to be this way to be able to deal with a varying number of active elevators) of the active elevators into the function CreateJSON. This returns a []byte variable, that you can pass to the HallRequestAssigner function. This function returns a map[string][][2]bool, which is a map, where each key is either "id_1", "id_2" or "id_3", and the value is a list of lists, where each of the inner lists contains two boolean values representing the hall calls that the respective elevators are responsible for. 

To test or play around with it, you can paste this in main: 


var reqs [4][3]bool 
	reqs[3][1] = true
	//reqs[3][1] = true
	testElevator := elevator.Elevator{
		Floor:     2,
		Dirn:      elevator.MD_Up,
		Requests:  reqs,
		Behaviour: elevator.EB_Moving,
	}
	testElevator1 := elevator.Elevator{
		Floor:     2,
		Dirn:      elevator.MD_Down,
		Requests:  reqs,
		Behaviour: elevator.EB_Moving,
	}
	testElevator2 := elevator.Elevator{
		Floor:     1,
		Dirn:      elevator.MD_Up,
		Requests:  reqs,
		Behaviour: elevator.EB_Idle,
	}

	//var HRAMatrix [4][2]bool
	//var elevator_list []map[string]interface{}
	var result_bytes []byte

	list_of_elevs := []elevator.Elevator{
		testElevator, 
		testElevator1,
		testElevator2,
	}

	result_bytes = hallrequestassigner.CreateJSON(list_of_elevs...)
	var optimal_path map[string][][2]bool
	optimal_path = hallrequestassigner.HallRequestAssigner(result_bytes)


	fmt.Print(string(result_bytes))
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
	}