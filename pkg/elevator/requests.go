package elevator

func requests_above(e Elevator) bool {
	for flr := e.Floor + 1; flr < N_FLOORS; flr++ {
		for btn := 0; flr < N_BUTTONS; btn++ {
			if e.Requests[flr][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for flr := 0; flr < e.Floor; flr++ {
		for btn := 0; flr < N_BUTTONS; btn++ {
			if e.Requests[flr][btn] == 1 {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] == 1 {
			return true
		}
	}

	return false
}

func requests_shouldStop(e Elevator) bool {
	switch e.Dirn {
	case MD_Down:
		return (e.Requests[e.Floor][BT_HallDown] == 1 ||
			e.Requests[e.Floor][BT_Cab] == 1 ||
			!requests_below(e))
	case MD_Up:
		return (e.Requests[e.Floor][BT_HallUp] == 1 ||
			e.Requests[e.Floor][BT_Cab] == 1 ||
			!requests_above(e))
	default:
		return true
	}
}

func GetDirectionAndBehaviour(elev *Elevator) (MotorDirection, ElevatorBehaviour) {
	switch elev.Dirn {
	case MD_Up:
		if requests_above(*elev) {
			return MD_Up, EB_Moving
		} else if requests_here(*elev) {
			return MD_Down, EB_DoorOpen
		} else if requests_below(*elev) {
			return MD_Down, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	case MD_Down:
		if requests_below(*elev) {
			return MD_Down, EB_Moving
		} else if requests_here(*elev) {
			return MD_Up, EB_DoorOpen
		} else if requests_above(*elev) {
			return MD_Down, EB_Moving
		} else {
			return MD_Down, EB_Idle
		}
	case MD_Stop:
		if requests_here(*elev) {
			return MD_Stop, EB_DoorOpen
		} else if requests_above(*elev) {
			return MD_Up, EB_Moving
		} else if requests_below(*elev) {
			return MD_Down, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	default:
		return MD_Stop, EB_Idle
	}
}
