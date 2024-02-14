package elevator

func GetDirectionAndBehaviour(button_floor int, elev_floor int) (MotorDirection, ElevatorBehaviour) {
	if button_floor == elev_floor {
		return MD_Stop, EB_Idle
	}

	if button_floor > elev_floor {
		return MD_Up, EB_Moving
	}

	return MD_Down, EB_Moving

}
