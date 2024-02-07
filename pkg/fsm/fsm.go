package fsm

import (
	"fmt"

	"OTP.com/Heis2e/pkg/elevator"
	"OTP.com/Heis2e/pkg/elevio"
	"OTP.com/Heis2e/pkg/timer"
)

func FSM(Button_ch chan elevio.ButtonEvent, Floor_sensor_ch chan int, Stop_button_ch chan bool, Obstruction_ch chan bool) {
	ElevatorPtr := new(elevator.Elevator)

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

func fsmOnFloorArrival(e *elevator.Elevator, newFloor int) {

	e.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	switch e.Behaviour {
	case elevator.EB_Moving:
		//if requestsShouldStop(elevator) { // ----------- Sjekker i køssystem om vi skal stoppe
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetDoorOpenLamp(true)
		//e = requestsClearAtCurrentFloor(e)  // ---------- Ber om at denne etasjen fjernes fra køer
		timer.Timer_start(3) // ----------- Hvilken input skal denne ha
		//setAllLights(e)             // ---------- Oppdaterer alle lys basert på køer og status
		e.Behaviour = elevator.EB_DoorOpen
	}
	//}
}

func fsmButtonPress(Buttonevent elevio.ButtonEvent, elev *elevator.Elevator) {

	switch elev.Behaviour {

	case elevator.EB_DoorOpen:
		if Buttonevent.Floor == elev.Floor {
			timer.Timer_start(elev.Stop_time)
		} else {
			elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1
		}

	case elevator.EB_Moving:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1

	case elevator.EB_Idle:
		elev.Requests[Buttonevent.Floor][Buttonevent.Button] = 1
		elev.Dirn, elev.Behaviour = elevator.GetDirectionAndBehaviour(Buttonevent.Floor, elev.Floor)
		if elev.Behaviour == elevator.EB_Moving {
			elevio.SetMotorDirection(elev.Dirn)
		}
	}
}

func HandleStopButtonPressed(e *elevator.Elevator) {
	//stop motor and consider opening door
	switch e.Floor {
	case -1:
		//stop motor
		elevio.SetMotorDirection(0)
	default:
		elevio.SetMotorDirection(0)
		elevio.SetDoorOpenLamp(true)
	}
	//set state stopped
	e.Behaviour = elevator.EB_Stopped
}
