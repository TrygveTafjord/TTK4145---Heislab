package fsm

import (
	"OTP.com/Heis2e/pkg/elevator"
	"OTP.com/Heis2e/pkg/elevio"
	"OTP.com/Heis2e/pkg/timer"
)

func FSM(Button_ch chan elevio.ButtonEvent, Floor_sensor_ch chan int, Stop_button_ch chan bool, Obstruction_ch chan bool){
	for {
		select {
		case Buttonevent := <-Button_ch:
			
		case Newfloor := <-Floor_sensor_ch:
			
		case Stopbutton := <-Stop_button_ch:
			
		case Obstruction := <-Obstruction_ch:
			
		}
	}
}




func fsmButtonPress(Buttonevent elevio.ButtonEvent, elev *elevator.Elevator){
	
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
		elev.Dirn, elev.Behaviour = elevator.GetDirectionAndBehaviour(Buttonevent.Floor, elev.floor)
		if (elev.Behaviour == elevator.EB_Moving){
				elevio.SetMotorDirection(elev.Dirn)
		}		
}
}