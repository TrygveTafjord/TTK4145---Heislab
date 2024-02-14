package elevator

func requests_clearAtCurrentFloor(e *Elevator) {

	e.Requests[e.Floor][BT_Cab] = 0
	switch e.Dirn {

	case MD_Up:
		if !requests_above(e) && !e.Requests[e.Floor][BT_HallUp] {
			e.Requests[e.Floor][BT_HallDown] = 0
		}
		e.Requests[e.Floor][BT_HallUp] = 0
		break

	case MD_Down:
		if !requests_below(e) && !e.Requests[e.Floor][BT_HallDown] {
			e.Requests[e.Floor][BT_HallUp] = 0
		}
		e.Requests[e.Floor][BT_HallDown] = 0
		break

	case MD_Stop:

	default:
		e.Requests[e.Floor][BT_HallUp] = 0
		e.Requests[e.Floor][BT_HallDown] = 0
		break
	}

}

func requests_shouldClearImmediately(e *Elevator, Buttonevent ButtonEvent) bool {

	switch e.Dirn {

	case MD_Up:
		if Buttonevent.Floor == e.Floor && Buttonevent.Button == BT_HallUp {
			return true
		}

	case MD_Down:
		if Buttonevent.Floor == e.Floor && Buttonevent.Button == BT_HallDown {
			return true
		}

	default:
	}
	return false
}
