package elevator

import (
	"fmt"
	"project.com/pkg/timer"
)

func FSM(Button_ch chan ButtonEvent, Floor_sensor_ch chan int, Stop_button_ch chan bool, Obstruction_ch chan bool, timerFinished chan bool) {
	ElevatorPtr := new(Elevator)

	for {
		select {
		case Buttonevent := <-Button_ch:
			fsmButtonPress(Buttonevent, ElevatorPtr, timerFinished)
		case Newfloor := <-Floor_sensor_ch:
			fsmOnFloorArrival(ElevatorPtr, Newfloor, timerFinished)
		case <-Stop_button_ch:
			HandleStopButtonPressed(ElevatorPtr)
			//set state stopped
			//case Obstruction := <-Obstruction_ch:
		case <-timerFinished:
			fmt.Print("It happened")
			HandleDeparture(ElevatorPtr)

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
		if requests_shouldStop(*e) { // ----------- Sjekker i køssystem om vi skal stoppe
			SetMotorDirection(MD_Stop)
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e) // ---------- Ber om at denne etasjen fjernes fra køer
			//timer.Timer_start(3)  //sett på en watchdog timer som skriver til kanal
			go timer.Run_timer(3, timerFinished)
			e.Behaviour = EB_DoorOpen
			setAllLights(e)
		}
	}

}

func fsmButtonPress(Buttonevent ButtonEvent, elev *Elevator, timerFinished chan bool) {

	switch elev.Behaviour {

	case EB_DoorOpen:

		if requests_shouldClearImmediately(elev, Buttonevent) {
			go timer.Run_timer(3, timerFinished)
		} else {
			elev.Requests[Buttonevent.Floor][Buttonevent.Button] = true
		}

	case EB_Moving:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = true

	case EB_Idle:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = true
		elev.Dirn, elev.Behaviour = GetDirectionAndBehaviour(elev)

		switch elev.Behaviour {

		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			go timer.Run_timer(3, timerFinished)
			requests_clearAtCurrentFloor(elev)

		case EB_Moving:
			SetMotorDirection(elev.Dirn)
		}
	case EB_Stopped:
	}
	setAllLights(elev)

}

func HandleStopButtonPressed(e *Elevator) {
	//stop motor and consider opening door
	switch e.Floor {
	case -1:
		//stop motor
		SetMotorDirection(0)
	default:
		SetMotorDirection(0)
		SetDoorOpenLamp(true)
	}
	//set state stopped
	e.Behaviour = EB_Stopped
}

func setAllLights(e *Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			SetButtonLamp(ButtonType(btn), floor, e.Requests[floor][btn] == true) //Ops
		}
	}
}
