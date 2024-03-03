package elevator

import (
	"project.com/pkg/timer"
)

func FSM(newAssignments_ch chan Elevator,
	floorSensor_ch chan int,
	stopButton_ch chan bool,
	obstruction_ch chan bool,
	timerFinished chan bool,
	clearedBtn_ch chan Elevator) {

	elevatorPtr := new(Elevator)

	for {
		select {
		case newAssignments := <-newAssignments_ch:
			*elevatorPtr = newAssignments
			fsmNewAssignments(elevatorPtr, timerFinished)
		case newFloor := <-floorSensor_ch:
			fsmOnFloorArrival(elevatorPtr, newFloor, timerFinished, clearedBtn_ch)
		case <-stopButton_ch:
			HandleStopButtonPressed(elevatorPtr)
			//elevatorPtr.Behaviour = EB_Stopped
			//case Obstruction := <-obstruction_ch:
		case <-timerFinished:
			HandleDeparture(elevatorPtr)
		}
	}
}

func HandleDeparture(e *Elevator) {
	e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

	switch e.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		requests_clearAtCurrentFloor(e)

	case EB_Moving:
		SetMotorDirection(e.Dirn)
		SetDoorOpenLamp(false)

	case EB_Idle:
		SetDoorOpenLamp(false)
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int, timerFinished chan bool, newElevState_ch chan Elevator) {

	e.Floor = newFloor
	SetFloorIndicator(newFloor)
	switch e.Behaviour {
	case EB_Moving:
		if requestShouldStop(*e) {
			SetMotorDirection(MD_Stop)
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e)
			newElevState_ch <- *e
			go timer.Run_timer(3, timerFinished)
			e.Behaviour = EB_DoorOpen
			setAllLights(e)
		}
	}
}

func fsmNewAssignments(e *Elevator, timerFinished chan bool) {

	e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

	switch e.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		go timer.Run_timer(3, timerFinished)
		requests_clearAtCurrentFloor(e)

	case EB_Moving:
		SetMotorDirection(e.Dirn)
	}
	setAllLights(e)
}

func HandleStopButtonPressed(e *Elevator) {
	switch e.Floor {
	case -1:
		SetMotorDirection(0)
	default:
		SetMotorDirection(0)
		SetDoorOpenLamp(true)
	}
	e.Behaviour = EB_Stopped
}

func setAllLights(e *Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			SetButtonLamp(ButtonType(btn), floor, e.Requests[floor][btn])
		}
	}
}
