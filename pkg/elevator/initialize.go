package elevator

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)
func ElevatorInit(elevStatusUpdate_ch chan Elevator, lastID string, ID string) {
	var e Elevator
	SetDoorOpenLamp(false)

	e.Id = ID
	e.similarity = 0

	floor := GetFloor()

	var cabCalls []bool

	//reset buttons
	if len(lastID) == 0 {
		for floor := 0; floor < 4; floor++ {
			for btn := 0; btn < 3; btn++ {
				SetButtonLamp(ButtonType(btn), floor, false)
			}
		}
		e.OrderClearedCounter = 0
		e.OrderCounter = 0
	}else{

		cabCalls, e.OrderClearedCounter, e.OrderCounter = readCSV(lastID)

		for floor := 0; floor < 4; floor++ {
			for btn := 0; btn < 2; btn++ {
				SetButtonLamp(ButtonType(btn), floor, false)
			}
			SetButtonLamp(ButtonType(BT_Cab), floor, cabCalls[floor])
			e.Requests[floor][BT_Cab] = cabCalls[floor]
		}
	}

	fmt.Print("The finished order matrix is: \n")
	for floor := 0; floor < 4; floor++ {
		for btn := 0; btn < 3; btn++ {
			fmt.Printf("%v", e.Requests[floor][btn])
		}
		fmt.Print("\n")
	}	


	if floor == -1 {
		SetMotorDirection(MD_Down)
		for floor == -1 {
			floor := GetFloor()
			if floor != (-1) {
				SetMotorDirection(MD_Stop)
				break
			}
		}
	}
	e.Floor = floor
	e.Dirn = MD_Stop
	e.Behaviour = EB_Idle
	elevStatusUpdate_ch <- e
}

func readCSV(previousID string) ([]bool, int, int){
	fmt.Printf("This is the ID in the readCSV function: %v \n", previousID)
	file, err := os.Open(previousID)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
	}
	defer file.Close()

	var returnSlice []bool

	csvReader := csv.NewReader(file)
	records, _ := csvReader.ReadAll()

	for i := 0; i < 4; i++ {
		fmt.Printf("Record %d: %v\n", i, records[i][0])
		if records[i][0] == "true"{
			returnSlice = append(returnSlice, true)
		}else{
			returnSlice = append(returnSlice, false)
		}
	}

	orderClearCounterString := records[4][0]
	OCC := strings.SplitN(orderClearCounterString, ":", 2)
	orderClearCounter, _ := strconv.Atoi(OCC[1])

	orderCounterString := records[4][0]
	OC := strings.SplitN(orderCounterString, ":", 2)
	orderCounter, _ := strconv.Atoi(OC[1])

	return returnSlice, orderClearCounter, orderCounter
}
