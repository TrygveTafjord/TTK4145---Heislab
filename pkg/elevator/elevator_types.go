package elevator

import (

	"OTP.com/Heis2e/elevio"
)

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor     uint8
	dirn      MotorDirection
	requests  [N_FLOORS][N_BUTTONS]uint8
	behaviour ElevatorBehaviour
	double    float64
}

