package elevator

import (
	//"fmt"


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
	*elevator = <-elevStatusUpdate_ch

	obstruction := GetObstruction()

	for {
		select {

		case newElev := <-elevStatusUpdate_ch:

			//if new_assignment{}
			if newElev.OrderCounter > elevator.OrderCounter {
				elevator.Requests = newElev.Requests
				elevator.Lights = newElev.Lights
				elevator.OrderCounter = newElev.OrderCounter
				elevator.OrderClearedCounter = newElev.OrderClearedCounter
				fsmNewAssignments(elevator, timer_ch)
				elevStatusUpdate_ch <- *elevator
			}
			if newElev.OrderClearedCounter > elevator.OrderClearedCounter {
				elevator.Lights = newElev.Lights
				elevator.OrderClearedCounter = newElev.OrderClearedCounter
				setAllLights(elevator)
			}

			//if should clear_light (Infobank vet alle rede at lyset er fjernet, vi trenger ikke sende den informasjonen)

		case newFloor := <-floorSensor_ch:
			fsmOnFloorArrival(elevator, newFloor, timer_ch, elevStatusUpdate_ch)
			elevStatusUpdate_ch <- *elevator

		case <-stopButton_ch:
			HandleStopButtonPressed(elevator)
		case <-timer_ch:
			HandleDeparture(elevator, timer_ch, obstruction)
			elevStatusUpdate_ch <- *elevator

		case obstr := <-obstruction_ch:
			if (!obstr && elevator.Behaviour == EB_DoorOpen) {
				go timer.Run_timer(3, timer_ch)
			}
			obstruction = obstr
		}
	}
}

func HandleDeparture(e *Elevator, timer_ch chan bool, obstruction bool) {
	if(obstruction && e.Behaviour == EB_DoorOpen){
		go timer.Run_timer(3, timer_ch)
	}else{
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

func fsmNewAssignments(e *Elevator, timer_ch chan bool) {
		// Her mistenker jeg det kommer bugs -Per 08.03
	if e.Behaviour == EB_DoorOpen {
		if requests_shouldClearImmediately(*e) {

			requests_clearAtCurrentFloor(e)

			e.OrderClearedCounter++

			go timer.Run_timer(3, timer_ch)
		}
		setAllLights(e)
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
			SetButtonLamp(ButtonType(btn), floor, e.Lights[floor][btn])
		}
	}
}
