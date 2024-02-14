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
			fmt.Print("hello from button pressed in FSM")
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

	switch e.Behaviour {
	case EB_Moving:
		//if requestsShouldStop(elevator) { // ----------- Sjekker i køssystem om vi skal stoppe
			SetMotorDirection(MD_Stop)
			SetDoorOpenLamp(true)
			//e = requestsClearAtCurrentFloor(e)  // ---------- Ber om at denne etasjen fjernes fra køer
			timer.Timer_start(3) // ----------- Hvilken input skal denne ha
			//setAllLights(e)             // ---------- Oppdaterer alle lys basert på køer og status
			e.Behaviour = EB_DoorOpen
	}
	//}
}

func fsmButtonPress(Buttonevent ButtonEvent, elev *Elevator) {

	switch elev.Behaviour {

	case EB_DoorOpen:
		if Buttonevent.Floor == elev.Floor {
			timer.Timer_start(elev.Stop_time)
		} else {
			elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1
		}

	case EB_Moving:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1

	case EB_Idle:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1
		//GetDirection()
		elev.Dirn, elev.Behaviour = GetDirectionAndBehaviour(elev)
		
		switch elev.Behaviour {

		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			timer.Timer_start(elev.Stop_time)
			//requests_clearAtCurrentFloor(elevator)

		case EB_Moving:
			SetMotorDirection(elev.Dirn)
	}
}
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
