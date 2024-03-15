package elevator

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
	State    State
	Requests [N_FLOORS][N_BUTTONS]bool
	Lights   [N_FLOORS][N_BUTTONS]bool
}

type State struct {
	Floor        int
	Dirn         MotorDirection
	Behaviour    ElevatorBehaviour
	OutOfService bool
}
