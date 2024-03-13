package elevator

import (
	//"fmt"
	//"fmt"

	"time"

	"project.com/pkg/timer"
)

func FSM(requestUpdate_ch chan [N_FLOORS][N_BUTTONS]bool, clearRequestToInfobank_ch chan [N_FLOORS][N_BUTTONS]bool, stateToInfobank_ch chan State, lightUpdate_ch chan [N_FLOORS][N_BUTTONS]bool, elevatorInit_ch chan Elevator) {

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
			if quickClear(elevator, timer_ch) {
				clearRequestToInfobank_ch <- elevator.Requests
				break	
			} 
			fsmNewRequests(elevator, timer_ch)
			stateToInfobank_ch <- elevator.State

		case lights := <-lightUpdate_ch:
			elevator.Lights = lights
			setAllLights(elevator)

		case newFloor := <-floorSensor_ch:
			
			
			ShouldStop := fsmOnFloorArrival(elevator, newFloor, timer_ch)
			stateToInfobank_ch <- elevator.State
			if ShouldStop {
				clearRequestToInfobank_ch <- elevator.Requests
			}
			

		case <-timer_ch:
			HandleDeparture(elevator, timer_ch, obstruction)
			stateToInfobank_ch <- elevator.State

		case obstruction = <-obstruction_ch:
			if !obstruction && elevator.State.Behaviour == EB_DoorOpen {
				go timer.Run_timer(3, timer_ch)
			}
/*
		case <-selfCheck_ch:
			diagnose := Selfdiagnose(elevator, &prevelevator, obstruction, &standstill)
			switch diagnose {
			case Healthy:
				elevator.State.Obstructed = false
				toInfobank_ch <- *elevator
			case Obstructed:
				elevator.State.Obstructed = true
				toInfobank_ch <- *elevator
			case Problem:
				//Reboot
			case Unchanged:
				//Nothing
			}
*/
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

func fsmOnFloorArrival(e *Elevator, newFloor int, timer_ch chan bool) bool {

	e.State.Floor = newFloor
	SetFloorIndicator(newFloor)
	setAllLights(e)

	if requestShouldStop(*e) {
		SetMotorDirection(MD_Stop)
			//e.Dirn = MD_Stop // Ole added march 12, needed for re-init
		SetDoorOpenLamp(true)
		requests_clearAtCurrentFloor(e)
		go timer.Run_timer(3, timer_ch)
		e.State.Behaviour = EB_DoorOpen
		setAllLights(e)
		return true
	}
	return false
}


func quickClear(e *Elevator, timer_ch chan bool) bool {
	if e.State.Behaviour == EB_DoorOpen {
		if requests_shouldClearImmediately(*e) {
			requests_clearAtCurrentFloor(e)
			go timer.Run_timer(3, timer_ch)
			setAllLights(e)
			return true
		}
		
	}
	return false
}

func fsmNewRequests(e *Elevator, timer_ch chan bool) {

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
		e.Lights[floor][BT_Cab] = e.Requests[floor][BT_Cab]
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

