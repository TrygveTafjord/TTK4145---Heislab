package elevator

import (
	"fmt"
	"time"

	"project.com/pkg/timer"
)

func FSM(requestUpdate_ch chan [N_FLOORS][N_BUTTONS]bool, clearRequestToInfobank_ch chan []ButtonEvent, stateToInfobank_ch chan State, lightUpdate_ch chan [N_FLOORS][N_BUTTONS]bool, elevatorInit_ch chan Elevator) {

	floorSensor_ch := make(chan int)
	obstruction_ch := make(chan bool)
	timer_ch := make(chan bool)
	selfCheck_ch := make(chan bool)

	go PollFloorSensor(floorSensor_ch)
	go PollObstructionSwitch(obstruction_ch)
	go PeriodicCheck(selfCheck_ch)

	elevator := new(Elevator)
	*elevator = <-elevatorInit_ch
	//prevelevator := *elevator
	obstruction := GetObstruction()
	//standstill := 0

	for {
		select {
		case requests := <-requestUpdate_ch:
			elevator.Requests = requests
			fsmNewRequests(elevator, timer_ch)
			if elevator.Requests != requests {
				clearRequestToInfobank_ch <- getClearedRequests(requests, elevator.Requests)
			}
			stateToInfobank_ch <- elevator.State

		case lights := <-lightUpdate_ch:
			elevator.Lights = lights
			fmt.Printf("Det kommer en light update, og den er: %v! \n \n", lights)
			setAllLights(elevator)

		case newFloor := <-floorSensor_ch:
			fmt.Print(" \n \n Floor sensor spoke! \n \n")
			requestsBeforeNewFloor := elevator.Requests
			fsmOnFloorArrival(elevator, newFloor, timer_ch)
			stateToInfobank_ch <- elevator.State
			if requestsBeforeNewFloor != elevator.Requests {
				clearRequestToInfobank_ch <- getClearedRequests(requestsBeforeNewFloor, elevator.Requests)
			}

		case <-timer_ch:
			HandleDeparture(elevator, timer_ch, obstruction)
			stateToInfobank_ch <- elevator.State

		case obstruction = <-obstruction_ch:
			if !obstruction && elevator.State.Behaviour == EB_DoorOpen {
				go timer.Run_timer(3, timer_ch)
				//sette lys?
			}
		}
	}
}

func HandleDeparture(e *Elevator, timer_ch chan bool, obstruction bool) {
	if obstruction && e.State.Behaviour == EB_DoorOpen {
		go timer.Run_timer(3, timer_ch)
	} else {
		e.State.Dirn, e.State.Behaviour = GetDirectionAndBehaviour(e)

		switch e.State.Behaviour {

		case EB_DoorOpen:
			fmt.Printf("DET SKJEDDE HANDLE DEPARTURE, HVAFAEN \n")
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e)
			go timer.Run_timer(3, timer_ch)

		case EB_Moving:
			SetMotorDirection(e.State.Dirn)
			SetDoorOpenLamp(false)

		case EB_Idle:
			SetDoorOpenLamp(false)
		}
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int, timer_ch chan bool) {

	e.State.Floor = newFloor
	SetFloorIndicator(newFloor)
	setAllLights(e)

	if requestShouldStop(*e) {
		SetMotorDirection(MD_Stop)
		e.State.Dirn = MD_Stop // Ole added march 12, needed for re-init
		SetDoorOpenLamp(true)
		requests_clearAtCurrentFloor(e)
		go timer.Run_timer(3, timer_ch)
		e.State.Behaviour = EB_DoorOpen
		setAllLights(e)
	}
}

func fsmNewRequests(e *Elevator, timer_ch chan bool) {
	if e.State.Behaviour == EB_DoorOpen {
		if requests_shouldClearImmediately(*e) {
			requests_clearAtCurrentFloor(e)
			go timer.Run_timer(3, timer_ch)
			setAllLights(e)
		}
		return
	}

	e.State.Dirn, e.State.Behaviour = GetDirectionAndBehaviour(e)
	switch e.State.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		go timer.Run_timer(3, timer_ch)
		requests_clearAtCurrentFloor(e)

	case EB_Moving:
		SetMotorDirection(e.State.Dirn)
	}
	setAllLights(e)
}

func HandleStopButtonPressed(e *Elevator) {
	switch e.State.Floor {
	case -1:
		SetMotorDirection(0)
	default:
		SetMotorDirection(0)
		SetDoorOpenLamp(true)
	}
	e.State.Behaviour = EB_Stopped
}

func setAllLights(e *Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			SetButtonLamp(ButtonType(btn), floor, e.Lights[floor][btn])
		}
	}
}

func PeriodicCheck(selfCheck_ch chan bool) {
	for {
		time.Sleep(1000 * time.Millisecond)
		selfCheck_ch <- true
	}
}

func Selfdiagnose(elevator *Elevator, prevElevator *Elevator, obstruction bool, standstill *int) Diagnose {
	hasRequests := Check_request(*elevator)

	if hasRequests && elevator.State.Behaviour == prevElevator.State.Behaviour {

		switch elevator.State.Behaviour {
		case EB_Idle:
			*prevElevator = *elevator
			return Problem

		case EB_DoorOpen:
			if elevator.State.Floor == prevElevator.State.Floor {
				*standstill += 1
			}
		case EB_Moving:
			if elevator.State.Floor == prevElevator.State.Floor {
				*standstill += 1
			}
		}
		*prevElevator = *elevator

		if *standstill > 10 && obstruction {
			return Obstructed
		} else if *standstill == 20 && !obstruction {
			return Problem
		}

	} else if obstruction {
		*prevElevator = *elevator
		return Unchanged

	} else {
		*standstill = 0
		*prevElevator = *elevator
	}
	return Healthy
}

func Check_request(elevator Elevator) bool {
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < N_BUTTONS; j++ {
			if elevator.Requests[i][j] {
				return true
			}
		}
	}
	return false
}

func getClearedRequests(oldRequests [N_FLOORS][N_BUTTONS]bool, newRequests [N_FLOORS][N_BUTTONS]bool) []ButtonEvent {
	var clearedRequests []ButtonEvent
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < N_BUTTONS; j++ {
			if oldRequests[i][j] != newRequests[i][j] {
				clearedRequests = append(clearedRequests, ButtonEvent{i, ButtonType(j)})
			}
		}
	}

	return clearedRequests

}
