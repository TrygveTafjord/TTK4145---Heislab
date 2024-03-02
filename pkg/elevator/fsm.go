package elevator

import (
	"project.com/pkg/timer"
)

func FSM(button_ch chan ButtonEvent, floorSensor_ch chan int, stopButton_ch chan bool, obstruction_ch chan bool, timerFinished chan bool) {
	elevatorPtr := new(Elevator)

	for {
		select {
		case buttonevent := <-button_ch:
			fsmButtonPress(buttonevent, elevatorPtr, timerFinished)
		case newFloor := <-floorSensor_ch:
			fsmOnFloorArrival(elevatorPtr, newFloor, timerFinished)
		case <-stopButton_ch:
			HandleStopButtonPressed(elevatorPtr)
			//set state stopped
			//case Obstruction := <-obstruction_ch: gutta
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
		//timer.Timer_start(e.Stop_time)
		requests_clearAtCurrentFloor(e)

	case EB_Moving:
		SetMotorDirection(e.Dirn)
		SetDoorOpenLamp(false)

	case EB_Idle:
		SetDoorOpenLamp(false)
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int, timerFinished chan bool) {

	e.Floor = newFloor
	SetFloorIndicator(newFloor)
	switch e.Behaviour {
	case EB_Moving:
		if requestShouldStop(*e) {
			SetMotorDirection(MD_Stop)
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e)
			go timer.Run_timer(3, timerFinished)
			e.Behaviour = EB_DoorOpen
			setAllLights(e)
		}
	}

}

func fsmButtonPress(buttonEvent ButtonEvent, e *Elevator, timerFinished chan bool) {

	switch e.Behaviour {

	case EB_DoorOpen:

		if requests_shouldClearImmediately(e, buttonEvent) {
			go timer.Run_timer(3, timerFinished)
		} else {
			e.Requests[buttonEvent.Floor][buttonEvent.Button] = true
		}

	case EB_Moving:
		e.Requests[buttonEvent.Floor][buttonEvent.Button] = true

	case EB_Idle:
		e.Requests[buttonEvent.Floor][buttonEvent.Button] = true
		e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

		switch e.Behaviour {

		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			go timer.Run_timer(3, timerFinished)
			requests_clearAtCurrentFloor(e)

		case EB_Moving:
			SetMotorDirection(e.Dirn)
		}
	case EB_Stopped:
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
