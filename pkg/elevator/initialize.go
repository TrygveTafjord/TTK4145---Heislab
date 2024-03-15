package elevator

// import (
// 	"encoding/csv"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"strings"

// 	"project.com/pkg/infobank"
// 	"project.com/pkg/elevator"
// )

// /*func ElevatorInit(elevStatusUpdate_ch chan Elevator, prevID string, ID string) {
// 	var e Elevator
// 	//SetDoorOpenLamp(false)

// 	e.Id = ID
// 	e.similarity = 0

// 	var cabCalls []bool

// 	for floor := 0; floor < 4; floor++ {
// 		for btn := 0; btn < 3; btn++ {
// 			SetButtonLamp(ButtonType(btn), floor, false)
// 		}
// 	}
// 	e.OrderClearedCounter = 0
// 	e.OrderCounter = 0
// 	e.Floor = GetFloor()

// 	if e.Floor == -1 {
// 		SetMotorDirection(MD_Down)
// 		for e.Floor == -1 {
// 			if GetFloor() != (-1) {
// 				SetMotorDirection(MD_Stop)
// 				break
// 			}
// 		}
// 		e.Floor = GetFloor()
// 	}

// 	//elevStatusUpdate_ch <- e //send dummy object to get to the good stuff in FSM

// 	//reset buttons
// 	if !isFirstProcess(prevID) {
// 		cabCalls, e.OrderClearedCounter, e.OrderCounter, e.Behaviour = readCSV(prevID)
// 		e.Floor = GetFloor()
// 		e.Behaviour = EB_Idle
// 		if e.Behaviour != EB_DoorOpen {
// 			e.Behaviour = EB_Idle
// 			if e.Floor == -1 {
// 				SetMotorDirection(MD_Down)
// 				for e.Floor == -1 {
// 					if GetFloor() != (-1) {
// 						SetMotorDirection(MD_Stop)
// 						break
// 					}
// 				}
// 				e.Floor = GetFloor()
// 			}
// 		}
// 		fmt.Printf("The cab calls are: %v", cabCalls)
// 	}
// 	e.Dirn = MD_Stop
// 	fmt.Printf("The elevator i am sending has an order counter of %v", e.OrderCounter)
// 	time.Sleep(1000 * time.Millisecond)
// 	elevStatusUpdate_ch <- e
// 	e.OrderCounter--
// 	elevStatusUpdate_ch <- e
// }*/

// func ElevatorInit(elevInitInfobank_ch chan infobank.ElevatorInfo, elevInitFSM_ch chan elevator.Elevator, lastID string, ID string) {
// 	var e Elevator
// 	//e.State.Standstill = 0
// 	//e.Id = ID
// 	//var cabCalls []bool
// 	//var direction int

// 	//reset buttons
	
// 	for floor := 0; floor < 4; floor++ {
// 		for btn := 0; btn < 3; btn++ {
// 			SetButtonLamp(ButtonType(btn), floor, false)
// 			e.Requests[floor][btn] = false
// 			e.Lights[floor][btn] = false
// 		}
// 	}
// 	//e.OrderClearedCounter = 0
// 	//e.OrderCounter = 0

// 	floor := GetFloor()
// 	if floor == -1 {
// 		SetMotorDirection(MD_Down)
// 		for floor == -1 {
// 			floor = GetFloor()
// 			if floor != (-1) {
// 				SetMotorDirection(MD_Stop)
// 				break
// 			}
// 		}
// 	}
// 	e.State.Floor = floor
// 	e.State.Behaviour = EB_Idle
// 	e.State.Obstructed = false
// 	e.State.Dirn = MD_Stop //Sketchy?????

// 	elevInitFSM_ch <- e
// 	// } else {

// 	// 	cabCalls, e.OrderClearedCounter, e.OrderCounter, e.Behaviour, direction = readCSV(lastID)

// 	// 	for floor := 0; floor < 4; floor++ {
// 	// 		for btn := 0; btn < 2; btn++ {
// 	// 			SetButtonLamp(ButtonType(btn), floor, false)
// 	// 		}
// 	// 		SetButtonLamp(ButtonType(BT_Cab), floor, cabCalls[floor])
// 	// 		e.Requests[floor][BT_Cab] = cabCalls[floor]
// 	// 		e.Lights[floor][BT_Cab] = cabCalls[floor]
// 	// 	}
// 	// }

// 	// switch direction {
//     // case 1:
// 	// 	e.Dirn = MD_Up
//     // case -1:
//     //     e.Dirn = MD_Down
//     // default:
//     //     e.Dirn = MD_Stop
// 	// }

// 	// floor := GetFloor()
// 	// if floor == -1 {
// 	// 	SetMotorDirection(e.Dirn)
// 	// 	for floor == -1 {
// 	// 		floor = GetFloor()
// 	// 		if floor != (-1) {
// 	// 			SetMotorDirection(MD_Stop)
// 	// 			break
// 	// 		}
// 	// 	}
// 	// }
// 	// e.Floor = floor
// 	// e.OrderCounter--
// 	// elevInitFSM_ch <- e
// 	// e.OrderCounter++
// 	// toFSM_ch <- e
// }

// func readCSV(previousID string) ([]bool, int, int, ElevatorBehaviour, int) {
// 	file, err := os.Open(previousID)
// 	if err != nil {
// 		fmt.Printf("Failed to open file: %v\n", err)
// 	}
// 	defer file.Close()

// 	var returnSlice []bool

// 	csvReader := csv.NewReader(file)
// 	records, _ := csvReader.ReadAll()

// 	for i := 0; i < 4; i++ {
// 		if records[i][0] == "true" {
// 			returnSlice = append(returnSlice, true)
// 		} else {
// 			returnSlice = append(returnSlice, false)
// 		}
// 	}

// 	orderClearCounterString := records[4][0]
// 	OCC := strings.SplitN(orderClearCounterString, ":", 2)
// 	orderClearCounter, _ := strconv.Atoi(OCC[1])

// 	orderCounterString := records[5][0]
// 	OC := strings.SplitN(orderCounterString, ":", 2)
// 	orderCounter, _ := strconv.Atoi(OC[1])

// 	behaviourString := records[6][0]
// 	BH := strings.SplitN(behaviourString, ":", 2)
// 	index, _ := strconv.Atoi(BH[1])
// 	behaviour := ElevatorBehaviour(index)

// 	directionString := records[7][0]
// 	DIR := strings.SplitN(directionString, ":", 2)
// 	direction, _ := strconv.Atoi(DIR[1])
	

// 	return returnSlice, orderClearCounter, orderCounter, behaviour, direction
// }

// /*func isFirstProcess(prevID string) bool {
// 	return len(prevID) == 0
// }*/

// /*func findNearestFloor() {
// 	SetMotorDirection(MD_Down)
// 			for e.Floor == -1 {
// 				if GetFloor() != (-1) {
// 					SetMotorDirection(MD_Stop)
// 					break
// 				}
// 			}
// 			e.Floor = GetFloor()
// }*/
