package initialize

import (
	// "encoding/csv"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	// "strconv"
	// "strings"

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

	//Is there a dead process?
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

func readCSV(previousID string) ([]bool, int, int, elevator.ElevatorBehaviour, int) {
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

	orderClearCounterString := records[4][0]
	OCC := strings.SplitN(orderClearCounterString, ":", 2)
	orderClearCounter, _ := strconv.Atoi(OCC[1])

	orderCounterString := records[5][0]
	OC := strings.SplitN(orderCounterString, ":", 2)
	orderCounter, _ := strconv.Atoi(OC[1])

	behaviourString := records[6][0]
	BH := strings.SplitN(behaviourString, ":", 2)
	index, _ := strconv.Atoi(BH[1])
	behaviour := elevator.ElevatorBehaviour(index)

	directionString := records[7][0]
	DIR := strings.SplitN(directionString, ":", 2)
	direction, _ := strconv.Atoi(DIR[1])

	return returnSlice, orderClearCounter, orderCounter, behaviour, direction
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
	e.State.Obstructed, e_IB.State.Obstructed = false, false
	e_IB.Id = ID
	e_IB.OrderClearedCounter = 0
	e_IB.OrderCounter = 0
	file.Close()
	return e, e_IB
}

func initReplacementObjects(ID string) (elevator.Elevator, infobank.ElevatorInfo) {
	var cabCalls []bool
	var direction int
	var e elevator.Elevator
	var e_IB infobank.ElevatorInfo

	cabCalls, e_IB.OrderClearedCounter, e_IB.OrderCounter, e.State.Behaviour, direction = readCSV(ID)

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
				elevator.SetMotorDirection(elevator.MD_Stop)
				break
			}
		}
	}
	e.State.Floor, e_IB.State.Floor = floor, floor
	e_IB.Id = ID
	return e, e_IB
}
