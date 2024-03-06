package elevator

import (
	"project.com/pkg/timer"
)

func FSM(elevStatusUpdate_ch chan Elevator) {

	floorSensor_ch := make(chan int)
	stopButton_ch := make(chan bool)
	obstruction_ch := make(chan bool)
	timer_ch := make(chan bool)

	go PollFloorSensor(floorSensor_ch)
	go PollStopButton(stopButton_ch)
	go PollObstructionSwitch(obstruction_ch)

	elevator := new(Elevator)

	initElevator(elevator, floorSensor_ch)

	elevStatusUpdate_ch <- *elevator

	for {
		select {

		case newElev := <-elevStatusUpdate_ch:
			elevator.Requests = newElev.Requests
			elevator.GlobalLights = newElev.GlobalLights
			elevator.OrderClearedCounter = newElev.OrderClearedCounter
			setAllLights(elevator)
			fsmNewAssignments(elevator, timer_ch, elevStatusUpdate_ch)
			elevStatusUpdate_ch <- *elevator

		case newFloor := <-floorSensor_ch:
			fsmOnFloorArrival(elevator, newFloor, timer_ch, elevStatusUpdate_ch)
			elevStatusUpdate_ch <- *elevator

		case <-stopButton_ch:
			HandleStopButtonPressed(elevator)

		case <-timer_ch:
			HandleDeparture(elevator, timer_ch)

		}
	}
}

func HandleDeparture(e *Elevator, timer_ch chan bool) {
	e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

	switch e.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		requests_clearAtCurrentFloor(e)
		go timer.Run_timer(3, timer_ch)

	case EB_Moving:
		SetMotorDirection(e.Dirn)
		SetDoorOpenLamp(false)

	case EB_Idle:
		SetDoorOpenLamp(false)
	}
}

func initElevator(e *Elevator, floorSensor_ch chan int) {
	floor := GetFloor()

	for floor := 0; floor < 4; floor++ {
		for btn := 0; btn < 3; btn++ {
			SetButtonLamp(ButtonType(btn), floor, false)
		}
	}
	if floor == -1 {
		SetMotorDirection(MD_Down)
	}
	for floor == (-1) {
		floor := GetFloor()

		if floor != (-1) {
			SetMotorDirection(MD_Stop)
			e.Floor = floor
			e.Dirn = MD_Stop
			e.Behaviour = EB_Idle
			break
		}
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int, timer_ch chan bool, elevStatusUpdate_ch chan Elevator) {

	e.Floor = newFloor
	SetFloorIndicator(newFloor)
	switch e.Behaviour {
	case EB_Moving:
		if requestShouldStop(*e) {
			SetMotorDirection(MD_Stop)
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e)
			e.OrderClearedCounter++
			go timer.Run_timer(3, timer_ch)
			e.Behaviour = EB_DoorOpen
			setAllLights(e)
		}
	}
}

func fsmNewAssignments(e *Elevator, timer_ch chan bool, elevStatusUpdate_ch chan Elevator) {

	if e.Behaviour == EB_DoorOpen {
		if requests_shouldClearImmediately(*e) {
			e.OrderClearedCounter++
			go timer.Run_timer(3, timer_ch)
		}
		return
	}

	e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

	switch e.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		go timer.Run_timer(3, timer_ch)
		requests_clearAtCurrentFloor(e)
		e.OrderClearedCounter++

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
			SetButtonLamp(ButtonType(btn), floor, e.GlobalLights[floor][btn])
		}
	}
}
