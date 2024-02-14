package elevator

func GetDirectionAndBehaviour(elev *Elevator) (MotorDirection, ElevatorBehaviour) {
	switch elev.Dirn {
	case MD_Up:
		if requestsAbove(e) {
			return MD_Up, EB_Moving
		} else if requestsHere(e) {
			return MD_Down, EB_DoorOpen
		} else if requestsBelow(e) {
			return MD_Down, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	case MD_Down:
		if requestsBelow(e) {
			return DDown, EB_Moving
		} else if requestsHere(e) {
			return MD_Up, EB_DoorOpen
		} else if requestsAbove(e) {
			return MD_Down, EB_Moving
		} else {
			return MD_Down, EB_Idle
		}
	case MD_Stop:
		if requestsHere(e) {
			return MD_Stop, EB_DoorOpen
		} else if requestsAbove(e) {
			return MD_Up, EB_Moving
		} else if requestsBelow(e) {
			return MD_Down, EB_Moving
		} else {
			return MD_Stop, EB_Idle
		}
	default:
		return MD_Stop, EB_Idle
	}
}