package elevator

import (
	"time"

	"project.com/pkg/timer"
)

func FSM(elevatorInit_ch chan Elevator,
	requestUpdate_ch chan [N_FLOORS][N_BUTTONS]bool,
	clearRequestToInfobank_ch chan []ButtonEvent,
	stateToInfobank_ch chan State,
	lightUpdate_ch chan [N_FLOORS][N_BUTTONS]bool,
	obstructedStateToInfobank_ch chan bool,
	updateDiagnostics_ch chan Elevator,
	obstructionDiagnose_ch chan bool) {

	floorSensor_ch := make(chan int)
	obstruction_ch := make(chan bool)
	timer_ch := make(chan bool)

	go PollFloorSensor(floorSensor_ch)
	go PollObstructionSwitch(obstruction_ch)

	elevator := new(Elevator)
	*elevator = <-elevatorInit_ch

	for {
		select {
		case requests := <-requestUpdate_ch:
			elevator.Requests = requests
			fsmNewRequests(elevator, timer_ch)
			if elevator.Requests != requests {
				clearRequestToInfobank_ch <- getClearedRequests(requests, elevator.Requests)
			}
			stateToInfobank_ch <- elevator.State
			updateDiagnostics_ch <- *elevator

		case lights := <-lightUpdate_ch:
			elevator.Lights = lights
			setAllLights(elevator)

		case newFloor := <-floorSensor_ch:
			requestsBeforeNewFloor := elevator.Requests
			fsmOnFloorArrival(elevator, newFloor, timer_ch)
			stateToInfobank_ch <- elevator.State
			if requestsBeforeNewFloor != elevator.Requests {
				clearRequestToInfobank_ch <- getClearedRequests(requestsBeforeNewFloor, elevator.Requests)
			}

			updateDiagnostics_ch <- *elevator

		case <-timer_ch:
			HandleDeparture(elevator, timer_ch)
			stateToInfobank_ch <- elevator.State
			updateDiagnostics_ch <- *elevator

		case obstruction := <-obstruction_ch:
			if !obstruction && elevator.State.Behaviour == EB_DoorOpen {
				go timer.Run_timer(3, timer_ch)

				if elevator.State.OutOfService {
					elevator.State.OutOfService = false
					updateDiagnostics_ch <- *elevator
					obstructedStateToInfobank_ch <- false
				}
			}

		case <-obstructionDiagnose_ch:
			elevator.State.OutOfService = true
			obstructedStateToInfobank_ch <- true
		}
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

func HandleDeparture(e *Elevator, timer_ch chan bool) {
	if GetObstruction() && e.State.Behaviour == EB_DoorOpen {
		go timer.Run_timer(3, timer_ch)
		return
	}

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

func fsmOnFloorArrival(e *Elevator, newFloor int, timer_ch chan bool) {

	e.State.Floor = newFloor
	SetFloorIndicator(newFloor)
	setAllLights(e)

	if requestShouldStop(*e) {
		SetMotorDirection(MD_Stop)
		e.State.Dirn = MD_Stop
		SetDoorOpenLamp(true)
		requests_clearAtCurrentFloor(e)
		go timer.Run_timer(3, timer_ch)
		e.State.Behaviour = EB_DoorOpen
		setAllLights(e)
	}
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
