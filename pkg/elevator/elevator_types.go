package elevator

import (
	"OTP.com/Heis2e/pkg/elevio"
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
	Floor     	int
	Dirn      	elevio.MotorDirection
	Requests  	[N_FLOORS][N_BUTTONS]uint8
	Behaviour 	ElevatorBehaviour
	Stop_time   float64
}

