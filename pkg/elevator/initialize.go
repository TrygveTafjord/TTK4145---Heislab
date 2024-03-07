package elevator

func ElevatorInit(elevStatusUpdate_ch chan Elevator, Id string) {
	var e Elevator

	e.Id = Id
	e.OrderClearedCounter = 0
	e.OrderCounter = 0

	floor := GetFloor()

	//reset buttons
	for floor := 0; floor < 4; floor++ {
		for btn := 0; btn < 3; btn++ {
			SetButtonLamp(ButtonType(btn), floor, false)
		}
	}

	if floor == -1 {
		SetMotorDirection(MD_Down)
		for floor == -1 {
			floor := GetFloor()
			if floor != (-1) {
				SetMotorDirection(MD_Stop)
				break
			}
		}
	}
	e.Floor = floor
	e.Dirn = MD_Stop
	e.Behaviour = EB_Idle
	elevStatusUpdate_ch <- e
}
