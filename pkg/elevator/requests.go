package elevator

func requestsAbove(e Elevator) bool {
	for flr := e.Floor + 1; flr < N_FLOORS; flr++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[flr][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e Elevator) bool {
	for flr := 0; flr < e.Floor; flr++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[flr][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}

	return false
}

func requestShouldStop(e Elevator) bool {
	switch e.Dirn {
	case MD_Down:
		return (e.Requests[e.Floor][BT_HallDown] ||
			e.Requests[e.Floor][BT_Cab] ||
			!requestsBelow(e))
	case MD_Up:
		return (e.Requests[e.Floor][BT_HallUp] ||
			e.Requests[e.Floor][BT_Cab] ||
			!requestsAbove(e))
	default:
		return true
	}
}

func GetDirectionAndBehaviour(e *Elevator) (MotorDirection, ElevatorBehaviour) {
	switch e.Dirn {
	case MD_Up:
		if requestsAbove(*e) {
			return MD_Up, EB_Moving
		} else if requestsHere(*e) {
			return MD_Stop, EB_DoorOpen
		} else if requestsBelow(*e) {
			return MD_Down, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	case MD_Down:
		if requestsBelow(*e) {
			return MD_Down, EB_Moving
		} else if requestsHere(*e) {
			return MD_Stop, EB_DoorOpen
		} else if requestsAbove(*e) {
			return MD_Up, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	case MD_Stop:
		if requestsHere(*e) {
			return MD_Stop, EB_DoorOpen
		} else if requestsAbove(*e) {
			return MD_Up, EB_Moving
		} else if requestsBelow(*e) {
			return MD_Down, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	default:
		return MD_Stop, EB_Idle
	}
}

func requests_clearAtCurrentFloor(e *Elevator) {

	e.Requests[e.Floor][BT_Cab] = false
	e.Lights[e.Floor][BT_Cab] = false

	switch e.Dirn {

	case MD_Up:
		if !requestsAbove(*e) && !(e.Requests[e.Floor][BT_HallUp]) {
			e.Requests[e.Floor][BT_HallDown] = false
			e.Lights[e.Floor][BT_HallDown] = false
		}
		e.Requests[e.Floor][BT_HallUp] = false
		e.Lights[e.Floor][BT_HallUp] = false

	case MD_Down:
		if !requestsBelow(*e) && !(e.Requests[e.Floor][BT_HallDown]) {
			e.Requests[e.Floor][BT_HallUp] = false
			e.Lights[e.Floor][BT_HallUp] = false
		}

		e.Requests[e.Floor][BT_HallDown] = false
		e.Lights[e.Floor][BT_HallDown] = false
	default:
		e.Requests[e.Floor][BT_HallUp] = false
		e.Requests[e.Floor][BT_HallDown] = false
		e.Lights[e.Floor][BT_HallUp] = false
		e.Lights[e.Floor][BT_HallDown] = false
	}
}

func requests_shouldClearImmediately(e Elevator) bool {
	var buttonsPressed []ButtonEvent

	for i := 0; i < N_BUTTONS; i++ {
		if e.Requests[e.Floor][i] {
			buttonsPressed = append(buttonsPressed, ButtonEvent{e.Floor, ButtonType(i)})
		}
	}
	if GetObstruction() {
		return false
	}

	for _, buttonevent := range buttonsPressed {

		switch e.Dirn {

		case MD_Up:
			if buttonevent.Button == BT_HallUp || buttonevent.Button == BT_Cab {
				return true
			}

		case MD_Down:
			if buttonevent.Button == BT_HallDown || buttonevent.Button == BT_Cab {
				return true
			}
		case MD_Stop:
			if buttonevent.Button == BT_HallDown || buttonevent.Button == BT_Cab || buttonevent.Button == BT_HallUp {
				return true
			}

		}

	}
	return false
}
