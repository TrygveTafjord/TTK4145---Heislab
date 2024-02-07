package elevator

import "OTP.com/Heis2e/pkg/elevio"

func GetDirectionAndBehaviour(button_floor int, elev_floor int) (elevio.MotorDirection, ElevatorBehaviour) {
	if button_floor == elev_floor {
		return elevio.MD_Stop, EB_Idle
	} 

	if button_floor > elev_floor {
		return elevio.MD_Up, EB_Moving
	} 

	return elevio.MD_Down, EB_Moving
	
}