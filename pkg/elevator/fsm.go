package elevator

import (
	"fmt"

	"OTP.com/Heis2e/pkg/timer"
)

func FSM(Button_ch chan ButtonEvent, Floor_sensor_ch chan int, Stop_button_ch chan bool, Obstruction_ch chan bool) {
	ElevatorPtr := new(Elevator)

	for {
		select {
		case Buttonevent := <-Button_ch:
			fsmButtonPress(Buttonevent, ElevatorPtr)
		case Newfloor := <-Floor_sensor_ch:
			fsmOnFloorArrival(ElevatorPtr, Newfloor)

		case <-Stop_button_ch:
			HandleStopButtonPressed(ElevatorPtr)

			//set state stopped
			//case Obstruction := <-Obstruction_ch:

		}
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int) {

	e.Floor = newFloor
	SetFloorIndicator(newFloor)
	fmt.Printf("floor check:  %v", newFloor)
	fmt.Printf("fsmOnFloorArrival status: %v\n", e.Requests, "Direction: %v\n", e.Dirn)

	switch e.Behaviour {
	case EB_Moving:
		if requests_shouldStop(*e) { // ----------- Sjekker i køssystem om vi skal stoppe
			SetMotorDirection(MD_Stop)
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e) // ---------- Ber om at denne etasjen fjernes fra køer
			timer.Timer_start(e.Stop_time)  //sett på en watchdog timer som skriver til kanal
			e.Behaviour = EB_DoorOpen
			setAllLights(e)
			for !timer.Timer_timedOut() {

			}

			e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

			switch e.Behaviour {

			case EB_DoorOpen:
				SetDoorOpenLamp(true)
				timer.Timer_start(e.Stop_time)
				requests_clearAtCurrentFloor(e)

			case EB_Moving:
				SetMotorDirection(e.Dirn)
			}

		}
	}

}

func fsmButtonPress(Buttonevent ButtonEvent, elev *Elevator) {

	switch elev.Behaviour {

	case EB_DoorOpen:

		if requests_shouldClearImmediately(elev, Buttonevent) {
			timer.Timer_start(elev.Stop_time)
		} else {
			elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1
		}

	case EB_Moving:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1

	case EB_Idle:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1
		elev.Dirn, elev.Behaviour = GetDirectionAndBehaviour(elev)

		switch elev.Behaviour {

		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			timer.Timer_start(elev.Stop_time)
			requests_clearAtCurrentFloor(elev)

		case EB_Moving:
			SetMotorDirection(elev.Dirn)
		}
	case EB_Stopped:
		fmt.Printf("The elevator is in state stopped, and a button was pressed, fix logic!")
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
			SetButtonLamp(ButtonType(btn), floor, e.Requests[floor][btn] == 1) //Ops
		}
	}
}
