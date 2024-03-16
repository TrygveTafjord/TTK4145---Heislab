package initialize

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"project.com/pkg/elevator"
	"project.com/pkg/infobank"
)

func ElevatorInit(elevInitInfobank_ch chan infobank.ElevatorInfo,
	elevInitFSM_ch chan elevator.Elevator,
	networkInit_ch chan string,
	ID string,
) {
	var FSMelevator elevator.Elevator
	var IBelevator infobank.ElevatorInfo

	_, err := os.Open(ID)

	if err != nil {
		FSMelevator, IBelevator = initDefaultObjects(ID)

	} else {
		FSMelevator, IBelevator = initReplacementObjects(ID)
	}

	networkInit_ch <- ID
	elevInitFSM_ch <- FSMelevator
	elevInitInfobank_ch <- IBelevator

}

func readCSV(previousID string) ([]bool, elevator.ElevatorBehaviour, int) {
	file, err := os.Open(previousID)
	if err != nil {
		fmt.Printf("Failed to open file in READ: %v\n", err)
	}
	defer file.Close()

	var returnSlice []bool

	csvReader := csv.NewReader(file)
	records, _ := csvReader.ReadAll()

	for i := 0; i < 4; i++ {
		if records[i][0] == "true" {
			returnSlice = append(returnSlice, true)
		} else {
			returnSlice = append(returnSlice, false)
		}
	}

	behaviourString := records[4][0]
	BH := strings.SplitN(behaviourString, ":", 2)
	index, _ := strconv.Atoi(BH[1])
	behaviour := elevator.ElevatorBehaviour(index)

	directionString := records[5][0]
	DIR := strings.SplitN(directionString, ":", 2)
	direction, _ := strconv.Atoi(DIR[1])

	return returnSlice, behaviour, direction
}

func initDefaultObjects(ID string) (elevator.Elevator, infobank.ElevatorInfo) {
	var e elevator.Elevator
	var e_IB infobank.ElevatorInfo

	file, err := os.Create(ID)
	if err != nil {
		fmt.Printf("Failed to create file in initialize: %v \n \n", err)
	}

	for floor := 0; floor < 4; floor++ {
		for btn := 0; btn < 3; btn++ {
			elevator.SetButtonLamp(elevator.ButtonType(btn), floor, false)
			e.Requests[floor][btn] = false
			e.Lights[floor][btn] = false
			e_IB.Requests[floor][btn] = false
			e_IB.Lights[floor][btn] = false
		}
	}

	floor := elevator.GetFloor()
	if floor == -1 {
		elevator.SetMotorDirection(elevator.MD_Down)
		for floor == -1 {
			floor = elevator.GetFloor()
			if floor != (-1) {
				elevator.SetMotorDirection(elevator.MD_Stop)
				break
			}
		}
	}
	e.State.Floor, e_IB.State.Floor = floor, floor
	e.State.Behaviour, e_IB.State.Behaviour = elevator.EB_Idle, elevator.EB_Idle
	e.State.OutOfService, e_IB.State.OutOfService = false, false
	e_IB.Id = ID

	file.Close()
	return e, e_IB
}

func initReplacementObjects(ID string) (elevator.Elevator, infobank.ElevatorInfo) {
	fmt.Printf("\n \n \n ---------REINIT-------- \n \n \n ")

	var cabCalls []bool
	var direction int
	var e elevator.Elevator
	var e_IB infobank.ElevatorInfo

	cabCalls, e.State.Behaviour, direction = readCSV(ID)

	for floor := 0; floor < 4; floor++ {
		for btn := 0; btn < 2; btn++ {
			elevator.SetButtonLamp(elevator.ButtonType(btn), floor, false)
		}
		e_IB.Requests[floor][elevator.BT_Cab] = cabCalls[floor]
		elevator.SetButtonLamp(elevator.ButtonType(elevator.BT_Cab), floor, cabCalls[floor])
	}

	switch direction {
	case 1:
		e.State.Dirn = elevator.MD_Up
		e_IB.State.Dirn = elevator.MD_Up
	case -1:
		e.State.Dirn = elevator.MD_Down
		e_IB.State.Dirn = elevator.MD_Down
	default:
		e.State.Dirn = elevator.MD_Stop
		e_IB.State.Dirn = elevator.MD_Stop
	}

	floor := elevator.GetFloor()
	if floor == -1 {
		elevator.SetMotorDirection(e.State.Dirn)
		for floor == -1 {
			floor = elevator.GetFloor()
			if floor != (-1) {
				break
			}
		}
	}
	elevator.SetMotorDirection(elevator.MD_Stop)

	e.State.Floor, e_IB.State.Floor = floor, floor
	e.State.OutOfService, e_IB.State.OutOfService = false, false
	e_IB.Id = ID
	
	return e, e_IB
}
